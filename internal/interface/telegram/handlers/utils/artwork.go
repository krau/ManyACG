package utils

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/model/command"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/pkg/imgtool"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/strutil"
	"github.com/krau/ManyACG/service"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

func PostAndCreateArtwork(
	ctx *telegohandler.Context,
	serv *service.Service,
	artwork *entity.CachedArtworkData,
	fromChatID telego.ChatID, toChatID telego.ChatID, messageID int,
) error {
	awInDb, err := serv.GetArtworkByURL(ctx, artwork.SourceURL)
	if err == nil {
		return oops.Errorf("artwork already exists in db: %s", awInDb.SourceURL)
	}
	if serv.CheckDeletedByURL(ctx, artwork.SourceURL) {
		return oops.Errorf("artwork is marked as deleted: %s", artwork.SourceURL)
	}
	showProgress := (fromChatID.ID != 0 || fromChatID.Username != "") && messageID != 0
	if showProgress {
		// // clear previous inline buttons
		// defer ctx.Bot().EditMessageReplyMarkup(ctx, &telego.EditMessageReplyMarkupParams{
		// 	ChatID:      telegoutil.ID(fromChatID),
		// 	MessageID:   messageID,
		// 	ReplyMarkup: nil,
		// })
		ctx.Bot().EditMessageReplyMarkup(ctx, telegoutil.EditMessageReplyMarkup(
			fromChatID,
			messageID,
			telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
				telegoutil.InlineKeyboardButton("正在存储图片...").WithCallbackData("noop"),
			}),
		))
	}

	for i, pic := range artwork.Pictures {
		file, clean, err := httpclient.DownloadWithCache(ctx, pic.Original, nil)
		if err != nil {
			return oops.Wrapf(err, "failed to download picture %d", i)
		}
		defer file.Close()
		defer clean()
		var ext string
		ext, err = strutil.GetFileExtFromURL(pic.Original)
		if err != nil {
			mtype, err := mimetype.DetectFile(file.Name())
			if err != nil {
				return oops.Wrapf(err, "failed to detect mime type for picture %d", i)
			}
			ext = mtype.Extension()
		}
		filename := fmt.Sprintf("%s%s", strutil.MD5Hash(pic.Original), ext)
		info, err := serv.StorageSaveAllSize(ctx, file, fmt.Sprintf("/%s/%s", artwork.SourceType, artwork.Artist.UID), filename)
		if err != nil {
			return oops.Wrapf(err, "failed to save picture %d", i)
		}
		artwork.Pictures[i].StorageInfo = *info
	}
	if showProgress {
		ctx.Bot().EditMessageReplyMarkup(ctx, telegoutil.EditMessageReplyMarkup(
			fromChatID,
			messageID,
			telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
				telegoutil.InlineKeyboardButton("正在发布到频道...").WithCallbackData("noop"),
			}),
		))
	}
	msgs, err := SendArtworkMediaGroup(ctx, toChatID, artwork)
	if err != nil {
		return oops.Wrapf(err, "failed to send artwork media group")
	}
	if len(msgs) == 0 {
		return oops.New("no messages sent")
	}
	for i := range artwork.Pictures {
		tginfo := shared.TelegramInfo{
			MessageID:    msgs[i].MessageID,
			MediaGroupID: msgs[i].MediaGroupID,
		}
		if photo := msgs[i].Photo; len(photo) > 0 {
			tginfo.PhotoFileID = photo[len(photo)-1].FileID
		}
		if doc := msgs[i].Document; doc != nil {
			tginfo.DocumentFileID = doc.FileID
		}
		artwork.Pictures[i].TelegramInfo = tginfo
	}
	ent, err := serv.CreateArtwork(ctx, &command.ArtworkCreation{
		Title:       artwork.Title,
		Description: artwork.Description,
		R18:         artwork.R18,
		SourceType:  artwork.SourceType,
		Artist: command.ArtworkArtistCreation{
			Name:     artwork.Artist.Name,
			UID:      artwork.Artist.UID,
			Username: artwork.Artist.Username,
		},
		SourceURL: artwork.SourceURL,
		Tags:      artwork.Tags,
		Pictures: func() []command.ArtworkPictureCreation {
			pics := make([]command.ArtworkPictureCreation, len(artwork.Pictures))
			for i, pic := range artwork.Pictures {
				pics[i] = command.ArtworkPictureCreation{
					Index:        pic.Index,
					Thumbnail:    pic.Thumbnail,
					Original:     pic.Original,
					Width:        pic.Width,
					Height:       pic.Height,
					Phash:        pic.Phash,
					ThumbHash:    pic.ThumbHash,
					TelegramInfo: &pic.TelegramInfo,
					StorageInfo:  &pic.StorageInfo,
				}
			}
			return pics
		}(),
	})
	if err != nil {
		return oops.Wrapf(err, "failed to create artwork in db")
	}
	log.Info("created artwork", "id", ent.ID, "url", ent.SourceURL, "title", ent.Title, "pics", len(ent.Pictures))
	return nil
}

func SendArtworkMediaGroup(ctx *telegohandler.Context, chatID telego.ChatID, artwork entity.ArtworkLike) ([]telego.Message, error) {
	pics := artwork.GetPictures()
	if len(pics) <= 10 {
		inputs, err := ArtworkInputMediaPhotos(ctx, service.FromContext(ctx), artwork, ArtworkHTMLCaption(metautil.FromContext(ctx), artwork), 0, len(pics))
		if err != nil {
			return nil, oops.Wrapf(err, "failed to create input media photos")
		}
		// Send the media group
		return ctx.Bot().SendMediaGroup(ctx, telegoutil.MediaGroup(
			chatID,
			inputs...,
		))
	}
	caption := ArtworkHTMLCaption(metautil.FromContext(ctx), artwork)
	messages := make([]telego.Message, len(pics))
	for i := 0; i < len(pics); i += 10 {
		end := i + 10
		if end > len(pics) {
			end = len(pics)
		}
		inputs, err := ArtworkInputMediaPhotos(ctx, service.FromContext(ctx), artwork, caption, i, end)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to create input media photos")
		}
		mediaGroup := telegoutil.MediaGroup(chatID, inputs...)
		if i > 0 {
			mediaGroup = mediaGroup.WithReplyParameters(&telego.ReplyParameters{
				ChatID:    chatID,
				MessageID: messages[i-1].MessageID,
			})
		}
		msgs, err := ctx.Bot().SendMediaGroup(ctx, mediaGroup)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to send media group")
		}
		copy(messages[i:], msgs)
	}
	return messages, nil
}

func ArtworkInputMediaPhotos(ctx context.Context,
	serv *service.Service,
	artwork entity.ArtworkLike,
	caption string,
	start, end int) ([]telego.InputMedia, error) {
	inputMediaPhotos := make([]telego.InputMedia, end-start)
	for i := start; i < end; i++ {
		picture := artwork.GetPictures()[i]
		var photo *telego.InputMediaPhoto
		if id := picture.GetTelegramInfo().PhotoFileID; id != "" {
			photo = telegoutil.MediaPhoto(telegoutil.FileFromID(id))
		}
		if photo == nil {
			var fileBytes []byte
			var err error
			fileBytes, err = httpclient.GetReqCachedFile(picture.GetOriginal())
			if err != nil {
				if picture.GetStorageInfo() == shared.ZeroStorageInfo || picture.GetStorageInfo().Original == nil {
					file, clean, err := httpclient.DownloadWithCache(ctx, picture.GetOriginal(), nil)
					if err != nil {
						return nil, oops.Wrapf(err, "failed to download file: %s", picture.GetOriginal())
					}
					defer file.Close()
					defer clean()
					fileBytes, err = os.ReadFile(file.Name())
					if err != nil {
						return nil, oops.Wrapf(err, "failed to read file: %s", picture.GetOriginal())
					}
				} else {
					rc, err := serv.StorageGetFile(ctx, *picture.GetStorageInfo().Original)
					if err != nil {
						return nil, oops.Wrapf(err, "failed to get file: %s", picture.GetOriginal())
					}
					defer rc.Close()
					var buf bytes.Buffer
					writer := bufio.NewWriter(&buf)
					_, err = writer.ReadFrom(rc)
					if err != nil {
						return nil, oops.Wrapf(err, "failed to read file: %s", picture.GetOriginal())
					}
					writer.Flush()
					fileBytes = buf.Bytes()
				}
			}
			fileBytes, err = imgtool.CompressForTelegram(fileBytes)
			if err != nil {
				return nil, oops.Wrapf(err, "failed to compress image: %s", picture.GetOriginal())
			}
			photo = telegoutil.MediaPhoto(telegoutil.File(telegoutil.NameReader(bytes.NewReader(fileBytes), uuid.New().String())))
		}
		if i == 0 {
			photo = photo.WithCaption(caption).WithParseMode(telego.ModeHTML)
		}
		if artwork.GetR18() {
			photo = photo.WithHasSpoiler()
		}
		inputMediaPhotos[i-start] = photo
	}
	return inputMediaPhotos, nil
}

func GetPicturePhotoInputFile(ctx context.Context, serv *service.Service, picture entity.PictureLike) (telego.InputFile, error) {
	if id := picture.GetTelegramInfo().PhotoFileID; id != "" {
		return telegoutil.FileFromID(id), nil
	}
	orgStorDetail := picture.GetStorageInfo().Original
	if orgStorDetail != nil {
		rsc, err := serv.StorageGetFile(ctx, *picture.GetStorageInfo().Original)
		if err != nil {
			return telego.InputFile{}, oops.Wrapf(err, "failed to get file from storage")
		}
		defer rsc.Close()
		// 将 rsc(ReadSeekCloser) 读取到内存中
		data := new(bytes.Buffer)
		_, err = data.ReadFrom(rsc)
		if err != nil {
			return telego.InputFile{}, oops.Wrapf(err, "failed to read file from storage")
		}
		compressed, err := imgtool.CompressForTelegram(data.Bytes())
		if err != nil {
			return telego.InputFile{}, oops.Wrapf(err, "failed to compress image")
		}

		return telegoutil.File(telegoutil.NameBytes(compressed, fmt.Sprintf("%s%s", strutil.MD5Hash(picture.GetOriginal()), ".jpg"))), nil
	}
	file, clean, err := httpclient.DownloadWithCache(ctx, picture.GetOriginal(), nil)
	if err != nil {
		return telego.InputFile{}, oops.Wrapf(err, "failed to download file: %s", picture.GetOriginal())
	}
	defer file.Close()
	defer clean()
	fileBytes, err := os.ReadFile(file.Name())
	if err != nil {
		return telego.InputFile{}, oops.Wrapf(err, "failed to read file: %s", picture.GetOriginal())
	}
	compressed, err := imgtool.CompressForTelegram(fileBytes)
	if err != nil {
		return telego.InputFile{}, oops.Wrapf(err, "failed to compress image")
	}
	return telegoutil.File(telegoutil.NameBytes(compressed, fmt.Sprintf("%s%s", strutil.MD5Hash(picture.GetOriginal()), ".jpg"))), nil
}

type SendArtworkInfoOptions struct {
	AppendCaption   string
	ReplyParameters *telego.ReplyParameters
	HasPermission   bool
}

type CreateArtworkInfoReplyMarkupOptions struct {
	CreatedArtwork bool
	HasPermission  bool
}

func CreateArtworkInfoReplyMarkup(ctx context.Context, meta *metautil.MetaData, serv *service.Service, artwork entity.ArtworkLike, controls *CreateArtworkInfoReplyMarkupOptions) (telego.ReplyMarkup, error) {
	if controls == nil {
		controls = &CreateArtworkInfoReplyMarkupOptions{}
	}
	if controls.CreatedArtwork {
		created, ok := artwork.(*entity.Artwork)
		if !ok {
			return nil, oops.New("artwork is not of type *entity.Artwork")
		}
		base := GetPostedArtworkInlineKeyboardButton(created, meta)
		if controls.HasPermission {
			return telegoutil.InlineKeyboard(
				base,
				telegoutil.InlineKeyboardRow(
					telegoutil.InlineKeyboardButton("更改R18").
						WithCallbackData(fmt.Sprintf("edit_artwork r18 %s %s", created.ID.Hex(), map[bool]string{true: "0", false: "1"}[created.R18])),
					telegoutil.InlineKeyboardButton("删除").
						WithCallbackData(fmt.Sprintf("delete_artwork %s", created.ID.Hex())),
				),
			), nil
		}
		return telegoutil.InlineKeyboard(base), nil
	}
	cbId, err := serv.CreateStringData(ctx, artwork.GetSourceURL())
	if err != nil {
		return nil, oops.Wrapf(err, "failed to create callback data")
	}
	previewKeyboard := []telego.InlineKeyboardButton{}
	if len(artwork.GetPictures()) > 1 {
		previewKeyboard = append(previewKeyboard, telegoutil.InlineKeyboardButton(fmt.Sprintf("删除这张(%d)", 1)).WithCallbackData(fmt.Sprintf("artwork_preview %s delete 0 0", cbId)))
		previewKeyboard = append(previewKeyboard, telegoutil.InlineKeyboardButton("下一张").WithCallbackData(fmt.Sprintf("artwork_preview %s preview 1 0", cbId)))
	}
	return telegoutil.InlineKeyboard(
		[]telego.InlineKeyboardButton{
			telegoutil.InlineKeyboardButton("发布").WithCallbackData(fmt.Sprintf("post_artwork %s", cbId)),
			telegoutil.InlineKeyboardButton("发布(反转R18)").WithCallbackData(fmt.Sprintf("post_artwork_r18 %s", cbId)),
		},
		[]telego.InlineKeyboardButton{
			telegoutil.InlineKeyboardButton("查重").WithCallbackData(fmt.Sprintf("search_picture %s", cbId)),
			telegoutil.InlineKeyboardButton("预览发布").WithURL(meta.BotDeepLink("info", cbId)),
		},
		previewKeyboard,
	), nil

}

// SendArtworkInfo 将作品信息附带操作按钮发送到指定聊天, 用于提供给管理员发布或修改作品
//
// 需要区分已发布的作品, 已标记为删除的作品, 和未发布的作品
func SendArtworkInfo(ctx *telegohandler.Context, meta *metautil.MetaData, serv *service.Service, sourceUrl string, chatID telego.ChatID, opts *SendArtworkInfoOptions) error {
	sourceUrl = serv.FindSourceURL(sourceUrl)
	if sourceUrl == "" {
		return oops.New("no valid source url found")
	}
	var artwork entity.ArtworkLike
	created := false
	if awent, err := serv.GetArtworkByURL(ctx, sourceUrl); err == nil {
		artwork = awent
		created = true
	} else if !errors.Is(err, errs.ErrRecordNotFound) {
		return oops.Wrapf(err, "failed to get artwork by url: %s", sourceUrl)
	}
	var deleted *entity.DeletedRecord
	if !created {
		if rec, err := serv.GetDeletedByURL(ctx, sourceUrl); err == nil {
			deleted = rec
		}
		cached, err := serv.GetOrFetchCachedArtwork(ctx, sourceUrl)
		if err != nil {
			return oops.Wrapf(err, "failed to get or fetch cached artwork by url: %s", sourceUrl)
		}
		artwork = cached
		// 再次检查是否已经发布, 主要解决某些源作品多个图片不同url时的问题
		if awent, err := serv.GetArtworkByURL(ctx, cached.SourceURL); err == nil {
			artwork = awent
			created = true
		}
	}
	if artwork == nil {
		return oops.New("no artwork found")
	}
	caption := ArtworkHTMLCaption(meta, artwork)
	caption += fmt.Sprintf("\n<i>该作品共有%d张图片</i>", len(artwork.GetPictures()))
	if deleted != nil {
		caption += fmt.Sprintf("\n<i>这是一个在 %s 被标记为删除的作品, 如果发布会取消删除</i>", deleted.DeletedAt.Format("2006-01-02 15:04:05"))
	}
	if opts != nil && opts.AppendCaption != "" {
		caption += "\n" + opts.AppendCaption
	}
	replyMarkup, err := CreateArtworkInfoReplyMarkup(ctx, meta, serv, artwork, &CreateArtworkInfoReplyMarkupOptions{
		CreatedArtwork: created,
		HasPermission:  opts != nil && opts.HasPermission,
	})
	if err != nil {
		return oops.Wrapf(err, "failed to create artwork info reply markup")
	}
	inputFile, err := GetPicturePhotoInputFile(ctx, serv, artwork.GetPictures()[0])
	if err != nil {
		return oops.Wrapf(err, "failed to get picture preview input file")
	}
	photo := telegoutil.Photo(chatID, inputFile).
		WithCaption(caption).WithReplyMarkup(replyMarkup).WithParseMode(telego.ModeHTML)
	if opts != nil && opts.ReplyParameters != nil {
		photo = photo.WithReplyParameters(opts.ReplyParameters)
	}
	if artwork.GetR18() {
		photo = photo.WithHasSpoiler()
	}
	_, err = ctx.Bot().SendPhoto(ctx, photo)
	if err != nil {
		return oops.Wrapf(err, "failed to send artwork info photo")
	}
	// [TODO] update artwork telegram info here
	// [TODO] lazy load input file and update preview
	return nil
}

type InputFileCloser struct {
	telego.InputFile
	CloseFunc func() error
}

func (i *InputFileCloser) Close() error {
	if i.CloseFunc != nil {
		return i.CloseFunc()
	}
	return nil
}

func GetPictureDocumentInputFile(ctx context.Context, serv *service.Service, picture entity.PictureLike) (*InputFileCloser, error) {
	if id := picture.GetTelegramInfo().DocumentFileID; id != "" {
		return &InputFileCloser{telegoutil.FileFromID(id), func() error { return nil }}, nil
	}
	orgStorDetail := picture.GetStorageInfo().Original
	if orgStorDetail != nil {
		rsc, err := serv.StorageGetFile(ctx, *picture.GetStorageInfo().Original)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to get file from storage")
		}
		ext, err := strutil.GetFileExtFromURL(picture.GetOriginal())
		if err != nil {
			mtype, err := mimetype.DetectReader(rsc)
			if mtype == nil {
				return nil, oops.New("failed to detect mime type")
			}
			ext = mtype.Extension()
			if err != nil {
				return nil, oops.Wrapf(err, "failed to detect mime type")
			}
			rsc.Seek(0, 0)
		}
		return &InputFileCloser{
			telegoutil.File(telegoutil.NameReader(rsc, fmt.Sprintf("%s%s", strutil.MD5Hash(picture.GetOriginal()), ext))),
			func() error {
				return rsc.Close()
			},
		}, nil
	}
	file, clean, err := httpclient.DownloadWithCache(ctx, picture.GetOriginal(), nil)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to download file: %s", picture.GetOriginal())
	}
	defer clean()
	return &InputFileCloser{
		telegoutil.File(file),
		func() error {
			defer clean()
			return file.Close()
		},
	}, nil
}
