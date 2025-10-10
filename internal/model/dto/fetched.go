package dto

import (
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

// IsHide implements shared.PictureLike.
func (f *FetchedPicture) IsHide() bool {
	return false
}

func (f *FetchedArtwork) GetID() string {
	return ""
}

// GetIndex implements shared.PictureLike.
func (f *FetchedPicture) GetIndex() uint {
	return f.Index
}

// GetArtistName implements shared.ArtworkLike.
func (f *FetchedArtwork) GetArtistName() string {
	return f.Artist.Name
}

// GetDescription implements shared.ArtworkLike.
func (f *FetchedArtwork) GetDescription() string {
	return f.Description
}

// GetPictures implements shared.ArtworkLike.
func (f *FetchedArtwork) GetPictures() []shared.PictureLike {
	var pictures []shared.PictureLike
	for _, pic := range f.Pictures {
		pictures = append(pictures, pic)
	}
	return pictures
}

// GetR18 implements shared.ArtworkLike.
func (f *FetchedArtwork) GetR18() bool {
	return f.R18
}

// GetSourceURL implements shared.ArtworkLike.
func (f *FetchedArtwork) GetSourceURL() string {
	return f.SourceURL
}

// GetTags implements shared.ArtworkLike.
func (f *FetchedArtwork) GetTags() []string {
	return f.Tags
}

// GetTitle implements shared.ArtworkLike.
func (f *FetchedArtwork) GetTitle() string {
	return f.Title
}

// GetOriginal implements shared.PictureLike.
func (f *FetchedPicture) GetOriginal() string {
	return f.Original
}

// GetSize implements shared.PictureLike.
func (f *FetchedPicture) GetSize() (width uint, height uint) {
	return f.Width, f.Height
}

// GetStorageInfo implements shared.PictureLike.
func (f *FetchedPicture) GetStorageInfo() shared.StorageInfo {
	return shared.StorageInfo{}
}

// GetTelegramInfo implements shared.PictureLike.
func (f *FetchedPicture) GetTelegramInfo() shared.TelegramInfo {
	return shared.TelegramInfo{}
}

// GetThumbnail implements shared.PictureLike.
func (f *FetchedPicture) GetThumbnail() string {
	return f.Thumbnail
}

var _ shared.ArtworkLike = (*FetchedArtwork)(nil)
var _ shared.PictureLike = (*FetchedPicture)(nil)
