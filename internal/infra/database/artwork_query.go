package database

import (
	"context"

	"github.com/krau/ManyACG/internal/app/query"
	"gorm.io/gorm"
)

type artworkQueryRepo struct {
	db *gorm.DB
}

// Count implements query.ArtworkQueryRepo.
func (a *artworkQueryRepo) Count(ctx context.Context, query query.ArtworkSearchQuery) (int, error) {
	panic("unimplemented")
}

// FindByID implements query.ArtworkQueryRepo.
func (a *artworkQueryRepo) FindByID(ctx context.Context, id string) (*query.ArtworkQueryResult, error) {
	panic("unimplemented")
}

// FindByURL implements query.ArtworkQueryRepo.
func (a *artworkQueryRepo) FindByURL(ctx context.Context, url string) (*query.ArtworkQueryResult, error) {
	panic("unimplemented")
}

// List implements query.ArtworkQueryRepo.
func (a *artworkQueryRepo) List(ctx context.Context, query query.ArtworkSearchQuery) (query.ArtworkSearchQueryResult, error) {
	panic("unimplemented")
}

var _ query.ArtworkQueryRepo = (*artworkQueryRepo)(nil)
