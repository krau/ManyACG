package service

import (
	"context"
	"fmt"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/imsearch"
	"github.com/krau/ManyACG/internal/infra/search"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/infra/storage"
	"github.com/krau/ManyACG/internal/infra/tagging"
	"github.com/krau/ManyACG/internal/repo"
	"github.com/krau/ManyACG/internal/shared"
)

type Service struct {
	repos    repo.Repositories
	searcher search.Searcher
	tagger   tagging.Tagger
	storages map[shared.StorageType]storage.Storage
	sources  map[shared.SourceType]source.ArtworkSource
	storCfg  runtimecfg.StorageConfig
	imsearch imsearch.Engine
}

type Option func(*Service)

func WithImsearch(e imsearch.Engine) Option {
	return func(s *Service) { s.imsearch = e }
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

// Imsearch returns the feature search engine (may be nil or disabled).
func (s *Service) Imsearch() imsearch.Engine { return s.imsearch }

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
