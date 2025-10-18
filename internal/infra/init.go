package infra

import (
	"context"
	"time"

	"github.com/krau/ManyACG/internal/infra/cache"
	"github.com/krau/ManyACG/internal/infra/database"
	"github.com/krau/ManyACG/internal/infra/kvstor"
	"github.com/samber/oops"
)

// Init initializes the infra package, return the closer function and error if any.
func Init(ctx context.Context) (func() error, error) {
	var errs []error
	var closerFuncs []func() error
	kvstor.Set("app:last_start_time", time.Now().Format(time.RFC3339)) // just for preheat
	closerFuncs = append(closerFuncs, func() error {
		return kvstor.Close()
	})
	initSource()
	if err := initStorage(ctx); err != nil {
		errs = append(errs, err)
	}
	database.Init(ctx)
	if err := cache.Init(); err != nil {
		errs = append(errs, err)
	} else {
		closerFuncs = append(closerFuncs, func() error {
			return cache.Close()
		})
	}

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
