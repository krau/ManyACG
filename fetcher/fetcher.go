package fetcher

import (
	"ManyACG/config"
	"ManyACG/errors"
	. "ManyACG/logger"
	"ManyACG/service"
	"ManyACG/sources"
	"ManyACG/storage"
	"ManyACG/telegram"
	"ManyACG/types"
	"context"
	es "errors"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

func StartScheduler(ctx context.Context) {
	artworkCh := make(chan *types.Artwork, config.Cfg.Fetcher.MaxConcurrent)
	enabledSources := ""
	for name, source := range sources.Sources {
		enabledSources += string(name) + " "
		go func(source sources.Source, limit int, artworkCh chan *types.Artwork, interval int) {
			if interval <= 0 {
				return
			}
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
	Logger.Infof("Enabled sources: %s", enabledSources)
	for artwork := range artworkCh {
		if telegram.IsChannelAvailable {
			err := telegram.PostAndCreateArtwork(ctx, artwork, telegram.Bot, config.Cfg.Telegram.Admins[0], 0)
			if err != nil {
				if es.Is(err, errors.ErrArtworkAlreadyExist) || es.Is(err, errors.ErrArtworkDeleted) {
					continue
				}
				Logger.Errorf(err.Error())
			}
			time.Sleep(time.Duration(int(config.Cfg.Telegram.Sleep)*len(artwork.Pictures)) * time.Second)
		} else {
			artworkInDB, err := service.GetArtworkByURL(ctx, artwork.SourceURL)
			if err == nil && artworkInDB != nil {
				Logger.Debugf("Artwork %s already exists", artwork.Title)
				continue
			}
			if err != nil && !es.Is(err, mongo.ErrNoDocuments) {
				Logger.Errorf(err.Error())
				continue
			}
			if service.CheckDeletedByURL(ctx, artwork.SourceURL) {
				Logger.Debugf("Artwork %s is deleted", artwork.Title)
				continue
			}

			saveSuccess := true
			for i, picture := range artwork.Pictures {
				info, err := storage.GetStorage().SavePicture(artwork, picture)
				if err != nil {
					Logger.Errorf("saving picture %d of artwork %s: %s", i, artwork.Title, err)
					saveSuccess = false
					break
				}
				artwork.Pictures[i].StorageInfo = info
			}
			if !saveSuccess {
				continue
			}
			_, err = service.CreateArtwork(ctx, artwork)
			if err != nil {
				Logger.Errorf(err.Error())
				continue
			}
			go func() {
				for _, picture := range artwork.Pictures {
					if err := service.ProcessPictureAndUpdate(ctx, picture); err != nil {
						Logger.Warnf("error when processing %d of artwork %s: %s", picture.Index, artwork.Title, err)
					}
				}
			}()
		}
	}
}

// TODO:
// func FetchOnce(ctx context.Context, limit int) {
// 	Logger.Info("Start fetching once")
// 	artworks := make([]*types.Artwork, 0)
// 	var wg sync.WaitGroup
// 	for name, source := range sources.Sources {
// 		wg.Add(1)
// 		Logger.Infof("Start fetching from %s", name)
// 		go func(source sources.Source, limit int) {
// 			defer wg.Done()
// 			artworksForURL, err := source.FetchNewArtworks(limit)
// 			if err != nil {
// 				Logger.Errorf("Error when fetching from %s: %s", name, err)
// 			}
// 			for _, artwork := range artworksForURL {
// 				if artwork != nil {
// 					artworks = append(artworks, artwork)
// 				}
// 			}
// 		}(source, limit)
// 	}
// 	wg.Wait()
// 	Logger.Infof("Fetched %d artworks", len(artworks))

// 	for _, artwork := range artworks {
// 		err := telegram.PostAndCreateArtwork(ctx, artwork, telegram.Bot, config.Cfg.Telegram.Admins[0], 0)
// 		if err != nil {
// 			if es.Is(err, errors.ErrArtworkAlreadyExist) || es.Is(err, errors.ErrArtworkDeleted) {
// 				continue
// 			}
// 			Logger.Errorf(err.Error())
// 		}
// 		time.Sleep(time.Duration(int(config.Cfg.Telegram.Sleep)*len(artwork.Pictures)) * time.Second)
// 	}
// }
