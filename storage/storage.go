package storage

import (
	"ManyACG-Bot/config"
	"ManyACG-Bot/storage/webdav"
)

var storage Storage

func GetStorage() Storage {
	if storage == nil {
		switch config.Cfg.Storage.Type {
		case "webdav":
			storage = new(webdav.Webdav)
		default:
			storage = new(webdav.Webdav)
		}
	}
	return storage
}