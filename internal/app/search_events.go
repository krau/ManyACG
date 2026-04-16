package app

import (
	"context"

	"github.com/duke-git/lancet/v2/retry"
	"github.com/krau/ManyACG/internal/infra/search"
	"github.com/krau/ManyACG/internal/model/converter"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/repo"
	"github.com/krau/ManyACG/pkg/log"
)

type dtoArtworkEventItem = dto.ArtworkEventItem

func registerArtworkEventSearcherHandlers(ctx context.Context, bus repo.EventBus[*dto.ArtworkEventItem], searcher search.Searcher) {
	filter := func(payload *dto.ArtworkEventItem) bool {
		return payload != nil && !payload.ID.IsZero()
	}
	bus.Subscribe(repo.EventTypeArtworkCreate, func(payload *dto.ArtworkEventItem) {
		retry.Retry(func() error {
			doc := converter.DtoArtworkEventItemToSearchDocument(payload)
			if doc == nil {
				return nil
			}
			err := searcher.AddDocuments(ctx, []*dto.ArtworkSearchDocument{doc})
			if err != nil {
				log.Error(err)
				return err
			}
			log.Debug("indexed artwork", "id", payload.ID, "title", payload.Title)
			return nil
		}, retry.Context(ctx))
	}, filter)
	bus.Subscribe(repo.EventTypeArtworkUpdate, func(payload *dto.ArtworkEventItem) {
		retry.Retry(func() error {
			doc := converter.DtoArtworkEventItemToSearchDocument(payload)
			if doc == nil {
				return nil
			}
			err := searcher.AddDocuments(ctx, []*dto.ArtworkSearchDocument{doc})
			if err != nil {
				log.Error(err)
				return err
			}
			log.Debug("re-indexed artwork", "id", payload.ID, "title", payload.Title)
			return nil
		}, retry.Context(ctx))
	}, filter)
	bus.Subscribe(repo.EventTypeArtworkDelete, func(payload *dto.ArtworkEventItem) {
		retry.Retry(func() error {
			err := searcher.DeleteDocuments(ctx, []string{payload.ID.Hex()})
			if err != nil {
				log.Error(err)
				return err
			}
			log.Debug("deleted indexed artwork", "id", payload.ID, "title", payload.Title)
			return nil
		}, retry.Context(ctx))
	}, filter)
}
