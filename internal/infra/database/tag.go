package database

import (
	"context"
	"errors"

	"github.com/krau/ManyACG/internal/model/entity"
	"gorm.io/gorm"
)

func (d *DB) GetTagByNameWithAlias(ctx context.Context, name string) (*entity.Tag, error) {
	tag, err := gorm.G[entity.Tag](d.db).Where("name = ?", name).First(ctx)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err == nil {
		return &tag, nil
	}
	alias, err := gorm.G[entity.TagAlias](d.db).Where("alias = ?", name).First(ctx)
	if err != nil {
		return nil, err
	}
	tag, err = gorm.G[entity.Tag](d.db).Where("id = ?", alias.TagID).First(ctx)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (d *DB) CreateTag(ctx context.Context, tag *entity.Tag) (*entity.Tag, error) {
	result := gorm.WithResult()
	err := gorm.G[entity.Tag](d.db, result).Create(ctx, tag)
	if err != nil {
		return nil, err
	}
	return tag, nil
}
