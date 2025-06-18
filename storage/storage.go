package storage

import (
	"context"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"

	"github.com/krau/ManyACG/storage/alist"
	"github.com/krau/ManyACG/storage/local"
	"github.com/krau/ManyACG/storage/telegram"
	"github.com/krau/ManyACG/storage/webdav"
	"github.com/krau/ManyACG/types"
)

var Storages = make(map[types.StorageType]Storage)

func InitStorage(ctx context.Context) {
	common.Logger.Info("Initializing storage")
	if config.Cfg.Storage.Local.Enable {
		Storages[types.StorageTypeLocal] = new(local.Local)
		Storages[types.StorageTypeLocal].Init(ctx)
	}
	if config.Cfg.Storage.Webdav.Enable {
		Storages[types.StorageTypeWebdav] = new(webdav.Webdav)
		Storages[types.StorageTypeWebdav].Init(ctx)
	}
	if config.Cfg.Storage.Alist.Enable {
		Storages[types.StorageTypeAlist] = new(alist.Alist)
		Storages[types.StorageTypeAlist].Init(ctx)
	}
	if config.Cfg.Storage.Telegram.Enable {
		Storages[types.StorageTypeTelegram] = new(telegram.TelegramStorage)
		Storages[types.StorageTypeTelegram].Init(ctx)
	}
}
