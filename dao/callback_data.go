package dao

import (
	"context"
	"time"

	"github.com/krau/ManyACG/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var callbackDataCollection *mongo.Collection

func GetCallbackDataByID(ctx context.Context, id primitive.ObjectID) (*model.CallbackDataModel, error) {
	var callbackData model.CallbackDataModel
	err := callbackDataCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&callbackData)
	if err != nil {
		return nil, err
	}
	return &callbackData, nil
}

func CreateCallbackData(ctx context.Context, data string) (*model.CallbackDataModel, error) {
	callbackData := &model.CallbackDataModel{
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
