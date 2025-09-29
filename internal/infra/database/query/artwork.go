package query

import (
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type ArtworksQuery struct {
	R18      shared.R18Type
	Tags     [][]objectuuid.ObjectUUID
	ArtistID objectuuid.ObjectUUID
	Limit    int
	Offset   int
}
