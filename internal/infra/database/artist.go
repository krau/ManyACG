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
	if err != nil {
		return nil, err
	}
	return &artist, nil
}

func (d *DB) GetArtistByUID(ctx context.Context, uid string, source shared.SourceType) (*entity.Artist, error) {
	artist, err := gorm.G[entity.Artist](d.db).Where("uid = ? AND type = ?", uid, source).First(ctx)
	if err != nil {
		return nil, err
	}
	return &artist, nil
}

func (d *DB) UpdateArtist(ctx context.Context, patch *entity.Artist) error {
	_, err := gorm.G[entity.Artist](d.db).Where("id = ?", patch.ID).Updates(ctx, *patch)
	return err
}

func (d *DB) CreateArtist(ctx context.Context, artist *entity.Artist) (*objectuuid.ObjectUUID, error) {
	result := gorm.WithResult()
	err := gorm.G[entity.Artist](d.db, result).Create(ctx, artist)
	if err != nil {
		return nil, err
	}
	return &artist.ID, nil
}
