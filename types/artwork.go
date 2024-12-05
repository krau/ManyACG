package types

import (
	"time"
)

type Artwork struct {
	ID          string     `json:"id" bson:"_id"`
	Title       string     `json:"title" bson:"title"`
	Description string     `json:"description" bson:"description"`
	R18         bool       `json:"r18" bson:"r18"`
	CreatedAt   time.Time  `json:"created_at" bson:"created_at"`
	SourceType  SourceType `json:"source_type" bson:"source_type"`
	SourceURL   string     `json:"source_url" bson:"source_url"`
	LikeCount   uint       `json:"like_count" bson:"like_count"`

	Artist   *Artist    `json:"artist" bson:"artist"`
	Tags     []string   `json:"tags" bson:"tags"`
	Pictures []*Picture `json:"pictures" bson:"pictures"`
}

type Artist struct {
	ID       string     `json:"id" bson:"_id"`
	Name     string     `json:"name" bson:"name"`
	Type     SourceType `json:"type" bson:"type"`
	UID      string     `json:"uid" bson:"uid"`
	Username string     `json:"username" bson:"username"`
}

type Picture struct {
	ID        string `json:"id" bson:"_id"`
	ArtworkID string `json:"artwork_id" bson:"artwork_id"`
	Index     uint   `json:"index" bson:"index"`
	Thumbnail string `json:"thumbnail" bson:"thumbnail"`
	Original  string `json:"original" bson:"original"`

	Width  uint   `json:"width" bson:"width"`
	Height uint   `json:"height" bson:"height"`
	Hash   string `json:"hash" bson:"hash"`
	// BlurScore float64 `json:"blur_score" bson:"blur_score"`

	TelegramInfo *TelegramInfo `json:"telegram_info" bson:"telegram_info"`
	StorageInfo  *StorageInfo  `json:"storage_info" bson:"storage_info"`
}

type TelegramInfo struct {
	PhotoFileID    string `json:"photo_file_id" bson:"photo_file_id"`
	DocumentFileID string `json:"document_file_id" bson:"document_file_id"`
	MessageID      int    `json:"message_id" bson:"message_id"`
	MediaGroupID   string `json:"media_group_id" bson:"media_group_id"`
}

type StorageInfo struct {
	Original *StorageDetail `json:"original" bson:"original"`
	Regular  *StorageDetail `json:"regular" bson:"regular"`
	Thumb    *StorageDetail `json:"thumb" bson:"thumb"`
}

type StorageDetail struct {
	Type StorageType `json:"type" bson:"type"`
	Path string      `json:"path" bson:"path"`
}

func (detail *StorageDetail) String() string {
	return string(detail.Type) + ":" + detail.Path
}
