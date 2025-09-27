package meilisearch

import (
	"os"

	"github.com/meilisearch/meilisearch-go"
)

var MeilisearchClient meilisearch.ServiceManager

func initMeilisearch() {
	if config.Cfg.Search.MeiliSearch.Host == "" || config.Cfg.Search.MeiliSearch.Key == "" || config.Cfg.Search.MeiliSearch.Index == "" {
		Logger.Fatalf("Meilisearch configuration is incomplete")
		os.Exit(1)
	}
	client := meilisearch.New(config.Cfg.Search.MeiliSearch.Host, meilisearch.WithAPIKey(config.Cfg.Search.MeiliSearch.Key))
	health, err := client.Health()
	if err != nil {
		Logger.Fatalf("Meilisearch health check failed: %v", err)
		os.Exit(1)
	}
	Logger.Infof("Meilisearch health check: %s", health.Status)
	MeilisearchClient = client
}
