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

func InitAll() {
	factoryMu.Lock()
	defer factoryMu.Unlock()
	for sourceType, factory := range factories {
		source := factory()
		sources[sourceType] = source
	}
}

func Sources() map[shared.SourceType]ArtworkSource {
	factoryMu.RLock()
	defer factoryMu.RUnlock()
	ss := make(map[shared.SourceType]ArtworkSource)
	maps.Copy(ss, sources)
	return ss
}
