package storage

import (
	"ManyACG/types"
	"context"
)

type Storage interface {
	Init()

	// filePath 本地文件路径, storagePath 存储路径.
	//
	// 存储实现可能会对传入的存储路径进行其他处理 (如添加前缀), 因此返回的 StorageDetail 中的 Path 可能与传入的 storagePath 不同.
	Save(ctx context.Context, filePath string, storagePath string) (*types.StorageDetail, error)
	GetFile(ctx context.Context, info *types.StorageDetail) ([]byte, error)
	Delete(ctx context.Context, info *types.StorageDetail) error
}
