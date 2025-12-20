package database

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/unvgo/ouid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// UpdateVideoTelegramInfoByID implements [repo.Video].
func (d *DB) UpdateVideoTelegramInfoByID(ctx context.Context, id ouid.OUID, tgInfo *shared.TelegramInfo) (*entity.Video, error) {
	video, err := d.GetVideoByID(ctx, id)
	if err != nil {
		return nil, err
	}
	video.TelegramInfo = datatypes.NewJSONType(*tgInfo)
	err = d.db.WithContext(ctx).Save(video).Error
	if err != nil {
		return nil, err
	}
	return video, nil
}

func (d *DB) GetVideoByID(ctx context.Context, id ouid.OUID) (*entity.Video, error) {
	video, err := gorm.G[entity.Video](d.db).Where("id = ?", id).First(ctx)
	if err != nil {
		return nil, err
	}
	return &video, nil
}
