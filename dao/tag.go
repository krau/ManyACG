package dao

import (
	"ManyACG-Bot/dao/model"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var tagCollection *mongo.Collection

func CreateTag(ctx context.Context, tag *model.TagModel) (*mongo.InsertOneResult, error) {
	return tagCollection.InsertOne(ctx, tag)
}

func CreateTags(ctx context.Context, tags []*model.TagModel) (*mongo.InsertManyResult, error) {
	var docs []interface{}
	for _, tag := range tags {
		docs = append(docs, tag)
	}
	return tagCollection.InsertMany(ctx, docs)
}

func GetTagByID(ctx context.Context, id primitive.ObjectID) (*model.TagModel, error) {
	var tag model.TagModel
	if err := tagCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&tag); err != nil {
		return nil, err
	}
	return &tag, nil
}

func GetTagByName(ctx context.Context, name string) (*model.TagModel, error) {
	var tag model.TagModel
	if err := tagCollection.FindOne(ctx, bson.M{"name": name}).Decode(&tag); err != nil {
		return nil, err
	}
	return &tag, nil
}
