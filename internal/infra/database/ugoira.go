package database

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// UpdateUgoiraTelegramInfoByID implements repo.Ugoira.
func (d *DB) UpdateUgoiraTelegramInfoByID(ctx context.Context, id objectuuid.ObjectUUID, tgInfo *shared.TelegramInfo) (*entity.UgoiraMeta, error) {
	ugo, err := d.GetUgoiraByID(ctx, id)
	if err != nil {
		return nil, err
	}
	ugo.TelegramInfo = datatypes.NewJSONType(*tgInfo)
	err = d.db.WithContext(ctx).Save(ugo).Error
	if err != nil {
		return nil, err
	}
	return ugo, nil
}

func (d *DB) GetUgoiraByID(ctx context.Context, id objectuuid.ObjectUUID) (*entity.UgoiraMeta, error) {
	ugo, err := gorm.G[entity.UgoiraMeta](d.db).Where("id = ?", id).First(ctx)
	if err != nil {
		return nil, err
	}
	return &ugo, nil
}
