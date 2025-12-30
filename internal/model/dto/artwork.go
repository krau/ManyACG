package dto

import (
	"time"

	"github.com/krau/ManyACG/internal/shared"
	"github.com/unvgo/ouid"
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
	IDs []ouid.OUID `json:"ids"`
}

type ArtworkEventItem struct {
	// [TODO] other media types
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	Title          string             `json:"title"`
	Description    string             `json:"description"`
	SourceType     shared.SourceType  `json:"source_type"`
	SourceURL      string             `json:"source_url"`
	ArtistName     string             `json:"artist_name"`
	ArtistUsername string             `json:"artist_username"`
	ArtistUID      string             `json:"artist_uid"`
	Tags           []string           `json:"tags"`
	Pictures       []PictureEventItem `json:"pictures"`
	ID             ouid.OUID          `json:"id"`
	ArtistID       ouid.OUID          `json:"artist_id"`
	R18            bool               `json:"r18"`
}

type PictureEventItem struct {
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Thumbnail  string    `json:"thumbnail"`
	Original   string    `json:"original"`
	Phash      string    `json:"phash"`      // phash
	ThumbHash  string    `json:"thumb_hash"` // thumbhash
	OrderIndex uint      `json:"index"`
	Width      uint      `json:"width"`
	Height     uint      `json:"height"`
	ID         ouid.OUID `json:"id"`
	ArtworkID  ouid.OUID `json:"artwork_id"`
}
