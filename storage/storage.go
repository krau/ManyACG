package storage

import (
	"ManyACG/config"
	. "ManyACG/logger"
	"ManyACG/storage/webdav"
)

var storage Storage

func GetStorage() Storage {
	if storage == nil {
		switch config.Cfg.Storage.Type {
		case "webdav":
			storage = new(webdav.Webdav)
			storage.Init()
		default:
			storage = new(webdav.Webdav)
			storage.Init()
		}
	}
	return storage
}

func InitStorage() {
	Logger.Info("Initializing storage")
	GetStorage()
}
