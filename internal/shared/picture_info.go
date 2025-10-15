package shared

import (
	"fmt"

	"github.com/krau/ManyACG/pkg/strutil"
)

type TelegramInfo struct {
	PhotoFileID    string `json:"photo_file_id"`
	DocumentFileID string `json:"document_file_id"`
	MessageID      int    `json:"message_id"`
	MediaGroupID   string `json:"media_group_id"`
}

var ZeroTelegramInfo = TelegramInfo{}

func (t TelegramInfo) IsZero() bool {
	return t == ZeroTelegramInfo
}

type StorageInfo struct {
	Original *StorageDetail `json:"original"`
	Regular  *StorageDetail `json:"regular"`
	Thumb    *StorageDetail `json:"thumb"`
}

var ZeroStorageInfo = StorageInfo{}

func (s StorageInfo) IsZero() bool {
	return s == ZeroStorageInfo
}

type StorageDetail struct {
	Type StorageType `json:"type"`
	Path string      `json:"path"`
	Mime string      `json:"mime,omitempty"`
}

func (s StorageDetail) IsZero() bool {
	return s == ZeroStorageDetail
}

var ZeroStorageDetail = StorageDetail{}

func (s StorageDetail) Hash() string {
	return strutil.MD5Hash(fmt.Sprintf("%s:%s", s.Type, s.Path))
}
