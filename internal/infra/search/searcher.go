package search

import (
	"context"
	"errors"
	"sync"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/search/meilisearch"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/pkg/log"
)

type Searcher interface {
	SearchArtworks(ctx context.Context, que *query.ArtworkSearch) (*dto.ArtworkSearchResult, error)
	FindSimilarArtworks(ctx context.Context, que *query.ArtworkSimilar) (*dto.ArtworkSearchResult, error)
}

var (
	defaultSearcher Searcher
	enabled         bool
	once            sync.Once
	ErrNotEnabled   = errors.New("search engine is not enabled")
)

func initDefault(ctx context.Context) {
	cfg := runtimecfg.Get().Search
	enabled = cfg.Enable
	if !enabled {
		return
	}
	switch cfg.Engine {
	case "meilisearch":
		s, err := meilisearch.NewSearcher(ctx, cfg.MeiliSearch)
		if err != nil {
			log.Fatal("failed to initialize meilisearch searcher", "err", err)
		}
		defaultSearcher = s
	}
}

func Enabled() bool {
	return enabled
}

type noopSearcher struct{}

func (s *noopSearcher) SearchArtworks(ctx context.Context, que *query.ArtworkSearch) (*dto.ArtworkSearchResult, error) {
	return nil, ErrNotEnabled
}

func (s *noopSearcher) FindSimilarArtworks(ctx context.Context, que *query.ArtworkSimilar) (*dto.ArtworkSearchResult, error) {
	return nil, ErrNotEnabled
}

func Default() Searcher {
	if !enabled {
		return &noopSearcher{}
	}
	return getDefault(context.Background())
}

func IsNotEnabledErr(err error) bool {
	return errors.Is(err, ErrNotEnabled)
}

func getDefault(ctx context.Context) Searcher {
	once.Do(func() {
		initDefault(ctx)
	})
	return defaultSearcher
}

func SearchArtworks(ctx context.Context, que *query.ArtworkSearch) (*dto.ArtworkSearchResult, error) {
	if !enabled {
		return nil, ErrNotEnabled
	}
	return getDefault(ctx).SearchArtworks(ctx, que)
}

func FindSimilarArtworks(ctx context.Context, que *query.ArtworkSimilar) (*dto.ArtworkSearchResult, error) {
	if !enabled {
		return nil, ErrNotEnabled
	}
	return getDefault(ctx).FindSimilarArtworks(ctx, que)
}
