package mesh

import (
	"sync"
	"testing"
	"time"
)

func TestMemoryTileCache_BasicOperations(t *testing.T) {
	cache := NewMemoryTileCache(10, 0)

	tile1 := &Tile{
		Coord: [3]int{0, 0, 0},
		Zoom:  0,
		Data:  []float64{1.0, 2.0, 3.0},
	}

	tile2 := &Tile{
		Coord: [3]int{1, 1, 0},
		Zoom:  0,
		Data:  []float64{4.0, 5.0, 6.0},
	}

	cache.Put([3]int{0, 0, 0}, tile1)
	cache.Put([3]int{1, 1, 0}, tile2)

	retrieved, found := cache.Get([3]int{0, 0, 0})
	if !found {
		t.Fatal("expected to find tile in cache")
	}
	if retrieved != tile1 {
		t.Fatal("retrieved tile does not match original")
	}

	retrieved, found = cache.Get([3]int{1, 1, 0})
	if !found {
		t.Fatal("expected to find tile in cache")
	}
	if retrieved != tile2 {
		t.Fatal("retrieved tile does not match original")
	}

	_, found = cache.Get([3]int{2, 2, 0})
	if found {
		t.Fatal("expected not to find non-existent tile")
	}

	stats := cache.Stats()
	if stats.Size != 2 {
		t.Fatalf("expected cache size 2, got %d", stats.Size)
	}
}

func TestMemoryTileCache_LRUEviction(t *testing.T) {
	cache := NewMemoryTileCache(3, 0)

	for i := 0; i < 3; i++ {
		tile := &Tile{
			Coord: [3]int{i, i, 0},
			Zoom:  0,
			Data:  []float64{float64(i)},
		}
		cache.Put([3]int{i, i, 0}, tile)
	}

	stats := cache.Stats()
	if stats.Size != 3 {
		t.Fatalf("expected cache size 3, got %d", stats.Size)
	}

	tile4 := &Tile{
		Coord: [3]int{3, 3, 0},
		Zoom:  0,
		Data:  []float64{4.0},
	}
	cache.Put([3]int{3, 3, 0}, tile4)

	stats = cache.Stats()
	if stats.Size != 3 {
		t.Fatalf("expected cache size 3 after eviction, got %d", stats.Size)
	}

	if stats.Evictions != 1 {
		t.Fatalf("expected 1 eviction, got %d", stats.Evictions)
	}

	_, found := cache.Get([3]int{0, 0, 0})
	if found {
		t.Fatal("expected first tile to be evicted")
	}

	_, found = cache.Get([3]int{3, 3, 0})
	if !found {
		t.Fatal("expected new tile to be in cache")
	}
}

func TestMemoryTileCache_HitsAndMisses(t *testing.T) {
	cache := NewMemoryTileCache(10, 0)

	tile := &Tile{
		Coord: [3]int{0, 0, 0},
		Zoom:  0,
		Data:  []float64{1.0},
	}

	cache.Put([3]int{0, 0, 0}, tile)

	_, found := cache.Get([3]int{0, 0, 0})
	if !found {
		t.Fatal("expected hit")
	}

	_, found = cache.Get([3]int{1, 1, 0})
	if found {
		t.Fatal("expected miss")
	}

	_, found = cache.Get([3]int{0, 0, 0})
	if !found {
		t.Fatal("expected hit")
	}

	stats := cache.Stats()
	if stats.Hits != 2 {
		t.Fatalf("expected 2 hits, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Fatalf("expected 1 miss, got %d", stats.Misses)
	}
}

func TestMemoryTileCache_Clear(t *testing.T) {
	cache := NewMemoryTileCache(10, 0)

	for i := 0; i < 5; i++ {
		tile := &Tile{
			Coord: [3]int{i, i, 0},
			Zoom:  0,
			Data:  []float64{float64(i)},
		}
		cache.Put([3]int{i, i, 0}, tile)
	}

	stats := cache.Stats()
	if stats.Size != 5 {
		t.Fatalf("expected cache size 5, got %d", stats.Size)
	}

	cache.Clear()

	stats = cache.Stats()
	if stats.Size != 0 {
		t.Fatalf("expected cache size 0 after clear, got %d", stats.Size)
	}

	_, found := cache.Get([3]int{0, 0, 0})
	if found {
		t.Fatal("expected not to find tile after clear")
	}
}

func TestMemoryTileCache_TTL(t *testing.T) {
	cache := NewMemoryTileCache(10, 100*time.Millisecond)

	tile := &Tile{
		Coord: [3]int{0, 0, 0},
		Zoom:  0,
		Data:  []float64{1.0},
	}
	cache.Put([3]int{0, 0, 0}, tile)

	_, found := cache.Get([3]int{0, 0, 0})
	if !found {
		t.Fatal("expected to find tile immediately after put")
	}

	time.Sleep(150 * time.Millisecond)

	_, found = cache.Get([3]int{0, 0, 0})
	if found {
		t.Fatal("expected tile to expire after TTL")
	}

	stats := cache.Stats()
	if stats.Misses != 1 {
		t.Fatalf("expected 1 miss after expiry, got %d", stats.Misses)
	}
}

func TestMemoryTileCache_ConcurrentAccess(t *testing.T) {
	cache := NewMemoryTileCache(1000, 0)
	var wg sync.WaitGroup

	numGoroutines := 10
	numOperations := 20

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for i := 0; i < numOperations; i++ {
				key := [3]int{goroutineID, i, 0}
				tile := &Tile{
					Coord: key,
					Zoom:  0,
					Data:  []float64{float64(goroutineID*numOperations + i)},
				}
				cache.Put(key, tile)

				_, found := cache.Get(key)
				if !found {
					t.Errorf("expected to find tile for key %v", key)
				}
			}
		}(g)
	}

	wg.Wait()

	stats := cache.Stats()
	if stats.Size != numGoroutines*numOperations {
		t.Fatalf("expected cache size %d, got %d", numGoroutines*numOperations, stats.Size)
	}
}

func TestMemoryTileCache_UpdateExisting(t *testing.T) {
	cache := NewMemoryTileCache(10, 0)

	tile1 := &Tile{
		Coord: [3]int{0, 0, 0},
		Zoom:  0,
		Data:  []float64{1.0},
	}

	tile2 := &Tile{
		Coord: [3]int{0, 0, 0},
		Zoom:  0,
		Data:  []float64{2.0},
	}

	cache.Put([3]int{0, 0, 0}, tile1)

	retrieved, found := cache.Get([3]int{0, 0, 0})
	if !found {
		t.Fatal("expected to find tile")
	}
	if retrieved.Data[0] != 1.0 {
		t.Fatalf("expected data 1.0, got %f", retrieved.Data[0])
	}

	cache.Put([3]int{0, 0, 0}, tile2)

	retrieved, found = cache.Get([3]int{0, 0, 0})
	if !found {
		t.Fatal("expected to find tile after update")
	}
	if retrieved.Data[0] != 2.0 {
		t.Fatalf("expected data 2.0 after update, got %f", retrieved.Data[0])
	}

	stats := cache.Stats()
	if stats.Size != 1 {
		t.Fatalf("expected cache size 1, got %d", stats.Size)
	}
}
