package handlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

func GetArtworkInfo(ctx *telegohandler.Context, message telego.Message) error {
	serv := service.FromContext(ctx)
	sourceURL := ctx.Value("source_url").(string)
	ogch := utils.GetMssageOriginChannel(&message)
	chatID := message.Chat.ChatID()
	meta := metautil.FromContext(ctx)
	if ogch != nil && (ogch.Chat.ID == meta.ChannelChatID().ID || strings.EqualFold(ogch.Chat.Username, strings.TrimPrefix(meta.ChannelChatID().Username, "@"))) {
		// handle the posted artwork in our channel
		artwork, err := serv.GetArtworkByURL(ctx, sourceURL)
		if errors.Is(err, errs.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			log.Errorf("get artwork by url failed: %s", err)
			return nil
		}
		ctx.Bot().SendMessage(ctx, telegoutil.Message(
			chatID,
			fmt.Sprintf("%s\n点击下方按钮在私聊中获取原图文件", sourceURL),
		).WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}).WithReplyMarkup(telegoutil.InlineKeyboard(
			utils.GetPostedArtworkInlineKeyboardButton(artwork, meta),
		)))
		return nil
	}
	hasPermission := utils.CheckPermissionInGroup(ctx, serv, message, shared.PermissionGetArtworkInfo)
	if !hasPermission {
		return nil
	}
	err := utils.SendArtworkInfo(ctx, ctx.Bot(), meta, serv, sourceURL, chatID, utils.SendArtworkInfoOptions{
		HasPermission:   hasPermission,
		ReplyParameters: &telego.ReplyParameters{MessageID: message.MessageID},
	})
	if err != nil {
		log.Errorf("send artwork info failed: %s", err)
		utils.ReplyMessage(ctx, message, err.Error())
	}
	return nil
}

func GetArtworkInfoCommand(ctx *telegohandler.Context, message telego.Message) error {
	serv := service.FromContext(ctx)
	sourceURL := utils.FindSourceURLInMessage(serv, &message)
	if sourceURL == "" {
		sourceURL = utils.FindSourceURLInMessage(serv, message.ReplyToMessage)
	}
	if sourceURL == "" {
		helpText := `
	<b>使用 /info 命令并在参数中提供作品链接, 或使用该命令回复一条包含支持的链接的消息, 将获取作品信息并发送全部图片</b>

	命令语法: /info [作品链接]
	`
		utils.ReplyMessageWithHTML(ctx, message, helpText)
		return nil
	}
	artwork, err := serv.GetOrFetchCachedArtwork(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessage(ctx, message, "获取作品信息失败")
		return oops.Wrapf(err, "get or fetch cached artwork failed: %s", sourceURL)
	}
	meta := metautil.MustFromContext(ctx)
	results, err := utils.SendArtworkMediaGroup(ctx, ctx.Bot(), serv, meta, message.Chat.ChatID(), artwork)
	if err != nil {
		utils.ReplyMessage(ctx, message, "发送作品图片时出现错误")
		return oops.Wrapf(err, "send artwork media group failed: %s", artwork.SourceURL)
	}
	data := artwork.Artwork.Data()
	for _, res := range results {
		if res.UgoiraIndex >= 0 {
			if len(data.UgoiraMetas) <= res.UgoiraIndex {
				log.Warn("ugoira index out of range", "index", res.UgoiraIndex, "len", len(data.UgoiraMetas), "title", artwork.GetTitle(), "url", artwork.GetSourceURL())
				continue
			}
			// data.UgoiraMetas[res.UgoiraIndex].TelegramInfo.PhotoFileID = res.FileID
			data.UgoiraMetas[res.UgoiraIndex].TelegramInfo.SetFileID(meta.BotID(), shared.TelegramMediaTypeVideo, res.FileID)
		} else if res.PictureIndex >= 0 {
			if len(data.Pictures) <= res.PictureIndex {
				log.Warn("picture index out of range", "index", res.PictureIndex, "len", len(data.Pictures), "title", artwork.GetTitle(), "url", artwork.GetSourceURL())
				continue
			}
			// data.Pictures[res.PictureIndex].TelegramInfo.PhotoFileID = res.FileID
			data.Pictures[res.PictureIndex].TelegramInfo.SetFileID(meta.BotID(), shared.TelegramMediaTypePhoto, res.FileID)
		}
	}
	if err := serv.UpdateCachedArtwork(ctx, data); err != nil {
		return oops.Wrapf(err, "failed to update cached artwork after send media group")
	}
	return nil
}
