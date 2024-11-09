package webdav

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	. "github.com/krau/ManyACG/logger"
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
		Logger.Fatalf("Failed to connect to webdav server: %v", err)
		os.Exit(1)
	}
}

func (w *Webdav) Save(ctx context.Context, filePath string, storagePath string) (*types.StorageDetail, error) {
	Logger.Debugf("saving file %s", filePath)
	storagePath = path.Join(basePath, storagePath)
	storageDir := filepath.Dir(storagePath)
	if err := Client.MkdirAll(storageDir, os.ModePerm); err != nil {
		Logger.Errorf("failed to create directory: %s", err)
		return nil, ErrFailedMkdirAll
	}

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		Logger.Errorf("failed to read file: %s", err)
		return nil, err
	}

	if err := Client.Write(storagePath, fileBytes, os.ModePerm); err != nil {
		Logger.Errorf("failed to write file: %s", err)
		return nil, ErrFailedWrite
	}

	cachePath := path.Join(config.Cfg.Storage.CacheDir, filepath.Base(storagePath))
	go common.MkCache(cachePath, fileBytes, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)

	return &types.StorageDetail{
		Type: types.StorageTypeWebdav,
		Path: storagePath,
	}, nil
}

func (w *Webdav) GetFile(ctx context.Context, detail *types.StorageDetail) ([]byte, error) {
	Logger.Debugf("Getting file %s", detail.Path)
	cachePath := path.Join(config.Cfg.Storage.CacheDir, filepath.Base(detail.Path))
	data, err := os.ReadFile(cachePath)
	if err == nil {
		return data, nil
	}
	data, err = Client.Read(detail.Path)
	if err != nil {
		Logger.Errorf("failed to read file: %s", err)
		return nil, ErrReadFile
	}
	go common.MkCache(cachePath, data, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)
	return data, nil
}

func (w *Webdav) Delete(ctx context.Context, detail *types.StorageDetail) error {
	Logger.Debugf("Deleting file %s", detail.Path)
	return Client.Remove(detail.Path)
}
