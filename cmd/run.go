package cmd

import (
	"ManyACG-Bot/config"
	"ManyACG-Bot/dao"
	. "ManyACG-Bot/logger"
	"ManyACG-Bot/sources"
	"ManyACG-Bot/storage"
	"ManyACG-Bot/telegram"
	"ManyACG-Bot/types"
	"context"
	"errors"
	"time"

	"github.com/mymmrac/telego"
	"go.mongodb.org/mongo-driver/mongo"
)

func Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	Logger.Info("Start running")
	dao.InitDB(ctx)
	defer func() {
		if err := dao.Client.Disconnect(ctx); err != nil {
			Logger.Panic(err)
		}
	}()

	artworkCh := make(chan *types.Artwork, config.Cfg.App.MaxConcurrent)
	for name, source := range sources.Sources {
		Logger.Infof("Start fetching from %s", name)
		go func(source sources.Source, limit int, artworkCh chan *types.Artwork, interval uint) {
			ticker := time.NewTicker(time.Duration(interval) * time.Minute)
			for {
				err := source.FetchNewArtworks(artworkCh, limit)
				if err != nil {
					Logger.Errorf("Error when fetching from %s: %s", name, err)
				}
				<-ticker.C
			}
		}(source, 30, artworkCh, source.Config().Intervel)
	}

	storage := storage.GetStorage()

	for artwork := range artworkCh {
		_, err := dao.GetArtworkByURL(context.TODO(), artwork.Source.URL)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			Logger.Errorf("Error when getting artwork %s: %s", artwork.Title, err)
			continue
		}
		if errors.Is(err, mongo.ErrNoDocuments) {
			messages, err := telegram.PostArtwork(telegram.Bot, artwork)
			if err != nil {
				Logger.Errorf("Error when posting artwork [%s](%s): %s", artwork.Title, artwork.Source.URL, err)
				continue
			}
			Logger.Infof("Posted artwork %s", artwork.Title)

			for i, picture := range artwork.Pictures {
				if picture.TelegramInfo == nil {
					picture.TelegramInfo = &types.TelegramInfo{}
				}
				picture.TelegramInfo.MessageID = messages[i].MessageID
				picture.TelegramInfo.MediaGroupID = messages[i].MediaGroupID
				if messages[i].Photo != nil {
					picture.TelegramInfo.PhotoFileID = messages[i].Photo[len(messages[i].Photo)-1].FileID
				}
				if messages[i].Document != nil {
					picture.TelegramInfo.DocumentFileID = messages[i].Document.FileID
				}
			}

			var storageErr error
			for i, picture := range artwork.Pictures {
				var info *types.StorageInfo
				info, storageErr = storage.SavePicture(artwork, picture)
				if storageErr != nil {
					Logger.Errorf("Error when saving picture %s: %s", picture.Original, storageErr)
					break
				}
				artwork.Pictures[i].StorageInfo = info
			}

			if storageErr != nil {
				Logger.Errorf("Error when saving pictures of artwork %s: %s", artwork.Title, storageErr)
				if telegram.Bot.DeleteMessages(&telego.DeleteMessagesParams{
					ChatID:     telegram.ChatID,
					MessageIDs: telegram.GetMessageIDs(messages),
				}) != nil {
					Logger.Errorf("Error when deleting messages: %s", err)
				}
				continue
			}

			_, err = dao.CreateArtwork(context.TODO(), artwork)
			if err != nil {
				Logger.Errorf("Error when creating artwork %s: %s", artwork.Source.URL, err)
				if telegram.Bot.DeleteMessages(&telego.DeleteMessagesParams{
					ChatID:     telegram.ChatID,
					MessageIDs: telegram.GetMessageIDs(messages),
				}) != nil {
					Logger.Errorf("Error when deleting messages: %s", err)
				}
				continue
			}
			time.Sleep(time.Duration(config.Cfg.Telegram.Sleep) * time.Second)
		} else {
			Logger.Infof("Artwork %s already exists", artwork.Title)
		}
	}
}
