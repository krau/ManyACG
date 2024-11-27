package fetcher

import (
	"context"
	es "errors"
	"time"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/errors"

	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/sources"
	"github.com/krau/ManyACG/storage"
	"github.com/krau/ManyACG/telegram"
	"github.com/krau/ManyACG/types"

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
			for {
				err := source.FetchNewArtworksWithCh(artworkCh, limit)
				if err != nil {
					common.Logger.Errorf("Error when fetching from %s: %s", name, err)
				}
				time.Sleep(time.Duration(interval) * time.Minute)
			}
		}(source, config.Cfg.Fetcher.Limit, artworkCh, source.Config().Intervel)
	}
	common.Logger.Infof("Enabled sources: %s", enabledSources)
	for artwork := range artworkCh {
		if artwork == nil {
			continue
		}
		if telegram.IsChannelAvailable {
			err := telegram.PostAndCreateArtwork(ctx, artwork, telegram.Bot, config.Cfg.Telegram.Admins[0], 0)
			if err != nil {
				if es.Is(err, errors.ErrArtworkAlreadyExist) || es.Is(err, errors.ErrArtworkDeleted) {
					continue
				}
				common.Logger.Errorf(err.Error())
			}
			time.Sleep(time.Duration(int(config.Cfg.Telegram.Sleep)*len(artwork.Pictures)) * time.Second)
		} else {
			artworkInDB, err := service.GetArtworkByURL(ctx, artwork.SourceURL)
			if err == nil && artworkInDB != nil {
				common.Logger.Debugf("Artwork %s already exists", artwork.Title)
				continue
			}
			if err != nil && !es.Is(err, mongo.ErrNoDocuments) {
				common.Logger.Errorf(err.Error())
				continue
			}
			if service.CheckDeletedByURL(ctx, artwork.SourceURL) {
				common.Logger.Debugf("Artwork %s is deleted", artwork.Title)
				continue
			}

			saveSuccess := true
			for i, picture := range artwork.Pictures {
				info, err := storage.SaveAll(ctx, artwork, picture)
				if err != nil {
					common.Logger.Errorf("saving picture %d of artwork %s: %s", i, artwork.Title, err)
					saveSuccess = false
					break
				}
				artwork.Pictures[i].StorageInfo = info
			}
			if !saveSuccess {
				continue
			}
			artwork, err = service.CreateArtwork(ctx, artwork)
			if err != nil {
				common.Logger.Errorf(err.Error())
				continue
			}
			go func() {
				for _, picture := range artwork.Pictures {
					service.AddProcessPictureTask(ctx, picture)
				}
			}()
		}
	}
}

// TODO:
// func FetchOnce(ctx context.Context, limit int) {
// 	common.Logger.Info("Start fetching once")
// 	artworks := make([]*types.Artwork, 0)
// 	var wg sync.WaitGroup
// 	for name, source := range sources.Sources {
// 		wg.Add(1)
// 		common.Logger.Infof("Start fetching from %s", name)
// 		go func(source sources.Source, limit int) {
// 			defer wg.Done()
// 			artworksForURL, err := source.FetchNewArtworks(limit)
// 			if err != nil {
// 				common.Logger.Errorf("Error when fetching from %s: %s", name, err)
// 			}
// 			for _, artwork := range artworksForURL {
// 				if artwork != nil {
// 					artworks = append(artworks, artwork)
// 				}
// 			}
// 		}(source, limit)
// 	}
// 	wg.Wait()
// 	common.Logger.Infof("Fetched %d artworks", len(artworks))

// 	for _, artwork := range artworks {
// 		err := telegram.PostAndCreateArtwork(ctx, artwork, telegram.Bot, config.Cfg.Telegram.Admins[0], 0)
// 		if err != nil {
// 			if es.Is(err, errors.ErrArtworkAlreadyExist) || es.Is(err, errors.ErrArtworkDeleted) {
// 				continue
// 			}
// 			common.Logger.Errorf(err.Error())
// 		}
// 		time.Sleep(time.Duration(int(config.Cfg.Telegram.Sleep)*len(artwork.Pictures)) * time.Second)
// 	}
// }
