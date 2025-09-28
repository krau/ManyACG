package service

import (
	"context"

	"github.com/krau/ManyACG/internal/infra/config"
	"github.com/krau/ManyACG/types"
)

func InitService(ctx context.Context) {
	go listenProcessPictureTask()
	if config.Get().Search.Enable {
		go syncArtworkToSearchEngine(ctx)
	}
	if config.Get().Tagger.Enable {
		go listenPredictArtworkTagsTask()
	}
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetArtworkByURL(ctx context.Context, url string, opts ...*types.AdapterOption) (*types.Artwork, error) {
	return GetArtworkByURL(ctx, url, opts...)
}
