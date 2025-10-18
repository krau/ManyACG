package cache

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/samber/oops"
)

var (
	ristrettoCache *ristretto.Cache[string, any]
	defaultTTL     time.Duration
)

func Init() error {
	cfg := runtimecfg.Get().Cache
	c, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: cfg.Ristretto.NumCounters,
		MaxCost:     cfg.Ristretto.MaxCost,
		BufferItems: 64,
		OnReject: func(item *ristretto.Item[any]) {
			log.Warnf("Cache item rejected: key=%d, value=%v", item.Key, item.Value)
		},
	})
	if err != nil {
		return oops.Wrapf(err, "failed to create ristretto cache")
	}
	ristrettoCache = c
	defaultTTL = time.Duration(cfg.DefaultTTL) * time.Second
	return nil
}

func Default() *ristretto.Cache[string, any] {
	return ristrettoCache
}

func Close() error {
	if ristrettoCache != nil {
		ristrettoCache.Close()
	}
	return nil
}

// Set set a value with default TTL
func Set(key string, value any) error {
	ok := ristrettoCache.SetWithTTL(key, value, 0, defaultTTL)
	if !ok {
		return oops.Errorf("failed to set cache value for key %s", key)
	}
	ristrettoCache.Wait()
	return nil
}

func SetWithoutTTL(key string, value any) error {
	ok := ristrettoCache.Set(key, value, 0)
	if !ok {
		return oops.Errorf("failed to set cache value for key %s", key)
	}
	ristrettoCache.Wait()
	return nil
}

func Get[T any](key string) (T, bool) {
	v, ok := ristrettoCache.Get(key)
	if !ok {
		var zero T
		return zero, false
	}
	vT, ok := v.(T)
	if !ok {
		var zero T
		return zero, false
	}
	return vT, true
}
