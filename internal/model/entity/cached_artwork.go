package entity

import (
	"time"

	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var _ shared.ArtworkLike = (*CachedArtwork)(nil)
var _ shared.ArtworkLike = (*CachedArtworkData)(nil)

type CachedArtworkData struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	R18         bool              `json:"r18"`
	SourceType  shared.SourceType `json:"source_type"`
	SourceURL   string            `json:"source_url"`

	Artist      *CachedArtist           `json:"artist"`
	Tags        []string                `json:"tags"`
	Pictures    []*CachedPicture        `json:"pictures"`
	UgoiraMetas []*CachedUgoiraMetaData `json:"ugoira_metas,omitempty"`

	Version int `json:"version"` // for future schema changes
}

// GetUgoiraMetas implements shared.UgoiraArtworkLike.
func (c *CachedArtworkData) GetUgoiraMetas() []shared.UgoiraMetaLike {
	var metas []shared.UgoiraMetaLike
	for _, m := range c.UgoiraMetas {
		metas = append(metas, m)
	}
	return metas
}

type CachedUgoiraMetaData struct {
	ID              string                `json:"id"`
	ArtworkID       string                `json:"artwork_id"`
	OrderIndex      uint                  `json:"index"`
	UgoiraMetaData  shared.UgoiraMetaData `json:"data"`
	OriginalStorage shared.StorageDetail  `json:"original_storage"`
	TelegramInfo    shared.TelegramInfo   `json:"telegram_info"`
}

// GetIndex implements shared.UgoiraMetaLike.
func (c *CachedUgoiraMetaData) GetIndex() uint {
	return c.OrderIndex
}

// GetOriginalStorage implements shared.UgoiraMetaLike.
func (c *CachedUgoiraMetaData) GetOriginalStorage() *shared.StorageDetail {
	return &c.OriginalStorage
}

// GetTelegramInfo implements shared.UgoiraMetaLike.
func (c *CachedUgoiraMetaData) GetTelegramInfo() shared.TelegramInfo {
	return c.TelegramInfo
}

// GetUgoiraMetaData implements shared.UgoiraMetaLike.
func (c *CachedUgoiraMetaData) GetUgoiraMetaData() shared.UgoiraMetaData {
	return c.UgoiraMetaData
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
func (c *CachedArtworkData) GetPictures() []shared.PictureLike {
	var pictures []shared.PictureLike
	for _, pic := range c.Pictures {
		pictures = append(pictures, pic)
	}
	return pictures
}

func (c *CachedArtworkData) GetViewablePictures() []*CachedPicture {
	var pictures []*CachedPicture
	for _, pic := range c.Pictures {
		if !pic.Hidden {
			pictures = append(pictures, pic)
		}
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

func (c *CachedArtworkData) GetID() string {
	return c.ID
}

type CachedArtist struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Type     shared.SourceType `json:"type"`
	UID      string            `json:"uid"`
	Username string            `json:"username"`
}

type CachedPicture struct {
	ID         string `json:"id"`
	ArtworkID  string `json:"artwork_id"`
	OrderIndex uint   `json:"index"`
	Thumbnail  string `json:"thumbnail"`
	Original   string `json:"original"`
	Hidden     bool   `json:"hidden"` // 设为 true 时不发布到 Artwork 中, 但仍在其他接口中返回

	Width     uint   `json:"width"`
	Height    uint   `json:"height"`
	Phash     string `json:"phash"`      // phash
	ThumbHash string `json:"thumb_hash"` // thumbhash

	StorageInfo  shared.StorageInfo  `json:"storage_info"`
	TelegramInfo shared.TelegramInfo `json:"telegram_info"`
}

// IsHide implements PictureLike.
func (c *CachedPicture) IsHide() bool {
	return c.Hidden
}

// GetIndex implements PictureLike.
func (c *CachedPicture) GetIndex() uint {
	return c.OrderIndex
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
func (c *CachedArtwork) GetPictures() []shared.PictureLike {
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

func (c *CachedArtwork) GetID() string {
	return c.ID.Hex()
}

func (c *CachedArtwork) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == objectuuid.Nil {
		c.ID = objectuuid.New()
	}
	if c.Artwork.Data().ID == "" {
		data := c.Artwork.Data()
		data.ID = c.ID.Hex()
		c.Artwork = datatypes.NewJSONType(data)
	}
	return nil
}
