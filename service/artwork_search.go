package service

// func HybridSearchArtworks(ctx context.Context, queryText string, hybridSemanticRatio float64, offset, limit int64, r18 types.R18Type, options ...*types.AdapterOption) ([]*types.Artwork, error) {
// 	if common.MeilisearchClient == nil {
// 		return nil, errs.ErrSearchEngineUnavailable
// 	}

// 	var filter string
// 	switch r18 {
// 	case types.R18TypeAll:
// 		filter = ""
// 	case types.R18TypeNone:
// 		filter = "r18 = false"
// 	case types.R18TypeOnly:
// 		filter = "r18 = true"
// 	}

// 	index := common.MeilisearchClient.Index(config.Cfg.Search.MeiliSearch.Index)
// 	resp, err := index.SearchWithContext(ctx, queryText, &meilisearch.SearchRequest{
// 		Offset:               offset,
// 		Limit:                limit,
// 		AttributesToRetrieve: []string{"id"},
// 		Filter:               filter,
// 		Hybrid: &meilisearch.SearchRequestHybrid{
// 			Embedder:      config.Cfg.Search.MeiliSearch.Embedder,
// 			SemanticRatio: hybridSemanticRatio,
// 		},
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	hits := resp.Hits
// 	artworkSearchDocs := make([]*types.ArtworkSearchDocument, 0, len(hits))
// 	hitsBytes, err := json.Marshal(hits)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = json.Unmarshal(hitsBytes, &artworkSearchDocs)
// 	if err != nil {
// 		return nil, err
// 	}
// 	artworkModels := make([]*types.ArtworkModel, 0, len(artworkSearchDocs))
// 	for _, doc := range artworkSearchDocs {
// 		objectID, err := primitive.ObjectIDFromHex(doc.ID)
// 		if err != nil {
// 			return nil, err
// 		}
// 		artworkModel, err := dao.GetArtworkByID(ctx, objectID)
// 		if err != nil {
// 			return nil, err
// 		}
// 		artworkModels = append(artworkModels, artworkModel)
// 	}
// 	return adapter.ConvertToArtworks(ctx, artworkModels, options...)
// }

// func SearchSimilarArtworks(ctx context.Context, artworkIdStr string, offset, limit int64, r18 types.R18Type, options ...*types.AdapterOption) ([]*types.Artwork, error) {
// 	if common.MeilisearchClient == nil {
// 		return nil, errs.ErrSearchEngineUnavailable
// 	}

// 	var filter string
// 	switch r18 {
// 	case types.R18TypeAll:
// 		filter = ""
// 	case types.R18TypeNone:
// 		filter = "r18 = false"
// 	case types.R18TypeOnly:
// 		filter = "r18 = true"
// 	}

// 	index := common.MeilisearchClient.Index(config.Cfg.Search.MeiliSearch.Index)
// 	var resp meilisearch.SimilarDocumentResult
// 	if err := index.SearchSimilarDocumentsWithContext(ctx, &meilisearch.SimilarDocumentQuery{
// 		AttributesToRetrieve: []string{"id"},
// 		Id:                   artworkIdStr,
// 		Embedder:             config.Cfg.Search.MeiliSearch.Embedder,
// 		Offset:               offset,
// 		Limit:                limit,
// 		Filter:               filter,
// 	}, &resp); err != nil {
// 		if strings.Contains(err.Error(), "not_found_similar_id") {
// 			// [TODO] need better error handling here but meilisearch-go does not provide a way to distinguish this error
// 			go func() {
// 				common.Logger.Warnf("No similar artworks found for ID %s", artworkIdStr)
// 				artworkID, err := primitive.ObjectIDFromHex(artworkIdStr)
// 				if err != nil {
// 					common.Logger.Errorf("Invalid artwork ID: %s", artworkIdStr)
// 					return
// 				}
// 				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 				defer cancel()
// 				artwork, err := dao.GetArtworkByID(ctx, artworkID)
// 				if err != nil {
// 					common.Logger.Errorf("Failed to get artwork by ID %s: %s", artworkIdStr, err)
// 					return
// 				}
// 				doc, err := adapter.ConvertToSearchDoc(ctx, artwork)
// 				if err != nil {
// 					common.Logger.Errorf("Failed to convert artwork to search doc: %s", err)
// 					return
// 				}
// 				artworkJson, err := json.Marshal(doc)
// 				if err != nil {
// 					common.Logger.Errorf("Failed to marshal artwork search doc: %s", err)
// 					return
// 				}
// 				task, err := index.UpdateDocumentsWithContext(ctx, artworkJson)
// 				if err != nil {
// 					common.Logger.Errorf("Failed to update artwork to Meilisearch: %s", err)
// 					return
// 				}
// 				common.Logger.Debugf("Committed update artwork task to Meilisearch: %d", task.TaskUID)
// 			}()
// 		}
// 		return nil, err
// 	}
// 	hits := resp.Hits
// 	artworkSearchDocs := make([]*types.ArtworkSearchDocument, 0, len(hits))
// 	hitsBytes, err := json.Marshal(hits)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = json.Unmarshal(hitsBytes, &artworkSearchDocs)
// 	if err != nil {
// 		return nil, err
// 	}
// 	artworkModels := make([]*types.ArtworkModel, 0, len(artworkSearchDocs))
// 	for _, doc := range artworkSearchDocs {
// 		objectID, err := primitive.ObjectIDFromHex(doc.ID)
// 		if err != nil {
// 			return nil, err
// 		}
// 		artworkModel, err := dao.GetArtworkByID(ctx, objectID)
// 		if err != nil {
// 			return nil, err
// 		}
// 		artworkModels = append(artworkModels, artworkModel)
// 	}
// 	return adapter.ConvertToArtworks(ctx, artworkModels, options...)
// }
