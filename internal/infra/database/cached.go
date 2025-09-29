package database

import (
	"context"

	"github.com/krau/ManyACG/internal/infra/database/model"
	"gorm.io/gorm"
)

func (d *DB) GetCachedArtworkByURL(ctx context.Context, sourceUrl string) (*model.CachedArtwork, error) {
	res, err := gorm.G[model.CachedArtwork](d.db).Where("source_url = ?", sourceUrl).First(ctx)
	return &res, err
}
