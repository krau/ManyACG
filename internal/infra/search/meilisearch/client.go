package meilisearch

import (
	"fmt"

	"github.com/meilisearch/meilisearch-go"
)

func NewMeilisearch(
	host, key, index string,
) (meilisearch.IndexManager, error) {
	client := meilisearch.New(host, meilisearch.WithAPIKey(key))
	_, err := client.Health()
	if err != nil {
		return nil, fmt.Errorf("meilisearch health check failed: %w", err)
	}
	indexClient := client.Index(index)
	return indexClient, nil
}
