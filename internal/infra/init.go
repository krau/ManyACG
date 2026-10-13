package infra

import (
	"context"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/database"
	"github.com/krau/ManyACG/internal/infra/kvstor"
	"github.com/samber/oops"
)

func Init(ctx context.Context, cfg runtimecfg.Config) (func() error, error) {
	var errs []error
	var closerFuncs []func() error
	if err := cfg.App.Mode.Valid(); err != nil {
		return nil, err
	}
	if !cfg.App.Mode.ReadOnly() {
		kvstor.Init(cfg.KVDB)
		closerFuncs = append(closerFuncs, func() error {
			return kvstor.Close()
		})
	}
	initSource()
	if err := initStorage(ctx); err != nil {
		errs = append(errs, err)
	}
	database.Init(ctx, cfg)

	return func() error {
		var errs []error
		for _, closer := range closerFuncs {
			if err := closer(); err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			return oops.Join(errs...)
		}
		return nil
	}, oops.Join(errs...)
}
