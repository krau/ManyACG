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
	base := d.db.WithContext(ctx).Model(&entity.Artwork{})

	if que.R18 != shared.R18TypeAll {
		base = base.Where("r18 = ?", que.R18 == shared.R18TypeR18)
	}
	if que.ArtistID != objectuuid.Nil {
		base = base.Where("artist_id = ?", que.ArtistID)
	}

	// 3) Tags: each inner slice is OR (one-of), outer slice is AND (must satisfy each group)
	if len(que.Tags) > 0 {
		for _, orTags := range que.Tags {
			if len(orTags) == 0 {
				continue
			}
			base = base.Where("EXISTS (SELECT 1 FROM artwork_tags at WHERE at.artwork_id = artworks.id AND at.tag_id IN ?)", orTags)
		}
	} else if len(que.Keywords) > 0 {
		// Keywords: 组内 OR，组间 AND
		for _, orKeywords := range que.Keywords {
			if len(orKeywords) == 0 {
				continue
			}
			var perKWExpr []string
			var perKWArgs []any

			for _, kw := range orKeywords {
				like := "%" + strings.ReplaceAll(strings.ReplaceAll(kw, "%", "\\%"), "_", "\\_") + "%"

				expr := "(" +
					"artworks.title LIKE ? OR artworks.description LIKE ? OR " +
					"EXISTS (SELECT 1 FROM artists ar WHERE ar.id = artworks.artist_id AND ar.name LIKE ?) OR " +
					"EXISTS (SELECT 1 FROM artwork_tags at JOIN tags t ON at.tag_id = t.id LEFT JOIN tag_aliases ta ON t.id = ta.tag_id WHERE at.artwork_id = artworks.id AND (t.name LIKE ? OR ta.alias LIKE ?))" +
					")"

				perKWExpr = append(perKWExpr, expr)
				perKWArgs = append(perKWArgs, like, like, like, like, like)
			}

			// 将当前 group 的所有 keyword 表达式用 OR 拼接，整个 group 作为一个 WHERE 条件（与其他 group AND）
			groupSQL := strings.Join(perKWExpr, " OR ")
			base = base.Where(groupSQL, perKWArgs...)
		}
	}

	countQuery := base.Session(&gorm.Session{})
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, err
	}
	if total == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	dataQuery := d.db.WithContext(ctx).Model(&entity.Artwork{})

	if que.R18 != shared.R18TypeAll {
		dataQuery = dataQuery.Where("r18 = ?", que.R18 == shared.R18TypeR18)
	}
	if que.ArtistID != objectuuid.Nil {
		dataQuery = dataQuery.Where("artist_id = ?", que.ArtistID)
	}
	if len(que.Tags) > 0 {
		for _, orTags := range que.Tags {
			if len(orTags) == 0 {
				continue
			}
			dataQuery = dataQuery.Where("EXISTS (SELECT 1 FROM artwork_tags at WHERE at.artwork_id = artworks.id AND at.tag_id IN ?)", orTags)
		}
	} else if len(que.Keywords) > 0 {
		for _, orKeywords := range que.Keywords {
			if len(orKeywords) == 0 {
				continue
			}
			var perKWExpr []string
			var perKWArgs []any
			for _, kw := range orKeywords {
				like := "%" + strings.ReplaceAll(strings.ReplaceAll(kw, "%", "\\%"), "_", "\\_") + "%"
				expr := "(" +
					"artworks.title LIKE ? OR artworks.description LIKE ? OR " +
					"EXISTS (SELECT 1 FROM artists ar WHERE ar.id = artworks.artist_id AND ar.name LIKE ?) OR " +
					"EXISTS (SELECT 1 FROM artwork_tags at JOIN tags t ON at.tag_id = t.id LEFT JOIN tag_aliases ta ON t.id = ta.tag_id WHERE at.artwork_id = artworks.id AND (t.name LIKE ? OR ta.alias LIKE ?))" +
					")"
				perKWExpr = append(perKWExpr, expr)
				perKWArgs = append(perKWArgs, like, like, like, like, like)
			}
			groupSQL := strings.Join(perKWExpr, " OR ")
			dataQuery = dataQuery.Where(groupSQL, perKWArgs...)
		}
	}

	dataQuery = dataQuery.Preload("Tags.Alias").Preload(clause.Associations)
	if que.Limit > 0 {
		dataQuery = dataQuery.Limit(que.Limit)
	}

	// 按时间排
	if !que.Random {
		if que.Offset > 0 {
			dataQuery = dataQuery.Offset(que.Offset)
		}
		var artworks []*entity.Artwork
		if err := dataQuery.Order("created_at DESC").Find(&artworks).Error; err != nil {
			return nil, err
		}
		return artworks, nil
	}

	var artworks []*entity.Artwork
	if int64(que.Limit) >= total {
		if err := dataQuery.Find(&artworks).Error; err != nil {
			return nil, err
		}
		return artworks, nil
	}

	// 总数<1000时直接 RANDOM()
	if total < 1000 {
		if err := dataQuery.Order("RANDOM()").Limit(que.Limit).Find(&artworks).Error; err != nil {
			return nil, err
		}
		return artworks, nil
	}
	maxOffset := total - int64(que.Limit)
	randOffset := rand.Int63n(maxOffset + 1)
	if err := dataQuery.Offset(int(randOffset)).Find(&artworks).Error; err != nil {
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
