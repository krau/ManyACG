package dao

import (
	"ManyACG/model"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	userCollection     *mongo.Collection
	likeCollection     *mongo.Collection
	favoriteCollection *mongo.Collection
)

func CreateUser(ctx context.Context, user *model.UserModel) (*mongo.InsertOneResult, error) {
	user.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	if user.Settings == nil {
		user.Settings = &model.UserSettings{}
	}
	return userCollection.InsertOne(ctx, user)
}

func GetUserByID(ctx context.Context, id primitive.ObjectID) (*model.UserModel, error) {
	user := &model.UserModel{}
	err := userCollection.FindOne(ctx, bson.M{"_id": id}).Decode(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByUsername(ctx context.Context, username string) (*model.UserModel, error) {
	user := &model.UserModel{}
	err := userCollection.FindOne(ctx, bson.M{"username": username}).Decode(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByTelegramID(ctx context.Context, telegramID int64) (*model.UserModel, error) {
	user := &model.UserModel{}
	err := userCollection.FindOne(ctx, bson.M{"telegram_id": telegramID}).Decode(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
