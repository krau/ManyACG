package webdav

import (
	"context"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"

	"github.com/krau/ManyACG/types"

	"github.com/studio-b12/gowebdav"
)

type Webdav struct{}

var Client *gowebdav.Client

var (
	basePath string
)

func (w *Webdav) Init() {
	webdavConfig := config.Cfg.Storage.Webdav
	basePath = strings.TrimSuffix(webdavConfig.Path, "/")
	Client = gowebdav.NewClient(webdavConfig.URL, webdavConfig.Username, webdavConfig.Password)
	if err := Client.Connect(); err != nil {
		common.Logger.Fatalf("Failed to connect to webdav server: %v", err)
		os.Exit(1)
	}
}

func (w *Webdav) Save(ctx context.Context, filePath string, storagePath string) (*types.StorageDetail, error) {
	common.Logger.Debugf("saving file %s", filePath)
	storagePath = path.Join(basePath, storagePath)
	storageDir := path.Dir(storagePath)
	if err := Client.MkdirAll(storageDir, os.ModePerm); err != nil {
		common.Logger.Errorf("failed to create directory: %s", err)
		return nil, ErrFailedMkdirAll
	}

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		common.Logger.Errorf("failed to read file: %s", err)
		return nil, err
	}

	if err := Client.Write(storagePath, fileBytes, os.ModePerm); err != nil {
		common.Logger.Errorf("failed to write file: %s", err)
		return nil, ErrFailedWrite
	}

	cachePath := filepath.Join(config.Cfg.Storage.CacheDir, filepath.Base(storagePath))
	go common.MkCache(cachePath, fileBytes, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)

	return &types.StorageDetail{
		Type: types.StorageTypeWebdav,
		Path: storagePath,
	}, nil
}

func (w *Webdav) GetFile(ctx context.Context, detail *types.StorageDetail) ([]byte, error) {
	common.Logger.Debugf("Getting file %s", detail.Path)
	cachePath := filepath.Join(config.Cfg.Storage.CacheDir, path.Base(detail.Path))
	data, err := os.ReadFile(cachePath)
	if err == nil {
		return data, nil
	}
	data, err = Client.Read(detail.Path)
	if err != nil {
		common.Logger.Errorf("failed to read file: %s", err)
		return nil, ErrReadFile
	}
	go common.MkCache(cachePath, data, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)
	return data, nil
}

func (w *Webdav) GetFileStream(ctx context.Context, detail *types.StorageDetail) (io.ReadCloser, error) {
	common.Logger.Debugf("Getting file %s", detail.Path)
	cachePath := filepath.Join(config.Cfg.Storage.CacheDir, path.Base(detail.Path))
	file, err := os.Open(cachePath)
	if err == nil {
		return file, nil
	}
	steam, err := Client.ReadStream(detail.Path)
	if err != nil {
		common.Logger.Errorf("failed to read file: %s", err)
		return nil, ErrReadFile
	}
	return steam, nil
}

func (w *Webdav) Delete(ctx context.Context, detail *types.StorageDetail) error {
	common.Logger.Debugf("Deleting file %s", detail.Path)
	return Client.Remove(detail.Path)
}
