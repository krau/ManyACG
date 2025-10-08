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
	cached, err := s.repos.CachedArtwork().GetCachedArtworkByURL(ctx, sourceURL)
	if err == nil {
		return cached, nil
	}
	fetched, err := s.FetchArtworkInfo(ctx, sourceURL)
	if err != nil {
		return nil, err
	}
	pics := make([]*entity.CachedPicture, len(fetched.Pictures))
	for i, pic := range fetched.Pictures {
		pics[i] = &entity.CachedPicture{
			OrderIndex: pic.Index,
			Thumbnail:  pic.Thumbnail,
			Original:   pic.Original,
			Width:      pic.Width,
			Height:     pic.Height,
		}
	}
	ent := &entity.CachedArtwork{
		SourceURL: sourceURL,
		Status:    shared.ArtworkStatusCached,
		Artwork: datatypes.NewJSONType(&entity.CachedArtworkData{
			Title:       fetched.Title,
			Description: fetched.Description,
			R18:         fetched.R18,
			Tags:        fetched.Tags,
			SourceURL:   sourceURL,
			SourceType:  fetched.SourceType,
			Artist: &entity.CachedArtist{
				Name:     fetched.Artist.Name,
				UID:      fetched.Artist.UID,
				Type:     fetched.Artist.Type,
				Username: fetched.Artist.Username,
			},
			Pictures: pics,
			Version:  1,
		}),
	}
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
