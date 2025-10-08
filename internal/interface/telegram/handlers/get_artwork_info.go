package handlers

import (
	"fmt"

	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/service"
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
	if ogch != nil && metautil.ChatIDEqual(ogch.Chat.ChatID(), meta.ChannelChatID()) {
		// handle the posted artwork in our channel
		artwork, err := serv.GetArtworkByURL(ctx, sourceURL)
		if err != nil {
			log.Errorf("failed to get posted artwork: %s", err)
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
	err := utils.SendArtworkInfo(ctx, meta, serv, sourceURL, chatID, utils.SendArtworkInfoOptions{
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
	msgs, err := utils.SendArtworkMediaGroup(ctx, message.Chat.ChatID(), artwork)
	if err != nil {
		utils.ReplyMessage(ctx, message, "发送作品图片时出现错误")
		return oops.Wrapf(err, "send artwork media group failed: %s", artwork.SourceURL)
	}
	if len(msgs) != len(artwork.GetPictures()) {
		log.Warnf("sent media group count mismatch: sent %d, expected %d", len(msgs), len(artwork.GetPictures()))
		return nil
	}
	data := artwork.Artwork.Data()
	for i, msg := range msgs {
		if len(msg.Photo) == 0 {
			continue
		}
		pic := data.Pictures[i]
		tginfo := pic.TelegramInfo
		tginfo.PhotoFileID = msg.Photo[len(msg.Photo)-1].FileID
		pic.TelegramInfo = tginfo
	}
	if err := serv.UpdateCachedArtwork(ctx, data); err != nil {
		return oops.Wrapf(err, "update cached artwork failed: %s", artwork.SourceURL)
	}
	return nil
}
