package handlers

import (
	"ManyACG/adapter"
	"ManyACG/common"
	. "ManyACG/logger"
	"ManyACG/service"
	"ManyACG/sources"
	"ManyACG/telegram"
	"ManyACG/types"
	"context"
	"errors"
	"math/rand"
	"strings"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/mongo"
)

func RandomPicture(ctx context.Context, bot *telego.Bot, message telego.Message) {
	cmd, _, args := telegoutil.ParseCommand(message.Text)
	argText := strings.ReplaceAll(strings.Join(args, " "), "\\", "")
	textArray := common.ParseStringTo2DArray(argText, "|", " ")
	r18 := cmd == "setu"
	r18Type := types.R18TypeNone
	if r18 {
		r18Type = types.R18TypeOnly
	}
	artwork, err := service.QueryArtworksByTexts(ctx, textArray, r18Type, 1, adapter.OnlyLoadPicture())
	if err != nil {
		Logger.Warnf("获取图片失败: %s", err)
		text := "获取图片失败" + err.Error()
		if errors.Is(err, mongo.ErrNoDocuments) {
			text = "未找到图片"
		}
		telegram.ReplyMessage(bot, message, text)
		return
	}
	if len(artwork) == 0 {
		telegram.ReplyMessage(bot, message, "未找到图片")
		return
	}
	pictures := artwork[0].Pictures
	picture := pictures[rand.Intn(len(pictures))]
	var file telego.InputFile
	if picture.TelegramInfo.PhotoFileID != "" {
		file = telegoutil.FileFromID(picture.TelegramInfo.PhotoFileID)
	} else {
		photoURL := picture.Original
		if artwork[0].SourceType == types.SourceTypePixiv {
			photoURL = sources.GetPixivRegularURL(photoURL)
		}
		file = telegoutil.FileFromURL(photoURL)
	}
	photo := telegoutil.Photo(message.Chat.ChatID(), file).
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}).WithCaption(artwork[0].Title).WithReplyMarkup(
		telegoutil.InlineKeyboard(telegram.GetPostedPictureInlineKeyboardButton(picture)),
	)
	if artwork[0].R18 {
		photo.WithHasSpoiler()
	}
	photoMessage, err := bot.SendPhoto(photo)
	if err != nil {
		telegram.ReplyMessage(bot, message, "发送图片失败: "+err.Error())
	}
	if photoMessage != nil {
		picture.TelegramInfo.PhotoFileID = photoMessage.Photo[len(photoMessage.Photo)-1].FileID
		if service.UpdatePictureTelegramInfo(ctx, picture, picture.TelegramInfo) != nil {
			Logger.Warnf("更新图片信息失败: %s", err)
		}
	}
}
