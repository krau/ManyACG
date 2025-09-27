package persist

import (
	"context"
	"fmt"

	"github.com/krau/ManyACG/internal/artwork/domain"
	"gorm.io/gorm"
)

// artworkDB impls the domain.ArtworkRepo interface
type artworkDB struct {
	db *gorm.DB
}

func NewArtworkDB(db *gorm.DB) *artworkDB {
	return &artworkDB{db: db}
}

func (r *artworkDB) Save(ctx context.Context, artwork *domain.Artwork) error {
	po := fromDomain(artwork)
	return r.db.WithContext(ctx).Save(po).Error
}

func (r *artworkDB) FindByID(ctx context.Context, id domain.ArtworkID) (*domain.Artwork, error) {
	var po Artwork
	if err := r.db.WithContext(ctx).Preload("Pictures").Preload("Tags").First(&po, "id = ?", id.Value()).Error; err != nil {
		return nil, fmt.Errorf("artworkDB.FindByID: %w", err)
	}
	return po.toDomain(), nil
}

func (r *artworkDB) FindBySourceURL(ctx context.Context, sourceURL string) (*domain.Artwork, error) {
	var po Artwork
	if err := r.db.WithContext(ctx).Preload("Pictures").Preload("Tags").First(&po, "source_url = ?", sourceURL).Error; err != nil {
		return nil, fmt.Errorf("artworkDB.FindBySourceURL: %w", err)
	}
	return po.toDomain(), nil
}
