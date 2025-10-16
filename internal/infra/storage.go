package infra

import (
	"context"

	"github.com/krau/ManyACG/internal/infra/storage"
	"github.com/krau/ManyACG/internal/infra/storage/local"
	"github.com/krau/ManyACG/internal/infra/storage/telegram"
	"github.com/krau/ManyACG/internal/infra/storage/webdav"
)

func initStorage(ctx context.Context) error {
	local.Init()
	telegram.Init()
	webdav.Init()

	return storage.InitAll(ctx)
}
