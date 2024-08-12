package storage

import (
	"ManyACG/config"
	. "ManyACG/logger"
	"ManyACG/storage/local"
	"ManyACG/storage/webdav"
	"ManyACG/types"
	"os"
)

var Storages = make(map[types.StorageType]Storage)

var defaultStorageType types.StorageType

// 获取默认的存储器
func GetStorage() Storage {
	if storage, ok := Storages[defaultStorageType]; ok {
		return storage
	}
	Logger.Panic("Default storage not found")
	os.Exit(1)
	return nil
}

func InitStorage() {
	Logger.Info("Initializing storage")
	defaultStorageType = types.StorageType(config.Cfg.Storage.Default)
	if config.Cfg.Storage.Local.Enable {
		Storages[types.StorageTypeLocal] = new(local.Local)
		Storages[types.StorageTypeLocal].Init()
	}
	if config.Cfg.Storage.Webdav.Enable {
		Storages[types.StorageTypeWebdav] = new(webdav.Webdav)
		Storages[types.StorageTypeWebdav].Init()
	}

	if _, ok := Storages[defaultStorageType]; !ok {
		switch defaultStorageType {
		case types.StorageTypeLocal:
			Storages[types.StorageTypeLocal] = new(local.Local)
			Storages[types.StorageTypeLocal].Init()
		case types.StorageTypeWebdav:
			Storages[types.StorageTypeWebdav] = new(webdav.Webdav)
			Storages[types.StorageTypeWebdav].Init()
		default:
			Storages[types.StorageTypeLocal] = new(local.Local)
			Storages[types.StorageTypeLocal].Init()
		}
	}
}
