package service

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"gorm.io/datatypes"
)

func (s *Service) DeleteCachedArtworkByURL(ctx context.Context, sourceURL string) error {
	cachedArt, err := s.repos.CachedArtwork().GetCachedArtworkByURL(ctx, sourceURL)
	if err != nil {
		return err
	}
	return s.repos.CachedArtwork().DeleteCachedArtworkByID(ctx, cachedArt.ID)
}

func (s *Service) GetOrFetchCachedArtwork(ctx context.Context, sourceURL string) (*entity.CachedArtwork, error) {
	return nil, nil
}

func (s *Service) UpdateCachedArtworkStatusByURL(ctx context.Context, sourceURL string, status shared.ArtworkStatus) error {
	cachedArt, err := s.repos.CachedArtwork().GetCachedArtworkByURL(ctx, sourceURL)
	if err != nil {
		return err
	}
	_, err = s.repos.CachedArtwork().UpdateCachedArtworkStatusByID(ctx, cachedArt.ID, status)
	return err
}

func (s *Service) HideCachedArtworkPicture(ctx context.Context, cachedArt *entity.CachedArtwork, index int) error {
	if index < 0 || index >= len(cachedArt.GetPictures()) {
		return nil
	}
	panic("not implemented")
}

func (s *Service) GetCachedArtworkByURL(ctx context.Context, sourceURL string) (*entity.CachedArtwork, error) {
	return s.repos.CachedArtwork().GetCachedArtworkByURL(ctx, sourceURL)
}

func (s *Service) UpdateCachedArtwork(ctx context.Context, data *entity.CachedArtworkData) error {
	cachedArt, err := s.repos.CachedArtwork().GetCachedArtworkByURL(ctx, data.SourceURL)
	if err != nil {
		return err
	}
	cachedArt.Artwork = datatypes.NewJSONType(data)
	panic("not implemented")
}
