package service

import (
	"context"
	"sync"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/search"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/infra/storage"
	"github.com/krau/ManyACG/internal/repo"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
)

type Service struct {
	repos    repo.Repositories
	searcher search.Searcher
	storages map[shared.StorageType]storage.Storage
	sources  map[shared.SourceType]source.ArtworkSource
	storCfg  runtimecfg.StorageConfig
}

type Option func(*Service)

func NewService(
	repos repo.Repositories,
	searcher search.Searcher,
	storageMap map[shared.StorageType]storage.Storage,
	sourceMap map[shared.SourceType]source.ArtworkSource,
	storCfg runtimecfg.StorageConfig,
	opts ...Option,
) *Service {
	s := &Service{
		repos:    repos,
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

var (
	defaultService *Service
	defaultOnce    sync.Once
)

func SetDefault(s *Service) { defaultOnce.Do(func() { defaultService = s }) }

func Default() *Service {
	if defaultService == nil {
		log.Fatal("service: Default service is not set")
	}
	return defaultService
}

type serviceCtxKey struct{}

var contextKey = serviceCtxKey{}

func WithContext(ctx context.Context, serv *Service) context.Context {
	return context.WithValue(ctx, contextKey, serv)
}

func FromContext(ctx context.Context) *Service {
	if serv, ok := ctx.Value(contextKey).(*Service); ok {
		return serv
	}

	return Default()
}

// func InitService(ctx context.Context) {
// 	go listenProcessPictureTask()
// 	// if config.Cfg.Search.Enable {
// 	// 	go syncArtworkToSearchEngine(ctx)
// 	// }
// 	if config.Cfg.Tagger.Enable {
// 		go listenPredictArtworkTagsTask()
// 	}
// }

// type Service struct{}

// func NewService() *Service {
// 	return &Service{}
// }

// func (s *Service) GetArtworkByURL(ctx context.Context, url string, opts ...*types.AdapterOption) (*types.Artwork, error) {
// 	return GetArtworkByURL(ctx, url, opts...)
// }
