package database

import (
	"context"
	"errors"

	"github.com/krau/ManyACG/internal/domain/entity/tag"
	"github.com/krau/ManyACG/internal/domain/repo"
	"github.com/krau/ManyACG/internal/infra/database/po"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/gorm"
)

type tagRepo struct {
	db *gorm.DB
}

// FindByNameWithAlias implements repo.TagRepo.
func (t *tagRepo) FindByNameWithAlias(ctx context.Context, find string) (*tag.Tag, error) {
	// First, try to find by name
	poTag, err := gorm.G[po.Tag](t.db).Preload("Alias", nil).Where("name = ?", find).First(ctx)
	if err == nil {
		return poTag.ToDomain(), nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	// If not found, try to find by alias
	poTag = po.Tag{}
	err = t.db.WithContext(ctx).Joins("JOIN tag_aliases ON tag_aliases.tag_id = tags.id").
		Where("tag_aliases.alias = ?", find).First(&poTag).Error
	if err != nil {
		return nil, err
	}
	return poTag.ToDomain(), nil
}

// Save implements repo.TagRepo.
func (t *tagRepo) Save(ctx context.Context, tag *tag.Tag) error {
	poTag := po.Tag{
		ID:    tag.ID,
		Name:  tag.Name,
		Alias: make([]po.TagAlias, 0, len(tag.Alias)),
	}
	for _, a := range tag.Alias {
		poTag.Alias = append(poTag.Alias, po.TagAlias{
			ID:    objectuuid.New(),
			TagID: tag.ID,
			Alias: a,
		})
	}
	return t.db.WithContext(ctx).Save(&poTag).Error
}

func NewTagRepo(db *gorm.DB) repo.TagRepo {
	return &tagRepo{db: db}
}
