package service

import (
	"context"

	"github.com/krau/ManyACG/dao"
	"github.com/krau/ManyACG/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateFavorite(ctx context.Context, userID, artworkID primitive.ObjectID) (*model.FavoriteModel, error) {
	res, err := dao.CreateFavorite(ctx, userID, artworkID)
	if err != nil {
		return nil, err
	}
	return &model.FavoriteModel{
		ID:        res.InsertedID.(primitive.ObjectID),
		UserID:    userID,
		ArtworkID: artworkID,
	}, nil
}

func GetFavorite(ctx context.Context, userID, artworkID primitive.ObjectID) (*model.FavoriteModel, error) {
	return dao.GetFavorite(ctx, userID, artworkID)
}

func DeleteFavorite(ctx context.Context, userID, artworkID primitive.ObjectID) error {
	_, err := dao.DeleteFavorite(ctx, userID, artworkID)
	return err
}
