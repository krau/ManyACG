package database

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/gorm"
)

func (d *DB) GetTagByName(ctx context.Context, name string) (*entity.Tag, error) {
	tag, err := gorm.G[entity.Tag](d.db).
		Where("name = ?", name).
		Preload("Alias", nil).
		First(ctx)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (d *DB) GetAliasTagByName(ctx context.Context, name string) (*entity.TagAlias, error) {
	alias, err := gorm.G[entity.TagAlias](d.db).
		Where("alias = ?", name).
		Preload("Tag", nil).
		First(ctx)
	if err != nil {
		return nil, err
	}
	return &alias, nil
}

func (d *DB) GetTagByID(ctx context.Context, id objectuuid.ObjectUUID) (*entity.Tag, error) {
	tag, err := gorm.G[entity.Tag](d.db).
		Where("id = ?", id).
		Preload("Alias", nil).
		First(ctx)
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

func (d *DB) UpdateTagAlias(ctx context.Context, id objectuuid.ObjectUUID, alias []*entity.TagAlias) error {
	var tag entity.Tag
	err := d.db.WithContext(ctx).Model(&entity.Tag{}).Where("id = ?", id).First(&tag).Error
	if err != nil {
		return err
	}
	return d.db.WithContext(ctx).Model(&tag).Association("Alias").Replace(alias)
}

func (d *DB) DeleteTagByID(ctx context.Context, id objectuuid.ObjectUUID) error {
	n, err := gorm.G[entity.Tag](d.db).Where("id = ?", id).Delete(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}