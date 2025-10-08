package entity

import "github.com/krau/ManyACG/internal/shared"

type PictureLike interface {
	IsHide() bool
	GetIndex() uint
	GetTelegramInfo() shared.TelegramInfo
	GetOriginal() string
	GetThumbnail() string
	GetSize() (width, height uint)
	GetStorageInfo() shared.StorageInfo
}

var _ PictureLike = (*Picture)(nil)
var _ PictureLike = (*CachedPicture)(nil)
