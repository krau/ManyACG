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

type Artwork struct {
	ID          string            `json:"id" bson:"_id"`
	Title       string            `json:"title" bson:"title"`
	Description string            `json:"description" bson:"description"`
	R18         bool              `json:"r18" bson:"r18"`
	CreatedAt   time.Time         `json:"created_at" bson:"created_at"`
	SourceType  shared.SourceType `json:"source_type" bson:"source_type"`
	SourceURL   string            `json:"source_url" bson:"source_url"`
	LikeCount   uint              `json:"like_count" bson:"like_count"`

	Artist   *Artist    `json:"artist" bson:"artist"`
	Tags     []string   `json:"tags" bson:"tags"`
	Pictures []*Picture `json:"pictures" bson:"pictures"`
}

type Artist struct {
	ID       string            `json:"id" bson:"_id"`
	Name     string            `json:"name" bson:"name"`
	Type     shared.SourceType `json:"type" bson:"type"`
	UID      string            `json:"uid" bson:"uid"`
	Username string            `json:"username" bson:"username"`
}

type Picture struct {
	ID        string `json:"id" bson:"_id"`
	ArtworkID string `json:"artwork_id" bson:"artwork_id"`
	Index     uint   `json:"index" bson:"index"`
	Thumbnail string `json:"thumbnail" bson:"thumbnail"`
	Original  string `json:"original" bson:"original"`

	Width     uint   `json:"width" bson:"width"`
	Height    uint   `json:"height" bson:"height"`
	Hash      string `json:"hash" bson:"hash"`
	ThumbHash string `json:"thumb_hash" bson:"thumb_hash"`

	TelegramInfo *shared.TelegramInfo `json:"telegram_info" bson:"telegram_info"`
	StorageInfo  *shared.StorageInfo  `json:"storage_info" bson:"storage_info"`
}
