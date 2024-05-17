package service

import (
	"ManyACG/dao"
	"ManyACG/dao/model"
	"ManyACG/types"
	"context"
)

func CreateCachedArtwork(ctx context.Context, artwork *types.Artwork, status types.ArtworkStatus) error {
	_, err := dao.CreateCachedArtwork(ctx, artwork, status)
	return err
}

func GetCachedArtworkByURL(ctx context.Context, sourceURL string) (*model.CachedArtworksModel, error) {
	cachedArtwork, err := dao.GetCachedArtworkByURL(ctx, sourceURL)
	if err != nil {
		return nil, err
	}
	return cachedArtwork, nil
}

func UpdateCachedArtworkByURL(ctx context.Context, sourceURL string, status types.ArtworkStatus) error {
	_, err := dao.UpdateCachedArtworkByURL(ctx, sourceURL, status)
	return err
}
