package utils

import (
	"context"
	"errors"
	"fmt"

	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/internal/infra/cache"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/pkg/imgtool"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/ioutil"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"github.com/krau/ManyACG/pkg/osutil"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

func GetPicturePhotoInputFile(ctx context.Context, serv *service.Service, picture shared.PictureLike) (*ioutil.Closer[telego.InputFile], error) {
	if id := picture.GetTelegramInfo().PhotoFileID; id != "" {
		return ioutil.NewCloser(telegoutil.FileFromID(id), nil), nil
	}
	orgStorDetail := picture.GetStorageInfo().Original
	if orgStorDetail != nil {
		file, err := serv.StorageGetFile(ctx, *picture.GetStorageInfo().Original)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to get file from storage")
		}
		defer file.Close()
		compressed, err := imgtool.CompressForTelegramFromFile(file.Name())
		if err != nil {
			return nil, oops.Wrapf(err, "failed to compress image")
		}
		return ioutil.NewCloser(telegoutil.File(compressed), func() error { return compressed.Close() }), nil
	}
	file, err := httpclient.DownloadWithCache(ctx, picture.GetOriginal(), nil)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to download file: %s", picture.GetOriginal())
	}
	defer file.Close()
	compressed, err := imgtool.CompressForTelegramFromFile(file.Name())
	if err != nil {
		return nil, oops.Wrapf(err, "failed to compress image")
	}
	return ioutil.NewCloser(telegoutil.File(compressed), func() error { return compressed.Close() }), nil
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

func CreateArtworkInfoReplyMarkup(ctx context.Context,
	meta *metautil.MetaData,
	serv *service.Service,
	artwork shared.ArtworkLike,
	controls *CreateArtworkInfoReplyMarkupOptions) (*telego.InlineKeyboardMarkup, error) {
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
	cbId := objectuuid.New().Hex()
	err := cache.Set(ctx, cbId, artwork.GetSourceURL())
	if err != nil {
		return nil, oops.Wrapf(err, "failed to create callback data")
	}
	return telegoutil.InlineKeyboard(
		ArtworkPostKeyboard(meta, cbId)...,
	), nil
}

func ArtworkPostKeyboard(meta *metautil.MetaData, cbId string) [][]telego.InlineKeyboardButton {
	return [][]telego.InlineKeyboardButton{
		{
			telegoutil.InlineKeyboardButton("发布").WithCallbackData(fmt.Sprintf("post_artwork %s", cbId)),
			telegoutil.InlineKeyboardButton("发布(反转R18)").WithCallbackData(fmt.Sprintf("post_artwork_r18 %s", cbId)),
		},
		{
			telegoutil.InlineKeyboardButton("查重").WithCallbackData(fmt.Sprintf("search_picture %s", cbId)),
			telegoutil.InlineKeyboardButton("预览").WithURL(meta.BotDeepLink("info", cbId)),
		},
	}
}

// SendArtworkInfo 将作品信息附带操作按钮发送到指定聊天, 用于提供给管理员发布或修改作品
//
// 需要区分已发布的作品, 已标记为删除的作品, 和未发布的作品
func SendArtworkInfo(ctx context.Context,
	bot *telego.Bot,
	meta *metautil.MetaData,
	serv *service.Service,
	sourceUrl string,
	chatID telego.ChatID,
	opts SendArtworkInfoOptions) error {
	sourceUrl = serv.FindSourceURL(sourceUrl)
	if sourceUrl == "" {
		return oops.New("no valid source url found")
	}
	var waitMsg *telego.Message
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
		cached, err := serv.GetCachedArtworkByURL(ctx, sourceUrl)
		if err == nil {
			artwork = cached
		} else if !errors.Is(err, errs.ErrRecordNotFound) {
			return oops.Wrapf(err, "failed to get cached artwork by url: %s", sourceUrl)
		}
		if artwork == nil {
			// 既没有发布也没有缓存, 则尝试抓取
			cbId := objectuuid.New().Hex()
			err := cache.Set(ctx, cbId, artwork.GetSourceURL())
			if err != nil {
				return oops.Wrapf(err, "failed to create callback data")
			}
			waitMsg, err = bot.SendMessage(ctx, telegoutil.
				Message(chatID, sourceUrl+"\n正在获取作品信息...").
				WithReplyParameters(opts.ReplyParameters).
				WithReplyMarkup(telegoutil.InlineKeyboard(
					ArtworkPostKeyboard(meta, cbId)...,
				)))
			if err != nil {
				return oops.Wrapf(err, "failed to send wait message")
			}
			cached, err = serv.GetOrFetchCachedArtwork(ctx, sourceUrl)
			if err != nil {
				return oops.Wrapf(err, "failed to get or fetch cached artwork by url: %s", sourceUrl)
			}
			artwork = cached
		}
		// 再次检查是否已经发布, 主要解决某些源作品多个图片不同url时的问题
		if awent, err := serv.GetArtworkByURL(ctx, cached.SourceURL); err == nil {
			artwork = awent
			created = true
		}
	}
	if artwork == nil {
		return oops.New("no artwork found")
	}
	caption := ArtworkHTMLCaption(artwork)
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
	defer func() {
		err := inputFile.Close()
		if err != nil {
			log.Errorf("failed to close input file: %s", err)
		}
	}()
	photo := telegoutil.MediaPhoto(inputFile.Value).
		WithCaption(caption).WithParseMode(telego.ModeHTML)
	if artwork.GetR18() {
		photo = photo.WithHasSpoiler()
	}

	updatePictureFileID := func(msg *telego.Message) error {
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
		return nil
	}

	if waitMsg != nil {
		editReq := telegoutil.EditMessageMedia(chatID, waitMsg.MessageID, photo).WithReplyMarkup(replyMarkup)
		msg, err := bot.EditMessageMedia(ctx, editReq)
		if err != nil {
			return oops.Wrapf(err, "failed to send artwork info photo")
		}
		return updatePictureFileID(msg)
	}
	sendPhoto := telegoutil.Photo(chatID, inputFile.Value).
		WithCaption(caption).WithParseMode(telego.ModeHTML).
		WithReplyParameters(opts.ReplyParameters).
		WithReplyMarkup(replyMarkup)
	if artwork.GetR18() {
		sendPhoto = sendPhoto.WithHasSpoiler()
	}
	msg, err := bot.SendPhoto(ctx, sendPhoto)
	if err != nil {
		return oops.Wrapf(err, "failed to send artwork info photo")
	}
	return updatePictureFileID(msg)
}

func GetPictureDocumentInputFile(ctx context.Context, serv *service.Service, artwork shared.ArtworkLike, picture shared.PictureLike) (*ioutil.Closer[telego.InputFile], error) {
	if id := picture.GetTelegramInfo().DocumentFileID; id != "" {
		return ioutil.NewCloser(telegoutil.FileFromID(id), func() error { return nil }), nil
	}
	orgStorDetail := picture.GetStorageInfo().Original
	if orgStorDetail != nil {
		rsc, err := serv.StorageGetFile(ctx, *picture.GetStorageInfo().Original)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to get file from storage")
		}
		return ioutil.NewCloser(telegoutil.File(telegoutil.NameReader(rsc, serv.PrettyFileName(artwork, picture))), func() error { return rsc.Close() }), nil
	}
	file, err := httpclient.DownloadWithCache(ctx, picture.GetOriginal(), nil)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to download file: %s", picture.GetOriginal())
	}
	return ioutil.NewCloser(telegoutil.File(telegoutil.NameReader(file, serv.PrettyFileName(artwork, picture))), func() error { return file.Close() }), nil
}

func GetUgoiraVideoDocumentInputFile(ctx context.Context, serv *service.Service, artwork shared.UgoiraArtworkLike, ugoira shared.UgoiraMetaLike) (*ioutil.Closer[telego.InputFile], error) {
	if id := ugoira.GetTelegramInfo().DocumentFileID; id != "" {
		return ioutil.NewCloser(telegoutil.FileFromID(id), func() error { return nil }), nil
	}
	data := ugoira.GetUgoiraMetaData()
	orgStorDetail := ugoira.GetOriginalStorage()
	if orgStorDetail != shared.ZeroStorageDetail {
		file, err := serv.StorageGetFile(ctx, orgStorDetail)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to get ugoira file from storage")
		}
		defer file.Close()
		videoPath, err := imgtool.UgoiraZipToMp4(file.Name(), data.Frames, file.Name()+".mp4")
		if err != nil {
			return nil, oops.Wrapf(err, "failed to convert ugoira to mp4")
		}
		videoFile, err := osutil.OpenTemp(videoPath)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to open temp video file")
		}
		return ioutil.NewCloser(telegoutil.File(videoFile), func() error { return videoFile.Close() }), nil
	}
	file, err := httpclient.DownloadWithCache(ctx, data.OriginalZip, nil)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to download ugoira file: %s", data.OriginalZip)
	}
	defer file.Close()
	videoPath, err := imgtool.UgoiraZipToMp4(file.Name(), data.Frames, file.Name()+".mp4")
	if err != nil {
		return nil, oops.Wrapf(err, "failed to convert ugoira to mp4")
	}
	videoFile, err := osutil.OpenTemp(videoPath)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to open temp video file")
	}
	return ioutil.NewCloser(telegoutil.File(videoFile), func() error { return videoFile.Close() }), nil
}
