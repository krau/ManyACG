package local

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/internal/shared"
)

type Local struct{}

var (
	basePath string
)

func (l *Local) Init(ctx context.Context) {
	basePath = strings.TrimSuffix(config.Cfg.Storage.Local.Path, "/")
	if basePath == "" {
		common.Logger.Panic("Local storage path not set,for example: manyacg/storage")
	}
	if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
		common.Logger.Panicf("Failed to create directory: %v", err)
	}
}

func (l *Local) Save(ctx context.Context, filePath string, storagePath string) (*shared.StorageDetail, error) {
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
	return &shared.StorageDetail{
		Type: shared.StorageTypeLocal,
		Path: storagePath,
	}, nil
}

func (l *Local) GetFile(ctx context.Context, detail *shared.StorageDetail) ([]byte, error) {
	common.Logger.Debugf("getting file %s", detail.Path)
	return os.ReadFile(detail.Path)
}

func (l *Local) GetFileStream(ctx context.Context, detail *shared.StorageDetail) (io.ReadCloser, error) {
	return os.Open(detail.Path)
}

func (l *Local) Delete(ctx context.Context, detail *shared.StorageDetail) error {
	common.Logger.Debugf("deleting file %s", detail.Path)
	return common.PurgeFile(detail.Path)
}
