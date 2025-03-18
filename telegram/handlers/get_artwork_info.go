package handlers

import (
	"time"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/telegram/utils"
	"github.com/krau/ManyACG/types"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

func GetArtworkInfo(ctx *telegohandler.Context, message telego.Message) error {
	hasPermission := CheckPermissionInGroup(ctx, message, types.PermissionGetArtworkInfo)
	sourceURL := utils.FindSourceURLForMessage(&message)
	var waitMessageID int
	if hasPermission {
		go func() {
			msg, err := utils.ReplyMessage(ctx, ctx.Bot(), message, "正在获取作品信息...")
			if err != nil {
				common.Logger.Warnf("发送消息失败: %s", err)
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

	err := utils.SendArtworkInfo(ctx, ctx.Bot(), &utils.SendArtworkInfoParams{
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
		common.Logger.Error(err)
		utils.ReplyMessage(ctx, ctx.Bot(), message, err.Error())
	}
	return nil
}

func GetArtworkInfoCommand(ctx *telegohandler.Context, message telego.Message) error {
	sourceURL := utils.FindSourceURLForMessage(&message)
	if sourceURL == "" {
		sourceURL = utils.FindSourceURLForMessage(message.ReplyToMessage)
	}
	if sourceURL == "" {
		helpText := `
<b>使用 /info 命令并在参数中提供作品链接, 或使用该命令回复一条包含支持的链接的消息, 将获取作品信息并发送全部图片</b>

命令语法: /info [作品链接]
`
		utils.ReplyMessageWithHTML(ctx, ctx.Bot(), message, helpText)
		return nil
	}
	if err := utils.SendFullArtworkInfo(ctx, ctx.Bot(), message, sourceURL); err != nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, err.Error())
	}
	return nil
}
