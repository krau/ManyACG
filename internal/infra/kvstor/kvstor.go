package kvstor

import (
	"errors"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	defaultDb *badger.DB
	once      sync.Once
)

func initDB() {
	cfg := runtimecfg.Get().KVDB
	switch cfg.Type {
	case "badger":
		var err error
		defaultDb, err = badger.Open(badger.DefaultOptions(cfg.Path).WithLoggingLevel(badger.WARNING))
		if err != nil {
			log.Fatal("Failed to initialize database", "err", err)
		}
	default:
		var err error
		defaultDb, err = badger.Open(badger.DefaultOptions("data/manyacg_kvdb").WithLoggingLevel(badger.WARNING))
		if err != nil {
			log.Fatal("Failed to initialize database", "err", err)
		}
	}
}

func getDefault() *badger.DB {
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
	val, err := msgpack.Marshal(value)
	if err != nil {
		return err
	}
	return db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), val)
		return err
	})
}

func Get[T any](key string) (T, error) {
	var zero T
	db := getDefault()
	var val []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if errors.Is(err, badger.ErrKeyNotFound) {
			return errs.ErrRecordNotFound
		}
		if err != nil {
			return err
		}
		val, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		return zero, err
	}
	var result T
	if err := msgpack.Unmarshal(val, &result); err != nil {
		return zero, err
	}
	return result, nil
}

func Delete(key string) error {
	db := getDefault()
	return db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(key))
		return err
	})
}

func Has(key string) (bool, error) {
	db := getDefault()
	err := db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(key))
		return err
	})
	if err == badger.ErrKeyNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
