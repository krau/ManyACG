package infra

import (
	"context"

	"github.com/krau/ManyACG/internal/infra/storage"
	_ "github.com/krau/ManyACG/internal/infra/storage/local"
	_ "github.com/krau/ManyACG/internal/infra/storage/telegram"
	_ "github.com/krau/ManyACG/internal/infra/storage/webdav"
)

func InitStorages(ctx context.Context) error {
	return storage.InitAll(ctx)
}
