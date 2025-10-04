package entity

import (
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/gorm"
)

type Tag struct {
	ID    objectuuid.ObjectUUID `gorm:"primaryKey;type:uuid" json:"id"`
	Name  string                `gorm:"type:text;not null;uniqueIndex" json:"name"`
	Alias []TagAlias            `gorm:"foreignKey:TagID;constraint:OnDelete:CASCADE" json:"alias"` // one-to-many relation

	// reverse relation via many2many
	Artworks []*Artwork `gorm:"many2many:artwork_tags" json:"-"`
}

type TagAlias struct {
	ID    objectuuid.ObjectUUID `gorm:"primaryKey;type:uuid" json:"id"`
	TagID objectuuid.ObjectUUID `gorm:"type:uuid;index" json:"tag_id"`
	Alias string                `gorm:"type:text;not null;index" json:"alias"`
}

func (t *Tag) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == objectuuid.Nil {
		t.ID = objectuuid.New()
	}
	return nil
}

func (ta *TagAlias) BeforeCreate(tx *gorm.DB) (err error) {
	if ta.ID == objectuuid.Nil {
		ta.ID = objectuuid.New()
	}
	return nil
}
