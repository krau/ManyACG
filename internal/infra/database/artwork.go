package database

import (
	"context"

	"github.com/krau/ManyACG/internal/domain/entity/artwork"
	"github.com/krau/ManyACG/internal/infra/database/po"
	"gorm.io/gorm"
)

type artworkDB struct {
	db *gorm.DB
}

func NewArtworkDB(db *gorm.DB) *artworkDB {
	return &artworkDB{db: db}
}

func (r *artworkDB) Save(ctx context.Context, artwork *artwork.Artwork) error {
	return r.db.WithContext(ctx).Save(po.ArtworkFromDomain(artwork)).Error
}
