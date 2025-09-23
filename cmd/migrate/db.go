package migrate

import (
	"gorm.io/datatypes"
)

type ArtworkModel struct {
	ID          string `gorm:"primaryKey;" json:"id,omitempty" bson:"_id,omitempty"`
	Title       string `json:"title" bson:"title"`
	Description string `json:"description" bson:"description"`
	R18         bool   `json:"r18" bson:"r18"`
	CreatedAt   int64  `json:"created_at" bson:"created_at"`
	SourceType  int    `json:"source_type" bson:"source_type"`
	SourceURL   string `gorm:"uniqueIndex" json:"source_url" bson:"source_url"`
	LikeCount   uint   `json:"like_count" bson:"like_count"`
}

type ArtistModel struct {
	ID       string `gorm:"primaryKey;" json:"id,omitempty" bson:"_id,omitempty"`
	Name     string `json:"name" bson:"name"`
	Type     int    `json:"type" bson:"type"`
	UID      string `json:"uid" bson:"uid"`
	Username string `json:"username" bson:"username"`
}

type TelegramInfo struct {
	PhotoFileID    string `json:"photo_file_id" bson:"photo_file_id"`
	DocumentFileID string `json:"document_file_id" bson:"document_file_id"`
	MessageID      int    `json:"message_id" bson:"message_id"`
	MediaGroupID   string `json:"media_group_id" bson:"media_group_id"`
}

type StorageType string

func (s StorageType) String() string {
	return string(s)
}

const (
	StorageTypeWebdav   StorageType = "webdav"
	StorageTypeLocal    StorageType = "local"
	StorageTypeAlist    StorageType = "alist"
	StorageTypeTelegram StorageType = "telegram"
)

var StorageTypes []StorageType = []StorageType{
	StorageTypeWebdav,
	StorageTypeLocal,
	StorageTypeAlist,
	StorageTypeTelegram,
}

type StorageDetail struct {
	Type StorageType `json:"type" bson:"type"`
	Path string      `json:"path" bson:"path"`
}

type StorageInfo struct {
	Original *StorageDetail `json:"original" bson:"original"`
	Regular  *StorageDetail `json:"regular" bson:"regular"`
	Thumb    *StorageDetail `json:"thumb" bson:"thumb"`
}

type PictureModel struct {
	ID        string `gorm:"primaryKey;" json:"id,omitempty" bson:"_id,omitempty"`
	ArtworkID string `json:"artwork_id" bson:"artwork_id"`
	Index     uint   `json:"index" bson:"index"`
	Thumbnail string `json:"thumbnail" bson:"thumbnail"`
	Original  string `json:"original" bson:"original"`
	Width     uint   `json:"width" bson:"width"`
	Height    uint   `json:"height" bson:"height"`
	Hash      string `json:"hash" bson:"hash"` // phash
	ThumbHash string `json:"thumb_hash" bson:"thumb_hash"`

	TelegramInfo datatypes.JSON `json:"telegram_info" bson:"telegram_info"`
	StorageInfo  datatypes.JSON `json:"storage_info" bson:"storage_info"`
}
