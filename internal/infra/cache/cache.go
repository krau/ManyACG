package cache

import (
	"context"
	"sync"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/dgraph-io/ristretto/v2"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	bigcachestor "github.com/eko/gocache/store/bigcache/v4"
	ristrettostor "github.com/eko/gocache/store/ristretto/v4"
	rueidisstor "github.com/eko/gocache/store/rueidis/v4"
	"github.com/redis/rueidis"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/pkg/log"
)

var (
	defaultCache *cache.Cache[any]
	once         sync.Once
)

func initCache(ctx context.Context) {
	cfg := runtimecfg.Get().Cache
	switch cfg.Type {
	case "bigcache":
		client, err := bigcache.New(ctx, bigcache.DefaultConfig(time.Duration(cfg.BigCache.Eviction)*time.Second))
		if err != nil {
			log.Fatal("Failed to initialize cache", "err", err)
		}
		stor := bigcachestor.NewBigcache(client)
		defaultCache = cache.New[any](stor)
	case "ristretto":
		client, err := ristretto.NewCache(&ristretto.Config[string, any]{
			NumCounters: cfg.Ristretto.NumCounters,
			MaxCost:     cfg.Ristretto.MaxCost,
			BufferItems: cfg.Ristretto.BufferItems,
		})
		if err != nil {
			log.Fatal("Failed to initialize cache", "err", err)
		}
		stor := ristrettostor.NewRistretto(client, store.WithSynchronousSet(), store.WithCost(1))
		defaultCache = cache.New[any](stor)
	case "redis":
		client, err := rueidis.NewClient(rueidis.ClientOption{
			InitAddress: cfg.Redis.InitAddress,
		})
		if err != nil {
			log.Fatal("Failed to initialize cache", "err", err)
		}
		stor := rueidisstor.NewRueidis(client)
		defaultCache = cache.New[any](stor)
	default:
		client, err := ristretto.NewCache(&ristretto.Config[string, any]{
			NumCounters: 1e5,
			MaxCost:     1e6,
			BufferItems: 64,
		})
		if err != nil {
			log.Fatal("Failed to initialize cache", "err", err)
		}
		stor := ristrettostor.NewRistretto(client, store.WithSynchronousSet(), store.WithCost(1))
		defaultCache = cache.New[any](stor)
	}
}

func getDefault(ctx context.Context) *cache.Cache[any] {
	once.Do(func() {
		initCache(ctx)
	})
	return defaultCache
}

// 只能缓存基本类型
func Set(ctx context.Context, key string, value any) error {
	return getDefault(ctx).Set(ctx, key, value, store.WithExpiration(time.Duration(runtimecfg.Get().Cache.DefaultTTL)*time.Second))
}

func SetWithTTL(ctx context.Context, key string, value any, ttl time.Duration) error {
	return getDefault(ctx).Set(ctx, key, value, store.WithExpiration(ttl))
}

func Get[T any](ctx context.Context, key string) (T, error) {
	got, err := getDefault(ctx).Get(ctx, key)
	if err != nil {
		return *new(T), err
	}
	val, ok := got.(T)
	if !ok {
		return *new(T), nil
	}
	return val, nil
}

// GetWithTTL returns the object stored in cache and its corresponding TTL
func GetWithTTL[T any](ctx context.Context, key string) (T, time.Duration, error) {
	got, ttl, err := getDefault(ctx).GetWithTTL(ctx, key)
	if err != nil {
		return *new(T), 0, err
	}
	val, ok := got.(T)
	if !ok {
		return *new(T), 0, nil
	}
	return val, ttl, nil
}

func Delete(ctx context.Context, key string) error {
	return getDefault(ctx).Delete(ctx, key)
}
