package source

import (
	"context"
	"sync"

	"github.com/krau/ManyACG/internal/shared"
)

type Factory func() ArtworkSource

type Config interface {
	Type() shared.SourceType
	Enabled() bool
}

var (
	sources   = make(map[shared.SourceType]ArtworkSource)
	factories = make(map[shared.SourceType]Factory)
	mu        sync.RWMutex
)

func Register(sourceType shared.SourceType, f Factory) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := factories[sourceType]; exists {
		panic("source: Register called twice for source type " + string(sourceType))
	}
	factories[sourceType] = f
}

func InitAll(ctx context.Context, cfgs []Config) error {
	mu.Lock()
	defer mu.Unlock()
	for _, cfg := range cfgs {
		if !cfg.Enabled() {
			continue
		}
		factory, exists := factories[cfg.Type()]
		if !exists {
			continue
		}
		source := factory()
		if err := source.Init(ctx, cfg); err != nil {
			return err
		}
		sources[cfg.Type()] = source
	}
	return nil
}
