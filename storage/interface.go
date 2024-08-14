package storage

import (
	"ManyACG/types"
	"context"
)

type Storage interface {
	Init()
	SavePicture(ctx context.Context, artwork *types.Artwork, picture *types.Picture) (*types.StorageInfo, error)
	GetFile(ctx context.Context, info *types.StorageInfo) ([]byte, error)
	DeletePicture(ctx context.Context, info *types.StorageInfo) error
}
