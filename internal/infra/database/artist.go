package database

import (
	"context"

	"github.com/krau/ManyACG/internal/domain/entity/artist"
	"github.com/krau/ManyACG/internal/domain/repo"
	"github.com/krau/ManyACG/internal/infra/database/po"
	"gorm.io/gorm"
)

type artistRepo struct {
	db *gorm.DB
}

// FindBySourceAndUID implements repo.ArtistRepo.
func (a *artistRepo) FindBySourceAndUID(ctx context.Context, sourceType string, uid string) (*artist.Artist, error) {
	artist, err := gorm.G[po.Artist](a.db).Where("source_type = ? AND source_uid = ?", sourceType, uid).First(ctx)
	if err != nil {
		return nil, err
	}
	return artist.ToDomain(), nil
}

// Save implements repo.ArtistRepo.
func (a *artistRepo) Save(ctx context.Context, artist *artist.Artist) error {
	return a.db.WithContext(ctx).Save(po.ArtistFromDomain(artist)).Error
}

func NewArtistRepo(db *gorm.DB) repo.ArtistRepo {
	return &artistRepo{db: db}
}
