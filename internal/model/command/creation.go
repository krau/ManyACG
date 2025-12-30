package command

import (
	"github.com/krau/ManyACG/internal/shared"
	"github.com/unvgo/ouid"
)

type ArtworkCreation struct {
	Artist      ArtworkArtistCreation
	Title       string
	Description string
	SourceType  shared.SourceType
	SourceURL   string
	Tags        []string
	Pictures    []ArtworkPictureCreation
	UgoiraMetas []*ArtworkUgoiraCreation
	Videos      []ArtworkVideoCreation
	ID          ouid.OUID
	R18         bool
}

type ArtworkArtistCreation struct {
	Name     string
	UID      string
	Username string
}

type ArtworkPictureCreation struct {
	StorageInfo  shared.StorageInfo
	TelegramInfo shared.TelegramInfo
	Thumbnail    string
	Original     string
	Phash        string
	ThumbHash    string
	Index        uint
	Width        uint
	Height       uint
}

type ArtworkUgoiraCreation struct {
	TelegramInfo    shared.TelegramInfo
	OriginalStorage shared.StorageDetail
	Data            shared.UgoiraMetaData
	Index           uint
}

type ArtworkVideoCreation struct {
	TelegramInfo    shared.TelegramInfo
	OriginalStorage shared.StorageDetail
	URL             string
	Poster          string
	MimeType        string
	Index           uint
	Width           uint
	Height          uint
	DurationMs      uint
}
