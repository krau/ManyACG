package local

import (
	"ManyACG/common"
	"ManyACG/config"
	. "ManyACG/logger"
	"ManyACG/types"
	"context"
	"os"
	"strings"
)

type Local struct{}

var (
	basePath string
)

func (l *Local) Init() {
	basePath = strings.TrimSuffix(config.Cfg.Storage.Local.Path, "/")
	if basePath == "" {
		Logger.Fatalf("Local storage path not set,for example: manyacg/storage")
		os.Exit(1)
	}
	if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
		Logger.Fatalf("Failed to create directory: %v", err)
		os.Exit(1)
	}
}

func (l *Local) Save(ctx context.Context, filePath string, storagePath string) (*types.StorageDetail, error) {
	Logger.Debugf("saving file %s", filePath)
	storagePath = basePath + storagePath
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		Logger.Errorf("failed to read file: %s", err)
		return nil, err
	}
	if err := common.MkFile(storagePath, fileBytes); err != nil {
		Logger.Errorf("failed to write file: %s", err)
		return nil, err
	}
	return &types.StorageDetail{
		Type: types.StorageTypeLocal,
		Path: storagePath,
	}, nil
}

func (l *Local) GetFile(ctx context.Context, detail *types.StorageDetail) ([]byte, error) {
	return os.ReadFile(detail.Path)
}

func (l *Local) Delete(ctx context.Context, detail *types.StorageDetail) error {
	return common.PurgeFile(detail.Path)
}
