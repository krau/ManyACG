package artwork

import (
	"slices"

	"github.com/krau/ManyACG/internal/common"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type ArtworkID struct {
	value objectuuid.ObjectUUID
}

func (a ArtworkID) Value() objectuuid.ObjectUUID {
	return a.value
}

func (a ArtworkID) IsZero() bool {
	return a.value == objectuuid.Nil
}

func NewArtworkID(id objectuuid.ObjectUUID) ArtworkID {
	return ArtworkID{value: id}
}

type ArtistID struct {
	value objectuuid.ObjectUUID
}

func (a ArtistID) Value() objectuuid.ObjectUUID {
	return a.value
}

func (a ArtistID) IsZero() bool {
	return a.value == objectuuid.Nil
}

func NewArtistID(id objectuuid.ObjectUUID) ArtistID {
	return ArtistID{value: id}
}

type TagIDs = objectuuid.ObjectUUIDs

func NewTagIDs(ids ...objectuuid.ObjectUUID) *TagIDs {
	return objectuuid.NewObjectUUIDs(ids...)
}

type Picture struct {
	ID        objectuuid.ObjectUUID
	ArtworkID ArtworkID
	Index     uint // order index in artwork
	Thumbnail string
	Original  string
	Width     uint
	Height    uint
	Phash     string // phash
	ThumbHash string // thumbhash

	TelegramInfo *TelegramInfo
	StorageInfo  *StorageInfo
}

type TelegramInfo struct {
	PhotoFileID    string `json:"photo_file_id"`
	DocumentFileID string `json:"document_file_id"`
	MessageID      int    `json:"message_id"`
	MediaGroupID   string `json:"media_group_id"`
}

type StorageInfo struct {
	Original *StorageDetail `json:"original"`
	Regular  *StorageDetail `json:"regular"`
	Thumb    *StorageDetail `json:"thumb"`
}

type StorageDetail struct {
	Type common.StorageType `json:"type"`
	Path string             `json:"path"`
}

type Artwork struct {
	ID          ArtworkID
	Title       string
	Description string
	R18         bool
	SourceType  common.SourceType
	SourceURL   string
	LikeCount   uint

	ArtistID ArtistID
	TagIDs   *TagIDs
	Pictures []Picture
}

func (a *Artwork) AddTags(tagIDs ...objectuuid.ObjectUUID) {
	if a.TagIDs == nil {
		a.TagIDs = NewTagIDs()
	}
	a.TagIDs.Add(tagIDs...)
}

func (a *Artwork) RemoveTags(tagIDs ...objectuuid.ObjectUUID) {
	if a.TagIDs == nil {
		return
	}
	a.TagIDs.Remove(tagIDs...)
}

func (a *Artwork) RemovePicture(pictureID objectuuid.ObjectUUID) {
	if a.Pictures == nil {
		return
	}
	index := slices.IndexFunc(a.Pictures, func(p Picture) bool {
		return p.ID == pictureID
	})
	if index == -1 {
		return
	}
	a.Pictures = append(a.Pictures[:index], a.Pictures[index+1:]...)
	a.reorderPictures()
}

func (a *Artwork) reorderPictures() {
	slices.SortFunc(a.Pictures, func(i, j Picture) int {
		return int(i.Index) - int(j.Index)
	})
	for i := range a.Pictures {
		a.Pictures[i].Index = uint(i)
	}
}

func (a *Artwork) UpdateTitle(title string) {
	a.Title = title
}

func (a *Artwork) UpdateDescription(description string) {
	a.Description = description
}

func (a *Artwork) UpdateR18(r18 bool) {
	a.R18 = r18
}

func (a *Artwork) UpdatePictureSize(pictureID objectuuid.ObjectUUID, width, height uint) {
	for i := range a.Pictures {
		if a.Pictures[i].ID == pictureID {
			a.Pictures[i].Width = width
			a.Pictures[i].Height = height
			return
		}
	}
}

func (a *Artwork) UpdatePicturePhash(pictureID objectuuid.ObjectUUID, phash string) {
	for i := range a.Pictures {
		if a.Pictures[i].ID == pictureID {
			a.Pictures[i].Phash = phash
			return
		}
	}
}

func (a *Artwork) UpdatePictureThumbHash(pictureID objectuuid.ObjectUUID, thumbHash string) {
	for i := range a.Pictures {
		if a.Pictures[i].ID == pictureID {
			a.Pictures[i].ThumbHash = thumbHash
			return
		}
	}
}

func (a *Artwork) UpdatePicture(pictureID objectuuid.ObjectUUID, f func(p *Picture)) {
	for i := range a.Pictures {
		if a.Pictures[i].ID == pictureID {
			f(&a.Pictures[i])
			return
		}
	}
}
