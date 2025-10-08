package service

import (
	"context"

	"github.com/krau/ManyACG/internal/infra/cache"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

// golang please support generic methods :(

func (s *Service) GetStringDataByID(ctx context.Context, id string) (string, error) {
	_, err := objectuuid.FromObjectIDHex(id)
	if err != nil {
		return "", err
	}
	data, err := cache.Get[string](ctx, id)
	if err != nil {
		return "", err
	}
	return data, nil
}

func (s *Service) CreateStringData(ctx context.Context, data string) (id string, err error) {
	id = objectuuid.New().Hex()
	err = cache.Set(ctx, id, data)
	return
}
