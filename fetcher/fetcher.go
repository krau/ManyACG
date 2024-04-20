package fetcher

import (
	"ManyACG-Bot/config"
	"ManyACG-Bot/dao"
	. "ManyACG-Bot/logger"
	"ManyACG-Bot/service"
	"ManyACG-Bot/sources"
	"ManyACG-Bot/storage"
	"ManyACG-Bot/telegram"
	"ManyACG-Bot/types"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/mymmrac/telego"
	"go.mongodb.org/mongo-driver/mongo"
)

func StartScheduler(ctx context.Context) {
	artworkCh := make(chan *types.Artwork, config.Cfg.Fetcher.MaxConcurrent)
	for name, source := range sources.Sources {
		Logger.Infof("Start fetching from %s", name)
		go func(source sources.Source, limit int, artworkCh chan *types.Artwork, interval uint) {
			ticker := time.NewTicker(time.Duration(interval) * time.Minute)
			for {
				err := source.FetchNewArtworksWithCh(artworkCh, limit)
				if err != nil {
					Logger.Errorf("Error when fetching from %s: %s", name, err)
				}
				<-ticker.C
			}
		}(source, config.Cfg.Fetcher.Limit, artworkCh, source.Config().Intervel)
	}

	storage := storage.GetStorage()

	for artwork := range artworkCh {
		_, err := dao.GetArtworkByURL(ctx, artwork.SourceURL)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			Logger.Errorf("Error when getting artwork %s: %s", artwork.Title, err)
			continue
		}
		if errors.Is(err, mongo.ErrNoDocuments) {
			messages, err := telegram.PostArtwork(telegram.Bot, artwork)
			if err != nil {
				Logger.Errorf("Error when posting artwork [%s](%s): %s", artwork.Title, artwork.SourceURL, err)
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

			var wg sync.WaitGroup
			var storageErrs []error
			for i, picture := range artwork.Pictures {
				wg.Add(1)
				go func(i int, picture *types.Picture) {
					defer wg.Done()
					info, err := storage.SavePicture(artwork, picture)
					if err != nil {
						Logger.Errorf("Error when saving picture %s: %s", picture.Original, err)
						storageErrs = append(storageErrs, err)
						return
					}
					artwork.Pictures[i].StorageInfo = info
				}(i, picture)
			}

			wg.Wait()

			if len(storageErrs) > 0 {
				Logger.Errorf("Error when saving pictures of artwork %s: %s", artwork.Title, storageErrs)
				go func() {
					if telegram.Bot.DeleteMessages(&telego.DeleteMessagesParams{
						ChatID:     telegram.ChannelChatID,
						MessageIDs: telegram.GetMessageIDs(messages),
					}) != nil {
						Logger.Errorf("Error when deleting messages: %s", err)
					}
				}()
				continue
			}

			_, err = service.CreateArtwork(ctx, artwork)
			if err != nil {
				Logger.Errorf("Error when creating artwork %s: %s", artwork.SourceURL, err)
				go func() {
					if telegram.Bot.DeleteMessages(&telego.DeleteMessagesParams{
						ChatID:     telegram.ChannelChatID,
						MessageIDs: telegram.GetMessageIDs(messages),
					}) != nil {
						Logger.Errorf("Error when deleting messages: %s", err)
					}
				}()
				continue
			}
			time.Sleep(time.Duration(config.Cfg.Telegram.Sleep) * time.Second)
		} else {
			Logger.Infof("Artwork %s already exists", artwork.Title)
		}
	}
}

func FetchOnce(ctx context.Context, limit int) {
	Logger.Info("Start fetching once")
	artworks := make([]*types.Artwork, 0)
	var wg sync.WaitGroup
	for name, source := range sources.Sources {
		wg.Add(1)
		Logger.Infof("Start fetching from %s", name)
		go func(source sources.Source, limit int) {
			defer wg.Done()
			artworksForURL, err := source.FetchNewArtworks(limit)
			if err != nil {
				Logger.Errorf("Error when fetching from %s: %s", name, err)
			}
			for _, artwork := range artworksForURL {
				if artwork != nil {
					artworks = append(artworks, artwork)
				}
			}
		}(source, limit)
	}
	wg.Wait()
	Logger.Infof("Fetched %d artworks", len(artworks))

	storage := storage.GetStorage()

	for _, artwork := range artworks {
		_, err := dao.GetArtworkByURL(ctx, artwork.SourceURL)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			Logger.Errorf("Error when getting artwork %s: %s", artwork.Title, err)
			continue
		}
		if errors.Is(err, mongo.ErrNoDocuments) {
			messages, err := telegram.PostArtwork(telegram.Bot, artwork)
			if err != nil {
				Logger.Errorf("Error when posting artwork [%s](%s): %s", artwork.Title, artwork.SourceURL, err)
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

			var wg sync.WaitGroup
			var storageErrs []error
			for i, picture := range artwork.Pictures {
				wg.Add(1)
				go func(i int, picture *types.Picture) {
					defer wg.Done()
					info, err := storage.SavePicture(artwork, picture)
					if err != nil {
						Logger.Errorf("Error when saving picture %s: %s", picture.Original, err)
						storageErrs = append(storageErrs, err)
						return
					}
					artwork.Pictures[i].StorageInfo = info
				}(i, picture)
			}

			wg.Wait()

			if len(storageErrs) > 0 {
				Logger.Errorf("Error when saving pictures of artwork %s: %s", artwork.Title, storageErrs)
				go func() {
					if telegram.Bot.DeleteMessages(&telego.DeleteMessagesParams{
						ChatID:     telegram.ChannelChatID,
						MessageIDs: telegram.GetMessageIDs(messages),
					}) != nil {
						Logger.Errorf("Error when deleting messages: %s", err)
					}
				}()
				continue
			}

			_, err = service.CreateArtwork(ctx, artwork)
			if err != nil {
				Logger.Errorf("Error when creating artwork %s: %s", artwork.SourceURL, err)
				go func() {
					if telegram.Bot.DeleteMessages(&telego.DeleteMessagesParams{
						ChatID:     telegram.ChannelChatID,
						MessageIDs: telegram.GetMessageIDs(messages),
					}) != nil {
						Logger.Errorf("Error when deleting messages: %s", err)
					}
				}()
				continue
			}
			time.Sleep(time.Duration(config.Cfg.Telegram.Sleep) * time.Second)
		} else {
			Logger.Infof("Artwork %s already exists", artwork.Title)
		}
	}
}
