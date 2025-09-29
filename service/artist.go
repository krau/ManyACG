package service

import (
	"context"

	"github.com/krau/ManyACG/internal/infra/database"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"github.com/krau/ManyACG/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetArtistByID(ctx context.Context, artistID primitive.ObjectID) (*types.Artist, error) {
	// artist, err := dao.GetArtistByID(ctx, artistID)
	// if err != nil {
	// 	return nil, err
	// }
	// return artist.ToArtist(), nil
	artist, err := database.Default().GetArtistByID(ctx, objectuuid.FromObjectID(objectuuid.ObjectID(artistID)))
	if err != nil {
		return nil, err
	}
	return &types.Artist{
		ID:       artist.ID.Hex(),
		Name:     artist.Name,
		UID:      artist.UID,
		Type:     types.SourceType(artist.Type),
		Username: artist.Username,
	}, nil
}

func GetArtistByUID(ctx context.Context, uid string, sourceType types.SourceType) (*types.Artist, error) {
	// artist, err := dao.GetArtistByUID(ctx, uid, sourceType)
	// if err != nil {
	// 	return nil, err
	// }
	// return artist.ToArtist(), nil
	artist, err := database.Default().GetArtistByUID(ctx, uid, shared.SourceType(sourceType))
	if err != nil {
		return nil, err
	}
	return &types.Artist{
		ID:       artist.ID.Hex(),
		Name:     artist.Name,
		UID:      artist.UID,
		Type:     types.SourceType(artist.Type),
		Username: artist.Username,
	}, nil
}
