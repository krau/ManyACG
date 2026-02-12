package meilisearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/goccy/go-json"
	"github.com/unvgo/ouid"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/meilisearch/meilisearch-go"
)

type SearcherMeilisearch struct {
	client meilisearch.IndexManager
	cfg    runtimecfg.MeiliSearchConfig
}

func meilisearchIndexSettings() *meilisearch.Settings {
	return &meilisearch.Settings{
		FilterableAttributes: []string{
			"r18",
			"tags",
			"artist",
		},
		SearchableAttributes: []string{
			"title",
			"artist",
			"tags",
			"description",
		},
	}
}

func settingsEqual(a, b *meilisearch.Settings) bool {
	if a == nil || b == nil {
		return a == b
	}
	// 比较二者的 FilterableAttributes 和 SearchableAttributes 是否相同，忽略顺序
	if len(a.FilterableAttributes) != len(b.FilterableAttributes) ||
		len(a.SearchableAttributes) != len(b.SearchableAttributes) {
		return false
	}
	filterableMap := make(map[string]struct{})
	for _, attr := range a.FilterableAttributes {
		filterableMap[attr] = struct{}{}
	}
	for _, attr := range b.FilterableAttributes {
		if _, ok := filterableMap[attr]; !ok {
			return false
		}
	}
	searchableMap := make(map[string]struct{})
	for _, attr := range a.SearchableAttributes {
		searchableMap[attr] = struct{}{}
	}
	for _, attr := range b.SearchableAttributes {
		if _, ok := searchableMap[attr]; !ok {
			return false
		}
	}
	return true
}

// AddDocuments implements search.Searcher.
func (m *SearcherMeilisearch) AddDocuments(ctx context.Context, docs []*dto.ArtworkSearchDocument) error {
	primaryKey := "id"
	_, err := m.client.AddDocumentsWithContext(ctx, docs, &meilisearch.DocumentOptions{
		PrimaryKey: &primaryKey,
	})
	return err
}

// DeleteDocuments implements search.Searcher.
func (m *SearcherMeilisearch) DeleteDocuments(ctx context.Context, ids []string) error {
	_, err := m.client.DeleteDocumentsWithContext(ctx, ids, nil)
	return err
}

// DeleteAllDocuments implements search.Searcher.
func (m *SearcherMeilisearch) DeleteAllDocuments(ctx context.Context) error {
	_, err := m.client.DeleteAllDocumentsWithContext(ctx, nil)
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
	// create index if not exists
	index := manager.Index(cfg.Index)
	_, err = index.FetchInfoWithContext(ctx)
	if err == nil {
		currentSettings, err := index.GetSettingsWithContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("meilisearch get settings failed: %w", err)
		}
		desiredSettings := meilisearchIndexSettings()
		if !settingsEqual(currentSettings, desiredSettings) {
			log.Info("meilisearch index settings are different from desired settings, updating settings")
			_, err = index.UpdateSettingsWithContext(ctx, desiredSettings)
			if err != nil {
				return nil, fmt.Errorf("meilisearch update settings failed: %w", err)
			}
		}
	}
	// 如果索引不存在，则创建索引并设置 settings
	if err != nil && strings.Contains(err.Error(), "index_not_found") {
		log.Info("meilisearch index not found, creating index", "index", cfg.Index)
		_, err = manager.CreateIndexWithContext(ctx, &meilisearch.IndexConfig{
			Uid:        cfg.Index,
			PrimaryKey: "id",
		})
		if err != nil {
			return nil, fmt.Errorf("meilisearch create index failed: %w", err)
		}
		index = manager.Index(cfg.Index)
		_, err = index.UpdateSettingsWithContext(ctx, meilisearchIndexSettings())
		if err != nil {
			return nil, fmt.Errorf("meilisearch update settings failed: %w", err)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("meilisearch fetch index info failed: %w", err)
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
	oids := make([]ouid.OUID, 0, len(docs))
	for _, doc := range docs {
		oid, err := ouid.FromObjectIDHex(doc.ID)
		if err != nil {
			return nil, fmt.Errorf("meilisearch parse objectid failed: %w", err)
		}
		oids = append(oids, oid)
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
		Embedder:             m.cfg.Embedder,
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
	oids := make([]ouid.OUID, 0, len(docs))
	for _, doc := range docs {
		oid, err := ouid.FromObjectIDHex(doc.ID)
		if err != nil {
			return nil, fmt.Errorf("meilisearch parse objectid failed: %w", err)
		}
		oids = append(oids, oid)
	}
	return &dto.ArtworkSearchResult{
		IDs: oids,
	}, nil
}
