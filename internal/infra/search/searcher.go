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
	"github.com/unvgo/ouid"
)

type Searcher interface {
	SearchArtworks(ctx context.Context, que *query.ArtworkSearch) (*dto.ArtworkSearchResult, error)
	FindSimilarArtworks(ctx context.Context, que *query.ArtworkSimilar) (*dto.ArtworkSearchResult, error)
}

type Indexer interface {
	GetMissingArtworkIDs(ctx context.Context, ids []ouid.OUID) ([]ouid.OUID, error)
	AddDocuments(ctx context.Context, docs []*dto.ArtworkSearchDocument) error
	AddDocumentsAndWait(ctx context.Context, docs []*dto.ArtworkSearchDocument) error
	DeleteDocuments(ctx context.Context, ids []string) error
}

var (
	defaultSearcher Searcher
	defaultIndexer  Indexer
	once            sync.Once
	ErrNotEnabled   = errors.New("search engine is not enabled")
)

func initDefault(ctx context.Context) {
	if !Enabled() {
		defaultSearcher = &noopSearcher{}
		return
	}
	cfg := runtimecfg.Get()
	switch cfg.Search.Engine {
	case "meilisearch":
		if cfg.App.Mode.ReadOnly() {
			s, err := meilisearch.NewReader(ctx, cfg.Search.MeiliSearch)
			if err != nil {
				log.Fatal("failed to initialize meilisearch reader", "err", err)
			}
			defaultSearcher = s
			return
		}
		s, err := meilisearch.NewSearcher(ctx, cfg.Search.MeiliSearch)
		if err != nil {
			log.Fatal("failed to initialize meilisearch searcher", "err", err)
		}
		defaultSearcher = s
		defaultIndexer = s
	case "mock":
		log.Warn("using mock searcher, this is only for testing purposes")
		s := mocksearch.NewSearcher(database.Default().Artwork())
		if cfg.App.Mode.ReadOnly() {
			defaultSearcher = &querySearcher{Searcher: s}
			return
		}
		defaultSearcher = s
		defaultIndexer = s
	default:
		log.Fatal("unknown search engine: %s", cfg.Search.Engine)
	}
}

func Enabled() bool {
	return runtimecfg.Get().Search.Enable
}

type querySearcher struct {
	Searcher
}

type noopSearcher struct{}

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

func DefaultIndexer(ctx context.Context) Indexer {
	once.Do(func() {
		initDefault(ctx)
	})
	return defaultIndexer
}
