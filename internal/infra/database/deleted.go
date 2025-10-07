package database

import (
	"context"
	"errors"

	"github.com/krau/ManyACG/internal/model/entity"
	"gorm.io/gorm"
)

func (d *DB) GetDeletedByURL(ctx context.Context, sourceURL string) (*entity.DeletedRecord, error) {
	res, err := gorm.G[entity.DeletedRecord](d.db).Where("source_url = ?", sourceURL).First(ctx)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (d *DB) CheckDeletedByURL(ctx context.Context, sourceURL string) bool {
	deleted, err := d.GetDeletedByURL(ctx, sourceURL)
	if err != nil {
		return false
	}
	return deleted != nil
}

func (d *DB) CreateDeletedRecord(ctx context.Context, deleted *entity.DeletedRecord) error {
	result := gorm.WithResult()
	err := gorm.G[entity.DeletedRecord](d.db, result).Create(ctx, deleted)
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) DeleteDeletedByURL(ctx context.Context, sourceURL string) error {
	n, err := gorm.G[entity.DeletedRecord](d.db).Where("source_url = ?", sourceURL).Delete(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if n == 0 {
		return nil
	}
	return nil
}
