package storage

import (
	"ManyACG-Bot/types"
)

type Storage interface {
	SavePicture(artwork *types.Artwork, picture *types.Picture) (*types.StorageInfo, error)
	GetFile(info *types.StorageInfo) ([]byte, error)
}
