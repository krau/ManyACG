package database

import (
	"context"

	"github.com/krau/ManyACG/internal/infra/database/model"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/gorm"
)

func (d *DB) GetArtistByID(ctx context.Context, id objectuuid.ObjectUUID) (*model.Artist, error) {
	artist, err := gorm.G[model.Artist](d.db).Where("id = ?", id).First(ctx)
	return &artist, err
}

func (d *DB) GetArtistByUID(ctx context.Context, uid string, source shared.SourceType) (*model.Artist, error) {
	artist, err := gorm.G[model.Artist](d.db).Where("uid = ? AND source = ?", uid, source).First(ctx)
	return &artist, err
}
