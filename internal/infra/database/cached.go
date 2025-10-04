package database

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/gorm"
)

func (d *DB) CreateCachedArtwork(ctx context.Context, data *entity.CachedArtwork) (*entity.CachedArtwork, error) {
	result := gorm.WithResult()
	err := gorm.G[entity.CachedArtwork](d.db, result).Create(ctx, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *DB) GetCachedArtworkByURL(ctx context.Context, sourceUrl string) (*entity.CachedArtwork, error) {
	res, err := gorm.G[entity.CachedArtwork](d.db).Where("source_url = ?", sourceUrl).First(ctx)
	return &res, err
}

func (d *DB) UpdateCachedArtworkStatusByURL(ctx context.Context, sourceUrl string, status shared.ArtworkStatus) error {
	return d.db.WithContext(ctx).Model(&entity.CachedArtwork{}).Where("source_url = ?", sourceUrl).Update("status", status).Error
}

func (d *DB) DeleteCachedArtworkByURL(ctx context.Context, sourceUrl string) error {
	n, err := gorm.G[entity.CachedArtwork](d.db).Where("source_url = ?", sourceUrl).Delete(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (d *DB) DeleteCachedArtworkByID(ctx context.Context, id objectuuid.ObjectUUID) error {
	n, err := gorm.G[entity.CachedArtwork](d.db).Where("id = ?", id).Delete(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (d *DB) SaveCachedArtwork(ctx context.Context, artwork *entity.CachedArtwork) (*entity.CachedArtwork, error) {
	err := d.db.WithContext(ctx).Save(artwork).Error
	if err != nil {
		return nil, err
	}
	return artwork, nil
}

func (d *DB) ResetPostingCachedArtworkStatus(ctx context.Context) error {
	return d.db.WithContext(ctx).Model(&entity.CachedArtwork{}).Where("status = ?", shared.ArtworkStatusPosting).Update("status", shared.ArtworkStatusCached).Error
}

// GetCachedArtworkByID implements repo.CachedArtwork.
func (d *DB) GetCachedArtworkByID(ctx context.Context, id objectuuid.ObjectUUID) (*entity.CachedArtwork, error) {
	panic("unimplemented")
}

// UpdateCachedArtworkStatusByID implements repo.CachedArtwork.
func (d *DB) UpdateCachedArtworkStatusByID(ctx context.Context, id objectuuid.ObjectUUID, status shared.ArtworkStatus) (*entity.CachedArtwork, error) {
	panic("unimplemented")
}