package service

import (
	"context"

	"github.com/krau/ManyACG/internal/infra/cache"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

// golang please support generic methods :(

func (s *Service) GetStringDataByID(ctx context.Context, id string) (string, error) {
	// objID, err := primitive.ObjectIDFromHex(id)
	// if err != nil {
	// 	return "", err
	// }
	// callbackData, err := dao.GetCallbackDataByID(ctx, objID)
	// if err != nil {
	// 	return "", err
	// }
	// data := callbackData.Data
	// return data, nil
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
	// callbackData, err := dao.CreateCallbackData(ctx, data)
	// if err != nil {
	// 	return
	// }
	// id = callbackData.ID.Hex()
	// return
	id = objectuuid.New().Hex()
	err = cache.Set(ctx, id, data)
	return
}
