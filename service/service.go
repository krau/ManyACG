package service

import (
	"context"

	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/types"
)

func InitService() {
	go listenProcessPictureTask()
	if config.Cfg.Search.Enable {
		go syncArtworkToSearchEngine()
	}
	if config.Cfg.Tagger.Enable {
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
