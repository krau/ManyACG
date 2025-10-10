package scheduler

import (
	"context"
	"errors"
	"time"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/model/converter"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/log"
)

type ArtworkPoster interface {
	PostAndCreateArtwork(ctx context.Context, serv *service.Service, artwork *entity.CachedArtworkData) error
}

func StartPoster(ctx context.Context, poster ArtworkPoster, serv *service.Service) {
	cfg := runtimecfg.Get().Scheduler
	if !cfg.Enable || poster == nil {
		return
	}
	sources := serv.Sources()
	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			seen := make(map[string]struct{})
			fetcheds := make([]*dto.FetchedArtwork, 0)
			for _, sou := range sources {
				artworks, err := sou.FetchNewArtworks(ctx, cfg.Limit)
				if err != nil {
					log.Errorf("scheduler: fetching new artworks from source %T: %s", sou, err)
				}
				if len(artworks) == 0 {
					continue
				}
				log.Infof("scheduler: fetched %d new artworks from source %T", len(artworks), sou)
				for _, artwork := range artworks {
					if artwork == nil || artwork.SourceURL == "" {
						continue
					}
					if _, ok := seen[artwork.SourceURL]; ok {
						continue
					}
					seen[artwork.SourceURL] = struct{}{}
					fetcheds = append(fetcheds, artwork)
				}
			}
			for _, fetchedArtwork := range fetcheds {
				exist, err := serv.GetCachedArtworkByURL(ctx, fetchedArtwork.SourceURL)
				if err != nil && !errors.Is(err, errs.ErrRecordNotFound) {
					log.Errorf("scheduler: checking cached artwork %s: %s", fetchedArtwork.SourceURL, err)
					continue
				}
				if err != nil {
					cached, err := serv.CreateCachedArtwork(ctx, converter.DtoFetchedArtworkToEntityCached(fetchedArtwork))
					if err != nil {
						log.Errorf("scheduler: creating cached artwork %s: %s", fetchedArtwork.SourceURL, err)
						continue
					}
					exist = cached
				}
				if exist == nil {
					continue
				}
				posted, err := serv.GetArtworkByURL(ctx, exist.SourceURL)
				if err == nil && posted != nil {
					continue
				}
				deleted := serv.CheckDeletedByURL(ctx, exist.SourceURL)
				if deleted {
					continue
				}
				data := exist.Artwork.Data()
				if data == nil || len(data.Pictures) == 0 {
					continue
				}
				if err := poster.PostAndCreateArtwork(ctx, serv, data); err != nil {
					log.Errorf("scheduler: posting artwork %s: %s", exist.SourceURL, err)
					continue
				}
				log.Infof("scheduler: posted artwork %s", exist.SourceURL)
			}
		}
	}
}

// func StartScheduler(ctx context.Context) {
// 	artworkCh := make(chan *types.Artwork, config.Cfg.Fetcher.MaxConcurrent)
// 	var enabledSources string
// 	for name, source := range sources.GetSources() {
// 		enabledSources += string(name) + " "
// 		go func(source types.Source, limit int, artworkCh chan *types.Artwork, interval int) {
// 			if interval <= 0 {
// 				return
// 			}
// 			for {
// 				err := source.FetchNewArtworksWithCh(artworkCh, limit)
// 				if err != nil {
// 					common.Logger.Errorf("Error when fetching from %s: %s", name, err)
// 				}
// 				time.Sleep(time.Duration(interval) * time.Minute)
// 			}
// 		}(source, config.Cfg.Fetcher.Limit, artworkCh, source.Config().Intervel)
// 	}
// 	common.Logger.Infof("Enabled sources: %s", enabledSources)
// 	for artwork := range artworkCh {
// 		if artwork == nil {
// 			continue
// 		}
// 		if telegram.IsChannelAvailable {
// 			err := telegram.PostAndCreateArtwork(ctx, artwork, telegram.Bot, config.Cfg.Telegram.Admins[0], 0)
// 			if err != nil {
// 				if es.Is(err, errs.ErrArtworkAlreadyExist) || es.Is(err, errs.ErrArtworkDeleted) {
// 					continue
// 				}
// 				common.Logger.Error(err.Error())
// 			}
// 			time.Sleep(time.Duration(int(config.Cfg.Telegram.Sleep)*len(artwork.Pictures)) * time.Second)
// 		} else {
// 			artworkInDB, err := service.GetArtworkByURL(ctx, artwork.SourceURL)
// 			if err == nil && artworkInDB != nil {
// 				common.Logger.Debugf("Artwork %s already exists", artwork.Title)
// 				continue
// 			}
// 			if err != nil && !es.Is(err, mongo.ErrNoDocuments) {
// 				common.Logger.Errorf(err.Error())
// 				continue
// 			}
// 			if service.CheckDeletedByURL(ctx, artwork.SourceURL) {
// 				common.Logger.Debugf("Artwork %s is deleted", artwork.Title)
// 				continue
// 			}

// 			saveSuccess := true
// 			for i, picture := range artwork.Pictures {
// 				info, err := storage.SaveAll(ctx, artwork, picture)
// 				if err != nil {
// 					common.Logger.Errorf("saving picture %d of artwork %s: %s", i, artwork.Title, err)
// 					saveSuccess = false
// 					break
// 				}
// 				artwork.Pictures[i].StorageInfo = info
// 			}
// 			if !saveSuccess {
// 				continue
// 			}
// 			artwork, err = service.CreateArtwork(ctx, artwork)
// 			if err != nil {
// 				common.Logger.Errorf(err.Error())
// 				continue
// 			}
// 			go func() {
// 				for _, picture := range artwork.Pictures {
// 					service.AddProcessPictureTask(ctx, picture)
// 				}
// 			}()
// 		}
// 	}
// }

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
