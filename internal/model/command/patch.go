package command

import (
	"github.com/krau/ManyACG/internal/shared"
	"github.com/unvgo/ouid"
)

type ArtworkBasicPatch struct {
	Title       *string
	Description *string
	R18         *shared.R18Type
	ID          ouid.OUID `gorm:"-"`
}

type ArtistPatch struct {
	Name     *string
	Username *string
	UID      *string
	ID       ouid.OUID `gorm:"-"`
}
