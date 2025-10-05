package webdav

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	config "github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/storage"
	"github.com/krau/ManyACG/internal/shared"

	"github.com/studio-b12/gowebdav"
)

type Webdav struct {
	cfg      config.StorageWebdavConfig
	basePath string
	client   *gowebdav.Client
}

func init() {
	if !config.Get().Storage.Webdav.Enable {
		return
	}
	storage.Register(shared.StorageTypeWebdav, func() storage.Storage {
		return &Webdav{
			cfg:      config.Get().Storage.Webdav,
			basePath: strings.TrimSuffix(config.Get().Storage.Webdav.Path, "/"),
		}
	})
}

func (w *Webdav) Init(ctx context.Context) error {
	client := gowebdav.NewClient(w.cfg.URL, w.cfg.Username, w.cfg.Password)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to webdav server: %w", err)
	}
	w.client = client
	return nil
}

func (w *Webdav) Save(ctx context.Context, r io.Reader, storagePath string) (*shared.StorageDetail, error) {
	storagePath = path.Join(w.basePath, storagePath)
	storageDir := path.Dir(storagePath)
	if err := w.client.MkdirAll(storageDir, os.ModePerm); err != nil {
		return nil, ErrFailedMkdirAll
	}

	if err := w.client.WriteStream(storagePath, r, os.ModePerm); err != nil {
		return nil, ErrFailedWrite
	}
	return &shared.StorageDetail{
		Type: shared.StorageTypeWebdav,
		Path: storagePath,
	}, nil
}

func (w *Webdav) GetFile(ctx context.Context, detail shared.StorageDetail) (io.ReadCloser, error) {
	return w.client.ReadStream(detail.Path)
}

func (w *Webdav) Delete(ctx context.Context, detail shared.StorageDetail) error {
	return w.client.Remove(detail.Path)
}
