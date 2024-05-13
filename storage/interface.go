package storage

import (
	"ManyACG/types"
)

type Storage interface {
	Init()
	SavePicture(artwork *types.Artwork, picture *types.Picture) (*types.StorageInfo, error)
	GetFile(info *types.StorageInfo) ([]byte, error)
	DeletePicture(info *types.StorageInfo) error
}
