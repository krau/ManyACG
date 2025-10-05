package handlers

import (
	"fmt"
	"strings"
	"time"

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
	hasPermission := utils.CheckPermissionInGroup(ctx, serv, message, shared.PermissionGetArtworkInfo)
	// sourceURL := utils.FindSourceURLInMessage(serv, &message)
	sourceURL := ctx.Value("source_url").(string)
	ogch := utils.GetMssageOriginChannel(&message)
	meta := metautil.FromContext(ctx)
	if ogch != nil && (ogch.Chat.ID == meta.ChannelChatID.ID || strings.EqualFold(ogch.Chat.Username, strings.TrimPrefix(meta.ChannelChatID.Username, "@"))) {
		// handle the posted artwork in our channel
		// time.Sleep(3 * time.Second)
		artwork, err := serv.GetArtworkByURL(ctx, sourceURL)
		if err != nil {
			log.Errorf("failed to get posted artwork: %s", err)
			return nil
		}
		ctx.Bot().SendMessage(ctx, telegoutil.Message(
			message.Chat.ChatID(),
			fmt.Sprintf("%s\n点击下方按钮在私聊中获取原图文件", sourceURL),
		).WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}).WithReplyMarkup(telegoutil.InlineKeyboard(
			utils.GetPostedArtworkInlineKeyboardButton(artwork, meta),
		)))
		return nil
	}
	var waitMessageID int
	if hasPermission {
		go func() {
			msg, err := utils.ReplyMessage(ctx, message, "正在获取作品信息...")
			if err != nil {
				log.Warnf("发送消息失败: %s", err)
				return
			}
			waitMessageID = msg.MessageID
		}()
	}
	defer func() {
		time.Sleep(1 * time.Second)
		if waitMessageID != 0 {
			ctx.Bot().DeleteMessage(ctx, telegoutil.Delete(message.Chat.ChatID(), waitMessageID))
		}
	}()
	chatID := message.Chat.ChatID()

	err := utils.SendArtworkInfo(ctx, utils.SendArtworkInfoOption{
		ChatID:        &chatID,
		SourceURL:     sourceURL,
		AppendCaption: "",
		Verify:        false,
		IgnoreDeleted: false,
		HasPermission: hasPermission,
		ReplyParams: &telego.ReplyParameters{
			MessageID: message.MessageID,
		},
	})
	if err != nil {
		log.Error(err)
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
	chatID := message.Chat.ChatID()

	if err := utils.SendArtworkInfo(ctx, utils.SendArtworkInfoOption{
		SendFull:  true,
		ChatID:    &chatID,
		SourceURL: sourceURL,
	}); err != nil {
		utils.ReplyMessage(ctx, message, err.Error())
	}
	return nil
}
