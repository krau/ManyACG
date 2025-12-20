package entity

import (
	"github.com/krau/ManyACG/internal/shared"
	"github.com/unvgo/ouid"
	"gorm.io/gorm"
)

type Artist struct {
	ID       ouid.OUID         `gorm:"primaryKey;type:uuid" json:"id"`
	Name     string            `gorm:"type:text;not null;index" json:"name"`
	Type     shared.SourceType `gorm:"type:text;not null;index" json:"type"`
	UID      string            `gorm:"type:text;not null;index" json:"uid"`
	Username string            `gorm:"type:text;not null;index" json:"username"`

	// reverse relation
	Artworks []*Artwork `gorm:"foreignKey:ArtistID" json:"-"` // json ignore to avoid circular reference
}

// GetName implements shared.ArtistLike.
func (a *Artist) GetName() string {
	return a.Name
}

// GetUID implements shared.ArtistLike.
func (a *Artist) GetUID() string {
	return a.UID
}

// GetUserName implements shared.ArtistLike.
func (a *Artist) GetUserName() string {
	return a.Username
}

func (a *Artist) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID.IsZero() {
		a.ID = ouid.New()
	}
	return nil
}
