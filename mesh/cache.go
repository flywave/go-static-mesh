package mesh

import (
	"container/list"
	"sync"
	"time"
)

type TileCache interface {
	Get(key [3]int) (*Tile, bool)
	Put(key [3]int, tile *Tile)
	Clear()
	Stats() CacheStats
}

type CacheStats struct {
	Hits      int
	Misses    int
	Size      int
	MaxSize   int
	Evictions int
}

type MemoryTileCache struct {
	data     map[[3]int]*list.Element
	lru      *list.List
	maxSize  int
	stats    CacheStats
	mu       sync.RWMutex
	ttl      time.Duration
	expiries map[[3]int]time.Time
}

func NewMemoryTileCache(maxSize int, ttl time.Duration) *MemoryTileCache {
	return &MemoryTileCache{
		data:     make(map[[3]int]*list.Element),
		lru:      list.New(),
		maxSize:  maxSize,
		stats:    CacheStats{MaxSize: maxSize},
		ttl:      ttl,
		expiries: make(map[[3]int]time.Time),
	}
}

func (c *MemoryTileCache) Get(key [3]int) (*Tile, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ttl > 0 {
		if expiry, exists := c.expiries[key]; exists && time.Now().After(expiry) {
			c.remove(key)
			c.stats.Misses++
			return nil, false
		}
	}

	elem, exists := c.data[key]
	if !exists {
		c.stats.Misses++
		return nil, false
	}

	c.lru.MoveToFront(elem)
	c.stats.Hits++
	return elem.Value.(*Tile), true
}

func (c *MemoryTileCache) Put(key [3]int, tile *Tile) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ttl > 0 {
		c.expiries[key] = time.Now().Add(c.ttl)
	}

	if elem, exists := c.data[key]; exists {
		elem.Value = tile
		c.lru.MoveToFront(elem)
		return
	}

	if c.lru.Len() >= c.maxSize {
		c.evict()
	}

	elem := c.lru.PushFront(tile)
	c.data[key] = elem
	c.stats.Size = c.lru.Len()
}

func (c *MemoryTileCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[[3]int]*list.Element)
	c.lru.Init()
	c.expiries = make(map[[3]int]time.Time)
	c.stats = CacheStats{MaxSize: c.stats.MaxSize}
}

func (c *MemoryTileCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

func (c *MemoryTileCache) remove(key [3]int) {
	if elem, exists := c.data[key]; exists {
		c.lru.Remove(elem)
		delete(c.data, key)
		delete(c.expiries, key)
		c.stats.Size = c.lru.Len()
	}
}

func (c *MemoryTileCache) evict() {
	if c.lru.Len() == 0 {
		return
	}

	elem := c.lru.Back()
	if elem != nil {
		tile := elem.Value.(*Tile)
		key := [3]int{tile.Coord[0], tile.Coord[1], tile.Coord[2]}
		c.lru.Remove(elem)
		delete(c.data, key)
		delete(c.expiries, key)
		c.stats.Evictions++
		c.stats.Size = c.lru.Len()
	}
}
