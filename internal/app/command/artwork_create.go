package command

import (
	"github.com/krau/ManyACG/internal/common"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type ArtworkCreation struct {
	Title       string
	Description string
	R18         bool
	SourceType  common.SourceType
	SourceURL   string
	Artist      common.ArtistInfo
	TagNames    []string
	Pictures    []common.PictureInfo
}

type ArtworkCreationResult struct {
	ArtworkID  objectuuid.ObjectUUID
	ArtistID   objectuuid.ObjectUUID
	PictureIDs objectuuid.ObjectUUIDs
	TagIDs     objectuuid.ObjectUUIDs
}
