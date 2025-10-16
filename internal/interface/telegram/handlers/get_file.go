package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"strconv"

	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/pkg/imgtool"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

func GetArtworkFiles(ctx *telegohandler.Context, message telego.Message) error {
	var sourceURL string
	serv := service.FromContext(ctx)
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLInMessage(serv, message.ReplyToMessage)
	} else {
		sourceURL = serv.FindSourceURL(message.Text)
	}
	meta := metautil.FromContext(ctx)
	if sourceURL == "" {
		getPictureByHash := func() *entity.Picture {
			if message.ReplyToMessage == nil {
				return nil
			}
			file, err := utils.GetMessagePhotoFile(ctx, message.ReplyToMessage)
			if err != nil {
				return nil
			}
			hash, err := imgtool.GetImagePhashFromReader(bytes.NewReader(file))
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
`, html.EscapeString("/files [作品链接]"))
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
	for i, picture := range artwork.GetPictures() {
		err := func() error {
			buildDocument := func() (*telego.SendDocumentParams, func() error, error) {
				file, err := utils.GetPictureDocumentInputFile(ctx, serv, meta, artwork, picture)
				if err != nil {
					return nil, nil, oops.Wrapf(err, "failed to get picture document input file")
				}
				document := telegoutil.Document(message.Chat.ChatID(), file.Value).
					WithReplyParameters(&telego.ReplyParameters{
						MessageID: message.MessageID,
					}).WithCaption(artwork.GetTitle() + "_" + strconv.Itoa(i+1)).WithDisableContentTypeDetection()
				if meta.ChannelAvailable() && picture.GetTelegramInfo().MessageID(meta.ChannelChatID().ID) != 0 {
					document.WithReplyMarkup(telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
						telegoutil.InlineKeyboardButton("详情").WithURL(meta.ChannelMessageURL(picture.GetTelegramInfo().MessageID(meta.ChannelChatID().ID))),
					}))
				} else {
					document.WithReplyMarkup(telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
						telegoutil.InlineKeyboardButton("详情").WithURL(artwork.GetSourceURL()),
					}))
				}
				return document, file.Close, nil
			}

			document, close, err := buildDocument()
			if err != nil {
				return oops.Wrapf(err, "failed to build document")
			}
			defer close()
			documentMessage, err := ctx.Bot().SendDocument(ctx, document)
			if err != nil {
				ctx.Bot().SendMessage(ctx, telegoutil.Messagef(
					message.Chat.ChatID(),
					"发送第 %d 张图片时失败",
					i+1,
				).WithReplyParameters(&telego.ReplyParameters{
					MessageID: message.MessageID,
				}))
				return oops.Wrapf(err, "failed to send document")
			}
			if documentMessage != nil && documentMessage.Document != nil {
				switch pic := picture.(type) {
				case *entity.Picture:
					tginfo := pic.GetTelegramInfo()
					tginfo.SetFileID(meta.BotID(), shared.TelegramMediaTypeDocument, documentMessage.Document.FileID)
					return serv.UpdatePictureTelegramInfo(ctx, pic.ID, &tginfo)
				case *entity.CachedPicture:
					cached, err := serv.GetCachedArtworkByURL(ctx, artwork.GetSourceURL())
					if err != nil {
						return oops.Wrapf(err, "failed to get cached artwork by url: %s", artwork.GetSourceURL())
					}
					data := cached.Artwork.Data()
					for _, p := range data.Pictures {
						if p.Original == pic.GetOriginal() {
							tginfo := pic.GetTelegramInfo()
							tginfo.SetFileID(meta.BotID(), shared.TelegramMediaTypeDocument, documentMessage.Document.FileID)
							p.TelegramInfo = tginfo
							return serv.UpdateCachedArtwork(ctx, data)
						}
					}
				default:
					log.Warnf("unknown picture type: %T", pic)
				}
			}
			return nil
		}()
		if err != nil {
			errs = append(errs, oops.Wrapf(err, "failed to send picture %d file", i+1))
		}
	}
	awUgoira, ok := artwork.(shared.UgoiraArtworkLike)
	if !ok || awUgoira.GetUgoiraMetas() == nil {
		return oops.Join(errs...)
	}
	for i, ugoira := range awUgoira.GetUgoiraMetas() {
		err := func() error {
			buildDocument := func() (*telego.SendDocumentParams, func() error, error) {
				file, err := utils.GetUgoiraVideoDocumentInputFile(ctx, serv, meta, awUgoira, ugoira)
				if err != nil {
					return nil, nil, oops.Wrapf(err, "failed to get ugoira video document input file")
				}
				document := telegoutil.Document(message.Chat.ChatID(), file.Value).
					WithReplyParameters(&telego.ReplyParameters{
						MessageID: message.MessageID,
					}).WithCaption(artwork.GetTitle() + "_" + strconv.Itoa(i+1)).WithDisableContentTypeDetection()
				if meta.ChannelAvailable() && ugoira.GetTelegramInfo().MessageID(meta.ChannelChatID().ID) != 0 {
					document.WithReplyMarkup(telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
						telegoutil.InlineKeyboardButton("详情").WithURL(meta.ChannelMessageURL(ugoira.GetTelegramInfo().MessageID(meta.ChannelChatID().ID))),
					}))
				} else {
					document.WithReplyMarkup(telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
						telegoutil.InlineKeyboardButton("详情").WithURL(artwork.GetSourceURL()),
					}))
				}
				return document, file.Close, nil
			}
			document, close, err := buildDocument()
			if err != nil {
				return oops.Wrapf(err, "failed to build document")
			}
			defer close()
			documentMessage, err := ctx.Bot().SendDocument(ctx, document)
			if err != nil {
				ctx.Bot().SendMessage(ctx, telegoutil.Messagef(
					message.Chat.ChatID(),
					"发送第 %d 个动图时失败",
					i+1,
				).WithReplyParameters(&telego.ReplyParameters{
					MessageID: message.MessageID,
				}))
				return oops.Wrapf(err, "failed to send document")
			}
			if documentMessage != nil && documentMessage.Document != nil {
				switch ugo := ugoira.(type) {
				case *entity.UgoiraMeta:
					tginfo := ugo.GetTelegramInfo()
					tginfo.SetFileID(meta.BotID(), shared.TelegramMediaTypeDocument, documentMessage.Document.FileID)
					return serv.UpdateUgoiraTelegramInfo(ctx, ugo.ID, &tginfo)
				case *entity.CachedUgoiraMeta:
					cached, err := serv.GetCachedArtworkByURL(ctx, artwork.GetSourceURL())
					if err != nil {
						return oops.Wrapf(err, "failed to get cached artwork by url: %s", artwork.GetSourceURL())
					}
					data := cached.Artwork.Data()
					for _, u := range data.UgoiraMetas {
						if u.MetaData.OriginalZip == ugo.MetaData.OriginalZip {
							tginfo := ugo.GetTelegramInfo()
							tginfo.SetFileID(meta.BotID(), shared.TelegramMediaTypeDocument, documentMessage.Document.FileID)
							u.TelegramInfo = tginfo
							return serv.UpdateCachedArtwork(ctx, data)
						}
					}
				default:
					log.Warnf("unknown ugoira type: %T", ugo)
				}
			}
			return nil
		}()
		if err != nil {
			errs = append(errs, oops.Wrapf(err, "failed to send ugoira %d file", i+1))
		}
	}
	return oops.Join(errs...)
}
