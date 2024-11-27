package dao

import (
	"context"

	"github.com/krau/ManyACG/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	likeCollection *mongo.Collection
)

func CreateLike(ctx context.Context, like *types.LikeModel) (*mongo.InsertOneResult, error) {
	return likeCollection.InsertOne(ctx, like)
}

func GetLike(ctx context.Context, userID, artworkID primitive.ObjectID) (*types.LikeModel, error) {
	like := &types.LikeModel{}
	err := likeCollection.FindOne(ctx, bson.M{"user_id": userID, "artwork_id": artworkID}).Decode(like)
	if err != nil {
		return nil, err
	}
	return like, nil
}
