package database

import (
	"context"

	"github.com/krau/ManyACG/internal/infra/database/model"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/gorm"
)

func (d *DB) GetDeletedByURL(ctx context.Context, sourceURL string) (*model.DeletedRecord, error) {
	res, err := gorm.G[model.DeletedRecord](d.db).Where("source_url = ?", sourceURL).First(ctx)
	return &res, err
}

func (d *DB) CheckDeletedByURL(ctx context.Context, sourceURL string) bool {
	deleted, err := d.GetDeletedByURL(ctx, sourceURL)
	if err != nil {
		return false
	}
	return deleted != nil
}

func (d *DB) CreateDeleted(ctx context.Context, deleted *model.DeletedRecord) (*objectuuid.ObjectUUID, error) {
	result := gorm.WithResult()
	err := gorm.G[model.DeletedRecord](d.db, result).Create(ctx, deleted)
	if err != nil {
		return nil, err
	}
	return &deleted.ID, nil
}

func (d *DB) CancelDeletedByURL(ctx context.Context, sourceURL string) error {
	n, err := gorm.G[model.DeletedRecord](d.db).Where("source_url = ?", sourceURL).Delete(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
