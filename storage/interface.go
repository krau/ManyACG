package storage

import (
	"ManyACG-Bot/types"
)

type Storage interface {
	Init()
	SavePicture(artwork *types.Artwork, picture *types.Picture) (*types.StorageInfo, error)
	GetFile(info *types.StorageInfo) ([]byte, error)
	DeletePicture(info *types.StorageInfo) error
}
