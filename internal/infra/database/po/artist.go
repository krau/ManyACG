package po

import (
	"github.com/krau/ManyACG/internal/domain/entity/artist"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type Artist struct {
	ID       objectuuid.ObjectUUID `gorm:"primaryKey;type:uuid" json:"id"`
	Name     string                `gorm:"type:text;not null;index" json:"name"`
	Type     shared.SourceType     `gorm:"type:text;not null;index" json:"type"`
	UID      string                `gorm:"type:text;not null;index" json:"uid"`
	Username string                `gorm:"type:text;not null;index" json:"username"`

	// reverse relation
	Artworks []*Artwork `gorm:"foreignKey:ArtistID" json:"artworks"`
}

func (a *Artist) ToDomain() *artist.Artist {
	if a == nil {
		return nil
	}
	return &artist.Artist{
		ID:       a.ID,
		Name:     a.Name,
		Type:     a.Type,
		UID:      a.UID,
		Username: a.Username,
	}
}

func ArtistFromDomain(a *artist.Artist) *Artist {
	if a == nil {
		panic("why you passing nil artist")
	}
	return &Artist{
		ID:       a.ID,
		Name:     a.Name,
		Type:     a.Type,
		UID:      a.UID,
		Username: a.Username,
	}
}
