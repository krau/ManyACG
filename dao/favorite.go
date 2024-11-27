package dao

import (
	"context"

	"github.com/krau/ManyACG/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	favoriteCollection *mongo.Collection
)

func CreateFavorite(ctx context.Context, userID, artworkID primitive.ObjectID) (*mongo.InsertOneResult, error) {
	favorite := &types.FavoriteModel{
		UserID:    userID,
		ArtworkID: artworkID,
	}
	return favoriteCollection.InsertOne(ctx, favorite)
}

func GetFavorite(ctx context.Context, userID, artworkID primitive.ObjectID) (*types.FavoriteModel, error) {
	favorite := &types.FavoriteModel{}
	err := favoriteCollection.FindOne(ctx, bson.M{"user_id": userID, "artwork_id": artworkID}).Decode(favorite)
	if err != nil {
		return nil, err
	}
	return favorite, nil
}

func DeleteFavorite(ctx context.Context, userID, artworkID primitive.ObjectID) (*mongo.DeleteResult, error) {
	return favoriteCollection.DeleteOne(ctx, bson.M{"user_id": userID, "artwork_id": artworkID})
}
