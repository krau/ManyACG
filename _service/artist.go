package service

import (
	"context"

	"github.com/krau/ManyACG/dao"
	"github.com/krau/ManyACG/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetArtistByID(ctx context.Context, artistID primitive.ObjectID) (*types.Artist, error) {
	artist, err := dao.GetArtistByID(ctx, artistID)
	if err != nil {
		return nil, err
	}
	return artist.ToArtist(), nil
}

func GetArtistByUID(ctx context.Context, uid string, sourceType types.SourceType) (*types.Artist, error) {
	artist, err := dao.GetArtistByUID(ctx, uid, sourceType)
	if err != nil {
		return nil, err
	}
	return artist.ToArtist(), nil
}
