package mocksearch

import (
	"context"

	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/repo"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

// mock searcher using random artworks from repo, only for testing purposes
type SearcherMock struct {
	repo repo.Artwork
}

// AddDocuments implements search.Searcher.
func (m *SearcherMock) AddDocuments(ctx context.Context, docs []*dto.ArtworkSearchDocument) error {
	return nil
}

// DeleteDocuments implements search.Searcher.
func (m *SearcherMock) DeleteDocuments(ctx context.Context, ids []string) error {
	return nil
}

func NewSearcher(awRepo repo.Artwork) *SearcherMock {
	return &SearcherMock{
		repo: awRepo,
	}
}

func (m *SearcherMock) SearchArtworks(ctx context.Context, que *query.ArtworkSearch) (*dto.ArtworkSearchResult, error) {
	res, err := m.repo.QueryArtworks(ctx, query.ArtworksDB{
		Paginate: query.Paginate{
			Limit: que.Limit,
		},
		Random: true,
	})
	if err != nil {
		return nil, err
	}
	ids := make([]objectuuid.ObjectUUID, 0, que.Limit)
	for _, aw := range res {
		ids = append(ids, aw.ID)
	}
	return &dto.ArtworkSearchResult{
		IDs: ids,
	}, nil
}

func (m *SearcherMock) FindSimilarArtworks(ctx context.Context, que *query.ArtworkSimilar) (*dto.ArtworkSearchResult, error) {
	res, err := m.repo.QueryArtworks(ctx, query.ArtworksDB{
		Paginate: query.Paginate{
			Limit: que.Limit,
		},
		Random: true,
	})
	if err != nil {
		return nil, err
	}
	ids := make([]objectuuid.ObjectUUID, 0, que.Limit)
	for _, aw := range res {
		ids = append(ids, aw.ID)
	}
	return &dto.ArtworkSearchResult{
		IDs: ids,
	}, nil
}
