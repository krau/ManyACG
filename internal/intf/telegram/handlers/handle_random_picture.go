package handlers

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/krau/ManyACG/internal/app/query"
	"github.com/krau/ManyACG/internal/domain/repo"
	"github.com/krau/ManyACG/internal/intf/telegram/utils"
	"github.com/krau/ManyACG/internal/pkg/log"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/strutil"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

func (h *BotHandlers) RandomPicture(ctx *telegohandler.Context, message telego.Message) error {
	cmd, _, args := telegoutil.ParseCommand(message.Text)
	argText := strings.ReplaceAll(strings.Join(args, " "), "\\", "")
	textArray := strutil.ParseTo2DArray(argText, "|", " ")
	r18 := cmd == "setu"
	r18Type := shared.R18TypeNone
	if r18 {
		r18Type = shared.R18TypeR18
	}
	artwork, err := h.app.Queries.ArtworkSearch.Handle(ctx, query.ArtworkSearchQuery{
		Keywords: textArray,
		R18:      r18Type,
		Limit:    1,
	})
	if err != nil {
		log.Error("failed to get random picture", "err", err)
		if errors.Is(err, repo.ErrNotFound) {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "未找到图片")
			return nil
		}
		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取图片失败")
		return nil
	}
	pic := artwork[0].Pictures[rand.Intn(len(artwork[0].Pictures))]
	var file telego.InputFile
	if pic.TelegramInfo.PhotoFileID != "" {
		file = telegoutil.FileFromID(pic.TelegramInfo.PhotoFileID)
	} else {
		// photoURL := utils.GetResizedPictureURL(pic.Original, 2560, 2560)
		file = telegoutil.FileFromURL(pic.Original)
	}
	caption := fmt.Sprintf("[%s](%s)", strutil.EscapeMarkdown(artwork[0].Title), artwork[0].SourceURL)
	photo := telegoutil.Photo(message.Chat.ChatID(), file).
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}).WithCaption(caption).WithParseMode(telego.ModeMarkdownV2)
	// .WithReplyMarkup(
	// telegoutil.InlineKeyboard(utils.GetPostedPictureInlineKeyboardButton(artwork[0], 0, ChannelChatID, BotUsername)),

	if artwork[0].R18 {
		photo.WithHasSpoiler()
	}
	photoMessage, err := ctx.Bot().SendPhoto(ctx, photo)
	if err != nil {
		log.Error("failed to send photo", "err", err)
		utils.ReplyMessage(ctx, ctx.Bot(), message, "发送图片失败")
		return nil
	}
	if photoMessage != nil {
		pic.TelegramInfo.PhotoFileID = photoMessage.Photo[len(photoMessage.Photo)-1].FileID
	}

	// pictures := artwork[0].Pictures
	// picture := pictures[rand.Intn(len(pictures))]
	// var file telego.InputFile
	// if picture.TelegramInfo.PhotoFileID != "" {
	// 	file = telegoutil.FileFromID(picture.TelegramInfo.PhotoFileID)
	// } else {
	// 	photoURL := fmt.Sprintf("%s/?url=%s&w=2560&h=2560&we&output=jpg", config.Get().WSRVURL, picture.Original)
	// 	file = telegoutil.FileFromURL(photoURL)
	// }
	// caption := fmt.Sprintf("[%s](%s)", common.EscapeMarkdown(artwork[0].Title), artwork[0].SourceURL)
	// photo := telegoutil.Photo(message.Chat.ChatID(), file).
	// 	WithReplyParameters(&telego.ReplyParameters{
	// 		MessageID: message.MessageID,
	// 	}).WithCaption(caption).WithParseMode(telego.ModeMarkdownV2).WithReplyMarkup(
	// 	telegoutil.InlineKeyboard(utils.GetPostedPictureInlineKeyboardButton(artwork[0], 0, ChannelChatID, BotUsername)),
	// )
	// if artwork[0].R18 {
	// 	photo.WithHasSpoiler()
	// }
	// photoMessage, err := ctx.Bot().SendPhoto(ctx, photo)
	// if err != nil {
	// 	utils.ReplyMessage(ctx, ctx.Bot(), message, "发送图片失败: "+err.Error())
	// }
	// if photoMessage != nil {
	// 	picture.TelegramInfo.PhotoFileID = photoMessage.Photo[len(photoMessage.Photo)-1].FileID
	// 	if service.UpdatePictureTelegramInfo(ctx, picture, picture.TelegramInfo) != nil {
	// 		common.Logger.Warnf("更新图片信息失败: %s", err)
	// 	}
	// }
	// return nil
}
