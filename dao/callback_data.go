package dao

import (
	"context"
	"time"

	"github.com/krau/ManyACG/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var callbackDataCollection *mongo.Collection

func GetCallbackDataByID(ctx context.Context, id primitive.ObjectID) (*types.CallbackDataModel, error) {
	var callbackData types.CallbackDataModel
	err := callbackDataCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&callbackData)
	if err != nil {
		return nil, err
	}
	return &callbackData, nil
}

func CreateCallbackData(ctx context.Context, data string) (*types.CallbackDataModel, error) {
	callbackData := &types.CallbackDataModel{
		Data:      data,
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
	}
	res, err := callbackDataCollection.InsertOne(ctx, callbackData)
	if err != nil {
		return nil, err
	}
	callbackData.ID = res.InsertedID.(primitive.ObjectID)
	return callbackData, nil
}
