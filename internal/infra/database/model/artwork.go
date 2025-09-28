package model

import (
	"time"

	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/gorm"
)

type Artwork struct {
	// keep ObjectID as 24-hex string
	ID          objectuuid.ObjectUUID `gorm:"primaryKey;type:uuid" json:"id"`
	Title       string                `gorm:"type:text;not null;index:idx_artwork_title,sort:asc" json:"title"`
	Description string                `gorm:"type:text" json:"description"`
	R18         bool                  `gorm:"not null;default:false" json:"r18"`
	CreatedAt   time.Time             `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time             `gorm:"not null;autoUpdateTime" json:"updated_at"`
	SourceType  shared.SourceType     `gorm:"type:text;not null" json:"source_type"`
	SourceURL   string                `gorm:"type:text;not null;uniqueIndex" json:"source_url"`
	LikeCount   uint                  `gorm:"not null;default:0" json:"like_count"`

	ArtistID objectuuid.ObjectUUID `gorm:"type:uuid;index" json:"artist_id"`
	Artist   *Artist               `gorm:"foreignKey:ArtistID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"artist"`

	// many2many relationship with tags
	Tags []*Tag `gorm:"many2many:artwork_tags;constraint:OnDelete:CASCADE" json:"tags"`

	// one-to-many pictures
	Pictures []*Picture `gorm:"foreignKey:ArtworkID;constraint:OnDelete:CASCADE" json:"pictures"`
}

func (a *Artwork) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == objectuuid.Nil {
		a.ID = objectuuid.New()
	}
	return nil
}
