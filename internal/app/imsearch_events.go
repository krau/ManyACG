package app

import (
	"context"

	"github.com/duke-git/lancet/v2/retry"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/repo"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/pkg/log"
)

func registerArtworkEventImsearchHandlers(ctx context.Context, bus repo.EventBus[*dto.ArtworkEventItem], serv *service.Service) {
	if serv == nil || serv.Imsearch() == nil || !serv.Imsearch().Enabled() {
		return
	}
	filter := func(payload *dto.ArtworkEventItem) bool {
		return payload != nil && !payload.ID.IsZero()
	}
	bus.Subscribe(repo.EventTypeArtworkCreate, func(payload *dto.ArtworkEventItem) {
		retry.Retry(func() error {
			return indexArtworkPictures(ctx, serv, payload)
		}, retry.Context(ctx))
	}, filter)
	bus.Subscribe(repo.EventTypeArtworkUpdate, func(payload *dto.ArtworkEventItem) {
		retry.Retry(func() error {
			return indexArtworkPictures(ctx, serv, payload)
		}, retry.Context(ctx))
	}, filter)
	bus.Subscribe(repo.EventTypeArtworkDelete, func(payload *dto.ArtworkEventItem) {
		retry.Retry(func() error {
			for _, p := range payload.Pictures {
				if err := serv.DeletePictureFromImsearch(ctx, p.ID); err != nil {
					log.Error("imsearch delete picture failed", "picture_id", p.ID, "err", err)
					return err
				}
			}
			log.Debug("imsearch deleted artwork pictures", "artwork_id", payload.ID, "count", len(payload.Pictures))
			return nil
		}, retry.Context(ctx))
	}, filter)
}

func indexArtworkPictures(ctx context.Context, serv *service.Service, payload *dto.ArtworkEventItem) error {
	for _, p := range payload.Pictures {
		pic, err := serv.GetPictureByID(ctx, p.ID)
		if err != nil {
			log.Error("imsearch load picture failed", "picture_id", p.ID, "err", err)
			return err
		}
		if err := serv.IndexPictureForImsearch(ctx, pic); err != nil {
			log.Error("imsearch index picture failed", "picture_id", p.ID, "err", err)
			return err
		}
	}
	log.Debug("imsearch indexed artwork pictures", "artwork_id", payload.ID, "count", len(payload.Pictures))
	return nil
}
