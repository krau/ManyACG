package objectuuid

import (
	"slices"
	"sync"
)

type ObjectUUIDs struct {
	data []ObjectUUID
	mu   sync.RWMutex
}

func NewObjectUUIDs(ids ...ObjectUUID) *ObjectUUIDs {
	m := &ObjectUUIDs{
		data: make([]ObjectUUID, 0),
	}
	m.UnsafeAdd(ids...)
	return m
}

func (m *ObjectUUIDs) UnsafeAdd(ids ...ObjectUUID) {
	if m.data == nil {
		m.data = make([]ObjectUUID, 0)
	}
	for _, id := range ids {
		if !slices.Contains(m.data, id) {
			m.data = append(m.data, id)
		}
	}
}

func (m *ObjectUUIDs) UnsafeRemove(ids ...ObjectUUID) {
	if len(m.data) == 0 {
		return
	}
	toRemove := make(map[ObjectUUID]struct{})
	for _, id := range ids {
		toRemove[id] = struct{}{}
	}
	newIDs := make([]ObjectUUID, 0)
	for _, id := range m.data {
		if _, found := toRemove[id]; !found {
			newIDs = append(newIDs, id)
		}
	}
	m.data = newIDs
}

func (m *ObjectUUIDs) Add(ids ...ObjectUUID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.data == nil {
		m.data = make([]ObjectUUID, 0)
	}
	for _, id := range ids {
		if !slices.Contains(m.data, id) {
			m.data = append(m.data, id)
		}
	}
}

func (m *ObjectUUIDs) Remove(ids ...ObjectUUID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.data) == 0 {
		return
	}
	toRemove := make(map[ObjectUUID]struct{}, len(ids))
	for _, id := range ids {
		toRemove[id] = struct{}{}
	}
	newIDs := make([]ObjectUUID, 0)
	for _, id := range m.data {
		if _, found := toRemove[id]; !found {
			newIDs = append(newIDs, id)
		}
	}
	m.data = newIDs
}

func (m *ObjectUUIDs) IsEmpty() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data) == 0
}

func (m *ObjectUUIDs) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data)
}

func (m *ObjectUUIDs) UnsafeValue() []ObjectUUID {
	return m.data
}

func (m *ObjectUUIDs) Contains(id ObjectUUID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.data == nil {
		return false
	}
	return slices.Contains(m.data, id)
}
