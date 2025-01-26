package local

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"

	"github.com/krau/ManyACG/types"
)

type Local struct{}

var (
	basePath string
)

func (l *Local) Init() {
	basePath = strings.TrimSuffix(config.Cfg.Storage.Local.Path, "/")
	if basePath == "" {
		common.Logger.Fatalf("Local storage path not set,for example: manyacg/storage")
		os.Exit(1)
	}
	if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
		common.Logger.Fatalf("Failed to create directory: %v", err)
		os.Exit(1)
	}
}

func (l *Local) Save(ctx context.Context, filePath string, storagePath string) (*types.StorageDetail, error) {
	common.Logger.Debugf("saving file %s", filePath)
	storagePath = filepath.Join(basePath, storagePath)
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		common.Logger.Errorf("failed to read file: %s", err)
		return nil, err
	}
	if err := common.MkFile(storagePath, fileBytes); err != nil {
		common.Logger.Errorf("failed to write file: %s", err)
		return nil, err
	}
	return &types.StorageDetail{
		Type: types.StorageTypeLocal,
		Path: storagePath,
	}, nil
}

func (l *Local) GetFile(ctx context.Context, detail *types.StorageDetail) ([]byte, error) {
	common.Logger.Debugf("getting file %s", detail.Path)
	return os.ReadFile(detail.Path)
}

func (l *Local) GetFileStream(ctx context.Context, detail *types.StorageDetail) (io.ReadCloser, error) {
	return os.Open(detail.Path)
}

func (l *Local) Delete(ctx context.Context, detail *types.StorageDetail) error {
	common.Logger.Debugf("deleting file %s", detail.Path)
	return common.PurgeFile(detail.Path)
}
