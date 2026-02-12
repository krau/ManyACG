package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/krau/ManyACG/internal/infra"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/database"
	"github.com/krau/ManyACG/internal/infra/search"
	"github.com/krau/ManyACG/internal/model/converter"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index all artworks to MeiliSearch",
	Long:  "Index all artworks from database to the specified MeiliSearch index",
	Run: func(cmd *cobra.Command, args []string) {
		IndexArtworks(cmd.Context())
	},
}

var (
	indexBatchSize int
)

func init() {
	rootCmd.AddCommand(indexCmd)
	indexCmd.Flags().IntVarP(&indexBatchSize, "batch", "b", 1000, "Batch size for indexing artworks")
}

func IndexArtworks(ctx context.Context) {
	// 读取配置
	cfg := runtimecfg.Get()

	// 初始化日志
	log.SetDefault(log.New(log.Config{}))

	// 检查搜索是否启用
	if !cfg.Search.Enable {
		log.Fatal("search engine is not enabled in config")
	}

	// 检查搜索引擎类型
	if cfg.Search.Engine != "meilisearch" {
		log.Fatal("only meilisearch is supported for indexing", "engine", cfg.Search.Engine)
	}

	// 初始化基础设施（数据库等）
	closer, err := infra.Init(ctx, cfg)
	if err != nil {
		log.Fatal("failed to initialize infrastructure", "err", err)
	}
	defer func() {
		if closer != nil {
			if err := closer(); err != nil {
				log.Error("failed to close infrastructure", "err", err)
			}
		}
	}()

	// 获取数据库实例
	db := database.Default()

	// 初始化搜索引擎
	searcher := search.Default(ctx)

	log.Info("Starting to index all artworks to MeiliSearch", "batchSize", indexBatchSize)

	// 统计总数
	total, err := db.Artwork().CountArtworks(ctx, shared.R18TypeAll)
	if err != nil {
		log.Fatal("failed to count artworks", "err", err)
	}

	log.Info("Total artworks to index", "count", total)

	if total == 0 {
		log.Info("No artworks found in database")
		return
	}

	// 批量读取并索引
	indexed := 0
	failed := 0
	offset := 0

	startTime := time.Now()

	for {
		// 查询一批 artwork
		que := query.ArtworksDB{
			ArtworksFilter: query.ArtworksFilter{
				R18: shared.R18TypeAll,
			},
			Paginate: query.Paginate{
				Limit:  indexBatchSize,
				Offset: offset,
			},
		}

		artworks, err := db.Artwork().QueryArtworks(ctx, que)
		if err != nil {
			log.Error("failed to query artworks", "err", err, "offset", offset)
			failed += indexBatchSize
			offset += indexBatchSize
			continue
		}

		if len(artworks) == 0 {
			break
		}

		// 转换为搜索文档
		docs := make([]*dto.ArtworkSearchDocument, 0, len(artworks))
		for _, artwork := range artworks {
			doc := converter.EntityArtworkToSearchDocument(artwork)
			if doc != nil {
				docs = append(docs, doc)
			}
		}

		// 批量添加到搜索引擎
		if len(docs) > 0 {
			err = searcher.AddDocuments(ctx, docs)
			if err != nil {
				log.Error("failed to add documents to search engine", "err", err, "count", len(docs))
				failed += len(docs)
			} else {
				indexed += len(docs)
				log.Info("Indexed batch", "count", len(docs), "progress", fmt.Sprintf("%d/%d", indexed, total))
			}
		}

		offset += indexBatchSize

		// 如果查询结果少于批次大小，说明已经到达末尾
		if len(artworks) < indexBatchSize {
			break
		}
	}

	duration := time.Since(startTime)

	log.Info("Indexing completed",
		"total", total,
		"indexed", indexed,
		"failed", failed,
		"duration", duration.String(),
		"rate", fmt.Sprintf("%.2f/s", float64(indexed)/duration.Seconds()))

	if failed > 0 {
		log.Warn("Some artworks failed to index", "failed", failed)
		os.Exit(1)
	}
}
