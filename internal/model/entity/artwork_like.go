package entity

type ArtworkLike interface {
	GetSourceURL() string
	GetTitle() string
	GetR18() bool
	GetArtistName() string
	GetDescription() string
	GetTags() []string
	GetPictures() []PictureLike
}

var _ ArtworkLike = (*Artwork)(nil)
var _ ArtworkLike = (*CachedArtworkData)(nil)
var _ ArtworkLike = (*CachedArtwork)(nil)
