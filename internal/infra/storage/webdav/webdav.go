package webdav

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/krau/ManyACG/internal/infra/config"
	"github.com/krau/ManyACG/internal/infra/storage"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/osutil"

	"github.com/studio-b12/gowebdav"
)

type Webdav struct {
	cfg      config.StorageWebdavConfig
	basePath string
	client   *gowebdav.Client
}

func init() {
	storage.Register(shared.StorageTypeWebdav, func() storage.Storage {
		return &Webdav{
			cfg:      config.Get().Storage.Webdav,
			basePath: strings.TrimSuffix(config.Get().Storage.Webdav.Path, "/"),
		}
	})
}

func (w *Webdav) Init(ctx context.Context) error {
	client := gowebdav.NewClient(w.cfg.URL, w.cfg.Username, w.cfg.Password)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to webdav server: %w", err)
	}
	w.client = client
	return nil
}

func (w *Webdav) Save(ctx context.Context, filePath string, storagePath string) (*shared.StorageDetail, error) {
	storagePath = path.Join(w.basePath, storagePath)
	storageDir := path.Dir(storagePath)
	if err := w.client.MkdirAll(storageDir, os.ModePerm); err != nil {
		return nil, ErrFailedMkdirAll
	}

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, ErrReadFile
	}

	if err := w.client.Write(storagePath, fileBytes, os.ModePerm); err != nil {
		return nil, ErrFailedWrite
	}

	cachePath := filepath.Join(config.Get().Storage.CacheDir, filepath.Base(storagePath))
	go osutil.MkCache(cachePath, fileBytes, time.Duration(config.Get().Storage.CacheTTL)*time.Second)

	return &shared.StorageDetail{
		Type: shared.StorageTypeWebdav,
		Path: storagePath,
	}, nil
}

func (w *Webdav) GetFile(ctx context.Context, detail *shared.StorageDetail) ([]byte, error) {
	cachePath := filepath.Join(config.Get().Storage.CacheDir, path.Base(detail.Path))
	data, err := os.ReadFile(cachePath)
	if err == nil {
		return data, nil
	}
	data, err = w.client.Read(detail.Path)
	if err != nil {
		return nil, ErrReadFile
	}
	go osutil.MkCache(cachePath, data, time.Duration(config.Get().Storage.CacheTTL)*time.Second)
	return data, nil
}

// func (w *Webdav) GetFileStream(ctx context.Context, detail *shared.StorageDetail) (io.ReadCloser, error) {
// 	cachePath := filepath.Join(config.Cfg.Storage.CacheDir, path.Base(detail.Path))
// 	file, err := os.Open(cachePath)
// 	if err == nil {
// 		return file, nil
// 	}
// 	steam, err := Client.ReadStream(detail.Path)
// 	if err != nil {
// 		common.Logger.Errorf("failed to read file: %s", err)
// 		return nil, ErrReadFile
// 	}
// 	return steam, nil
// }

func (w *Webdav) Delete(ctx context.Context, detail *shared.StorageDetail) error {
	return w.client.Remove(detail.Path)
}
