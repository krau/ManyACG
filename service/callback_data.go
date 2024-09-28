package service

import (
	"context"

	"github.com/krau/ManyACG/dao"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetCallbackDataByID(ctx context.Context, id string) (string, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return "", err
	}
	callbackData, err := dao.GetCallbackDataByID(ctx, objID)
	if err != nil {
		return "", err
	}
	data := callbackData.Data
	return data, nil
}

func CreateCallbackData(ctx context.Context, data string) (id string, err error) {
	callbackData, err := dao.CreateCallbackData(ctx, data)
	if err != nil {
		return
	}
	id = callbackData.ID.Hex()
	return
}
