package tile

import (
	"testing"

	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

func TestNewTileProvider(t *testing.T) {
	url := "https://example.com/{z}/{x}/{y}.png"
	provider := NewTileProvider(url, TileProviderModeXYZ)

	if provider == nil {
		t.Fatal("NewTileProvider returned nil")
	}

	if provider.config == nil {
		t.Fatal("provider config is nil")
	}

	if provider.config.Mode != TileProviderModeXYZ {
		t.Errorf("expected XYZ mode, got %d", provider.config.Mode)
	}
}

func TestNewTileProviderTMS(t *testing.T) {
	url := "https://example.com/{z}/{x}/{y}.png"
	provider := NewTileProvider(url, TileProviderModeTMS)

	if provider == nil {
		t.Fatal("NewTileProvider returned nil")
	}

	if provider.config.Mode != TileProviderModeTMS {
		t.Errorf("expected TMS mode, got %d", provider.config.Mode)
	}
}

func TestNewTileProviderWithConfig(t *testing.T) {
	grid := geo.NewTileGrid(map[string]interface{}{
		"srs":       3857,
		"tile_size": []uint32{256, 256},
	})

	config := &TileProviderConfig{
		Mode:        TileProviderModeXYZ,
		URL:         "https://example.com/{z}/{x}/{y}.png",
		MaxRetries:  3,
		Timeout:     30,
		UserAgent:   "TestAgent",
		Attribution: "Test Attribution",
		Grid:        grid,
		Bounds:      vec2d.Rect{Min: vec2d.T{-90, -180}, Max: vec2d.T{90, 180}},
	}

	provider := NewTileProviderWithConfig(config)

	if provider == nil {
		t.Fatal("NewTileProviderWithConfig returned nil")
	}

	if provider.config.Mode != TileProviderModeXYZ {
		t.Errorf("expected XYZ mode, got %d", provider.config.Mode)
	}

	if provider.config.MaxRetries != 3 {
		t.Errorf("expected MaxRetries 3, got %d", provider.config.MaxRetries)
	}

	if provider.grid != grid {
		t.Error("grid not set correctly")
	}
}

func TestTileProviderAttribution(t *testing.T) {
	config := &TileProviderConfig{
		Mode:        TileProviderModeXYZ,
		URL:         "https://example.com/{z}/{x}/{y}.png",
		Attribution: "Test Attribution",
	}

	provider := NewTileProviderWithConfig(config)

	attribution := provider.Attribution()
	if attribution != "Test Attribution" {
		t.Errorf("expected attribution 'Test Attribution', got '%s'", attribution)
	}
}

func TestTileProviderGrid(t *testing.T) {
	grid := geo.NewTileGrid(map[string]interface{}{
		"srs":       3857,
		"tile_size": []uint32{256, 256},
	})

	config := &TileProviderConfig{
		Mode: TileProviderModeXYZ,
		URL:  "https://example.com/{z}/{x}/{y}.png",
		Grid: grid,
	}

	provider := NewTileProviderWithConfig(config)

	retrievedGrid := provider.Grid()
	if retrievedGrid != grid {
		t.Error("grid not retrieved correctly")
	}
}

func TestTileProviderBounds(t *testing.T) {
	provider := NewTileProvider("https://example.com/{z}/{x}/{y}.png", TileProviderModeXYZ)

	retrievedBounds := provider.Bounds()

	expectedMin := vec2d.T{-85.0, -180.0}
	expectedMax := vec2d.T{85.0, 180.0}

	if retrievedBounds.Min != expectedMin || retrievedBounds.Max != expectedMax {
		t.Errorf("expected bounds [%v, %v], got [%v, %v]", expectedMin, expectedMax, retrievedBounds.Min, retrievedBounds.Max)
	}
}

func TestTileProviderSrs(t *testing.T) {
	provider := NewTileProvider("https://example.com/{z}/{x}/{y}.png", TileProviderModeXYZ)

	retrievedSrs := provider.Srs()

	if retrievedSrs == nil {
		t.Error("SRS should not be nil")
	}
}

func TestTileProviderModeString(t *testing.T) {
	if TileProviderModeXYZ != 0 {
		t.Errorf("TileProviderModeXYZ should be 0, got %d", TileProviderModeXYZ)
	}

	if TileProviderModeTMS != 1 {
		t.Errorf("TileProviderModeTMS should be 1, got %d", TileProviderModeTMS)
	}
}

func TestTileProviderConfigDefaults(t *testing.T) {
	provider := NewTileProvider("https://example.com/{z}/{x}/{y}.png", TileProviderModeXYZ)

	if provider.client == nil {
		t.Error("HTTP client should not be nil")
	}

	if provider.config.Headers == nil {
		t.Error("Headers map should be initialized")
	}
}

func TestTileProviderBuildTileURL(t *testing.T) {
	provider := NewTileProvider("https://example.com/{z}/{x}/{y}.png", TileProviderModeXYZ)

	url := provider.buildTileURL(1, 2, 3)

	expectedURL := "https://example.com/3/1/2.png"
	if url != expectedURL {
		t.Errorf("expected URL '%s', got '%s'", expectedURL, url)
	}
}
