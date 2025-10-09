package dto

import (
	"time"

	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type ArtworkSearchDocument struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Artist      string   `json:"artist"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	R18         bool     `json:"r18"`
}

type ArtworkSearchResult struct {
	IDs []objectuuid.ObjectUUID `json:"ids"`
}

type ArtworkEventItem struct {
	ID             objectuuid.ObjectUUID `json:"id"`
	Title          string                `json:"title"`
	Description    string                `json:"description"`
	R18            bool                  `json:"r18"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
	SourceType     shared.SourceType     `json:"source_type"`
	SourceURL      string                `json:"source_url"`
	ArtistID       objectuuid.ObjectUUID `json:"artist_id"`
	ArtistName     string                `json:"artist_name"`
	ArtistUsername string                `json:"artist_username"`
	ArtistUID      string                `json:"artist_uid"`
	Tags           []string              `json:"tags"`
	Pictures       []PictureEventItem    `json:"pictures"`
}

type PictureEventItem struct {
	ID         objectuuid.ObjectUUID `json:"id"`
	ArtworkID  objectuuid.ObjectUUID `json:"artwork_id"`
	OrderIndex uint                  `json:"index"`
	Thumbnail  string                `json:"thumbnail"`
	Original   string                `json:"original"`
	Width      uint                  `json:"width"`
	Height     uint                  `json:"height"`
	Phash      string                `json:"phash"`      // phash
	ThumbHash  string                `json:"thumb_hash"` // thumbhash
	CreatedAt  time.Time             `json:"created_at"`
	UpdatedAt  time.Time             `json:"updated_at"`
}
