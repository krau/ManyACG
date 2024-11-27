package service

import (
	"context"

	"github.com/krau/ManyACG/types"
)

func InitService() {
	go listenProcessPictureTask()
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetArtworkByURL(ctx context.Context, url string, opts ...*types.AdapterOption) (*types.Artwork, error) {
	return GetArtworkByURL(ctx, url, opts...)
}
