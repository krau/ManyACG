package database

import (
	"context"

	"github.com/krau/ManyACG/internal/model/command"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (d *DB) CreateArtwork(ctx context.Context, artwork *entity.Artwork) (*objectuuid.ObjectUUID, error) {
	result := gorm.WithResult()
	err := gorm.G[entity.Artwork](d.db, result).Create(ctx, artwork)
	if err != nil {
		return nil, err
	}
	return &artwork.ID, nil
}

func (d *DB) GetArtworkByID(ctx context.Context, id objectuuid.ObjectUUID) (*entity.Artwork, error) {
	aw, err := gorm.G[entity.Artwork](d.db).Preload(clause.Associations, nil).Where("id = ?", id).First(ctx)
	if err != nil {
		return nil, err
	}
	return &aw, nil
}

func (d *DB) GetArtworksByIDs(ctx context.Context, ids []objectuuid.ObjectUUID) ([]*entity.Artwork, error) {
	if len(ids) == 0 {
		return []*entity.Artwork{}, nil
	}
	var artworks []*entity.Artwork
	err := d.db.WithContext(ctx).Model(&entity.Artwork{}).
		Preload(clause.Associations).
		Where("id IN ?", ids).
		Find(&artworks).Error
	if err != nil {
		return nil, err
	}
	return artworks, nil
}

func (d *DB) GetArtworkByURL(ctx context.Context, url string) (*entity.Artwork, error) {
	aw, err := gorm.G[entity.Artwork](d.db).Where("source_url = ?", url).First(ctx)
	if err != nil {
		return nil, err
	}
	return &aw, nil
}

func (d *DB) QueryArtworks(ctx context.Context, que query.ArtworksDB) ([]*entity.Artwork, error) {
	var artworks []*entity.Artwork

	query := d.db.WithContext(ctx).Model(&entity.Artwork{}).
		Preload("Artist").
		Preload("Tags").
		Preload("Tags.Alias").
		Preload("Pictures")

	if que.R18 != shared.R18TypeAll {
		query = query.Where("r18 = ?", que.R18 == shared.R18TypeR18)
	}
	if que.ArtistID != objectuuid.Nil {
		query = query.Where("artist_id = ?", que.ArtistID)
	}
	// [[or1,or2],[or3,or4]] means (or1 OR or2) AND (or3 OR or4)
	if len(que.Tags) > 0 {
		for _, orTags := range que.Tags {
			if len(orTags) == 0 {
				continue
			}
			subQuery := query.Table("artwork_tags").
				Select("DISTINCT artwork_id").
				Where("tag_id IN ?", orTags)
			query = query.Where("id IN (?)", subQuery)
		}
	}
	if que.Limit > 0 {
		query = query.Limit(que.Limit)
	}
	if que.Offset > 0 {
		query = query.Offset(que.Offset)
	}
	err := query.Order("created_at DESC").Find(&artworks).Error
	if err != nil {
		return nil, err
	}
	return artworks, nil
}

func (d *DB) DeleteArtworkByID(ctx context.Context, id objectuuid.ObjectUUID) error {
	n, err := gorm.G[entity.Artwork](d.db).
		Where("id = ?", id).
		Delete(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (d *DB) UpdateArtwork(ctx context.Context, patch command.ArtworkBasicPatch) error {
	if patch.ID == objectuuid.Nil {
		return gorm.ErrInvalidData
	}
	return d.db.WithContext(ctx).Model(&entity.Artwork{}).
		Where("id = ?", patch.ID).
		Updates(patch).Error
}

func (d *DB) UpdateArtworkPictures(ctx context.Context, id objectuuid.ObjectUUID, pics []*entity.Picture) error {
	if id == objectuuid.Nil {
		return gorm.ErrInvalidData
	}
	if len(pics) == 0 {
		return gorm.ErrInvalidData
	}
	for _, pic := range pics {
		pic.ArtworkID = id
	}
	var existing entity.Artwork
	if err := d.db.WithContext(ctx).First(&existing, id).Error; err != nil {
		return err
	}
	return d.db.WithContext(ctx).Model(&existing).Association("Pictures").Replace(pics)
}

func (d *DB) UpdateArtworkTags(ctx context.Context, id objectuuid.ObjectUUID, tags []*entity.Tag) error {
	if id == objectuuid.Nil {
		return gorm.ErrInvalidData
	}
	var existing entity.Artwork
	if err := d.db.WithContext(ctx).First(&existing, id).Error; err != nil {
		return err
	}
	return d.db.WithContext(ctx).Model(&existing).Association("Tags").Replace(tags)
}
