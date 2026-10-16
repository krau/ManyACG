package source

import (
	"maps"
	"sync"

	"github.com/krau/ManyACG/internal/shared"
)

type Factory func() ArtworkSource

var (
	sources   = make(map[shared.SourceType]ArtworkSource)
	factories = make(map[shared.SourceType]Factory)
	pollers   = make(map[shared.SourceType]ArtworkPoller)
	factoryMu sync.RWMutex
)

func Register(sourceType shared.SourceType, f Factory) {
	factoryMu.Lock()
	defer factoryMu.Unlock()
	if _, exists := factories[sourceType]; exists {
		panic("source: Register called twice for source type " + string(sourceType))
	}
	factories[sourceType] = f
}

func RegisterPoller(sourceType shared.SourceType, poller ArtworkPoller) {
	factoryMu.Lock()
	defer factoryMu.Unlock()
	if _, exists := pollers[sourceType]; exists {
		panic("source: RegisterPoller called twice for source type " + string(sourceType))
	}
	pollers[sourceType] = poller
}

func Pollers() map[shared.SourceType]ArtworkPoller {
	factoryMu.RLock()
	defer factoryMu.RUnlock()
	return maps.Clone(pollers)
}

func InitAll() {
	factoryMu.RLock()
	registered := maps.Clone(factories)
	factoryMu.RUnlock()
	built := make(map[shared.SourceType]ArtworkSource, len(registered))
	for sourceType, factory := range registered {
		built[sourceType] = factory()
	}
	factoryMu.Lock()
	defer factoryMu.Unlock()
	maps.Copy(sources, built)
}

func Sources() map[shared.SourceType]ArtworkSource {
	factoryMu.RLock()
	defer factoryMu.RUnlock()
	return maps.Clone(sources)
}
