package command

import (
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type ArtworkBasicPatch struct {
	ID          objectuuid.ObjectUUID
	Title       *string
	Description *string
	R18         *shared.R18Type
}
