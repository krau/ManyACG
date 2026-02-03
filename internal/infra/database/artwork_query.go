package database

import (
	"context"
	"math/rand"
	"strings"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/shared"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func applyBaseFilters(que query.ArtworksDB) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if que.R18 != shared.R18TypeAll {
			db = db.Where("r18 = ?", que.R18 == shared.R18TypeR18)
		}
		if !que.ArtistID.IsZero() {
			db = db.Where("artist_id = ?", que.ArtistID)
		}
		if que.HasPicture {
			db = db.Where("EXISTS (SELECT 1 FROM pictures p WHERE p.artwork_id = artworks.id)")
		}
		if que.HasVideo {
			db = db.Where("EXISTS (SELECT 1 FROM videos v WHERE v.artwork_id = artworks.id)")
		}
		if que.HasUgoira {
			db = db.Where("EXISTS (SELECT 1 FROM ugoira_metas u WHERE u.artwork_id = artworks.id)")
		}

		// Tags：Group AND，Group 内 OR（IDs）
		if len(que.Tags) > 0 {
			for _, orTags := range que.Tags {
				if len(orTags) > 0 {
					db = db.Where(`
						EXISTS (
							SELECT 1 FROM artwork_tags at
							WHERE at.artwork_id = artworks.id AND at.tag_id IN ?
						)
					`, orTags)
				}
			}
		}
		return db
	}
}

func buildKeywordArgs(kw string) []any {
	like := "%" + strings.ReplaceAll(strings.ReplaceAll(kw, "%", "\\%"), "_", "\\_") + "%"
	return []any{like, like, like, like, like}
}

type preloadOrder struct {
	Association string
	OrderBy     string
}

var artworkMediaPreloadOrders = []preloadOrder{
	{Association: "Pictures", OrderBy: "order_index ASC"},
	{Association: "UgoiraMetas", OrderBy: "order_index ASC"},
	{Association: "Videos", OrderBy: "order_index ASC"},
}

func applyArtworkPreloads() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db = db.Preload("Tags.Alias")
		for _, item := range artworkMediaPreloadOrders {
			assoc := item.Association
			orderBy := item.OrderBy
			db = db.Preload(assoc, func(db *gorm.DB) *gorm.DB {
				return db.Order(orderBy)
			})
		}
		return db.Preload(clause.Associations)
	}
}

func (d *DB) QueryArtworks(ctx context.Context, que query.ArtworksDB) ([]*entity.Artwork, error) {
	// 基础过滤
	baseQuery := d.db.WithContext(ctx).Model(&entity.Artwork{}).
		Scopes(applyBaseFilters(que))

	// Keyword 处理（Group AND 内部 OR）
	if len(que.Keywords) > 0 {
		keywordOrExpr := `
		artworks.title LIKE ? OR
		artworks.description LIKE ? OR
		EXISTS (
			SELECT 1 FROM artists ar
			WHERE ar.id = artworks.artist_id AND ar.name LIKE ?
		) OR
		EXISTS (
			SELECT 1 FROM artwork_tags at
			JOIN tags t ON t.id = at.tag_id
			LEFT JOIN tag_aliases ta ON ta.tag_id = t.id
			WHERE at.artwork_id = artworks.id
			  AND (t.name LIKE ? OR ta.alias LIKE ?)
		)
	`
		for _, orKeywords := range que.Keywords {
			if len(orKeywords) == 0 {
				continue
			}

			var orParts []string
			var args []any

			for _, kw := range orKeywords {
				orParts = append(orParts, "("+keywordOrExpr+")")
				args = append(args, buildKeywordArgs(kw)...)
			}

			groupSQL := "(" + strings.Join(orParts, " OR ") + ")"
			baseQuery = baseQuery.Where(groupSQL, args...)
		}
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, err
	}
	if total == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	dataQuery := baseQuery.Scopes(applyArtworkPreloads())

	if que.Limit > 0 {
		dataQuery = dataQuery.Limit(que.Limit)
	}

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

// func (d *DB) QueryArtworks(ctx context.Context, que query.ArtworksDB) ([]*entity.Artwork, error) {
// 	base := d.db.WithContext(ctx).Model(&entity.Artwork{})

// 	if que.R18 != shared.R18TypeAll {
// 		base = base.Where("r18 = ?", que.R18 == shared.R18TypeR18)
// 	}
// 	if !que.ArtistID.IsZero() {
// 		base = base.Where("artist_id = ?", que.ArtistID)
// 	}

// 	mainQueryExpr := "(" +
// 		"artworks.title LIKE ? OR artworks.description LIKE ? OR " +
// 		"EXISTS (SELECT 1 FROM artists ar WHERE ar.id = artworks.artist_id AND ar.name LIKE ?) OR " +
// 		"EXISTS (SELECT 1 FROM artwork_tags at JOIN tags t ON at.tag_id = t.id LEFT JOIN tag_aliases ta ON t.id = ta.tag_id WHERE at.artwork_id = artworks.id AND (t.name LIKE ? OR ta.alias LIKE ?))" +
// 		")"
// 	// Tags: each inner slice is OR (one-of), outer slice is AND (must satisfy each group)
// 	if len(que.Tags) > 0 {
// 		for _, orTags := range que.Tags {
// 			if len(orTags) == 0 {
// 				continue
// 			}
// 			base = base.Where("EXISTS (SELECT 1 FROM artwork_tags at WHERE at.artwork_id = artworks.id AND at.tag_id IN ?)", orTags)
// 		}
// 	} else if len(que.Keywords) > 0 {
// 		// Keywords: 组内 OR，组间 AND
// 		for _, orKeywords := range que.Keywords {
// 			if len(orKeywords) == 0 {
// 				continue
// 			}
// 			var perKWExpr []string
// 			var perKWArgs []any

// 			for _, kw := range orKeywords {
// 				like := "%" + strings.ReplaceAll(strings.ReplaceAll(kw, "%", "\\%"), "_", "\\_") + "%"

// 				perKWExpr = append(perKWExpr, mainQueryExpr)
// 				perKWArgs = append(perKWArgs, like, like, like, like, like)
// 			}

// 			// 将当前 group 的所有 keyword 表达式用 OR 拼接，整个 group 作为一个 WHERE 条件（与其他 group AND）
// 			groupSQL := strings.Join(perKWExpr, " OR ")
// 			base = base.Where(groupSQL, perKWArgs...)
// 		}
// 	}

// 	countQuery := base.Session(&gorm.Session{})
// 	var total int64
// 	if err := countQuery.Count(&total).Error; err != nil {
// 		return nil, err
// 	}
// 	if total == 0 {
// 		return nil, gorm.ErrRecordNotFound
// 	}

// 	dataQuery := d.db.WithContext(ctx).Model(&entity.Artwork{})

// 	if que.R18 != shared.R18TypeAll {
// 		dataQuery = dataQuery.Where("r18 = ?", que.R18 == shared.R18TypeR18)
// 	}
// 	if que.ArtistID != objectuuid.Nil {
// 		dataQuery = dataQuery.Where("artist_id = ?", que.ArtistID)
// 	}
// 	if len(que.Tags) > 0 {
// 		for _, orTags := range que.Tags {
// 			if len(orTags) == 0 {
// 				continue
// 			}
// 			dataQuery = dataQuery.Where("EXISTS (SELECT 1 FROM artwork_tags at WHERE at.artwork_id = artworks.id AND at.tag_id IN ?)", orTags)
// 		}
// 	} else if len(que.Keywords) > 0 {
// 		for _, orKeywords := range que.Keywords {
// 			if len(orKeywords) == 0 {
// 				continue
// 			}
// 			var perKWExpr []string
// 			var perKWArgs []any
// 			for _, kw := range orKeywords {
// 				like := "%" + strings.ReplaceAll(strings.ReplaceAll(kw, "%", "\\%"), "_", "\\_") + "%"
// 				perKWExpr = append(perKWExpr, mainQueryExpr)
// 				perKWArgs = append(perKWArgs, like, like, like, like, like)
// 			}
// 			groupSQL := strings.Join(perKWExpr, " OR ")
// 			dataQuery = dataQuery.Where(groupSQL, perKWArgs...)
// 		}
// 	}

// 	dataQuery = dataQuery.Preload("Tags.Alias").
// 		Preload("Pictures", func(db *gorm.DB) *gorm.DB {
// 			return db.Order("order_index ASC")
// 		}).Preload(clause.Associations)
// 	if que.Limit > 0 {
// 		dataQuery = dataQuery.Limit(que.Limit)
// 	}

// 	// 按时间排
// 	if !que.Random {
// 		if que.Offset > 0 {
// 			dataQuery = dataQuery.Offset(que.Offset)
// 		}
// 		var artworks []*entity.Artwork
// 		if err := dataQuery.Order("created_at DESC").Find(&artworks).Error; err != nil {
// 			return nil, err
// 		}
// 		return artworks, nil
// 	}

// 	var artworks []*entity.Artwork
// 	if int64(que.Limit) >= total {
// 		if err := dataQuery.Find(&artworks).Error; err != nil {
// 			return nil, err
// 		}
// 		return artworks, nil
// 	}

// 	// 总数<1000时直接 RANDOM()
// 	if total < 1000 {
// 		if err := dataQuery.Order("RANDOM()").Limit(que.Limit).Find(&artworks).Error; err != nil {
// 			return nil, err
// 		}
// 		return artworks, nil
// 	}
// 	maxOffset := total - int64(que.Limit)
// 	randOffset := rand.Int63n(maxOffset + 1)
// 	if err := dataQuery.Offset(int(randOffset)).Find(&artworks).Error; err != nil {
// 		return nil, err
// 	}
// 	return artworks, nil
// }
