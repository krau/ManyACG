package scheduler

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/model/converter"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/log"
)

// Poster 应该完成所有创建工作, 包括文件存储等
type ArtworkPoster interface {
	PostAndCreateArtwork(ctx context.Context, artwork *entity.CachedArtworkData) error
}

func StartPoster(ctx context.Context, poster ArtworkPoster, serv *service.Service) {
	cfg := runtimecfg.Get().Scheduler
	if !cfg.Enable || poster == nil {
		return
	}
	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
	defer ticker.Stop()
	sources := serv.Sources()
	limit := cfg.Limit
	if limit <= 0 {
		limit = math.MaxInt
	}
	timeout := time.Duration(cfg.Interval-10) * time.Second
	doTask := func() {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		log.Info("scheduler: start fetching new artworks")
		seen := make(map[string]struct{})
		fetcheds := make([]*dto.FetchedArtwork, 0)
		for _, sou := range sources {
			artworks, err := sou.FetchNewArtworks(ctx, limit)
			if err != nil {
				log.Error("fetching new artworks from source", "source", fmt.Sprintf("%T", sou), "err", err)
			}
			if len(artworks) == 0 {
				continue
			}
			log.Info("fetched artworks", "count", len(artworks), "source", fmt.Sprintf("%T", sou))
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
				log.Error("checking cached artwork", "url", fetchedArtwork.SourceURL, "err", err)
				continue
			}
			if err != nil {
				cached, err := serv.CreateCachedArtwork(ctx, converter.DtoFetchedArtworkToEntityCached(fetchedArtwork))
				if err != nil {
					log.Error("creating cached artwork", "url", fetchedArtwork.SourceURL, "err", err)
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
			if err := poster.PostAndCreateArtwork(ctx, data); err != nil {
				log.Error("posting artwork", "url", exist.SourceURL, "err", err)
				continue
			}
			log.Info("posted artwork", "url", exist.SourceURL)
		}
	}
	doTask()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			doTask()
		}
	}
}
