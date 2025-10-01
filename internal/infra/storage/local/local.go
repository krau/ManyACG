package local

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	config "github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/storage"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/osutil"
)

type Local struct {
	basePath string
}

func init() {
	storage.Register(shared.StorageTypeLocal, func() storage.Storage {
		return &Local{
			basePath: strings.TrimSuffix(config.Get().Storage.Local.Path, "/"),
		}
	})
}

func (l *Local) Init(ctx context.Context) error {
	if err := os.MkdirAll(l.basePath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

func (l *Local) Save(ctx context.Context, filePath string, storagePath string) (*shared.StorageDetail, error) {
	storagePath = filepath.Join(l.basePath, storagePath)
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	if err := osutil.MkFile(storagePath, fileBytes); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}
	return &shared.StorageDetail{
		Type: shared.StorageTypeLocal,
		Path: storagePath,
	}, nil
}

func (l *Local) GetFile(ctx context.Context, detail *shared.StorageDetail) ([]byte, error) {
	return os.ReadFile(detail.Path)
}

func (l *Local) Delete(ctx context.Context, detail *shared.StorageDetail) error {
	return osutil.PurgeFile(detail.Path)
}
