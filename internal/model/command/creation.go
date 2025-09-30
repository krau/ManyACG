package command

import "github.com/krau/ManyACG/internal/shared"

type ArtworkCreation struct {
	Title       string
	Description string
	R18         bool
	SourceType  shared.SourceType
	SourceURL   string

	Artist ArtworkArtistCreation

	Tags []string

	Pictures []ArtworkPictureCreation
}

type ArtworkArtistCreation struct {
	ID       string
	Name     string
	UID      string
	Username string
}

type ArtworkPictureCreation struct {
	Index     uint
	Thumbnail string
	Original  string

	Width     uint
	Height    uint
	Phash     string
	ThumbHash string

	TelegramInfo *shared.TelegramInfo
	StorageInfo  *shared.StorageInfo
}
