package storage

import (
	"context"

	"github.com/krau/ManyACG/internal/shared"
)

type Storage interface {
	Init(ctx context.Context) error
	// from: 本地文件路径, to: 存储路径 (如 "2023/01/01/xxxx.jpg")
	//
	// 存储实现可能会对传入的存储路径进行其他处理 (如添加前缀), 因此返回的 StorageDetail 中的 Path 可能与传入的 storagePath 不同.
	Save(ctx context.Context, from, to string) (*shared.StorageDetail, error)
	GetFile(ctx context.Context, info *shared.StorageDetail) ([]byte, error)
	// GetFileStream(ctx context.Context, info *types.StorageDetail) (io.ReadCloser, error)
	Delete(ctx context.Context, info *shared.StorageDetail) error
}
