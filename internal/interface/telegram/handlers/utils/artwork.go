package utils

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/pkg/imgtool"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/strutil"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

func SendArtworkMediaGroup(ctx *telegohandler.Context, chatID telego.ChatID, artwork shared.ArtworkLike) ([]telego.Message, error) {
	pics := artwork.GetPictures()
	caption := ArtworkHTMLCaption(metautil.FromContext(ctx), artwork)
	if len(pics) <= 10 {
		inputs, err := ArtworkInputMediaPhotos(ctx, service.FromContext(ctx), artwork, caption, 0, len(pics))
		if err != nil {
			return nil, oops.Wrapf(err, "failed to create input media photos")
		}
		// Send the media group
		return ctx.Bot().SendMediaGroup(ctx, telegoutil.MediaGroup(
			chatID,
			inputs...,
		))
	}
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
	artwork shared.ArtworkLike,
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
			photo = telegoutil.MediaPhoto(telegoutil.File(telegoutil.NameReader(bytes.NewReader(fileBytes), serv.PrettyFileName(artwork, picture))))
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

func GetPicturePhotoInputFile(ctx context.Context, serv *service.Service, picture shared.PictureLike) (telego.InputFile, error) {
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

func CreateArtworkInfoReplyMarkup(ctx context.Context, meta *metautil.MetaData, serv *service.Service, artwork shared.ArtworkLike, controls *CreateArtworkInfoReplyMarkupOptions) (*telego.InlineKeyboardMarkup, error) {
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
	return telegoutil.InlineKeyboard(
		[]telego.InlineKeyboardButton{
			telegoutil.InlineKeyboardButton("发布").WithCallbackData(fmt.Sprintf("post_artwork %s", cbId)),
			telegoutil.InlineKeyboardButton("发布(反转R18)").WithCallbackData(fmt.Sprintf("post_artwork_r18 %s", cbId)),
		},
		[]telego.InlineKeyboardButton{
			telegoutil.InlineKeyboardButton("查重").WithCallbackData(fmt.Sprintf("search_picture %s", cbId)),
			telegoutil.InlineKeyboardButton("预览").WithURL(meta.BotDeepLink("info", cbId)),
		},
	), nil

}

// SendArtworkInfo 将作品信息附带操作按钮发送到指定聊天, 用于提供给管理员发布或修改作品
//
// 需要区分已发布的作品, 已标记为删除的作品, 和未发布的作品
func SendArtworkInfo(ctx *telegohandler.Context,
	meta *metautil.MetaData,
	serv *service.Service,
	sourceUrl string,
	chatID telego.ChatID,
	opts SendArtworkInfoOptions) error {
	sourceUrl = serv.FindSourceURL(sourceUrl)
	if sourceUrl == "" {
		return oops.New("no valid source url found")
	}

	waitMsg, err := ctx.Bot().SendMessage(ctx, telegoutil.Message(chatID, "正在获取作品信息...").WithReplyParameters(opts.ReplyParameters))
	if err != nil {
		return oops.Wrapf(err, "failed to send wait message")
	}
	var artwork shared.ArtworkLike
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
	if opts.AppendCaption != "" {
		caption += "\n" + opts.AppendCaption
	}
	replyMarkup, err := CreateArtworkInfoReplyMarkup(ctx, meta, serv, artwork, &CreateArtworkInfoReplyMarkupOptions{
		CreatedArtwork: created,
		HasPermission:  opts.HasPermission,
	})
	if err != nil {
		return oops.Wrapf(err, "failed to create artwork info reply markup")
	}
	inputFile, err := GetPicturePhotoInputFile(ctx, serv, artwork.GetPictures()[0])
	if err != nil {
		return oops.Wrapf(err, "failed to get picture preview input file")
	}
	photo := telegoutil.MediaPhoto(inputFile).
		WithCaption(caption).WithParseMode(telego.ModeHTML)
	if artwork.GetR18() {
		photo = photo.WithHasSpoiler()
	}
	editReq := telegoutil.EditMessageMedia(chatID, waitMsg.MessageID, photo).WithReplyMarkup(replyMarkup)
	msg, err := ctx.Bot().EditMessageMedia(ctx, editReq)
	if err != nil {
		return oops.Wrapf(err, "failed to send artwork info photo")
	}
	if msg != nil && msg.Photo != nil {
		fileId := msg.Photo[len(msg.Photo)-1].FileID
		pic := artwork.GetPictures()[0]
		switch p := pic.(type) {
		case *entity.Picture:
			tginfo := p.GetTelegramInfo()
			tginfo.PhotoFileID = fileId
			return serv.UpdatePictureTelegramInfo(ctx, p.ID, &tginfo)
		case *entity.CachedPicture:
			switch aw := artwork.(type) {
			case *entity.CachedArtworkData:
				for _, p := range aw.Pictures {
					if p.GetOriginal() != pic.GetOriginal() {
						continue
					}
					tginfo := p.GetTelegramInfo()
					tginfo.PhotoFileID = fileId
					p.TelegramInfo = tginfo
					break
				}
				return serv.UpdateCachedArtwork(ctx, aw)
			case *entity.CachedArtwork:
				data := aw.Artwork.Data()
				for _, p := range data.Pictures {
					if p.GetOriginal() != pic.GetOriginal() {
						continue
					}
					tginfo := p.GetTelegramInfo()
					tginfo.PhotoFileID = fileId
					p.TelegramInfo = tginfo
					break
				}
				return serv.UpdateCachedArtwork(ctx, data)
			default:
				return oops.Errorf("unknown artwork type: %T", artwork)
			}
		}
	}
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

func GetPictureDocumentInputFile(ctx context.Context, serv *service.Service, artwork shared.ArtworkLike, picture shared.PictureLike) (*InputFileCloser, error) {
	if id := picture.GetTelegramInfo().DocumentFileID; id != "" {
		return &InputFileCloser{telegoutil.FileFromID(id), func() error { return nil }}, nil
	}
	orgStorDetail := picture.GetStorageInfo().Original
	if orgStorDetail != nil {
		rsc, err := serv.StorageGetFile(ctx, *picture.GetStorageInfo().Original)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to get file from storage")
		}
		return &InputFileCloser{
			telegoutil.File(telegoutil.NameReader(rsc, serv.PrettyFileName(artwork, picture))),
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
