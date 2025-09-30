package service

import (
	"context"

	"github.com/krau/ManyACG/internal/infra/database"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

func GetArtistByID(ctx context.Context, artistID objectuuid.ObjectUUID) (*entity.Artist, error) {
	// artist, err := dao.GetArtistByID(ctx, artistID)
	// if err != nil {
	// 	return nil, err
	// }
	// return artist.ToArtist(), nil
	artist, err := database.Default().GetArtistByID(ctx, artistID)
	if err != nil {
		return nil, err
	}
	return artist, nil
}

func GetArtistByUID(ctx context.Context, uid string, sourceType shared.SourceType) (*entity.Artist, error) {
	// artist, err := dao.GetArtistByUID(ctx, uid, sourceType)
	// if err != nil {
	// 	return nil, err
	// }
	// return artist.ToArtist(), nil
	artist, err := database.Default().GetArtistByUID(ctx, uid, shared.SourceType(sourceType))
	if err != nil {
		return nil, err
	}
	return artist, nil
}
