package kvstor

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/vmihailenco/msgpack/v5"
	"go.etcd.io/bbolt"
)

type bboltDB struct {
	db             *bbolt.DB
	bucket         string
	ttlBucket      string
	ttlBatchLimit  int
	ttlSweepPeriod time.Duration
}

var (
	defaultDb  *bboltDB
	initOnce   sync.Once
	reaperOnce sync.Once
	reaperStop chan struct{}
)

type item[T any] struct {
	Value     T
	CreatedAt time.Time
	ExpiresAt int64
}

func newItem[T any](value T) item[T] {
	return item[T]{
		Value:     value,
		CreatedAt: time.Now(),
	}
}

func newItemWithTTL[T any](value T, ttl time.Duration) item[T] {
	item := newItem(value)
	if ttl > 0 {
		item.ExpiresAt = item.CreatedAt.Add(ttl).UnixNano()
	}
	return item
}

func Init(cfg runtimecfg.KVDBConfig) {
	dbPath := cfg.Path
	initOnce.Do(func() {
		bdb, err := bbolt.Open(dbPath, 0600, nil)
		if err != nil {
			log.Fatal("Failed to initialize kvdb", "err", err)
		}
		defaultDb = &bboltDB{
			db:             bdb,
			bucket:         cfg.Bucket,
			ttlBucket:      cfg.TTLBucket,
			ttlBatchLimit:  cfg.TTLBatchLimit,
			ttlSweepPeriod: time.Duration(cfg.TTLSweepPeriod) * time.Second,
		}
		startTTLReaper(defaultDb)
	})
}

func Close() error {
	if defaultDb != nil {
		if stop := reaperStop; stop != nil {
			close(stop)
		}
		return defaultDb.db.Close()
	}
	return nil
}

// 使用 msgpack 序列化
func Set(key string, value any) error {
	return set(key, value, 0)
}

// SetWithTTL stores the value with a given TTL; a non-positive TTL behaves like Set.
func SetWithTTL(key string, value any, ttl time.Duration) error {
	if ttl <= 0 {
		return Set(key, value)
	}
	return set(key, value, ttl)
}

func Get[T any](key string) (T, error) {
	var zero T
	db := defaultDb.db
	var val []byte
	err := db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(defaultDb.bucket))
		if bucket == nil {
			return errs.ErrRecordNotFound
		}
		val = bucket.Get([]byte(key))
		if val == nil {
			return errs.ErrRecordNotFound
		}
		return nil
	})
	if err != nil {
		return zero, err
	}
	var result item[T]
	if err := msgpack.Unmarshal(val, &result); err != nil {
		return zero, err
	}
	if result.ExpiresAt > 0 && time.Now().UnixNano() > result.ExpiresAt {
		if err := deleteExpired(key, result.ExpiresAt); err != nil {
			log.Warn("kvstore ttl cleanup failed", "key", key, "err", err)
		}
		return zero, errs.ErrRecordNotFound
	}
	return result.Value, nil
}

func Delete(key string) error {
	db := defaultDb.db
	return db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(defaultDb.bucket))
		if bucket == nil {
			return errs.ErrRecordNotFound
		}
		raw := bucket.Get([]byte(key))
		if raw != nil {
			var existing item[any]
			if err := msgpack.Unmarshal(raw, &existing); err == nil && existing.ExpiresAt > 0 {
				ttlBucket := tx.Bucket([]byte(defaultDb.ttlBucket))
				if ttlBucket != nil {
					_ = ttlBucket.Delete(encodeTTLKey(existing.ExpiresAt, key))
				}
			}
		}
		return bucket.Delete([]byte(key))
	})
}

func set[T any](key string, value T, ttl time.Duration) error {
	db := defaultDb.db
	return db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(defaultDb.bucket))
		if err != nil {
			return err
		}
		ttlBucket, err := tx.CreateBucketIfNotExists([]byte(defaultDb.ttlBucket))
		if err != nil {
			return err
		}

		if prev := bucket.Get([]byte(key)); prev != nil {
			var existing item[any]
			if err := msgpack.Unmarshal(prev, &existing); err == nil && existing.ExpiresAt > 0 {
				if err := ttlBucket.Delete(encodeTTLKey(existing.ExpiresAt, key)); err != nil {
					return err
				}
			}
		}

		entry := newItemWithTTL(value, ttl)
		val, err := msgpack.Marshal(entry)
		if err != nil {
			return err
		}
		if err := bucket.Put([]byte(key), val); err != nil {
			return err
		}
		if entry.ExpiresAt > 0 {
			if err := ttlBucket.Put(encodeTTLKey(entry.ExpiresAt, key), []byte{1}); err != nil {
				return err
			}
		}
		return nil
	})
}

func deleteExpired(key string, expiresAt int64) error {
	db := defaultDb.db
	return db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(defaultDb.bucket))
		if bucket == nil {
			return nil
		}
		ttlBucket := tx.Bucket([]byte(defaultDb.ttlBucket))
		if ttlBucket != nil {
			ttlBucket.Delete(encodeTTLKey(expiresAt, key))
		}
		return bucket.Delete([]byte(key))
	})
}

func startTTLReaper(db *bboltDB) {
	reaperOnce.Do(func() {
		reaperStop = make(chan struct{})
		go func() {
			ticker := time.NewTicker(db.ttlSweepPeriod)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					if err := sweepExpired(db); err != nil {
						log.Warn("kvstore ttl sweep failed", "err", err)
					}
				case <-reaperStop:
					return
				}
			}
		}()
	})
}

func sweepExpired(db *bboltDB) error {
	now := time.Now().UnixNano()
	return db.db.Update(func(tx *bbolt.Tx) error {
		ttlBucket := tx.Bucket([]byte(defaultDb.ttlBucket))
		if ttlBucket == nil {
			return nil
		}
		bucket := tx.Bucket([]byte(defaultDb.bucket))
		if bucket == nil {
			return nil
		}
		cursor := ttlBucket.Cursor()
		processed := 0
		for k, _ := cursor.First(); k != nil && processed < db.ttlBatchLimit; k, _ = cursor.Next() {
			if len(k) < 8 {
				continue
			}
			expiresAt := int64(binary.BigEndian.Uint64(k[:8]))
			if expiresAt > now {
				break
			}
			rawKey := k[8:]
			if err := bucket.Delete(rawKey); err != nil {
				return err
			}
			if err := cursor.Delete(); err != nil {
				return err
			}
			processed++
		}
		return nil
	})
}

func encodeTTLKey(expiresAt int64, key string) []byte {
	buf := make([]byte, 8+len(key))
	binary.BigEndian.PutUint64(buf[:8], uint64(expiresAt))
	copy(buf[8:], key)
	return buf
}
