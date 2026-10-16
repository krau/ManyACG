package kvstor

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
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

func Open(cfg runtimecfg.KVDBConfig) (KVStore, error) {
	if cfg.Type == "redis" {
		rc := cfg.Redis
		var opt rueidis.ClientOption
		if rc.URL != "" {
			var err error
			opt, err = rueidis.ParseURL(rc.URL)
			if err != nil {
				return nil, err
			}
		} else {
			opt = rueidis.ClientOption{
				InitAddress: rc.Addrs,
				Username:    rc.Username,
				Password:    rc.Password,
				SelectDB:    rc.DB,
			}
		}
		if rc.TLS {
			opt.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: rc.TLSInsecure}
		}
		client, err := rueidis.NewClient(opt)
		if err != nil {
			return nil, err
		}
		return &redisDB{client: client, prefix: rc.Prefix}, nil
	}
	bdb, err := bbolt.Open(cfg.Path, 0600, nil)
	if err != nil {
		return nil, err
	}
	db := &bboltDB{
		db:             bdb,
		bucket:         cfg.Bucket,
		ttlBucket:      cfg.TTLBucket,
		ttlBatchLimit:  cfg.TTLBatchLimit,
		ttlSweepPeriod: time.Duration(cfg.TTLSweepPeriod) * time.Second,
		stop:           make(chan struct{}),
	}
	db.startTTLReaper()
	return db, nil
}

func Get[T any](ctx context.Context, store KVStore, key string) (T, error) {
	var zero T
	val, err := store.Get(ctx, key)
	if err != nil {
		return zero, err
	}
	typedVal, ok := val.(T)
	if !ok {
		return zero, oops.New("type assertion failed in kvstor.Get")
	}
	return typedVal, nil
}
