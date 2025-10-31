package kvstor

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/redis/rueidis"
	"github.com/samber/oops"
	"go.etcd.io/bbolt"
)

type KVStore interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string) (any, error)
	Delete(ctx context.Context, key string) error
	Close() error
}

var (
	defaultDb  KVStore
	initOnce   sync.Once
	reaperStop chan struct{}
)

func Init(cfg runtimecfg.KVDBConfig) {
	switch cfg.Type {
	case "bbolt":
		dbPath := cfg.Path
		initOnce.Do(func() {
			bdb, err := bbolt.Open(dbPath, 0600, nil)
			if err != nil {
				log.Fatal("Failed to initialize kvdb", "err", err)
			}
			reaperStop = make(chan struct{})
			bbdb := &bboltDB{
				db:             bdb,
				bucket:         cfg.Bucket,
				ttlBucket:      cfg.TTLBucket,
				ttlBatchLimit:  cfg.TTLBatchLimit,
				ttlSweepPeriod: time.Duration(cfg.TTLSweepPeriod) * time.Second,
				stop:           reaperStop,
			}
			defaultDb = bbdb
			bbdb.startTTLReaper()
		})
	case "redis":
		initOnce.Do(func() {
			var (
				client rueidis.Client
				err    error
			)
			rc := cfg.Redis
			if rc.URL != "" {
				opt := rueidis.MustParseURL(rc.URL)
				if rc.TLS {
					opt.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: rc.TLSInsecure}
				}
				client, err = rueidis.NewClient(opt)
			} else {
				opt := rueidis.ClientOption{
					InitAddress: rc.Addrs,
					Username:    rc.Username,
					Password:    rc.Password,
					SelectDB:    rc.DB,
				}
				if rc.TLS {
					opt.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: rc.TLSInsecure}
				}
				client, err = rueidis.NewClient(opt)
			}
			if err != nil {
				log.Fatal("Failed to initialize redis kvdb", "err", err)
			}
			defaultDb = &redisDB{client: client}
		})
	default:
		dbPath := cfg.Path
		initOnce.Do(func() {
			bdb, err := bbolt.Open(dbPath, 0600, nil)
			if err != nil {
				log.Fatal("Failed to initialize kvdb", "err", err)
			}
			reaperStop = make(chan struct{})
			bbdb := &bboltDB{
				db:             bdb,
				bucket:         cfg.Bucket,
				ttlBucket:      cfg.TTLBucket,
				ttlBatchLimit:  cfg.TTLBatchLimit,
				ttlSweepPeriod: time.Duration(cfg.TTLSweepPeriod) * time.Second,
				stop:           reaperStop,
			}
			defaultDb = bbdb
			bbdb.startTTLReaper()
		})
	}

}

func Close() error {
	if defaultDb != nil {
		if stop := reaperStop; stop != nil {
			close(stop)
		}
		return defaultDb.Close()
	}
	return nil
}

func Set(ctx context.Context, key string, value any) error {
	return defaultDb.Set(ctx, key, value, 0)
}

// SetWithTTL stores the value with a given TTL; a non-positive TTL behaves like Set.
func SetWithTTL(ctx context.Context, key string, value any, ttl time.Duration) error {
	if ttl <= 0 {
		return Set(ctx, key, value)
	}
	return defaultDb.Set(ctx, key, value, ttl)
}

func Get[T any](ctx context.Context, key string) (T, error) {
	var zero T
	val, err := defaultDb.Get(ctx, key)
	if err != nil {
		return zero, err
	}
	typedVal, ok := val.(T)
	if !ok {
		return zero, oops.New("type assertion failed in kvstor.Get")
	}
	return typedVal, nil
}

func Delete(ctx context.Context, key string) error {
	return defaultDb.Delete(ctx, key)
}
