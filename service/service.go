package service

import (
	"github.com/krau/ManyACG/internal/infra/search"
	"github.com/krau/ManyACG/internal/repo"
)

type Service struct {
	repos    repo.Repositories
	searcher search.Searcher
}

type Option func(*Service)

func NewService(
	repos repo.Repositories,
	searcher search.Searcher,
	opts ...Option,
) *Service {
	s := &Service{
		repos:    repos,
		searcher: searcher,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
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
