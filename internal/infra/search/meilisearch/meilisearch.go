package meilisearch

import (
	"context"
	"fmt"

	"github.com/goccy/go-json"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"github.com/meilisearch/meilisearch-go"
)

type SearcherMeilisearch struct {
	client meilisearch.IndexManager
	cfg    runtimecfg.MeiliSearchConfig
}

// AddDocuments implements search.Searcher.
func (m *SearcherMeilisearch) AddDocuments(ctx context.Context, docs []*dto.ArtworkSearchDocument) error {
	_, err := m.client.AddDocumentsWithContext(ctx, docs, "id")
	return err
}

// DeleteDocuments implements search.Searcher.
func (m *SearcherMeilisearch) DeleteDocuments(ctx context.Context, ids []string) error {
	_, err := m.client.DeleteDocumentsWithContext(ctx, ids)
	return err
}

func NewSearcher(ctx context.Context, cfg runtimecfg.MeiliSearchConfig) (*SearcherMeilisearch, error) {
	if err := cfg.Valid(); err != nil {
		return nil, err
	}
	manager := meilisearch.New(cfg.Host, meilisearch.WithAPIKey(cfg.Key))
	_, err := manager.HealthWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("meilisearch health check failed: %w", err)
	}
	return &SearcherMeilisearch{client: manager.Index(cfg.Index), cfg: cfg}, nil
}

func (m *SearcherMeilisearch) SearchArtworks(ctx context.Context, que *query.ArtworkSearch) (*dto.ArtworkSearchResult, error) {
	filter := map[shared.R18Type]string{
		shared.R18TypeAll:  "",
		shared.R18TypeNone: "r18 = false",
		shared.R18TypeR18:  "r18 = true",
	}[que.R18]
	// [TODO] handle more filters
	req := &meilisearch.SearchRequest{
		AttributesToRetrieve: []string{"id"},
		Filter:               filter,
		Offset:               int64(que.Offset),
		Limit:                int64(que.Limit),
	}
	if que.Hybrid {
		req.Hybrid = &meilisearch.SearchRequestHybrid{
			Embedder:      m.cfg.Embedder,
			SemanticRatio: que.HybridSemanticRatio,
		}
	}
	resp, err := m.client.SearchWithContext(ctx, que.Query, req)
	if err != nil {
		return nil, fmt.Errorf("meilisearch search failed: %w", err)
	}
	hits := resp.Hits
	docs := make([]*dto.ArtworkSearchDocument, 0, len(hits))
	hitsBytes, err := json.Marshal(hits)
	if err != nil {
		return nil, fmt.Errorf("meilisearch marshal hits failed: %w", err)
	}
	err = json.Unmarshal(hitsBytes, &docs)
	if err != nil {
		return nil, fmt.Errorf("meilisearch unmarshal hits failed: %w", err)
	}
	oids := make([]objectuuid.ObjectUUID, 0, len(docs))
	for i, doc := range docs {
		oid, err := objectuuid.FromObjectIDHex(doc.ID)
		if err != nil {
			return nil, fmt.Errorf("meilisearch parse objectid failed: %w", err)
		}
		oids[i] = oid
	}
	return &dto.ArtworkSearchResult{
		IDs: oids,
	}, nil
}

func (m *SearcherMeilisearch) FindSimilarArtworks(ctx context.Context, que *query.ArtworkSimilar) (*dto.ArtworkSearchResult, error) {
	filter := map[shared.R18Type]string{
		shared.R18TypeAll:  "",
		shared.R18TypeNone: "r18 = false",
		shared.R18TypeR18:  "r18 = true",
	}[que.R18]
	req := &meilisearch.SimilarDocumentQuery{
		AttributesToRetrieve: []string{"id"},
		Id:                   que.ArtworkID.Hex(),
		Filter:               filter,
		Offset:               int64(que.Offset),
		Limit:                int64(que.Limit),
	}
	var resp meilisearch.SimilarDocumentResult
	err := m.client.SearchSimilarDocumentsWithContext(ctx, req, &resp)
	if err != nil {
		return nil, fmt.Errorf("meilisearch similar search failed: %w", err)
	}
	hits := resp.Hits
	docs := make([]*dto.ArtworkSearchDocument, 0, len(hits))
	hitsBytes, err := json.Marshal(hits)
	if err != nil {
		return nil, fmt.Errorf("meilisearch marshal hits failed: %w", err)
	}
	err = json.Unmarshal(hitsBytes, &docs)
	if err != nil {
		return nil, fmt.Errorf("meilisearch unmarshal hits failed: %w", err)
	}
	oids := make([]objectuuid.ObjectUUID, 0, len(docs))
	for i, doc := range docs {
		oid, err := objectuuid.FromObjectIDHex(doc.ID)
		if err != nil {
			return nil, fmt.Errorf("meilisearch parse objectid failed: %w", err)
		}
		oids[i] = oid
	}
	return &dto.ArtworkSearchResult{
		IDs: oids,
	}, nil
}
