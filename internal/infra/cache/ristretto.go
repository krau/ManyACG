package cache

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/samber/oops"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	ristrettoCache *ristretto.Cache[string, []byte]
	defaultTTL     time.Duration
)

func Init() error {
	cfg := runtimecfg.Get().Cache
	c, err := ristretto.NewCache(&ristretto.Config[string, []byte]{
		NumCounters: cfg.Ristretto.NumCounters,
		MaxCost:     cfg.Ristretto.MaxCost,
		BufferItems: 64,
		OnReject: func(item *ristretto.Item[[]byte]) {
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

func Default() *ristretto.Cache[string, []byte] {
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
	val, err := msgpack.Marshal(value)
	if err != nil {
		return oops.Wrapf(err, "failed to marshal cache value for key %s", key)
	}
	ok := ristrettoCache.SetWithTTL(key, val, 0, defaultTTL)
	if !ok {
		return oops.Errorf("failed to set cache value for key %s", key)
	}
	ristrettoCache.Wait()
	return nil
}

func SetWithoutTTL(key string, value any) error {
	val, err := msgpack.Marshal(value)
	if err != nil {
		return oops.Wrapf(err, "failed to marshal cache value for key %s", key)
	}
	ok := ristrettoCache.Set(key, val, 0)
	if !ok {
		return oops.Errorf("failed to set cache value for key %s", key)
	}
	ristrettoCache.Wait()
	return nil
}

func Get[T any](key string) (T, bool) {
	v, ok := ristrettoCache.Get(key)
	if !ok {
		return *new(T), false
	}
	var value T
	if err := msgpack.Unmarshal(v, &value); err != nil {
		log.Errorf("failed to unmarshal cache value for key %s: %v", key, err)
		return *new(T), false
	}
	return value, true
}
