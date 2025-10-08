package search

import (
	"context"
	"errors"
	"sync"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/database"
	"github.com/krau/ManyACG/internal/infra/search/meilisearch"
	"github.com/krau/ManyACG/internal/infra/search/mocksearch"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/pkg/log"
)

type Searcher interface {
	SearchArtworks(ctx context.Context, que *query.ArtworkSearch) (*dto.ArtworkSearchResult, error)
	FindSimilarArtworks(ctx context.Context, que *query.ArtworkSimilar) (*dto.ArtworkSearchResult, error)
	AddDocuments(ctx context.Context, docs []*dto.ArtworkSearchDocument) error
	DeleteDocuments(ctx context.Context, ids []string) error
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
		defaultSearcher = &noopSearcher{}
		return
	}
	switch cfg.Engine {
	case "meilisearch":
		s, err := meilisearch.NewSearcher(ctx, cfg.MeiliSearch)
		if err != nil {
			log.Fatal("failed to initialize meilisearch searcher", "err", err)
		}
		defaultSearcher = s
	case "mock":
		log.Warn("using mock searcher, this is only for testing purposes")
		defaultSearcher = mocksearch.NewSearcher(database.Default().Artwork())
	default:
		log.Fatal("unknown search engine: %s", cfg.Engine)
	}
}

func Enabled() bool {
	return enabled
}

type noopSearcher struct{}

// AddDocuments implements Searcher.
func (s *noopSearcher) AddDocuments(ctx context.Context, docs []*dto.ArtworkSearchDocument) error {
	return ErrNotEnabled
}

// DeleteDocuments implements Searcher.
func (s *noopSearcher) DeleteDocuments(ctx context.Context, ids []string) error {
	return ErrNotEnabled
}

func (s *noopSearcher) SearchArtworks(ctx context.Context, que *query.ArtworkSearch) (*dto.ArtworkSearchResult, error) {
	return nil, ErrNotEnabled
}

func (s *noopSearcher) FindSimilarArtworks(ctx context.Context, que *query.ArtworkSimilar) (*dto.ArtworkSearchResult, error) {
	return nil, ErrNotEnabled
}

func IsNotEnabledErr(err error) bool {
	return errors.Is(err, ErrNotEnabled)
}

func Default(ctx context.Context) Searcher {
	once.Do(func() {
		initDefault(ctx)
	})
	return defaultSearcher
}
