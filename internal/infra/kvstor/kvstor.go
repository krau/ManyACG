package kvstor

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/redis/rueidis"
	"github.com/samber/oops"
	"github.com/vmihailenco/msgpack/v5"
	"go.etcd.io/bbolt"
)

type KVStore interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string) (any, error)
	Delete(ctx context.Context, key string) error
	Close() error
}

func Open(cfg runtimecfg.KVDBConfig) (KVStore, error) {
	switch cfg.Type {
	case "bbolt":
		return openBbolt(cfg)
	case "redis":
		return openRedis(cfg)
	default:
		return nil, oops.Errorf("unsupported kvdb type %q", cfg.Type)
	}
}

func openRedis(cfg runtimecfg.KVDBConfig) (KVStore, error) {
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

func openBbolt(cfg runtimecfg.KVDBConfig) (KVStore, error) {
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
	if ok {
		return typedVal, nil
	}
	// Values are decoded back from msgpack as maps, so struct types need a round trip.
	raw, err := msgpack.Marshal(val)
	if err != nil {
		return zero, oops.Wrapf(err, "failed to encode value of %T", val)
	}
	var decoded T
	if err := msgpack.Unmarshal(raw, &decoded); err != nil {
		return zero, oops.New("type assertion failed in kvstor.Get")
	}
	return decoded, nil
}
