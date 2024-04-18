package webdav

import (
	"ManyACG-Bot/common"
	"ManyACG-Bot/config"
	. "ManyACG-Bot/logger"
	"ManyACG-Bot/types"
	"os"
	"path/filepath"
	"strings"
)

type Webdav struct{}

func (w *Webdav) SavePicture(artwork *types.Artwork, picture *types.Picture) (*types.StorageInfo, error) {
	Logger.Debugf("saving picture %d of artwork %s", picture.Index, artwork.Title)
	fileName := artwork.Title + "_" + filepath.Base(picture.Original)
	fileDir := strings.TrimSuffix(config.Cfg.Storage.Webdav.Path, "/") + "/" + string(artwork.Source.Type) + "/" + artwork.Artist.Name + "/"
	if err := Client.MkdirAll(fileDir, os.ModePerm); err != nil {
		return nil, err
	}
	fileBytes, err := common.DownloadFromURL(picture.Original)
	if err != nil {
		return nil, err
	}

	filePath := fileDir + fileName
	if err := Client.Write(filePath, fileBytes, os.ModePerm); err != nil {
		return nil, err
	}
	return &types.StorageInfo{
		Type: types.StorageTypeWebdav,
		Path: filePath,
	}, nil

}

func (w *Webdav) GetFile(info *types.StorageInfo) ([]byte, error) {
	return Client.Read(info.Path)
}
