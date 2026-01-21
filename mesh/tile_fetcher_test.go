package mesh

import (
	"image"
	"image/color"
	"testing"

	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

type MockProvider struct{}

func (m *MockProvider) GetImageTile(coord [3]int) (image.Image, error) {
	img := image.NewRGBA(image.Rect(0, 0, 256, 256))
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x % 256),
				G: uint8(y % 256),
				B: uint8((coord[0] + coord[1]) % 256),
				A: 255,
			})
		}
	}
	return img, nil
}

func TestNewTileFetcher(t *testing.T) {
	provider := &MockProvider{}
	fetcher := NewTileFetcher(provider)

	if fetcher == nil {
		t.Error("NewTileFetcher returned nil")
	}

	if fetcher.provider != provider {
		t.Error("Provider not set correctly")
	}
}

func TestTileFetcherFetchTiles(t *testing.T) {
	provider := &MockProvider{}
	fetcher := NewTileFetcher(provider)

	bounds := vec2d.Rect{
		Min: vec2d.T{39.9, -75.2},
		Max: vec2d.T{40.1, -74.8},
	}

	srs := geo.NewProj(4326)
	zoom := 13

	tiles, err := fetcher.FetchTiles(bounds, zoom, srs)
	if err != nil {
		t.Errorf("FetchTiles returned error: %v", err)
	}

	if len(tiles) == 0 {
		t.Error("FetchTiles returned no tiles")
	}

	for _, tile := range tiles {
		if tile.Image == nil {
			t.Errorf("Tile %v has nil image", tile.Coord)
		}

		if tile.Image.Bounds().Dx() != 256 || tile.Image.Bounds().Dy() != 256 {
			t.Errorf("Tile %v has unexpected size", tile.Coord)
		}
	}
}

func TestTileFetcherCalculateTileCoords(t *testing.T) {
	provider := &MockProvider{}
	fetcher := NewTileFetcher(provider)

	bounds := vec2d.Rect{
		Min: vec2d.T{39.9, -75.2},
		Max: vec2d.T{40.1, -74.8},
	}

	srs := geo.NewProj(4326)
	zoom := 13

	coords := fetcher.calculateTileCoords(bounds, zoom, srs)

	if len(coords) == 0 {
		t.Error("calculateTileCoords returned no coordinates")
	}

	for _, coord := range coords {
		if coord[2] != zoom {
			t.Errorf("Expected zoom %d, got %d", zoom, coord[2])
		}

		if coord[0] < 0 || coord[1] < 0 {
			t.Errorf("Invalid tile coordinates: %v", coord)
		}
	}
}

func TestTileFetcherCalculateTileBounds(t *testing.T) {
	provider := &MockProvider{}
	fetcher := NewTileFetcher(provider)

	coord := [3]int{100, 200, 14}
	srs := geo.NewProj(4326)

	bounds := fetcher.calculateTileBounds(coord, coord[2], srs)

	if bounds.Min[0] >= bounds.Max[0] {
		t.Error("Invalid latitude bounds")
	}

	if bounds.Min[1] >= bounds.Max[1] {
		t.Error("Invalid longitude bounds")
	}

	if bounds.Min[0] < -85.0511 || bounds.Max[0] > 85.0511 {
		t.Error("Latitude bounds out of valid range")
	}

	if bounds.Min[1] < -180 || bounds.Max[1] > 180 {
		t.Error("Longitude bounds out of valid range")
	}
}

func TestTileFetcherFetchTilesEmptyBounds(t *testing.T) {
	provider := &MockProvider{}
	fetcher := NewTileFetcher(provider)

	bounds := vec2d.Rect{
		Min: vec2d.T{0, 0},
		Max: vec2d.T{0, 0},
	}

	srs := geo.NewProj(4326)
	zoom := 14

	tiles, err := fetcher.FetchTiles(bounds, zoom, srs)
	if err != nil {
		t.Logf("FetchTiles returned error for empty bounds: %v", err)
	}

	if tiles == nil {
		t.Error("Expected tiles slice, got nil")
	}

	if len(tiles) == 0 {
		t.Log("No tiles for empty bounds (expected behavior)")
	}
}

func TestTileFetcherFetchTilesOutOfBounds(t *testing.T) {
	provider := &MockProvider{}
	fetcher := NewTileFetcher(provider)

	bounds := vec2d.Rect{
		Min: vec2d.T{-90, -180},
		Max: vec2d.T{90, 180},
	}

	srs := geo.NewProj(4326)
	zoom := 2

	tiles, err := fetcher.FetchTiles(bounds, zoom, srs)
	if err != nil {
		t.Errorf("FetchTiles returned error for large bounds: %v", err)
	}

	if len(tiles) == 0 {
		t.Error("FetchTiles returned no tiles")
	}
}
