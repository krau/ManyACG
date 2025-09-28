package model

import (
	"time"

	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/gorm"
)

type DeletedRecord struct {
	ID        objectuuid.ObjectUUID `gorm:"primaryKey;type:uuid" json:"id"`
	ArtworkID objectuuid.ObjectUUID `gorm:"type:uuid;uniqueIndex" json:"artwork_id"`
	SourceURL string                `gorm:"type:text;not null;uniqueIndex" json:"source_url"`
	DeletedAt time.Time             `gorm:"not null;autoCreateTime" json:"deleted_at"`
}

func (d *DeletedRecord) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == objectuuid.Nil {
		d.ID = objectuuid.New()
	}
	return nil
}
