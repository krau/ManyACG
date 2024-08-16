package dao

import (
	"ManyACG/model"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var unauthUserCollection *mongo.Collection

func CreateUnauthUser(ctx context.Context, user *model.UnauthUserModel) (*mongo.InsertOneResult, error) {
	return unauthUserCollection.InsertOne(ctx, user)
}

func GetUnauthUserByID(ctx context.Context, id primitive.ObjectID) (*model.UnauthUserModel, error) {
	user := &model.UnauthUserModel{}
	err := unauthUserCollection.FindOne(ctx, bson.M{"_id": id}).Decode(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUnauthUserByUsername(ctx context.Context, username string) (*model.UnauthUserModel, error) {
	user := &model.UnauthUserModel{}
	err := unauthUserCollection.FindOne(ctx, bson.M{"username": username}).Decode(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func UpdateUnauthUser(ctx context.Context, id primitive.ObjectID, user *model.UnauthUserModel) (*mongo.UpdateResult, error) {
	return unauthUserCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": user})
}

func DeleteUnauthUser(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	return unauthUserCollection.DeleteOne(ctx, bson.M{"_id": id})
}
