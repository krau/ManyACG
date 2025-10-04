package entity

import (
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/gorm"
)

type Artist struct {
	ID       objectuuid.ObjectUUID `gorm:"primaryKey;type:uuid" json:"id"`
	Name     string                `gorm:"type:text;not null;index" json:"name"`
	Type     shared.SourceType     `gorm:"type:text;not null;index" json:"type"`
	UID      string                `gorm:"type:text;not null;index" json:"uid"`
	Username string                `gorm:"type:text;not null;index" json:"username"`

	// reverse relation
	Artworks []*Artwork `gorm:"foreignKey:ArtistID" json:"-"` // json ignore to avoid circular reference
}

func (a *Artist) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == objectuuid.Nil {
		a.ID = objectuuid.New()
	}
	return nil
}
