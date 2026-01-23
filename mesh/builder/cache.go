package builder

import "sync"

type TileCache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}) error
	Clear() error
	GetSize() int
}

type MemoryTileCache struct {
	cache   map[string]interface{}
	mu      sync.RWMutex
	maxSize int
}

func NewMemoryTileCache(maxSize int) *MemoryTileCache {
	return &MemoryTileCache{
		cache:   make(map[string]interface{}),
		maxSize: maxSize,
	}
}

func (m *MemoryTileCache) Get(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.cache[key]
	return val, ok
}

func (m *MemoryTileCache) Set(key string, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.maxSize > 0 && len(m.cache) >= m.maxSize {
		m.evict()
	}

	m.cache[key] = value
	return nil
}

func (m *MemoryTileCache) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache = make(map[string]interface{})
	return nil
}

func (m *MemoryTileCache) GetSize() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.cache)
}

func (m *MemoryTileCache) evict() {
	for key := range m.cache {
		delete(m.cache, key)
		break
	}
}
