package service

import (
	"context"
	"errors"

	"github.com/krau/ManyACG/dao"
	manyacgErrors "github.com/krau/ManyACG/errors"
	"github.com/krau/ManyACG/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateLike(ctx context.Context, userID, artworkID primitive.ObjectID) error {
	likeModel, err := dao.GetLike(ctx, userID, artworkID)
	if errors.Is(err, mongo.ErrNoDocuments) {
		like := &model.LikeModel{
			UserID:    userID,
			ArtworkID: artworkID,
		}
		_, err := dao.CreateLike(ctx, like)
		if err != nil {
			return err
		}
		_, err = dao.IncrementArtworkLikeCountByID(ctx, artworkID)
		return err
	}
	if err != nil {
		return err
	}
	if likeModel != nil {
		return manyacgErrors.ErrLikeExists
	}
	return nil
}
