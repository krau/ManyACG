package handlers

import (
	"ManyACG/common"
	"ManyACG/telegram"
	"context"
	"fmt"

	"github.com/mymmrac/telego"
)

func CalculatePicture(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if message.ReplyToMessage == nil {
		telegram.ReplyMessage(bot, message, "请使用该命令回复一条图片消息")
		return
	}
	var waitMessageID int
	msg, err := telegram.ReplyMessage(bot, message, "少女做高数中...(((φ(◎ロ◎;)φ)))")
	if err == nil {
		waitMessageID = msg.MessageID
	}
	fileBytes, err := telegram.GetMessagePhotoFileBytes(bot, message.ReplyToMessage)
	if err != nil {
		telegram.ReplyMessage(bot, message, "获取图片文件失败: "+err.Error())
		return
	}
	hash, err := common.GetImagePhash(fileBytes)
	if err != nil {
		telegram.ReplyMessage(bot, message, "计算图片信息失败: "+err.Error())
		return
	}
	blurScore, err := common.GetImageBlurScore(fileBytes)
	if err != nil {
		telegram.ReplyMessage(bot, message, "计算图片信息失败: "+err.Error())
		return
	}
	width, height, err := common.GetImageSize(fileBytes)
	if err != nil {
		telegram.ReplyMessage(bot, message, "计算图片信息失败: "+err.Error())
		return
	}
	text := fmt.Sprintf(
		"<b>Hash</b>: <code>%s</code>\n<b>模糊度</b>: %.2f\n<b>尺寸</b>: %d x %d",
		hash, blurScore, width, height,
	)
	if waitMessageID == 0 {
		telegram.ReplyMessageWithHTML(bot, message, text)
		return
	}
	bot.EditMessageText(&telego.EditMessageTextParams{
		ChatID:    message.Chat.ChatID(),
		MessageID: waitMessageID,
		Text:      text,
		ParseMode: telego.ModeHTML,
	})
}
