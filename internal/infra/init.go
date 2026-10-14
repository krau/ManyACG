package infra

import (
	"context"
	"sync"

	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/database"
	"github.com/krau/ManyACG/internal/infra/kvstor"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/infra/source/remote"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/samber/oops"
)

type Resources struct {
	KV             kvstor.KVStore
	ArtworkFetcher source.ArtworkFetcher
	FileDownloader httpclient.Fetcher
	SourceNamers   map[shared.SourceType]source.ArtworkNamer
	closers        []func() error
	closeOnce      sync.Once
	closeErr       error
}

func (r *Resources) Close() error {
	r.closeOnce.Do(func() {
		var errs []error
		for i := len(r.closers) - 1; i >= 0; i-- {
			if err := r.closers[i](); err != nil {
				errs = append(errs, err)
			}
		}
		r.closeErr = oops.Join(errs...)
	})
	return r.closeErr
}

func Init(ctx context.Context, cfg runtimecfg.Config) (*Resources, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	resources := &Resources{}
	writable := cfg.App.Capabilities.Has(runtimecfg.CapabilityWrite)
	readable := cfg.App.Capabilities.Has(runtimecfg.CapabilityRead)
	if writable {
		store, err := kvstor.Open(cfg.KVDB)
		if err != nil {
			return nil, err
		}
		resources.KV = store
		resources.closers = append(resources.closers, store.Close)
	}
	if writable || cfg.App.Capabilities.Has(runtimecfg.CapabilityFetch) {
		initSource(resources.KV)
	}
	if readable || writable {
		resources.SourceNamers = sourceNamers(cfg.Source)
		if err := initStorage(ctx); err != nil {
			return nil, oops.Join(err, resources.Close())
		}
		database.Init(ctx, cfg)
		resources.closers = append(resources.closers, database.Default().Close)
	}
	if writable && len(cfg.Source.Remote.Fetch.Endpoints) > 0 {
		fetcher, err := remote.New(cfg.Source.Remote.Fetch, cfg.Source.Remote.Token)
		if err != nil {
			return nil, oops.Join(err, resources.Close())
		}
		resources.ArtworkFetcher = fetcher
		resources.closers = append(resources.closers, fetcher.Close)
	}
	if (writable || readable) && len(cfg.Source.Remote.Download.Endpoints) > 0 {
		downloader, err := remote.New(cfg.Source.Remote.Download, cfg.Source.Remote.Token)
		if err != nil {
			return nil, oops.Join(err, resources.Close())
		}
		resources.FileDownloader = downloader
		resources.closers = append(resources.closers, downloader.Close)
	}
	return resources, nil
}
