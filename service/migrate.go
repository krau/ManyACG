package service

import (
	"ManyACG/dao"
	. "ManyACG/logger"
	"context"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func ProcessPicturesAndUpdate(ctx context.Context, bot *telego.Bot, message *telego.Message) {
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

func AddLikeCountField(ctx context.Context) error {
	return dao.AddLikeCountToArtwork(ctx)
}

func Migrate(ctx context.Context) {
	if err := AddLikeCountField(ctx); err != nil {
		Logger.Errorf("Failed to add likes field: %v", err)
	}
	Logger.Noticef("Migration completed")
}
