package database

import (
	"context"
	"math/rand"
	"strings"

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
	var artwork entity.Artwork
	err := d.db.WithContext(ctx).Model(&entity.Artwork{}).
		Preload("Tags.Alias").
		Preload(clause.Associations).
		Where("id = ?", id).
		First(&artwork).Error
	if err != nil {
		return nil, err
	}
	return &artwork, nil
}

func (d *DB) GetArtworksByIDs(ctx context.Context, ids []objectuuid.ObjectUUID) ([]*entity.Artwork, error) {
	if len(ids) == 0 {
		return []*entity.Artwork{}, nil
	}
	var artworks []*entity.Artwork
	err := d.db.WithContext(ctx).Model(&entity.Artwork{}).
		Preload("Tags.Alias").
		Preload(clause.Associations).
		Where("id IN ?", ids).
		Find(&artworks).Error
	if err != nil {
		return nil, err
	}
	return artworks, nil
}

func (d *DB) GetArtworkByURL(ctx context.Context, url string) (*entity.Artwork, error) {
	var artwork entity.Artwork
	err := d.db.WithContext(ctx).Model(&entity.Artwork{}).
		Preload("Tags.Alias").
		Preload(clause.Associations).
		Where("source_url = ?", url).
		First(&artwork).Error
	if err != nil {
		return nil, err
	}
	return &artwork, nil
}

func (d *DB) QueryArtworks(ctx context.Context, que query.ArtworksDB) ([]*entity.Artwork, error) {
	query := d.db.WithContext(ctx).Model(&entity.Artwork{}).
		Preload("Tags.Alias").
		Preload(clause.Associations)

	if que.R18 != shared.R18TypeAll {
		query = query.Where("r18 = ?", que.R18 == shared.R18TypeR18)
	}
	if que.ArtistID != objectuuid.Nil {
		query = query.Where("artist_id = ?", que.ArtistID)
	}

	// 标签 IDs 筛选
	if len(que.Tags) > 0 {
		for _, orTags := range que.Tags {
			if len(orTags) == 0 {
				continue
			}
			subQuery := d.db.Table("artwork_tags").
				Select("DISTINCT artwork_id").
				Where("tag_id IN ?", orTags)
			query = query.Where("id IN (?)", subQuery)
		}
	} else if len(que.Keywords) > 0 {
		// 关键词 LIKE 筛选
		for _, orKeywords := range que.Keywords {
			if len(orKeywords) == 0 {
				continue
			}

			subQuery := d.db.Table("artworks AS a").
				Select("DISTINCT a.id").
				Joins("LEFT JOIN artists AS ar ON a.artist_id = ar.id").
				Joins("LEFT JOIN artwork_tags AS at ON a.id = at.artwork_id").
				Joins("LEFT JOIN tags AS t ON at.tag_id = t.id").
				Joins("LEFT JOIN tag_aliases AS ta ON t.id = ta.tag_id")

			var orExpr []string
			var orArgs []any
			for _, kw := range orKeywords {
				like := "%" + strings.ReplaceAll(strings.ReplaceAll(kw, "%", "\\%"), "_", "\\_") + "%"
				orExpr = append(orExpr, "(a.title LIKE ? OR a.description LIKE ? OR ar.name LIKE ? OR t.name LIKE ? OR ta.alias LIKE ?)")
				orArgs = append(orArgs, like, like, like, like, like)
			}

			subQuery = subQuery.Where(strings.Join(orExpr, " OR "), orArgs...)
			query = query.Where("id IN (?)", subQuery)
		}
	}

	if que.Limit > 0 {
		query = query.Limit(que.Limit)
	}

	if !que.Random {
		if que.Offset > 0 {
			query = query.Offset(que.Offset)
		}
		var artworks []*entity.Artwork
		err := query.Order("created_at DESC").Find(&artworks).Error
		if err != nil {
			return nil, err
		}
		return artworks, nil
	}

	var total int64
	countQuery := query.Session(&gorm.Session{})
	err := countQuery.Count(&total).Error
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var artworks []*entity.Artwork
	if int64(que.Limit) >= total {
		err := query.Find(&artworks).Error
		if err != nil {
			return nil, err
		}
		return artworks, nil
	}
	if total < 1000 {
		err = query.Order("RANDOM()").Limit(que.Limit).Find(&artworks).Error
		if err != nil {
			return nil, err
		}
		return artworks, nil
	}
	maxOffset := total - int64(que.Limit)
	randOffset := rand.Int63n(maxOffset + 1)
	err = query.Offset(int(randOffset)).Find(&artworks).Error
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

func (d *DB) DeleteArtworkByURL(ctx context.Context, url string) error {
	n, err := gorm.G[entity.Artwork](d.db).
		Where("source_url = ?", url).
		Delete(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateArtwork updates non-zero fields in the patch.
func (d *DB) UpdateArtwork(ctx context.Context, patch *entity.Artwork) error {
	if patch.ID == objectuuid.Nil {
		return gorm.ErrInvalidData
	}
	_, err := gorm.G[entity.Artwork](d.db).Where("id = ?", patch.ID).Updates(ctx, *patch)
	return err
}

// UpdateArtworkByMap updates all given fields in the patch map.
func (d *DB) UpdateArtworkByMap(ctx context.Context, id objectuuid.ObjectUUID, patch map[string]any) error {
	if id == objectuuid.Nil {
		return gorm.ErrInvalidData
	}
	if len(patch) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Model(&entity.Artwork{}).Where("id = ?", id).Updates(patch).Error
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

func (d *DB) CountArtworks(ctx context.Context, r18 shared.R18Type) (int64, error) {
	var count int64
	query := d.db.WithContext(ctx).Model(&entity.Artwork{})
	if r18 != shared.R18TypeAll {
		query = query.Where("r18 = ?", r18 == shared.R18TypeR18)
	}
	err := query.Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
