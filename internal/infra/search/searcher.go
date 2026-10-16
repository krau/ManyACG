package search

import (
	"context"
	"errors"
	"fmt"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/search/meilisearch"
	"github.com/krau/ManyACG/internal/infra/search/mocksearch"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/repo"
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

var ErrNotEnabled = errors.New("search engine is not enabled")

func New(ctx context.Context, cfg runtimecfg.SearchConfig, artworks repo.Artwork, initializeIndex bool) (Searcher, Indexer, error) {
	if !cfg.Enable {
		return &noopSearcher{}, nil, nil
	}
	switch cfg.Engine {
	case "meilisearch":
		if !initializeIndex {
			s, err := meilisearch.Open(ctx, cfg.MeiliSearch)
			return s, nil, err
		}
		s, err := meilisearch.NewSearcher(ctx, cfg.MeiliSearch)
		if err != nil {
			return nil, nil, err
		}
		return s, s, nil
	case "mock":
		log.Warn("using mock searcher, this is only for testing purposes")
		s := mocksearch.NewSearcher(artworks)
		if !initializeIndex {
			return s, nil, nil
		}
		return s, s, nil
	default:
		return nil, nil, fmt.Errorf("unknown search engine %q", cfg.Engine)
	}
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
