package kvstor

import (
	"sync"
	"time"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/vmihailenco/msgpack/v5"
	"go.etcd.io/bbolt"
)

var (
	defaultDb *bbolt.DB
	once      sync.Once
)

type item[T any] struct {
	Value     T
	CreatedAt time.Time
	Version   int
}

func newItem[T any](value T) item[T] {
	return item[T]{
		Value:     value,
		CreatedAt: time.Now(),
		Version:   1,
	}
}

func initDB() {
	cfg := runtimecfg.Get().KVDB
	dbPath := cfg.Path
	if dbPath == "" {
		dbPath = "data/bbolt.db"
	}
	var err error
	defaultDb, err = bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		log.Fatal("Failed to initialize database", "err", err)
	}
}

func getDefault() *bbolt.DB {
	once.Do(func() {
		initDB()
	})
	return defaultDb
}

func Close() error {
	if defaultDb != nil {
		return defaultDb.Close()
	}
	return nil
}

// 使用 msgpack 序列化
func Set(key string, value any) error {
	db := getDefault()
	item := newItem(value)
	val, err := msgpack.Marshal(item)
	if err != nil {
		return err
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("manyacg"))
		if err != nil {
			return err
		}
		return bucket.Put([]byte(key), val)
	})
	return err
}

func Get[T any](key string) (T, error) {
	var zero T
	db := getDefault()
	var val []byte
	err := db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("manyacg"))
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
	return result.Value, nil
}

func Delete(key string) error {
	db := getDefault()
	return db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("manyacg"))
		if bucket == nil {
			return errs.ErrRecordNotFound
		}
		return bucket.Delete([]byte(key))
	})
}
