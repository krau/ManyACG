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
	// 	serv := service.FromContext(ctx)
	// 	sourceURL := utils.FindSourceURLInMessage(serv, &message)
	// 	if sourceURL == "" {
	// 		sourceURL = utils.FindSourceURLInMessage(serv, message.ReplyToMessage)
	// 	}
	// 	if sourceURL == "" {
	// 		helpText := `
	// <b>使用 /info 命令并在参数中提供作品链接, 或使用该命令回复一条包含支持的链接的消息, 将获取作品信息并发送全部图片</b>

	// 命令语法: /info [作品链接]
	// `
	// 		utils.ReplyMessageWithHTML(ctx, message, helpText)
	// 		return nil
	// 	}
	// 	chatID := message.Chat.ChatID()

	//	if err := utils.SendArtworkInfo(ctx, utils.SendArtworkInfoOption{
	//		SendFull:  true,
	//		ChatID:    &chatID,
	//		SourceURL: sourceURL,
	//	}); err != nil {
	//
	//		utils.ReplyMessage(ctx, message, err.Error())
	//	}
	//
	// return nil
	panic("not implemented")
}
