package entity

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

// GetArtistName implements ArtworkLike.
func (c *CachedArtworkData) GetArtistName() string {
	return c.Artist.Name
}

// GetDescription implements ArtworkLike.
func (c *CachedArtworkData) GetDescription() string {
	return c.Description
}

// GetPictures implements ArtworkLike.
func (c *CachedArtworkData) GetPictures() []PictureLike {
	var pictures []PictureLike
	for _, pic := range c.Pictures {
		pictures = append(pictures, pic)
	}
	return pictures
}

// GetSourceURL implements ArtworkLike.
func (c *CachedArtworkData) GetSourceURL() string {
	return c.SourceURL
}

// GetTags implements ArtworkLike.
func (c *CachedArtworkData) GetTags() []string {
	return c.Tags
}

// GetTitle implements ArtworkLike.
func (c *CachedArtworkData) GetTitle() string {
	return c.Title
}

// GetR18 implements ArtworkLike.
func (c *CachedArtworkData) GetR18() bool {
	return c.R18
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

	StorageInfo  shared.StorageInfo  `json:"storage_info"`
	TelegramInfo shared.TelegramInfo `json:"telegram_info"`
}

// GetOriginal implements PictureLike.
func (c *CachedPicture) GetOriginal() string {
	return c.Original
}

// GetSize implements PictureLike.
func (c *CachedPicture) GetSize() (width uint, height uint) {
	return c.Width, c.Height
}

// GetStorageInfo implements PictureLike.
func (c *CachedPicture) GetStorageInfo() shared.StorageInfo {
	return c.StorageInfo
}

// GetTelegramInfo implements PictureLike.
func (c *CachedPicture) GetTelegramInfo() shared.TelegramInfo {
	return c.TelegramInfo
}

// GetThumbnail implements PictureLike.
func (c *CachedPicture) GetThumbnail() string {
	return c.Thumbnail
}

type CachedArtwork struct {
	ID        objectuuid.ObjectUUID                  `gorm:"primaryKey;type:uuid" json:"id"`
	SourceURL string                                 `gorm:"type:text;uniqueIndex" json:"source_url"`
	CreatedAt time.Time                              `gorm:"autoCreateTime" json:"created_at"`
	Artwork   datatypes.JSONType[*CachedArtworkData] `gorm:"type:json" json:"artwork"`
	Status    shared.ArtworkStatus                   `gorm:"type:text;index" json:"status"`
}

// GetArtistName implements ArtworkLike.
func (c *CachedArtwork) GetArtistName() string {
	return c.Artwork.Data().GetArtistName()
}

// GetDescription implements ArtworkLike.
func (c *CachedArtwork) GetDescription() string {
	return c.Artwork.Data().GetDescription()
}

// GetPictures implements ArtworkLike.
func (c *CachedArtwork) GetPictures() []PictureLike {
	return c.Artwork.Data().GetPictures()
}

// GetSourceURL implements ArtworkLike.
func (c *CachedArtwork) GetSourceURL() string {
	return c.Artwork.Data().GetSourceURL()
}

// GetTags implements ArtworkLike.
func (c *CachedArtwork) GetTags() []string {
	return c.Artwork.Data().GetTags()
}

// GetTitle implements ArtworkLike.
func (c *CachedArtwork) GetTitle() string {
	return c.Artwork.Data().GetTitle()
}

func (c *CachedArtwork) GetR18() bool {
	return c.Artwork.Data().GetR18()
}

func (c *CachedArtwork) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == objectuuid.Nil {
		c.ID = objectuuid.New()
	}
	return nil
}
