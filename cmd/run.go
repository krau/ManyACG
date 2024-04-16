package cmd

import (
	"ManyACG-Bot/dao"
	. "ManyACG-Bot/logger"
	"ManyACG-Bot/sources"
	"ManyACG-Bot/telegram"
	"ManyACG-Bot/types"
	"context"
	"errors"
	"time"

	"github.com/mymmrac/telego"
	"go.mongodb.org/mongo-driver/mongo"
)

func Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	Logger.Info("Start running")
	dao.InitDB(ctx)
	defer func() {
		if err := dao.Client.Disconnect(ctx); err != nil {
			Logger.Panic(err)
		}
	}()

	artworkCh := make(chan types.Artwork, 30)
	for name, source := range sources.Sources {
		Logger.Infof("Start fetching from %s", name)
		go func(source sources.Source, limit int, ch chan types.Artwork, interval uint) {
			ticker := time.NewTicker(time.Duration(interval) * time.Minute)
			for {
				artworks, err := source.FetchNewArtworks(limit)
				if err != nil {
					Logger.Errorf("Error when fetching from %s: %s", name, err)
					// 在 go 1.22 之后, 循环变量不会再共用, 所以上面的警告应忽视
				}
				if len(artworks) > 0 {
					Logger.Infof("Fetched %d artworks from %s", len(artworks), name)
					for _, artwork := range artworks {
						ch <- artwork
					}
				}
				<-ticker.C
			}
		}(source, 30, artworkCh, source.Config().Intervel)
	}

	for artwork := range artworkCh {
		artworkDB, err := dao.GetArtworkByURL(ctx, artwork.Source.URL)
		if artworkDB != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			Logger.Infof("Artwork %s already exists", artwork.Source.URL)
			continue
		}
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			Logger.Errorf("Error when getting artwork %s: %s", artwork.Source.URL, err)
			continue
		}

		messages, err := telegram.PostArtwork(telegram.Bot, &artwork)
		if err != nil {
			Logger.Errorf("Error when posting artwork %s: %s", artwork.Source.URL, err)
			continue
		}
		Logger.Infof("Posted artwork %s", artwork.Source.URL)

		artworkDB = &artwork

		for i, picture := range artwork.Pictures {
			picture.TelegramInfo.MessageID = messages[i].MessageID
			picture.TelegramInfo.MediaGroupID = messages[i].MediaGroupID
			if messages[i].Photo != nil {
				picture.TelegramInfo.PhotoFileID = messages[i].Photo[len(messages[i].Photo)-1].FileID
			}
			if messages[i].Document != nil {
				picture.TelegramInfo.DocumentFileID = messages[i].Document.FileID
			}
		}

		_, err = dao.CreateArtwork(ctx, artworkDB)
		if err != nil {
			Logger.Errorf("Error when creating artwork %s: %s", artwork.Source.URL, err)
			messageIDs := make([]int, 0)
			for _, message := range messages {
				messageIDs = append(messageIDs, message.MessageID)
			}
			telegram.Bot.DeleteMessages(&telego.DeleteMessagesParams{
				ChatID:     telegram.ChatID,
				MessageIDs: messageIDs,
			})
		}
	}
}
