package infra

import (
	"context"
	"sync"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/database"
	"github.com/krau/ManyACG/internal/infra/kvstor"
	"github.com/samber/oops"
)

type Resources struct {
	KV        kvstor.KVStore
	closers   []func() error
	closeOnce sync.Once
	closeErr  error
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
	resources := &Resources{}
	if !cfg.App.Mode.ReadOnly() {
		store, err := kvstor.Open(cfg.KVDB)
		if err != nil {
			return nil, err
		}
		resources.KV = store
		resources.closers = append(resources.closers, store.Close)
	}
	initSource(cfg, resources.KV)
	if err := initStorage(ctx); err != nil {
		return nil, oops.Join(err, resources.Close())
	}
	database.Init(ctx, cfg)
	resources.closers = append(resources.closers, database.Default().Close)
	return resources, nil
}
