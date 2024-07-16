package dao

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var etcDataCollection *mongo.Collection

func GetEtcData(ctx context.Context, key string) (bson.M, error) {
	var etcData bson.M
	err := etcDataCollection.FindOne(ctx, bson.M{"key": key}).Decode(&etcData)
	return etcData, err
}

func SetEtcData(ctx context.Context, key string, value any) (*mongo.UpdateResult, error) {
	return etcDataCollection.UpdateOne(ctx, bson.M{"key": key}, bson.M{"$set": bson.M{"value": value}}, options.Update().SetUpsert(true))
}

func DeleteEtcData(ctx context.Context, key string) (*mongo.DeleteResult, error) {
	return etcDataCollection.DeleteOne(ctx, bson.M{"key": key})
}
