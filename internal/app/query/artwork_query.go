package query

import (
	"github.com/krau/ManyACG/internal/shared"
)

// query a single artwork
type ArtworkQuery struct {
	ID        string
	SourceURL string
}

type ArtistInfo struct {
	shared.ArtistInfo
	ID string `json:"id"`
}

type PictureInfo struct {
	ID        string `json:"id"`
	Width     uint   `json:"width"`
	Height    uint   `json:"height"`
	Index     uint   `json:"index"`
	Phash     string `json:"phash"`
	ThumbHash string `json:"thumb_hash"`
	FileName  string `json:"file_name"`
	Thumbnail string `json:"thumbnail"`
	Regular   string `json:"regular"`
	MessageID int    `json:"message_id"`
}

type ArtworkQueryResult struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	R18         bool              `json:"r18"`
	SourceType  shared.SourceType `json:"source_type"`
	SourceURL   string            `json:"source_url"`
	Artist      *ArtistInfo       `json:"artist"`
	Tags        []string          `json:"tags"`
	CreatedAt   string            `json:"created_at"`
	Pictures    []*PictureInfo    `json:"pictures"`
}

type ArtworkListQuery struct {
	R18      shared.R18Type
	ArtistID string
	Tags     [][]string // AND of OR tags
	Keywords [][]string // AND of OR keywords
	Hybrid   bool       // use hybrid search
	Query    string     // fulltext search query
	Limit    int
	Offset   int
}

type ArtworkListQueryResult = []*ArtworkQueryResult
