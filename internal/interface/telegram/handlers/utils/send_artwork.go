package utils

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/kvstor"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/pkg/mediatool"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/ioutil"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/osutil"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
	"github.com/unvgo/ouid"
)

func GetPicturePhotoInputFile(ctx context.Context, serv *service.Service, meta *metautil.MetaData, picture shared.PictureLike) (*ioutil.Closer[telego.InputFile], error) {
	if id := picture.GetTelegramInfo().PhotoFileID(meta.BotID()); id != "" {
		return ioutil.NewCloser(telegoutil.FileFromID(id), nil), nil
	}
	orgStorDetail := picture.GetStorageInfo().Original
	if orgStorDetail != nil {
		file, err := serv.StorageGetFile(ctx, *picture.GetStorageInfo().Original)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to get file from storage")
		}
		defer file.Close()
		compressed, err := mediatool.CompressImgForTelegramFromFile(file.Name())
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
	compressed, err := mediatool.CompressImgForTelegramFromFile(file.Name())
	if err != nil {
		return nil, oops.Wrapf(err, "failed to compress image")
	}
	return ioutil.NewCloser(telegoutil.File(compressed), func() error { return compressed.Close() }), nil
}

func GetUgoiraVideoInputFile(ctx context.Context, serv *service.Service, meta *metautil.MetaData, ugoira shared.UgoiraMetaLike) (*ioutil.Closer[telego.InputFile], error) {
	if id := ugoira.GetTelegramInfo().VideoFileID(meta.BotID()); id != "" {
		return ioutil.NewCloser(telegoutil.FileFromID(id), func() error { return nil }), nil
	}
	storDetail := ugoira.GetOriginalStorage()
	if storDetail != shared.ZeroStorageDetail {
		file, err := serv.StorageGetFile(ctx, storDetail)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to get file from storage")
		}
		defer file.Close()
		videoPath, err := mediatool.UgoiraZipToMp4(file.Name(), ugoira.GetUgoiraMetaData().Frames, file.Name()+".mp4")
		if err != nil {
			return nil, oops.Wrapf(err, "failed to convert ugoira to mp4")
		}
		videoFile, err := osutil.OpenTemp(videoPath)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to open mp4 file")
		}
		return ioutil.NewCloser(telegoutil.File(videoFile), func() error { return videoFile.Close() }), nil
	}
	file, err := httpclient.DownloadWithCache(ctx, ugoira.GetUgoiraMetaData().OriginalZip, nil)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to download file: %s", ugoira.GetUgoiraMetaData().OriginalZip)
	}
	defer file.Close()
	videoPath, err := mediatool.UgoiraZipToMp4(file.Name(), ugoira.GetUgoiraMetaData().Frames, file.Name()+".mp4")
	if err != nil {
		return nil, oops.Wrapf(err, "failed to convert ugoira to mp4")
	}
	videoFile, err := osutil.OpenTemp(videoPath)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to open mp4 file")
	}
	return ioutil.NewCloser(telegoutil.File(videoFile), func() error { return videoFile.Close() }), nil
}

func GetVideoVideoInputFile(ctx context.Context, serv *service.Service, meta *metautil.MetaData, video shared.VideoLike) (*ioutil.Closer[telego.InputFile], error) {
	if id := video.GetTelegramInfo().VideoFileID(meta.BotID()); id != "" {
		return ioutil.NewCloser(telegoutil.FileFromID(id), func() error { return nil }), nil
	}
	orgStorDetail := video.GetOriginalStorage()
	if orgStorDetail != shared.ZeroStorageDetail {
		file, err := serv.StorageGetFile(ctx, orgStorDetail)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to get file from storage")
		}
		return ioutil.NewCloser(telegoutil.File(file), func() error { return file.Close() }), nil
	}
	file, err := httpclient.DownloadWithCache(ctx, video.GetURL(), nil)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to download file: %s", video.GetURL())
	}
	return ioutil.NewCloser(telegoutil.File(file), func() error { return file.Close() }), nil
}

// MediaLike 转换为对应的 InputMedia
//
// 第一个 Media 意为 MediaLike, 第二个 Media 意为tg的 MediaInputFile
func GetMediaInputMedia(ctx context.Context,
	serv *service.Service,
	meta *metautil.MetaData,
	artwork shared.ArtworkLike,
	media shared.MediaLike) (*ioutil.Closer[telego.InputMedia], error) {
	switch m := media.(type) {
	case shared.PictureLike:
		closer, err := GetPicturePhotoInputFile(ctx, serv, meta, m)
		if err != nil {
			return nil, err
		}
		media := telegoutil.MediaPhoto(closer.Value)
		return ioutil.NewCloser(telego.InputMedia(media), func() error { return closer.Close() }), nil
	case shared.UgoiraMetaLike:
		closer, err := GetUgoiraVideoInputFile(ctx, serv, meta, m)
		if err != nil {
			return nil, err
		}
		media := telegoutil.MediaVideo(closer.Value)
		return ioutil.NewCloser(telego.InputMedia(media), func() error { return closer.Close() }), nil
	case shared.VideoLike:
		closer, err := GetVideoVideoInputFile(ctx, serv, meta, m)
		if err != nil {
			return nil, err
		}
		media := telegoutil.MediaVideo(closer.Value)
		return ioutil.NewCloser(telego.InputMedia(media), func() error { return closer.Close() }), nil
	default:
		return nil, oops.New("unsupported media type")
	}
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
	cbId := ouid.New().Hex()
	err := kvstor.SetWithTTL(ctx, cbId, artwork.GetSourceURL(), time.Hour*24*7)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to create callback data")
	}
	return telegoutil.InlineKeyboard(
		ArtworkPostKeyboard(meta, cbId)...,
	), nil
}

func ArtworkPostKeyboard(meta *metautil.MetaData, cbId string) [][]telego.InlineKeyboardButton {
	base := [][]telego.InlineKeyboardButton{
		{
			telegoutil.InlineKeyboardButton("发布").WithCallbackData(fmt.Sprintf("post_artwork %s", cbId)),
			telegoutil.InlineKeyboardButton("发布(反转R18)").WithCallbackData(fmt.Sprintf("post_artwork_r18 %s", cbId)),
		},
		{
			telegoutil.InlineKeyboardButton("查重").WithCallbackData(fmt.Sprintf("search_picture %s", cbId)),
			telegoutil.InlineKeyboardButton("预览").WithURL(meta.BotDeepLink("info", cbId)),
		},
	}
	if extra := runtimecfg.Get().Telegram.ExtraTarget; len(extra) > 0 {
		row := []telego.InlineKeyboardButton{}
		for _, target := range extra {
			// 两个一行
			btn := telegoutil.InlineKeyboardButton(fmt.Sprintf("发到 %s", target.Title))
			btn = btn.WithCallbackData(fmt.Sprintf("sendto %s %d", cbId, target.ChatID))
			row = append(row, btn)
			if len(row) >= 2 {
				base = append(base, row)
				row = []telego.InlineKeyboardButton{}
			}
		}
		if len(row) > 0 {
			base = append(base, row)
		}
	}
	return base
}

// 将作品信息附带操作按钮发送到指定聊天, 用于提供给管理员发布或修改作品
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
			cbId := ouid.New().Hex()
			err := kvstor.SetWithTTL(ctx, cbId, sourceUrl, time.Hour*24*7)
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
	caption += fmt.Sprintf("\n<i>该作品共有%d个媒体</i>", artwork.MediasCount())
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
	inputMedia, err := GetMediaInputMedia(ctx, serv, meta, artwork, artwork.FirstMedia())
	if err != nil {
		return oops.Wrapf(err, "failed to get picture preview input file")
	}
	defer func() {
		err := inputMedia.Close()
		if err != nil {
			log.Errorf("failed to close input file: %s", err)
		}
	}()
	media := inputMedia.Value
	// golang's type assertion is so annoying
	var mediaPhoto *telego.InputMediaPhoto
	var mediaVideo *telego.InputMediaVideo
	switch m := media.(type) {
	case *telego.InputMediaPhoto:
		m.WithCaption(caption).WithParseMode(telego.ModeHTML)
		if artwork.GetR18() {
			m.WithHasSpoiler()
		}
		media = m
		mediaPhoto = m
	case *telego.InputMediaVideo:
		m.WithCaption(caption).WithParseMode(telego.ModeHTML)
		if artwork.GetR18() {
			m.WithHasSpoiler()
		}
		media = m
		mediaVideo = m
	default:
		return oops.New("unsupported media type")
	}

	updateMediaFileID := func(msg *telego.Message) error {
		if msg != nil && msg.Photo != nil && len(artwork.GetPictures()) > 0 {
			fileId := msg.Photo[len(msg.Photo)-1].FileID
			pic := artwork.GetPictures()[0]
			switch p := pic.(type) {
			case *entity.Picture:
				tginfo := p.GetTelegramInfo()
				tginfo.SetFileID(meta.BotID(), shared.TelegramMediaTypePhoto, fileId)
				return serv.UpdatePictureTelegramInfo(ctx, p.ID, &tginfo)
			case *entity.CachedPicture:
				switch aw := artwork.(type) {
				case *entity.CachedArtworkData:
					for _, p := range aw.Pictures {
						if p.GetOriginal() != pic.GetOriginal() {
							continue
						}
						tginfo := p.GetTelegramInfo()
						tginfo.SetFileID(meta.BotID(), shared.TelegramMediaTypePhoto, fileId)
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
						tginfo.SetFileID(meta.BotID(), shared.TelegramMediaTypePhoto, fileId)
						p.TelegramInfo = tginfo
						break
					}
					return serv.UpdateCachedArtwork(ctx, data)
				default:
					return oops.Errorf("unknown artwork type: %T", artwork)
				}
			}
		}
		if msg != nil && msg.Video != nil && (len(artwork.GetVideos()) > 0 || len(artwork.GetUgoiraMetas()) > 0) {
			// 如果 artwork 有 ugoira 则这里应该是 ugoira, 详见各个 FirstMedia 实现
			if len(artwork.GetUgoiraMetas()) > 0 {
				ugoira := artwork.GetUgoiraMetas()[0]
				switch v := ugoira.(type) {
				case *entity.UgoiraMeta:
					tginfo := v.GetTelegramInfo()
					tginfo.SetFileID(meta.BotID(), shared.TelegramMediaTypeVideo, msg.Video.FileID)
					return serv.UpdateUgoiraTelegramInfo(ctx, v.ID, &tginfo)
				case *entity.CachedUgoiraMeta:
					switch aw := artwork.(type) {
					case *entity.CachedArtworkData:
						for _, um := range aw.UgoiraMetas {
							if um.GetUgoiraMetaData().OriginalZip != ugoira.GetUgoiraMetaData().OriginalZip {
								continue
							}
							tginfo := um.GetTelegramInfo()
							tginfo.SetFileID(meta.BotID(), shared.TelegramMediaTypeVideo, msg.Video.FileID)
							um.TelegramInfo = tginfo
							break
						}
						return serv.UpdateCachedArtwork(ctx, aw)
					case *entity.CachedArtwork:
						data := aw.Artwork.Data()
						for _, um := range data.UgoiraMetas {
							if um.GetUgoiraMetaData().OriginalZip != ugoira.GetUgoiraMetaData().OriginalZip {
								continue
							}
							tginfo := um.GetTelegramInfo()
							tginfo.SetFileID(meta.BotID(), shared.TelegramMediaTypeVideo, msg.Video.FileID)
							um.TelegramInfo = tginfo
							break
						}
						return serv.UpdateCachedArtwork(ctx, data)
					default:
						return oops.Errorf("unknown artwork type: %T", artwork)
					}
				}
				return nil
			}
			video := artwork.GetVideos()[0]
			switch v := video.(type) {
			case *entity.Video:
				tginfo := v.GetTelegramInfo()
				tginfo.SetFileID(meta.BotID(), shared.TelegramMediaTypeVideo, msg.Video.FileID)
				return serv.UpdateVideoTelegramInfo(ctx, v.ID, &tginfo)
			case *entity.CachedVideo:
				switch aw := artwork.(type) {
				case *entity.CachedArtworkData:
					for _, vid := range aw.Videos {
						if vid.GetURL() != video.GetURL() {
							continue
						}
						tginfo := vid.GetTelegramInfo()
						tginfo.SetFileID(meta.BotID(), shared.TelegramMediaTypeVideo, msg.Video.FileID)
						vid.TelegramInfo = tginfo
						break
					}
					return serv.UpdateCachedArtwork(ctx, aw)
				case *entity.CachedArtwork:
					data := aw.Artwork.Data()
					for _, vid := range data.Videos {
						if vid.GetURL() != video.GetURL() {
							continue
						}
						tginfo := vid.GetTelegramInfo()
						tginfo.SetFileID(meta.BotID(), shared.TelegramMediaTypeVideo, msg.Video.FileID)
						vid.TelegramInfo = tginfo
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
		editReq := telegoutil.EditMessageMedia(chatID, waitMsg.MessageID, media).WithReplyMarkup(replyMarkup)
		msg, err := bot.EditMessageMedia(ctx, editReq)
		if err != nil {
			return oops.Wrapf(err, "failed to send artwork info media")
		}
		return updateMediaFileID(msg)
	}
	if mediaPhoto != nil {
		sendPhoto := telegoutil.Photo(chatID, mediaPhoto.Media).
			WithCaption(mediaPhoto.Caption).
			WithParseMode(telego.ModeHTML).
			WithReplyParameters(opts.ReplyParameters).
			WithReplyMarkup(replyMarkup)
		if artwork.GetR18() {
			sendPhoto = sendPhoto.WithHasSpoiler()
		}
		msg, err := bot.SendPhoto(ctx, sendPhoto)
		if err != nil {
			return oops.Wrapf(err, "failed to send artwork info photo")
		}
		return updateMediaFileID(msg)
	}
	if mediaVideo != nil {
		sendVideo := telegoutil.Video(chatID, mediaVideo.Media).
			WithCaption(mediaVideo.Caption).
			WithParseMode(telego.ModeHTML).
			WithReplyParameters(opts.ReplyParameters).
			WithReplyMarkup(replyMarkup)
		if artwork.GetR18() {
			sendVideo = sendVideo.WithHasSpoiler()
		}
		msg, err := bot.SendVideo(ctx, sendVideo)
		if err != nil {
			return oops.Wrapf(err, "failed to send artwork info video")
		}
		return updateMediaFileID(msg)
	}

	return nil
}

func GetPictureDocumentInputFile(ctx context.Context, serv *service.Service, meta *metautil.MetaData, artwork shared.ArtworkLike, picture shared.PictureLike) (*ioutil.Closer[telego.InputFile], error) {
	if id := picture.GetTelegramInfo().DocumentFileID(meta.BotID()); id != "" {
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

func GetUgoiraVideoDocumentInputFile(ctx context.Context, serv *service.Service, meta *metautil.MetaData, artwork shared.ArtworkLike, ugoira shared.UgoiraMetaLike) (*ioutil.Closer[telego.InputFile], error) {
	if id := ugoira.GetTelegramInfo().DocumentFileID(meta.BotID()); id != "" {
		return ioutil.NewCloser(telegoutil.FileFromID(id), func() error { return nil }), nil
	}

	data := ugoira.GetUgoiraMetaData()

	buildInput := func(file interface {
		Name() string
		Close() error
	}) (*ioutil.Closer[telego.InputFile], error) {
		defer file.Close()

		outputPath := file.Name()[0:len(file.Name())-len(filepath.Ext(file.Name()))] + ".mp4"
		videoPath, err := mediatool.UgoiraZipToMp4(file.Name(), data.Frames, outputPath)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to convert ugoira to mp4")
		}
		videoFile, err := osutil.OpenTemp(videoPath)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to open temp video file")
		}
		return ioutil.NewCloser(telegoutil.File(videoFile), func() error { return videoFile.Close() }), nil
	}

	orgStorDetail := ugoira.GetOriginalStorage()
	if orgStorDetail != shared.ZeroStorageDetail {
		file, err := serv.StorageGetFile(ctx, orgStorDetail)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to get ugoira file from storage")
		}
		return buildInput(file)
	}

	file, err := httpclient.DownloadWithCache(ctx, data.OriginalZip, nil)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to download ugoira file: %s", data.OriginalZip)
	}
	return buildInput(file)
}

func GetVideoDocumentInputFile(ctx context.Context, serv *service.Service, meta *metautil.MetaData, artwork shared.ArtworkLike, video shared.VideoLike) (*ioutil.Closer[telego.InputFile], error) {
	if id := video.GetTelegramInfo().DocumentFileID(meta.BotID()); id != "" {
		return ioutil.NewNoopCloser(telegoutil.FileFromID(id)), nil
	}
	orgStorDetail := video.GetOriginalStorage()
	if !orgStorDetail.IsZero() {
		rsc, err := serv.StorageGetFile(ctx, orgStorDetail)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to get video file from storage")
		}
		return ioutil.NewCloser(telegoutil.File(rsc), func() error { return rsc.Close() }), nil
	}
	file, err := httpclient.DownloadWithCache(ctx, video.GetURL(), nil)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to download video file: %s", video.GetURL())
	}
	return ioutil.NewCloser(telegoutil.File(file), func() error { return file.Close() }), nil
}
