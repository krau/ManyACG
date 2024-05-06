package service

import (
	"ManyACG-Bot/dao"
	. "ManyACG-Bot/logger"
	"context"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func ProcessOldPicturesAndUpdate(ctx context.Context, bot *telego.Bot, message *telego.Message) {
	pictures, err := dao.GetNotProcessedPictures(ctx)
	sendMessage := bot != nil && message != nil
	if err != nil {
		Logger.Errorf("Failed to get not processed pictures: %v", err)
		if sendMessage {
			bot.SendMessage(telegoutil.Messagef(
				message.Chat.ChatID(),
				"Failed to get not processed pictures: %s",
				err.Error(),
			))
		}
		return
	}
	if sendMessage {
		bot.SendMessage(telegoutil.Messagef(
			message.Chat.ChatID(),
			"Found %d not processed pictures",
			len(pictures),
		))
	}

	failed := 0
	for _, picture := range pictures {
		if err := ProcessPictureAndUpdate(ctx, picture.ToPicture()); err != nil {
			failed++
		}
	}
	Logger.Infof("Processed %d pictures, %d failed", len(pictures)-failed, failed)
	if sendMessage {
		bot.SendMessage(telegoutil.Messagef(
			message.Chat.ChatID(),
			"Processed %d pictures, %d failed",
			len(pictures)-failed,
			failed,
		))
	}
}

func GetNotProcessedPictureCount(ctx context.Context) int {
	pictures, _ := dao.GetNotProcessedPictures(ctx)
	return len(pictures)
}
