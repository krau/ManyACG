package persist

import (
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/datatypes"
)

type Tag struct {
	ID    objectuuid.ObjectUUID       `gorm:"primaryKey;type:uuid" json:"id"`
	Name  string                      `gorm:"type:text;not null;uniqueIndex" json:"name"`
	Alias datatypes.JSONSlice[string] `gorm:"type:json" json:"alias"` // stores []string as JSON

	// reverse relation via many2many
	Artworks []*Artwork `gorm:"many2many:artwork_tags" json:"artworks"`
}
