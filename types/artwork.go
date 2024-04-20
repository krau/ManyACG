package types

import (
	"time"
)

type Artwork struct {
	Title       string
	Description string
	R18         bool
	CreatedAt   time.Time
	SourceType  SourceType
	SourceURL   string
	Artist      *Artist
	Tags        []string
	Pictures    []*Picture
}

type Artist struct {
	Name     string
	Type     SourceType
	UID      int
	Username string
}

type Picture struct {
	Index     uint   // 图片在作品中的顺序
	Thumbnail string // 缩略图 URL
	Original  string // 原图 URL

	Width     uint
	Height    uint
	Hash      string
	BlurScore float64

	TelegramInfo *TelegramInfo
	StorageInfo  *StorageInfo
}

type TelegramInfo struct {
	PhotoFileID    string
	DocumentFileID string
	MessageID      int
	MediaGroupID   string
}

type StorageInfo struct {
	Type StorageType
	Path string
}
