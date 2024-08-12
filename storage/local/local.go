package local

import (
	"ManyACG/common"
	"ManyACG/config"
	. "ManyACG/logger"
	"ManyACG/sources"
	"ManyACG/types"
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

func (l *Local) SavePicture(artwork *types.Artwork, picture *types.Picture) (*types.StorageInfo, error) {
	Logger.Debugf("Saving picture %d of artwork %s", picture.Index, artwork.Title)
	fileName, err := sources.GetFileName(artwork, picture)
	if err != nil {
		return nil, err
	}
	artistName := common.ReplaceFileNameInvalidChar(artwork.Artist.Username)
	fileDir := basePath + "/" + string(artwork.SourceType) + "/" + artistName + "/"
	fileBytes, err := common.DownloadWithCache(picture.Original, nil)
	if err != nil {
		Logger.Errorf("Failed to download file: %s", err)
		return nil, err
	}
	filePath := fileDir + fileName
	if err := common.MkFile(filePath, fileBytes); err != nil {
		Logger.Errorf("Failed to write file: %s", err)
		return nil, err
	}
	Logger.Infof("Picture %d of artwork %s saved to %s", picture.Index, artwork.Title, filePath)
	storageInfo := &types.StorageInfo{
		Type: types.StorageTypeLocal,
		Path: filePath,
	}
	fileBytes = nil
	return storageInfo, nil
}

func (l *Local) GetFile(info *types.StorageInfo) ([]byte, error) {
	return os.ReadFile(info.Path)
}

func (l *Local) DeletePicture(info *types.StorageInfo) error {
	return common.PurgeFile(info.Path)
}
