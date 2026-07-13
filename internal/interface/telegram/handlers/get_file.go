package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/pkg/mediatool"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/ioutil"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

func GetArtworkFiles(ctx *telegohandler.Context, message telego.Message) error {
	var sourceURL string
	serv, err := requireService(ctx)
	if err != nil {
		return err
	}
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLInMessage(serv, message.ReplyToMessage)
	} else {
		sourceURL = serv.FindSourceURL(message.Text)
	}
	meta, err := requireMeta(ctx)
	if err != nil {
		return err
	}
	if sourceURL == "" {
		getPictureByHash := func() *entity.Picture {
			if message.ReplyToMessage == nil {
				return nil
			}
			file, err := utils.GetMessagePhotoFile(ctx, message.ReplyToMessage)
			if err != nil {
				return nil
			}
			hash, err := mediatool.GetImagePhashFromReader(bytes.NewReader(file))
			if err != nil {
				return nil
			}
			pictures, err := serv.QueryPicturesByPhash(ctx, query.PicturesPhash{
				Input:    hash,
				Limit:    1,
				Distance: 10,
			})
			if err != nil || len(pictures) == 0 {
				return nil
			}
			return pictures[0]
		}
		picture := getPictureByHash()
		if picture == nil {
			helpText := fmt.Sprintf(`
<b>使用 /files 命令回复一条含有图片或支持的链接的消息, 或在参数中提供作品链接, 将发送作品全部原图文件</b>

命令语法: %s
`, utils.EscapeHTML("/files [作品链接]"))
			utils.ReplyMessageWithHTML(ctx, message, helpText)
			return nil
		}
		return getArtworkFiles(ctx, serv, meta, message, picture.Artwork)
	}

	artwork, err := serv.GetArtworkByURL(ctx, sourceURL)
	if err != nil && !errors.Is(err, errs.ErrRecordNotFound) {
		utils.ReplyMessage(ctx, message, "获取作品信息失败")
		return oops.Wrapf(err, "failed to get artwork by url: %s", sourceURL)
	}

	if artwork == nil {
		artwork, err := serv.GetOrFetchCachedArtwork(ctx, sourceURL)
		if err != nil {
			utils.ReplyMessage(ctx, message, "获取作品信息失败")
			return oops.Wrapf(err, "failed to get artwork by url: %s", sourceURL)
		}
		return getArtworkFiles(ctx, serv, meta, message, artwork)
	}
	return getArtworkFiles(ctx, serv, meta, message, artwork)
}

func getArtworkFiles(ctx *telegohandler.Context,
	serv *service.Service,
	meta *metautil.MetaData,
	message telego.Message,
	artwork shared.ArtworkLike) error {
	msg, err := utils.ReplyMessage(ctx, message, "正在发送文件, 请稍等...")
	if err == nil {
		defer func() {
			ctx.Bot().DeleteMessage(ctx, telegoutil.Delete(msg.Chat.ChatID(), msg.MessageID))
		}()
	}
	var errs []error

	type fileMediaItem struct {
		picture shared.PictureLike
		ugoira  shared.UgoiraMetaLike
		video   shared.VideoLike
		index   int
		kind    string
	}
	const (
		fileMediaPicture = "picture"
		fileMediaUgoira  = "ugoira"
		fileMediaVideo   = "video"
	)

	items := make([]fileMediaItem, 0)
	for i, picture := range artwork.GetPictures() {
		items = append(items, fileMediaItem{picture: picture, index: i, kind: fileMediaPicture})
	}
	for i, ugoira := range artwork.GetUgoiraMetas() {
		items = append(items, fileMediaItem{ugoira: ugoira, index: i, kind: fileMediaUgoira})
	}
	for i, video := range artwork.GetVideos() {
		items = append(items, fileMediaItem{video: video, index: i, kind: fileMediaVideo})
	}
	if len(items) == 0 {
		return nil
	}

	var cachedData *entity.CachedArtworkData
	var cachedUpdated bool
	getCachedData := func() (*entity.CachedArtworkData, error) {
		if cachedData != nil {
			return cachedData, nil
		}
		cached, err := serv.GetCachedArtworkByURL(ctx, artwork.GetSourceURL())
		if err != nil {
			return nil, oops.Wrapf(err, "failed to get cached artwork by url: %s", artwork.GetSourceURL())
		}
		cachedData = cached.Artwork.Data()
		return cachedData, nil
	}

	sendBatch := func(start, end int) error {
		inputs := make([]telego.InputMedia, 0, end-start)
		closers := make([]func() error, 0, end-start)
		for i := start; i < end; i++ {
			item := items[i]
			var file *ioutil.Closer[telego.InputFile]
			var err error
			switch item.kind {
			case fileMediaPicture:
				file, err = utils.GetPictureDocumentInputFile(ctx, serv, meta, artwork, item.picture)
			case fileMediaUgoira:
				file, err = utils.GetUgoiraVideoDocumentInputFile(ctx, serv, meta, artwork, item.ugoira)
			case fileMediaVideo:
				file, err = utils.GetVideoDocumentInputFile(ctx, serv, meta, artwork, item.video)
			default:
				err = oops.Errorf("unknown media kind: %s", item.kind)
			}
			if err != nil {
				for _, close := range closers {
					close()
				}
				return err
			}
			closers = append(closers, file.Close)
			caption := artwork.GetTitle() + "_" + strconv.Itoa(item.index+1)
			doc := telegoutil.MediaDocument(file.Value).
				WithCaption(caption).
				WithDisableContentTypeDetection()
			inputs = append(inputs, telego.InputMedia(doc))
		}
		defer func() {
			for _, close := range closers {
				close()
			}
		}()

		group := telegoutil.MediaGroup(message.Chat.ChatID(), inputs...).WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		})
		msgs, err := ctx.Bot().SendMediaGroup(ctx, group)
		if err != nil {
			return oops.Wrapf(err, "failed to send media group")
		}

		for i, msg := range msgs {
			if msg.Document == nil {
				continue
			}
			item := items[start+i]
			switch item.kind {
			case fileMediaPicture:
				switch pic := item.picture.(type) {
				case *entity.Picture:
					tginfo := pic.GetTelegramInfo()
					tginfo.SetFileID(meta.BotID(), shared.TelegramMediaTypeDocument, msg.Document.FileID)
					if err := serv.UpdatePictureTelegramInfo(ctx, pic.ID, &tginfo); err != nil {
						return err
					}
				case *entity.CachedPicture:
					data, err := getCachedData()
					if err != nil {
						return err
					}
					for _, p := range data.Pictures {
						if p.Original == pic.GetOriginal() {
							tginfo := pic.GetTelegramInfo()
							tginfo.SetFileID(meta.BotID(), shared.TelegramMediaTypeDocument, msg.Document.FileID)
							p.TelegramInfo = tginfo
							cachedUpdated = true
							break
						}
					}
				default:
					log.Warnf("unknown picture type: %T", pic)
				}
			case fileMediaUgoira:
				switch ugo := item.ugoira.(type) {
				case *entity.UgoiraMeta:
					tginfo := ugo.GetTelegramInfo()
					tginfo.SetFileID(meta.BotID(), shared.TelegramMediaTypeDocument, msg.Document.FileID)
					if err := serv.UpdateUgoiraTelegramInfo(ctx, ugo.ID, &tginfo); err != nil {
						return err
					}
				case *entity.CachedUgoiraMeta:
					data, err := getCachedData()
					if err != nil {
						return err
					}
					for _, u := range data.UgoiraMetas {
						if u.MetaData.OriginalZip == ugo.MetaData.OriginalZip {
							tginfo := ugo.GetTelegramInfo()
							tginfo.SetFileID(meta.BotID(), shared.TelegramMediaTypeDocument, msg.Document.FileID)
							u.TelegramInfo = tginfo
							cachedUpdated = true
							break
						}
					}
				default:
					log.Warnf("unknown ugoira type: %T", ugo)
				}
			case fileMediaVideo:
				switch vid := item.video.(type) {
				case *entity.Video:
					tginfo := vid.GetTelegramInfo()
					tginfo.SetFileID(meta.BotID(), shared.TelegramMediaTypeDocument, msg.Document.FileID)
					if err := serv.UpdateVideoTelegramInfo(ctx, vid.ID, &tginfo); err != nil {
						return err
					}
				case *entity.CachedVideo:
					data, err := getCachedData()
					if err != nil {
						return err
					}
					for _, v := range data.Videos {
						if v.URL == vid.URL {
							tginfo := vid.GetTelegramInfo()
							tginfo.SetFileID(meta.BotID(), shared.TelegramMediaTypeDocument, msg.Document.FileID)
							v.TelegramInfo = tginfo
							cachedUpdated = true
							break
						}
					}
				default:
					log.Warnf("unknown video type: %T", vid)
				}
			}
		}
		return nil
	}

	for i := 0; i < len(items); i += 10 {
		end := min(i+10, len(items))
		if err := sendBatch(i, end); err != nil {
			errs = append(errs, oops.Wrapf(err, "failed to send files %d-%d", i+1, end))
		}
	}

	if cachedUpdated && cachedData != nil {
		if err := serv.UpdateCachedArtwork(ctx, cachedData); err != nil {
			errs = append(errs, oops.Wrapf(err, "failed to update cached artwork"))
		}
	}

	return oops.Join(errs...)
}
