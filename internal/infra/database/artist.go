package database

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/gorm"
)

func (d *DB) GetArtistByID(ctx context.Context, id objectuuid.ObjectUUID) (*entity.Artist, error) {
	artist, err := gorm.G[entity.Artist](d.db).Where("id = ?", id).First(ctx)
	return &artist, err
}

func (d *DB) GetArtistByUID(ctx context.Context, uid string, source shared.SourceType) (*entity.Artist, error) {
	artist, err := gorm.G[entity.Artist](d.db).Where("uid = ? AND source = ?", uid, source).First(ctx)
	return &artist, err
}
