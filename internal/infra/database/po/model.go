package po

import (
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/datatypes"
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

type Tag struct {
	ID    objectuuid.ObjectUUID       `gorm:"primaryKey;type:uuid" json:"id"`
	Name  string                      `gorm:"type:text;not null;uniqueIndex" json:"name"`
	Alias datatypes.JSONSlice[string] `gorm:"type:json" json:"alias"` // stores []string as JSON

	// reverse relation via many2many
	Artworks []*Artwork `gorm:"many2many:artwork_tags" json:"artworks"`
}
