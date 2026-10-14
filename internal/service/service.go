package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/imseek"
	"github.com/krau/ManyACG/internal/infra/search"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/infra/storage"
	"github.com/krau/ManyACG/internal/infra/tagging"
	"github.com/krau/ManyACG/internal/repo"
	"github.com/krau/ManyACG/internal/shared"
)

type Service struct {
	repos      repo.Repositories
	searcher   search.Searcher
	indexer    search.Indexer
	tagger     tagging.Tagger
	storages   map[shared.StorageType]storage.Storage
	sources    map[shared.SourceType]source.ArtworkSource
	pollers    map[shared.SourceType]source.ArtworkPoller
	fetcher    source.ArtworkFetcher
	downloader httpclient.Fetcher
	storCfg    runtimecfg.StorageConfig
	imseek     imseek.Engine

	indexRepairMu       sync.Mutex
	indexRepairProgress ArtworkIndexProgress
}

type Option func(*Service)

func WithImseek(e imseek.Engine) Option {
	return func(s *Service) { s.imseek = e }
}

func WithIndexer(indexer search.Indexer) Option {
	return func(s *Service) { s.indexer = indexer }
}

func WithSourcePollers(pollers map[shared.SourceType]source.ArtworkPoller) Option {
	return func(s *Service) { s.pollers = pollers }
}

// WithArtworkFetcher delegates artwork detail fetching to a remote node.
// Without it the local sources are used.
func WithArtworkFetcher(fetcher source.ArtworkFetcher) Option {
	return func(s *Service) { s.fetcher = fetcher }
}

// WithDownloader delegates origin file downloads to a remote node.
// Without it files are downloaded from the origin by this node.
func WithDownloader(downloader httpclient.Fetcher) Option {
	return func(s *Service) { s.downloader = downloader }
}

func NewService(
	repos repo.Repositories,
	searcher search.Searcher,
	tagger tagging.Tagger,
	storageMap map[shared.StorageType]storage.Storage,
	sourceMap map[shared.SourceType]source.ArtworkSource,
	storCfg runtimecfg.StorageConfig,
	opts ...Option,
) *Service {
	s := &Service{
		repos:    repos,
		tagger:   tagger,
		searcher: searcher,
		storages: storageMap,
		sources:  sourceMap,
		storCfg:  storCfg,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Imseek returns the feature search engine (may be nil or disabled).
func (s *Service) Imseek() imseek.Engine { return s.imseek }

type serviceCtxKey struct{}

var contextKey = serviceCtxKey{}

func WithContext(ctx context.Context, serv *Service) context.Context {
	return context.WithValue(ctx, contextKey, serv)
}

func FromContext(ctx context.Context) *Service {
	if serv, ok := ctx.Value(contextKey).(*Service); ok {
		return serv
	}
	return nil
}

func MustFromContext(ctx context.Context) *Service {
	serv := FromContext(ctx)
	if serv == nil {
		panic(fmt.Sprintf("service: missing service in context (%T)", ctx))
	}
	return serv
}
