package entity

import (
	"time"

	"github.com/unvgo/ouid"
	"gorm.io/gorm"
)

type DeletedRecord struct {
	DeletedAt time.Time `gorm:"not null;autoCreateTime" json:"deleted_at"`
	SourceURL string    `gorm:"type:text;not null;uniqueIndex" json:"source_url"`
	ID        ouid.OUID `gorm:"primaryKey;type:uuid" json:"id"`
}

func (d *DeletedRecord) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID.IsZero() {
		d.ID = ouid.New()
	}
	return nil
}
