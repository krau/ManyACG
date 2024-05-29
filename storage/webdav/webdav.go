package webdav

import (
	"ManyACG/common"
	"ManyACG/config"
	. "ManyACG/logger"
	"ManyACG/sources"
	"ManyACG/types"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/studio-b12/gowebdav"
)

type Webdav struct{}

var Client *gowebdav.Client

func (w *Webdav) Init() {
	webdavConfig := config.Cfg.Storage.Webdav
	Client = gowebdav.NewClient(webdavConfig.URL, webdavConfig.Username, webdavConfig.Password)
	if err := Client.Connect(); err != nil {
		Logger.Fatalf("Failed to connect to webdav server: %v", err)
		os.Exit(1)
	}
}

func (w *Webdav) SavePicture(artwork *types.Artwork, picture *types.Picture) (*types.StorageInfo, error) {
	Logger.Debugf("saving picture %d of artwork %s", picture.Index, artwork.Title)
	fileName := sources.GetFileName(artwork, picture)
	artistName := common.ReplaceFileNameInvalidChar(artwork.Artist.Username)
	fileDir := strings.TrimSuffix(config.Cfg.Storage.Webdav.Path, "/") + "/" + string(artwork.SourceType) + "/" + artistName + "/"
	if err := Client.MkdirAll(fileDir, os.ModePerm); err != nil {
		Logger.Errorf("failed to create directory: %s", err)
		return nil, ErrFailedMkdirAll
	}
	fileBytes, err := common.DownloadFromURL(picture.Original)
	if err != nil {
		Logger.Errorf("failed to download file: %s", err)
		return nil, ErrFailedDownload
	}
	filePath := fileDir + fileName
	if err := Client.Write(filePath, fileBytes, os.ModePerm); err != nil {
		Logger.Errorf("failed to write file: %s", err)
		return nil, ErrFailedWrite
	}
	Logger.Infof("picture %d of artwork %s saved to %s", picture.Index, artwork.Title, filePath)
	storageInfo := &types.StorageInfo{
		Type: types.StorageTypeWebdav,
		Path: filePath,
	}
	if config.Cfg.Storage.CacheDir == "" {
		return storageInfo, nil
	}
	cachePath := strings.TrimSuffix(config.Cfg.Storage.CacheDir, "/") + "/" + filepath.Base(filePath)
	if err := common.MkFile(cachePath, fileBytes); err != nil {
		Logger.Warnf("failed to save cache file: %s", err)
	} else {
		go common.PurgeFileAfter(cachePath, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)
	}
	return storageInfo, nil
}

func (w *Webdav) GetFile(info *types.StorageInfo) ([]byte, error) {
	Logger.Debugf("Getting file %s", info.Path)
	if config.Cfg.Storage.CacheDir != "" {
		return w.getFileWithCache(info)
	}
	return Client.Read(info.Path)
}

func (w *Webdav) getFileWithCache(info *types.StorageInfo) ([]byte, error) {
	cachePath := strings.TrimSuffix(config.Cfg.Storage.CacheDir, "/") + "/" + filepath.Base(info.Path)
	data, err := os.ReadFile(cachePath)
	if err == nil {
		return data, nil
	}
	data, err = Client.Read(info.Path)
	if err != nil {
		Logger.Errorf("failed to read file: %s", err)
		return nil, ErrReadFile
	}
	if err := common.MkFile(cachePath, data); err != nil {
		Logger.Errorf("failed to save cache file: %s", err)
	} else {
		go common.PurgeFileAfter(cachePath, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)
	}
	return data, nil
}

func (w *Webdav) DeletePicture(info *types.StorageInfo) error {
	Logger.Debugf("deleting file %s", info.Path)
	return Client.Remove(info.Path)
}
