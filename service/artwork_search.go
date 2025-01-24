package service

import (
	"context"

	"github.com/bytedance/sonic"
	"github.com/krau/ManyACG/adapter"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/dao"
	"github.com/krau/ManyACG/errs"
	"github.com/krau/ManyACG/types"
	"github.com/meilisearch/meilisearch-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func HybridSearchArtworks(ctx context.Context, queryText string, hybridSemanticRatio float64, limit int64, options ...*types.AdapterOption) ([]*types.Artwork, error) {
	if common.MeilisearchClient == nil {
		return nil, errs.ErrNotEnabledHybridSearch
	}
	index := common.MeilisearchClient.Index(config.Cfg.Search.MeiliSearch.Index)
	resp, err := index.SearchWithContext(ctx, queryText, &meilisearch.SearchRequest{
		Limit: limit,
		Hybrid: &meilisearch.SearchRequestHybrid{
			SemanticRatio: hybridSemanticRatio,
		},
	})
	if err != nil {
		return nil, err
	}
	hits := resp.Hits
	artworkSearchDocs := make([]*types.ArtworkSearchDocument, 0, len(hits))
	hitsBytes, err := sonic.Marshal(hits)
	if err != nil {
		return nil, err
	}
	err = sonic.Unmarshal(hitsBytes, &artworkSearchDocs)
	if err != nil {
		return nil, err
	}
	artworkModels := make([]*types.ArtworkModel, 0, len(artworkSearchDocs))
	for _, doc := range artworkSearchDocs {
		objectID, err := primitive.ObjectIDFromHex(doc.ID)
		if err != nil {
			return nil, err
		}
		artworkModel, err := dao.GetArtworkByID(ctx, objectID)
		if err != nil {
			return nil, err
		}
		artworkModels = append(artworkModels, artworkModel)
	}
	return adapter.ConvertToArtworks(ctx, artworkModels, options...)
}
