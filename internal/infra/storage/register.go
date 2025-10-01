package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/krau/ManyACG/internal/shared"
)

type Factory func() Storage

var (
	storages  = make(map[shared.StorageType]Storage)
	factories = make(map[shared.StorageType]Factory)
	factoryMu sync.RWMutex
)

func Register(storageType shared.StorageType, f Factory) {
	factoryMu.Lock()
	defer factoryMu.Unlock()
	if _, exists := factories[storageType]; exists {
		panic("storage: Register called twice for storage type " + string(storageType))
	}
	factories[storageType] = f
}

func InitAll(ctx context.Context) error {
	factoryMu.Lock()
	defer factoryMu.Unlock()
	for storageType, factory := range factories {
		storage := factory()
		if err := storage.Init(ctx); err != nil {
			return fmt.Errorf("failed to init storage %s: %w", storageType, err)
		}
		storages[storageType] = storage
	}
	return nil
}
