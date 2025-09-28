package model

import (
	"time"

	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CachedArtworkData struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	R18         bool              `json:"r18"`
	SourceType  shared.SourceType `json:"source_type"`
	SourceURL   string            `json:"source_url"`

	Artist   *CachedArtist    `json:"artist"`
	Tags     []string         `json:"tags"`
	Pictures []*CachedPicture `json:"pictures"`

	Version int `json:"version"` // for future schema changes
}

type CachedArtist struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Type     shared.SourceType `json:"type"`
	UID      string            `json:"uid"`
	Username string            `json:"username"`
}

type CachedPicture struct {
	ID        string `json:"id"`
	ArtworkID string `json:"artwork_id"`
	Index     uint   `json:"index"`
	Thumbnail string `json:"thumbnail"`
	Original  string `json:"original"`
	Hidden    bool   `json:"hidden"` // don't post when true

	Width     uint   `json:"width"`
	Height    uint   `json:"height"`
	Phash     string `json:"phash"`      // phash
	ThumbHash string `json:"thumb_hash"` // thumbhash
}

type CachedArtwork struct {
	ID        objectuuid.ObjectUUID                  `gorm:"primaryKey;type:uuid" json:"id"`
	SourceURL string                                 `gorm:"type:text;uniqueIndex" json:"source_url"`
	CreatedAt time.Time                              `gorm:"autoCreateTime" json:"created_at"`
	Artwork   datatypes.JSONType[*CachedArtworkData] `gorm:"type:json" json:"artwork"`
	Status    shared.ArtworkStatus                   `gorm:"type:text;index" json:"status"`
}

func (c *CachedArtwork) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == objectuuid.Nil {
		c.ID = objectuuid.New()
	}
	return nil
}
