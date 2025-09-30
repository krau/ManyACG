package command

import (
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type ArtworkBasicPatch struct {
	ID          objectuuid.ObjectUUID `gorm:"-"`
	Title       *string
	Description *string
	R18         *shared.R18Type
}

type ArtistPatch struct {
	ID       objectuuid.ObjectUUID `gorm:"-"`
	Name     *string
	Username *string
	UID      *string
}
