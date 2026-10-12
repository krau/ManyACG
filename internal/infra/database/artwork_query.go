package database

import (
	"context"
	"math/rand/v2"
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
			db = db.Where("EXISTS (SELECT 1 FROM ugoira_meta u WHERE u.artwork_id = artworks.id)")
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

var keywordLikeEscaper = strings.NewReplacer("!", "!!", "%", "!%", "_", "!_")

var keywordIDBranches = [...]struct {
	query      string
	column     string
	pageDirect bool
}{
	{"SELECT id FROM artworks WHERE ", "artworks.title", true},
	{"SELECT id FROM artworks WHERE ", "artworks.description", true},
	{"SELECT a.id FROM artists ar JOIN artworks a ON a.artist_id = ar.id WHERE ", "ar.name", false},
	{"SELECT at.artwork_id FROM tags t JOIN artwork_tags at ON at.tag_id = t.id WHERE ", "t.name", false},
	{"SELECT at.artwork_id FROM tag_aliases ta JOIN tags t ON t.id = ta.tag_id JOIN artwork_tags at ON at.tag_id = t.id WHERE ", "ta.alias", false},
}

func buildKeywordGroupQuery(keywords []string, page bool) (string, []any) {
	var patterns []any
	for _, kw := range keywords {
		if kw == "" {
			continue
		}
		if patterns == nil {
			patterns = make([]any, 0, len(keywords))
		}
		patterns = append(patterns, "%"+keywordLikeEscaper.Replace(kw)+"%")
	}
	if len(patterns) == 0 {
		return "", nil
	}

	const inPrefix = "artworks.id IN ("
	const match = " LIKE ? ESCAPE '!'"
	prefix, separator := inPrefix, " UNION "
	if page {
		prefix, separator = "(", " OR "
	}
	size := len(prefix) + 1 + (len(keywordIDBranches)-1)*len(separator)
	for _, branch := range keywordIDBranches {
		size += len(patterns)*(len(branch.column)+len(match)) + (len(patterns)-1)*len(" OR ")
		if !page {
			size += len(branch.query)
		} else if !branch.pageDirect {
			size += len(inPrefix) + len(branch.query) + 1
		}
	}
	var sql strings.Builder
	sql.Grow(size)
	sql.WriteString(prefix)
	args := make([]any, 0, len(patterns)*len(keywordIDBranches))
	// 分页直接匹配正文以保留 LIMIT 短路；计数先收集各字段命中 ID。
	for i, branch := range keywordIDBranches {
		if i > 0 {
			sql.WriteString(separator)
		}
		direct := page && branch.pageDirect
		if page && !direct {
			sql.WriteString(inPrefix)
		}
		if !direct {
			sql.WriteString(branch.query)
		}
		for j := range patterns {
			if j > 0 {
				sql.WriteString(" OR ")
			}
			sql.WriteString(branch.column)
			sql.WriteString(match)
		}
		args = append(args, patterns...)
		if page && !direct {
			sql.WriteByte(')')
		}
	}
	sql.WriteByte(')')
	return sql.String(), args
}

func applyKeywordFilters(db *gorm.DB, groups [][]string, page bool) *gorm.DB {
	for _, keywords := range groups {
		groupSQL, args := buildKeywordGroupQuery(keywords, page)
		if groupSQL != "" {
			db = db.Where(groupSQL, args...)
		}
	}
	return db
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

const randomScanThreshold = 1000

func (d *DB) buildArtworkQuery(ctx context.Context, que query.ArtworksDB) *gorm.DB {
	base := d.db.WithContext(ctx).Model(&entity.Artwork{}).Scopes(applyBaseFilters(que))
	return applyKeywordFilters(base, que.Keywords, false)
}

func (d *DB) QueryArtworks(ctx context.Context, que query.ArtworksDB) ([]*entity.Artwork, error) {
	base := d.db.WithContext(ctx).Model(&entity.Artwork{}).Scopes(applyBaseFilters(que))
	baseQuery := applyKeywordFilters(base, que.Keywords, !que.Random)

	if !que.Random {
		// 非随机查询直接分页取数, 无需统计总数
		dataQuery := baseQuery.Scopes(applyArtworkPreloads())
		if que.Limit > 0 {
			dataQuery = dataQuery.Limit(que.Limit)
		}
		if que.Offset > 0 {
			dataQuery = dataQuery.Offset(que.Offset)
		}
		var artworks []*entity.Artwork
		if err := dataQuery.Order("created_at DESC").Find(&artworks).Error; err != nil {
			return nil, err
		}
		return artworks, nil
	}

	// 随机采样依赖总数: 结果集较小时直接随机排序, 较大时随机偏移取窗口
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

	var artworks []*entity.Artwork
	switch {
	case int64(que.Limit) >= total:
		if err := dataQuery.Order("created_at DESC").Find(&artworks).Error; err != nil {
			return nil, err
		}
	case total < randomScanThreshold:
		if err := dataQuery.Order("RANDOM()").Find(&artworks).Error; err != nil {
			return nil, err
		}
	default:
		// 随机偏移取连续的 Limit 条: 避免对大结果集全量排序, 代价为一次偏移扫描
		randOffset := rand.Int64N(total - int64(que.Limit) + 1)
		if err := dataQuery.Order("created_at DESC").Offset(int(randOffset)).Find(&artworks).Error; err != nil {
			return nil, err
		}
	}
	return artworks, nil
}
