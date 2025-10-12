package service

import (
	"context"

	"github.com/krau/ManyACG/internal/model/converter"
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
	cached, err := s.repos.CachedArtwork().GetCachedArtworkByURL(ctx, sourceURL)
	if err == nil {
		return cached, nil
	}
	fetched, err := s.FetchArtworkInfo(ctx, sourceURL)
	if err != nil {
		return nil, err
	}
	cached, err = s.repos.CachedArtwork().GetCachedArtworkByURL(ctx, fetched.SourceURL)
	if err == nil {
		return cached, nil
	}
	ent := converter.DtoFetchedArtworkToEntityCached(fetched)
	created, err := s.repos.CachedArtwork().CreateCachedArtwork(ctx, ent)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *Service) UpdateCachedArtworkStatusByURL(ctx context.Context, sourceURL string, status shared.ArtworkStatus) error {
	cachedArt, err := s.repos.CachedArtwork().GetCachedArtworkByURL(ctx, sourceURL)
	if err != nil {
		return err
	}
	cachedArt.Status = status
	_, err = s.repos.CachedArtwork().SaveCachedArtwork(ctx, cachedArt)
	return err
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
	_, err = s.repos.CachedArtwork().SaveCachedArtwork(ctx, cachedArt)
	return err
}

func (s *Service) CreateCachedArtwork(ctx context.Context, ent *entity.CachedArtwork) (*entity.CachedArtwork, error) {
	return s.repos.CachedArtwork().CreateCachedArtwork(ctx, ent)
}

func (s *Service) HideCachedArtworkPicture(ctx context.Context, cachedArt *entity.CachedArtwork, picIndex int) error {
	data := cachedArt.Artwork.Data()
	for _, pic := range data.Pictures {
		if pic.OrderIndex == uint(picIndex) {
			pic.Hidden = true
		}
	}
	cachedArt.Artwork = datatypes.NewJSONType(data)
	_, err := s.repos.CachedArtwork().SaveCachedArtwork(ctx, cachedArt)
	return err
}
