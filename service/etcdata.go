package service

import (
	"ManyACG/dao"
	"context"
	"errors"

	. "ManyACG/logger"

	"go.mongodb.org/mongo-driver/mongo"
)

func GetEtcData(ctx context.Context, key string) any {
	result, err := dao.GetEtcData(ctx, key)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			Logger.Errorf("Error when getting etc data: %s", err)
		}
		return nil
	}
	return result["value"]
}

func SetEtcData(ctx context.Context, key string, value interface{}) (*mongo.UpdateResult, error) {
	return dao.SetEtcData(ctx, key, value)
}

func DeleteEtcData(ctx context.Context, key string) (*mongo.DeleteResult, error) {
	return dao.DeleteEtcData(ctx, key)
}
