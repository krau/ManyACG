package database

import (
	"context"

	"github.com/krau/ManyACG/internal/infra/database/model"
	"github.com/krau/ManyACG/internal/shared"
	"gorm.io/gorm"
)

func (d *DB) CreateCachedArtwork(ctx context.Context, data *model.CachedArtwork) (*model.CachedArtwork, error) {
	result := gorm.WithResult()
	err := gorm.G[model.CachedArtwork](d.db, result).Create(ctx, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *DB) GetCachedArtworkByURL(ctx context.Context, sourceUrl string) (*model.CachedArtwork, error) {
	res, err := gorm.G[model.CachedArtwork](d.db).Where("source_url = ?", sourceUrl).First(ctx)
	return &res, err
}

func (d *DB) UpdateCachedArtworkStatusByURL(ctx context.Context, sourceUrl string, status shared.ArtworkStatus) error {
	return d.db.WithContext(ctx).Model(&model.CachedArtwork{}).Where("source_url = ?", sourceUrl).Update("status", status).Error
}

func (d *DB) DeleteCachedArtworkByURL(ctx context.Context, sourceUrl string) error {
	n, err := gorm.G[model.CachedArtwork](d.db).Where("source_url = ?", sourceUrl).Delete(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
