package dao

import (
	"context"

	"ManyACG-Bot/dao/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var pictureCollection *mongo.Collection

func CreatePicture(ctx context.Context, picture *model.PictureModel) (*mongo.InsertOneResult, error) {
	return pictureCollection.InsertOne(ctx, picture)
}

func CreatePictures(ctx context.Context, pictures []*model.PictureModel) (*mongo.InsertManyResult, error) {
	var docs []interface{}
	for _, picture := range pictures {
		docs = append(docs, picture)
	}
	return pictureCollection.InsertMany(ctx, docs)
}

func DeletePicturesByIDs(ctx context.Context, ids []primitive.ObjectID) (*mongo.DeleteResult, error) {
	return pictureCollection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": ids}})
}
