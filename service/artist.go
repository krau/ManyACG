package service

import (
	"ManyACG/dao"
	"ManyACG/types"
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetArtistByID(ctx context.Context, artistID primitive.ObjectID) (*types.Artist, error) {
	artist, err := dao.GetArtistByID(ctx, artistID)
	if err != nil {
		return nil, err
	}
	return artist.ToArtist(), nil
}
