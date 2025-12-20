package query

import (
	"github.com/krau/ManyACG/internal/shared"
	"github.com/unvgo/ouid"
)

type Paginate struct {
	Limit  int
	Offset int
}

type ArtworksFilter struct {
	R18        shared.R18Type
	Tags       [][]ouid.OUID
	Keywords   [][]string
	ArtistID   ouid.OUID
	HasPicture bool
	HasVideo   bool
	HasUgoira  bool
}

// 只需要查数据库
type ArtworksDB struct {
	ArtworksFilter
	Paginate
	Random bool // 随机排序, 默认按 created_at 降序
}

// 需要其他设施
type ArtworkSearch struct {
	Query               string
	Hybrid              bool
	HybridSemanticRatio float64
	R18                 shared.R18Type
	// [TODO] more filters, need to migrate data
	Paginate
}

type ArtworkSimilar struct {
	ArtworkID ouid.OUID
	R18       shared.R18Type
	Paginate
}
