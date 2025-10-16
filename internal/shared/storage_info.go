package shared

import (
	"fmt"

	"github.com/krau/ManyACG/pkg/strutil"
)

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
	return strutil.MD5Hash(fmt.Sprintf("%s:%s:%s", s.Type, s.Path, s.Mime))
}
