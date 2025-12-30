package entity

import (
	"time"

	"github.com/krau/ManyACG/internal/shared"
	"github.com/unvgo/ouid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var _ shared.ArtworkLike = (*CachedArtwork)(nil)
var _ shared.ArtworkLike = (*CachedArtworkData)(nil)

type CachedArtworkData struct {
	Artist      *CachedArtist     `json:"artist"`
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	SourceType  shared.SourceType `json:"source_type"`
	SourceURL   string            `json:"source_url"`

	Tags        []string            `json:"tags"`
	Pictures    []*CachedPicture    `json:"pictures"`
	UgoiraMetas []*CachedUgoiraMeta `json:"ugoira_metas,omitempty"`
	Videos      []*CachedVideo      `json:"videos,omitempty"`

	Version int  `json:"version"` // for future schema changes
	R18     bool `json:"r18"`
}

// FirstMedia implements [shared.ArtworkLike].
func (c *CachedArtworkData) FirstMedia() shared.MediaLike {
	if len(c.Pictures) > 0 {
		return c.Pictures[0]
	}
	if len(c.UgoiraMetas) > 0 {
		return c.UgoiraMetas[0]
	}
	if len(c.Videos) > 0 {
		return c.Videos[0]
	}
	return nil
}

// GetVideos implements [shared.ArtworkLike].
func (c *CachedArtworkData) GetVideos() []shared.VideoLike {
	var videos []shared.VideoLike
	for _, v := range c.Videos {
		videos = append(videos, v)
	}
	return videos
}

// MediasCount implements [shared.ArtworkLike].
func (c *CachedArtworkData) MediasCount() int {
	count := len(c.Pictures)
	count += len(c.Videos)
	count += len(c.UgoiraMetas)
	return count
}

func (c *CachedArtworkData) GetType() shared.SourceType {
	return c.SourceType
}

// GetUgoiraMetas implements shared.UgoiraArtworkLike.
func (c *CachedArtworkData) GetUgoiraMetas() []shared.UgoiraMetaLike {
	var metas []shared.UgoiraMetaLike
	for _, m := range c.UgoiraMetas {
		metas = append(metas, m)
	}
	return metas
}

var _ shared.UgoiraMetaLike = (*CachedUgoiraMeta)(nil)
var _ shared.PictureLike = (*CachedPicture)(nil)
var _ shared.VideoLike = (*CachedVideo)(nil)

type CachedVideo struct {
	TelegramInfo    shared.TelegramInfo  `json:"telegram_info"`
	OriginalStorage shared.StorageDetail `json:"original_storage"`
	ID              string               `json:"id"`
	ArtworkID       string               `json:"artwork_id"`
	Poster          string               `json:"poster"`
	URL             string               `json:"original"`
	MimeType        string               `json:"mime_type"`
	OrderIndex      uint                 `json:"index"`
	Width           uint                 `json:"width"`
	Height          uint                 `json:"height"`
	Duration        uint                 `json:"duration"` // in milliseconds
}

// GetDuration implements [shared.VideoLike].
func (c *CachedVideo) GetDuration() uint {
	return c.Duration
}

// GetHeight implements [shared.VideoLike].
func (c *CachedVideo) GetHeight() uint {
	return c.Height
}

// GetIndex implements [shared.VideoLike].
func (c *CachedVideo) GetIndex() uint {
	return c.OrderIndex
}

// GetMimeType implements [shared.VideoLike].
func (c *CachedVideo) GetMimeType() string {
	return c.MimeType
}

// GetPoster implements [shared.VideoLike].
func (c *CachedVideo) GetPoster() string {
	return c.Poster
}

// GetStorageInfo implements [shared.VideoLike].
func (c *CachedVideo) GetOriginalStorage() shared.StorageDetail {
	return c.OriginalStorage
}

// GetTelegramInfo implements [shared.VideoLike].
func (c *CachedVideo) GetTelegramInfo() shared.TelegramInfo {
	return c.TelegramInfo
}

// GetURL implements [shared.VideoLike].
func (c *CachedVideo) GetURL() string {
	return c.URL
}

// GetWidth implements [shared.VideoLike].
func (c *CachedVideo) GetWidth() uint {
	return c.Width
}

type CachedUgoiraMeta struct {
	TelegramInfo    shared.TelegramInfo   `json:"telegram_info"`
	OriginalStorage shared.StorageDetail  `json:"original_storage"`
	ID              string                `json:"id"`
	ArtworkID       string                `json:"artwork_id"`
	MetaData        shared.UgoiraMetaData `json:"data"`
	OrderIndex      uint                  `json:"index"`
}

// GetIndex implements shared.UgoiraMetaLike.
func (c *CachedUgoiraMeta) GetIndex() uint {
	return c.OrderIndex
}

// GetOriginalStorage implements shared.UgoiraMetaLike.
func (c *CachedUgoiraMeta) GetOriginalStorage() shared.StorageDetail {
	return c.OriginalStorage
}

// GetTelegramInfo implements shared.UgoiraMetaLike.
func (c *CachedUgoiraMeta) GetTelegramInfo() shared.TelegramInfo {
	return c.TelegramInfo
}

// GetUgoiraMetaData implements shared.UgoiraMetaLike.
func (c *CachedUgoiraMeta) GetUgoiraMetaData() shared.UgoiraMetaData {
	return c.MetaData
}

// GetArtist implements ArtworkLike.
func (c *CachedArtworkData) GetArtist() shared.ArtistLike {
	return c.Artist
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

// GetName implements shared.ArtistLike.
func (c *CachedArtist) GetName() string {
	return c.Name
}

// GetUID implements shared.ArtistLike.
func (c *CachedArtist) GetUID() string {
	return c.UID
}

// GetUserName implements shared.ArtistLike.
func (c *CachedArtist) GetUserName() string {
	return c.Username
}

type CachedPicture struct {
	StorageInfo  shared.StorageInfo  `json:"storage_info"`
	TelegramInfo shared.TelegramInfo `json:"telegram_info"`
	ID           string              `json:"id"`
	ArtworkID    string              `json:"artwork_id"`
	Thumbnail    string              `json:"thumbnail"`
	Original     string              `json:"original"`
	Phash        string              `json:"phash"`      // phash
	ThumbHash    string              `json:"thumb_hash"` // thumbhash

	OrderIndex uint `json:"index"`

	Width  uint `json:"width"`
	Height uint `json:"height"`
	Hidden bool `json:"hidden"` // 设为 true 时不发布到 Artwork 中, 但仍在其他接口中返回

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
	CreatedAt time.Time                              `gorm:"autoCreateTime" json:"created_at"`
	Artwork   datatypes.JSONType[*CachedArtworkData] `gorm:"type:json" json:"artwork"`
	SourceURL string                                 `gorm:"type:text;uniqueIndex" json:"source_url"`
	Status    shared.ArtworkStatus                   `gorm:"type:text;index" json:"status"`
	ID        ouid.OUID                              `gorm:"primaryKey;type:uuid" json:"id"`
}

// FirstMedia implements [shared.ArtworkLike].
func (c *CachedArtwork) FirstMedia() shared.MediaLike {
	return c.Artwork.Data().FirstMedia()
}

// GetVideos implements [shared.ArtworkLike].
func (c *CachedArtwork) GetVideos() []shared.VideoLike {
	return c.Artwork.Data().GetVideos()
}

// GetType implements shared.ArtworkLike.
func (c *CachedArtwork) GetType() shared.SourceType {
	return c.Artwork.Data().GetType()
}

// GetArtist implements ArtworkLike.
func (c *CachedArtwork) GetArtist() shared.ArtistLike {
	return c.Artwork.Data().GetArtist()
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

func (c *CachedArtwork) MediasCount() int {
	count := len(c.Artwork.Data().Pictures)
	count += len(c.Artwork.Data().Videos)
	count += len(c.Artwork.Data().UgoiraMetas)
	return count
}

// GetUgoiraMetas implements shared.UgoiraArtworkLike.
func (c *CachedArtwork) GetUgoiraMetas() []shared.UgoiraMetaLike {
	return c.Artwork.Data().GetUgoiraMetas()
}

func (c *CachedArtwork) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID.IsZero() {
		c.ID = ouid.New()
	}
	if c.Artwork.Data().ID == "" {
		data := c.Artwork.Data()
		data.ID = c.ID.Hex()
		c.Artwork = datatypes.NewJSONType(data)
	}
	return nil
}
