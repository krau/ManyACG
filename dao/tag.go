package dao

import (
	"ManyACG/model"
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

func GetTagCount(ctx context.Context) (int64, error) {
	return tagCollection.CountDocuments(ctx, bson.M{})
}

func QueryTagsByName(ctx context.Context, name string) ([]*model.TagModel, error) {
	if name == "" {
		return nil, mongo.ErrNoDocuments
	}
	var tags []*model.TagModel
	cursor, err := tagCollection.Find(ctx, bson.M{"name": primitive.Regex{Pattern: name, Options: "i"}})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &tags); err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return tags, nil
}

func GetRandomTags(ctx context.Context, limit int) ([]*model.TagModel, error) {
	var tags []*model.TagModel
	cursor, err := tagCollection.Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$sample", Value: bson.D{{Key: "size", Value: limit}}}},
	})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &tags); err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return tags, nil
}
