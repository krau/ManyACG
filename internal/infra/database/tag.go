package database

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"gorm.io/gorm"
)

func (d *DB) GetTagByName(ctx context.Context, name string) (*entity.Tag, error) {
	tag, err := gorm.G[entity.Tag](d.db).Where("name = ?", name).First(ctx)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (d *DB) GetTagByNameWithAlias(ctx context.Context, name string) (*entity.Tag, error) {
	tag, err := d.GetTagByName(ctx, name)
	if err == nil {
		return tag, nil
	}
	alias, err := gorm.G[entity.TagAlias](d.db).Where("alias = ?", name).First(ctx)
	if err != nil {
		return nil, err
	}
	aliasTag, err := gorm.G[entity.Tag](d.db).Where("id = ?", alias.TagID).First(ctx)
	if err != nil {
		return nil, err
	}
	return &aliasTag, nil
}

func (d *DB) CreateTag(ctx context.Context, tag *entity.Tag) (*entity.Tag, error) {
	result := gorm.WithResult()
	err := gorm.G[entity.Tag](d.db, result).Create(ctx, tag)
	if err != nil {
		return nil, err
	}
	return tag, nil
}
