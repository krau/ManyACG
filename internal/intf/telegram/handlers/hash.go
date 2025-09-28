package handlers

import (
	"bytes"
	"fmt"
	"image"

	"github.com/krau/ManyACG/internal/common"
	"github.com/krau/ManyACG/internal/intf/telegram/utils"
	"github.com/krau/ManyACG/internal/pkg/imgtool"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
)

func CalculatePicture(ctx *telegohandler.Context, message telego.Message) error {
	if message.ReplyToMessage == nil {
		helpText := `
<b>使用 /hash 命令回复一条图片消息, 将计算图片信息</b>
		`
		utils.ReplyMessageWithHTML(ctx, ctx.Bot(), message, helpText)
		return nil
	}
	var waitMessageID int
	msg, err := utils.ReplyMessage(ctx, ctx.Bot(), message, "少女做高数中...(((φ(◎ロ◎;)φ)))")
	if err == nil {
		waitMessageID = msg.MessageID
	}
	file, err := utils.GetMessagePhotoFile(ctx, ctx.Bot(), message.ReplyToMessage)
	if err != nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取图片文件失败: "+err.Error())
		return nil
	}
	img, _, err := image.Decode(bytes.NewReader(file))
	if err != nil {
		common.Logger.Error("解码图片失败: %v", err)
		utils.ReplyMessage(ctx, ctx.Bot(), message, "解码图片失败")
		return nil
	}
	hash, err := imgtool.GetImagePhash(img)
	if err != nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "计算图片信息失败: "+err.Error())
		return nil
	}
	width, height, err := imgtool.GetImageSize(img)
	if err != nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "计算图片信息失败: "+err.Error())
		return nil
	}
	text := fmt.Sprintf(
		"<b>Hash</b>: <code>%s</code>\n<b>尺寸</b>: %d x %d",
		hash, width, height,
	)
	if waitMessageID == 0 {
		utils.ReplyMessageWithHTML(ctx, ctx.Bot(), message, text)
		return nil
	}
	ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
		ChatID:    message.Chat.ChatID(),
		MessageID: waitMessageID,
		Text:      text,
		ParseMode: telego.ModeHTML,
	})
	return nil
}
