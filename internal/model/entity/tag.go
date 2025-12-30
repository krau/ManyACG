package entity

import (
	"github.com/unvgo/ouid"
	"gorm.io/gorm"
)

type Tag struct {
	Name  string     `gorm:"type:text;not null;uniqueIndex" json:"name"`
	Alias []TagAlias `gorm:"foreignKey:TagID;constraint:OnDelete:CASCADE" json:"alias"` // one-to-many relation

	// reverse relation via many2many
	Artworks []*Artwork `gorm:"many2many:artwork_tags" json:"-"`
	ID       ouid.OUID  `gorm:"primaryKey;type:uuid" json:"id"`
}

type TagAlias struct {
	Alias string    `gorm:"type:text;not null;index" json:"alias"`
	ID    ouid.OUID `gorm:"primaryKey;type:uuid" json:"id"`
	TagID ouid.OUID `gorm:"type:uuid;index" json:"tag_id"`
}

func (t *Tag) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID.IsZero() {
		t.ID = ouid.New()
	}
	if t.Alias == nil {
		t.Alias = []TagAlias{}
	}
	return nil
}

func (ta *TagAlias) BeforeCreate(tx *gorm.DB) (err error) {
	if ta.ID.IsZero() {
		ta.ID = ouid.New()
	}
	return nil
}
