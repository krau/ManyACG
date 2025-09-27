package artist

import (
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type Artist struct {
	ID       objectuuid.ObjectUUID
	Name     string
	Type     shared.SourceType
	UID      string
	Username string
}

func NewArtist(id objectuuid.ObjectUUID, name string, sourceType shared.SourceType, uid, username string) *Artist {
	return &Artist{
		ID:       id,
		Name:     name,
		Type:     sourceType,
		UID:      uid,
		Username: username,
	}
}

func (a *Artist) Update(name, username string) {
	a.Name = name
	a.Username = username
}
