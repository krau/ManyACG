package service

import (
	"context"

	"github.com/krau/ManyACG/dao"
	"github.com/krau/ManyACG/internal/common/errs"
	sources "github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/types"
)

func CreateCachedArtwork(ctx context.Context, artwork *types.Artwork, status types.ArtworkStatus) error {
	_, err := dao.CreateCachedArtwork(ctx, artwork, status)
	return err
}

func GetCachedArtworkByURL(ctx context.Context, sourceURL string) (*types.CachedArtworksModel, error) {
	cachedArtwork, err := dao.GetCachedArtworkByURL(ctx, sourceURL)
	if err != nil {
		return nil, err
	}
	return cachedArtwork, nil
}

func UpdateCachedArtworkStatusByURL(ctx context.Context, sourceURL string, status types.ArtworkStatus) error {
	_, err := dao.UpdateCachedArtworkStatusByURL(ctx, sourceURL, status)
	return err
}

func UpdateCachedArtwork(ctx context.Context, artwork *types.CachedArtworksModel) error {
	_, err := dao.UpdateCachedArtwork(ctx, artwork)
	return err
}

// GetCachedArtworkByURLWithCache get cached artwork by sourceURL, if not exist, fetch from source and cache it
func GetCachedArtworkByURLWithCache(ctx context.Context, sourceURL string) (*types.CachedArtworksModel, error) {
	cachedArtwork, err := dao.GetCachedArtworkByURL(ctx, sourceURL)
	if err != nil {
		artwork, err := sources.GetArtworkInfo(ctx, sourceURL)
		if err != nil {
			return nil, err
		}
		err = CreateCachedArtwork(ctx, artwork, types.ArtworkStatusCached)
		if err != nil {
			return nil, err
		}
		cachedArtwork, err = dao.GetCachedArtworkByURL(ctx, artwork.SourceURL)
		if err != nil {
			return nil, err
		}
	}
	return cachedArtwork, nil
}

func DeleteCachedArtworkByURL(ctx context.Context, sourceURL string) error {
	_, err := dao.DeleteCachedArtworkByURL(ctx, sourceURL)
	return err
}

func DeleteCachedArtworkPicture(ctx context.Context, cachedArtwork *types.CachedArtworksModel, pictureIndex int) error {
	if pictureIndex < 0 || pictureIndex > len(cachedArtwork.Artwork.Pictures) {
		return errs.ErrIndexOOB
	}
	cachedArtwork.Artwork.Pictures = append(cachedArtwork.Artwork.Pictures[:pictureIndex], cachedArtwork.Artwork.Pictures[pictureIndex+1:]...)
	for i := pictureIndex; i < len(cachedArtwork.Artwork.Pictures); i++ {
		cachedArtwork.Artwork.Pictures[i].Index = uint(i)
	}
	err := UpdateCachedArtwork(ctx, cachedArtwork)
	if err != nil {
		return err
	}
	return nil
}
