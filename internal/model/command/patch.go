package command

import (
	"github.com/krau/ManyACG/internal/shared"
	"github.com/unvgo/ouid"
)

type ArtworkBasicPatch struct {
	ID          ouid.OUID `gorm:"-"`
	Title       *string
	Description *string
	R18         *shared.R18Type
}

type ArtistPatch struct {
	ID       ouid.OUID `gorm:"-"`
	Name     *string
	Username *string
	UID      *string
}
