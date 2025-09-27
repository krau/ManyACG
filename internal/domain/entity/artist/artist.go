package artist

import (
	"github.com/krau/ManyACG/internal/common"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type Artist struct {
	ID       objectuuid.ObjectUUID
	Name     string
	Type     common.SourceType
	UID      string
	Username string
}

func NewArtist(id objectuuid.ObjectUUID, name string, sourceType common.SourceType, uid, username string) *Artist {
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
