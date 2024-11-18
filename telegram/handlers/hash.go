package handlers

import (
	"context"
	"fmt"
	"image"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/telegram/utils"

	"github.com/mymmrac/telego"
)

func CalculatePicture(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if message.ReplyToMessage == nil {
		utils.ReplyMessage(bot, message, "请使用该命令回复一条图片消息")
		return
	}
	var waitMessageID int
	msg, err := utils.ReplyMessage(bot, message, "少女做高数中...(((φ(◎ロ◎;)φ)))")
	if err == nil {
		waitMessageID = msg.MessageID
	}
	file, err := utils.GetMessagePhotoFile(bot, message.ReplyToMessage)
	if err != nil {
		utils.ReplyMessage(bot, message, "获取图片文件失败: "+err.Error())
		return
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		common.Logger.Error("解码图片失败: %v", err)
		utils.ReplyMessage(bot, message, "解码图片失败")
		return
	}
	hash, err := common.GetImagePhash(img)
	if err != nil {
		utils.ReplyMessage(bot, message, "计算图片信息失败: "+err.Error())
		return
	}
	blurScore, err := common.GetImageBlurScore(img)
	if err != nil {
		utils.ReplyMessage(bot, message, "计算图片信息失败: "+err.Error())
		return
	}
	width, height, err := common.GetImageSize(img)
	if err != nil {
		utils.ReplyMessage(bot, message, "计算图片信息失败: "+err.Error())
		return
	}
	text := fmt.Sprintf(
		"<b>Hash</b>: <code>%s</code>\n<b>模糊度</b>: %.2f\n<b>尺寸</b>: %d x %d",
		hash, blurScore, width, height,
	)
	if waitMessageID == 0 {
		utils.ReplyMessageWithHTML(bot, message, text)
		return
	}
	bot.EditMessageText(&telego.EditMessageTextParams{
		ChatID:    message.Chat.ChatID(),
		MessageID: waitMessageID,
		Text:      text,
		ParseMode: telego.ModeHTML,
	})
}
