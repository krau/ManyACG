package database

import (
	"context"

	"github.com/krau/ManyACG/internal/infra/database/model"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/gorm"
)

func (d *DB) CreateApiKey(ctx context.Context, apiKey *model.ApiKey) (*objectuuid.ObjectUUID, error) {
	result := gorm.WithResult()
	err := gorm.G[model.ApiKey](d.db, result).Create(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	return &apiKey.ID, nil
}

func (d *DB) GetApiKeyByKey(ctx context.Context, key string) (*model.ApiKey, error) {
	res, err := gorm.G[model.ApiKey](d.db).Where("key = ?", key).First(ctx)
	return &res, err
}

func (d *DB) IncreaseApiKeyUsed(ctx context.Context, key string) error {
	return d.db.WithContext(ctx).Model(&model.ApiKey{}).Where("key = ?", key).Update("used", gorm.Expr("used + ?", 1)).Error
}

func (d *DB) AddApiKeyQuota(ctx context.Context, key string, quota int) error {
	return d.db.WithContext(ctx).Model(&model.ApiKey{}).Where("key = ?", key).Update("quota", gorm.Expr("quota + ?", quota)).Error
}

func (d *DB) DeleteApiKey(ctx context.Context, key string) error {
	n, err := gorm.G[model.ApiKey](d.db).Where("key = ?", key).Delete(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
