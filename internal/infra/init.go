package infra

import (
	"context"
	"sync"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/database"
	"github.com/krau/ManyACG/internal/infra/kvstor"
	"github.com/krau/ManyACG/internal/infra/source/remote"
	"github.com/samber/oops"
)

type Resources struct {
	KV             kvstor.KVStore
	ArtworkFetcher *remote.Client
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
	if err := cfg.App.Mode.Valid(); err != nil {
		return nil, err
	}
	if err := cfg.Source.Remote.Valid(cfg.App.Mode, cfg.Rest.Enable); err != nil {
		return nil, err
	}
	resources := &Resources{}
	if !cfg.App.Mode.ReadOnly() {
		store, err := kvstor.Open(cfg.KVDB)
		if err != nil {
			return nil, err
		}
		resources.KV = store
		resources.closers = append(resources.closers, store.Close)
		if len(cfg.Source.Remote.Endpoints) > 0 {
			fetcher, err := remote.New(cfg.Source.Remote)
			if err != nil {
				return nil, oops.Join(err, resources.Close())
			}
			resources.ArtworkFetcher = fetcher
			resources.closers = append(resources.closers, fetcher.Close)
		}
	}
	initSource(resources.KV)
	if err := initStorage(ctx); err != nil {
		return nil, oops.Join(err, resources.Close())
	}
	database.Init(ctx, cfg)
	resources.closers = append(resources.closers, database.Default().Close)
	return resources, nil
}
