package dto

import (
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
)

type FetchedArtwork struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	R18         bool              `json:"r18"`
	SourceType  shared.SourceType `json:"source_type"`
	SourceURL   string            `json:"source_url"`

	Artist   *FetchedArtist    `json:"artist"`
	Tags     []string          `json:"tags"`
	Pictures []*FetchedPicture `json:"pictures"`
}

type FetchedArtist struct {
	Name     string            `json:"name"`
	Type     shared.SourceType `json:"type"`
	UID      string            `json:"uid"`
	Username string            `json:"username"`
}

type FetchedPicture struct {
	Index     uint   `json:"index"`
	Thumbnail string `json:"thumbnail"`
	Original  string `json:"original"`

	Width  uint `json:"width"`
	Height uint `json:"height"`
}

// IsHide implements entity.PictureLike.
func (f *FetchedPicture) IsHide() bool {
	return false
}

// GetIndex implements entity.PictureLike.
func (f *FetchedPicture) GetIndex() uint {
	return f.Index
}

// GetArtistName implements entity.ArtworkLike.
func (f *FetchedArtwork) GetArtistName() string {
	return f.Artist.Name
}

// GetDescription implements entity.ArtworkLike.
func (f *FetchedArtwork) GetDescription() string {
	return f.Description
}

// GetPictures implements entity.ArtworkLike.
func (f *FetchedArtwork) GetPictures() []entity.PictureLike {
	var pictures []entity.PictureLike
	for _, pic := range f.Pictures {
		pictures = append(pictures, pic)
	}
	return pictures
}

// GetR18 implements entity.ArtworkLike.
func (f *FetchedArtwork) GetR18() bool {
	return f.R18
}

// GetSourceURL implements entity.ArtworkLike.
func (f *FetchedArtwork) GetSourceURL() string {
	return f.SourceURL
}

// GetTags implements entity.ArtworkLike.
func (f *FetchedArtwork) GetTags() []string {
	return f.Tags
}

// GetTitle implements entity.ArtworkLike.
func (f *FetchedArtwork) GetTitle() string {
	return f.Title
}

// GetOriginal implements entity.PictureLike.
func (f *FetchedPicture) GetOriginal() string {
	return f.Original
}

// GetSize implements entity.PictureLike.
func (f *FetchedPicture) GetSize() (width uint, height uint) {
	return f.Width, f.Height
}

// GetStorageInfo implements entity.PictureLike.
func (f *FetchedPicture) GetStorageInfo() shared.StorageInfo {
	return shared.StorageInfo{}
}

// GetTelegramInfo implements entity.PictureLike.
func (f *FetchedPicture) GetTelegramInfo() shared.TelegramInfo {
	return shared.TelegramInfo{}
}

// GetThumbnail implements entity.PictureLike.
func (f *FetchedPicture) GetThumbnail() string {
	return f.Thumbnail
}

var _ entity.ArtworkLike = (*FetchedArtwork)(nil)
var _ entity.PictureLike = (*FetchedPicture)(nil)
