package dao

import (
	"context"
	"errors"

	"github.com/krau/ManyACG/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var apiKeyCollection *mongo.Collection

func CreateApiKey(ctx context.Context, apiKey *types.ApiKeyModel) (*mongo.InsertOneResult, error) {
	if apiKey.Key == "" || apiKey.Quota <= 0 {
		return nil, errors.New("invalid api key")
	}
	return apiKeyCollection.InsertOne(ctx, apiKey)
}

func GetApiKeyByKey(ctx context.Context, key string) (*types.ApiKeyModel, error) {
	var apiKey types.ApiKeyModel
	err := apiKeyCollection.FindOne(ctx, bson.M{"key": key}).Decode(&apiKey)
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func IncreaseApiKeyUsed(ctx context.Context, key string) error {
	_, err := apiKeyCollection.UpdateOne(ctx, bson.M{"key": key}, bson.M{"$inc": bson.M{"used": 1}})
	return err
}

func AddApiKeyQuota(ctx context.Context, key string, quota int) error {
	_, err := apiKeyCollection.UpdateOne(ctx, bson.M{"key": key}, bson.M{"$inc": bson.M{"quota": quota}})
	return err
}

func DeleteApiKey(ctx context.Context, key string) error {
	_, err := apiKeyCollection.DeleteOne(ctx, bson.M{"key": key})
	return err
}
