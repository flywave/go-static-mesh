package tile

import (
	"testing"
	"time"

	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

func TestNewCesiumQuantizedMeshProvider(t *testing.T) {
	provider := NewCesiumQuantizedMeshProvider("https://terrain.example.com/{z}/{x}/{y}.terrain")

	if provider == nil {
		t.Fatal("NewCesiumQuantizedMeshProvider returned nil")
	}

	if provider.url != "https://terrain.example.com/{z}/{x}/{y}.terrain" {
		t.Errorf("Expected URL 'https://terrain.example.com/{z}/{x}/{y}.terrain', got '%s'", provider.url)
	}

	if provider.srs.Eq(geo.NewProj(4326)) != true {
		t.Error("Expected SRS to be EPSG:4326")
	}

	if provider.maxRetries != 3 {
		t.Errorf("Expected maxRetries 3, got %d", provider.maxRetries)
	}

	if provider.maxCacheSize != 100 {
		t.Errorf("Expected maxCacheSize 100, got %d", provider.maxCacheSize)
	}
}

func TestNewCesiumQuantizedMeshProviderWithConfig(t *testing.T) {
	config := NewDecoderConfig()
	config.MaxRetries = 5
	config.Timeout = 60
	config.VertexNormals = true
	config.WaterMask = false
	config.Metadata = false

	provider := NewCesiumQuantizedMeshProviderWithConfig(
		"https://terrain.example.com/{z}/{x}/{y}.terrain",
		config,
	)

	if provider == nil {
		t.Fatal("NewCesiumQuantizedMeshProviderWithConfig returned nil")
	}

	if provider.maxRetries != 5 {
		t.Errorf("Expected maxRetries 5, got %d", provider.maxRetries)
	}

	if provider.timeout != 60*time.Second {
		t.Errorf("Expected timeout 60s, got %v", provider.timeout)
	}

	if provider.extensionFlag != 1 {
		t.Errorf("Expected extensionFlag 1 (Ext_Light), got %d", provider.extensionFlag)
	}
}

func TestCesiumQuantizedMeshProvider_Attribution(t *testing.T) {
	provider := NewCesiumQuantizedMeshProvider("https://terrain.example.com/{z}/{x}/{y}.terrain")
	attribution := provider.Attribution()

	if attribution != "" {
		t.Errorf("Expected empty attribution, got '%s'", attribution)
	}
}

func TestCesiumQuantizedMeshProvider_Grid(t *testing.T) {
	provider := NewCesiumQuantizedMeshProvider("https://terrain.example.com/{z}/{x}/{y}.terrain")
	grid := provider.Grid()

	if grid == nil {
		t.Fatal("Grid returned nil")
	}
}

func TestCesiumQuantizedMeshProvider_Bounds(t *testing.T) {
	provider := NewCesiumQuantizedMeshProvider("https://terrain.example.com/{z}/{x}/{y}.terrain")
	bounds := provider.Bounds()

	if bounds.Min[0] != -85.0 || bounds.Min[1] != -180.0 {
		t.Errorf("Expected Min [-85.0, -180.0], got %v", bounds.Min)
	}

	if bounds.Max[0] != 85.0 || bounds.Max[1] != 180.0 {
		t.Errorf("Expected Max [85.0, 180.0], got %v", bounds.Max)
	}
}

func TestCesiumQuantizedMeshProvider_Srs(t *testing.T) {
	provider := NewCesiumQuantizedMeshProvider("https://terrain.example.com/{z}/{x}/{y}.terrain")
	srs := provider.Srs()

	if srs == nil {
		t.Fatal("Srs returned nil")
	}

	if !srs.Eq(geo.NewProj(4326)) {
		t.Error("Expected SRS to be EPSG:4326")
	}
}

func TestCesiumQuantizedMeshProvider_CacheManagement(t *testing.T) {
	provider := NewCesiumQuantizedMeshProvider("https://terrain.example.com/{z}/{x}/{y}.terrain")

	provider.SetMaxCacheSize(50)
	if provider.GetCacheSize() != 0 {
		t.Errorf("Expected cache size 0, got %d", provider.GetCacheSize())
	}

	provider.ClearCache()
	if provider.GetCacheSize() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", provider.GetCacheSize())
	}
}

func TestDecoderConfig(t *testing.T) {
	config := NewDecoderConfig()

	if config == nil {
		t.Fatal("NewDecoderConfig returned nil")
	}

	if !config.ExtensionHeader {
		t.Error("Expected ExtensionHeader to be true")
	}

	if !config.VertexNormals {
		t.Error("Expected VertexNormals to be true")
	}

	if config.WaterMask {
		t.Error("Expected WaterMask to be false")
	}

	if config.Metadata {
		t.Error("Expected Metadata to be false")
	}

	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries 3, got %d", config.MaxRetries)
	}

	if config.Timeout != 30 {
		t.Errorf("Expected Timeout 30, got %d", config.Timeout)
	}
}

func TestCesiumQuantizedMeshProvider_calculateTileBounds(t *testing.T) {
	provider := NewCesiumQuantizedMeshProvider("https://terrain.example.com/{z}/{x}/{y}.terrain")

	testCases := []struct {
		coord        [3]int
		wantLat      float64
		wantLon      float64
		latTolerance float64
		lonTolerance float64
	}{
		{[3]int{0, 0, 0}, 85.0511287798066, -180.0, 0.0001, 0.0001},
		{[3]int{1, 0, 1}, 85.0511287798066, 0.0, 0.0001, 0.0001},
		{[3]int{0, 1, 1}, 0.0, -180.0, 0.0001, 0.0001},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			bounds := provider.calculateTileBounds(tc.coord)

			lonDiff := bounds.Min[1] - tc.wantLon
			if lonDiff < 0 {
				lonDiff = -lonDiff
			}
			if lonDiff > tc.lonTolerance {
				t.Errorf("Expected Min[1] %f (tolerance %f), got %f (diff %f)",
					tc.wantLon, tc.lonTolerance, bounds.Min[1], lonDiff)
			}

			latDiff := bounds.Max[0] - tc.wantLat
			if latDiff < 0 {
				latDiff = -latDiff
			}
			if latDiff > tc.latTolerance {
				t.Errorf("Expected Max[0] %f (tolerance %f), got %f (diff %f)",
					tc.wantLat, tc.latTolerance, bounds.Max[0], latDiff)
			}
		})
	}
}

func TestCesiumQuantizedMeshProvider_calculateTileCoordsForBounds(t *testing.T) {
	provider := NewCesiumQuantizedMeshProvider("https://terrain.example.com/{z}/{x}/{y}.terrain")

	bounds := vec2d.Rect{
		Min: vec2d.T{40.0, -74.0},
		Max: vec2d.T{41.0, -73.0},
	}

	coords := provider.calculateTileCoordsForBounds(12, bounds)

	if len(coords) == 0 {
		t.Fatal("Expected non-empty coords list")
	}

	zoom := 12
	if coords[0][2] != zoom {
		t.Errorf("Expected zoom %d, got %d", zoom, coords[0][2])
	}
}

func TestCesiumQuantizedMeshProvider_mergeMeshes(t *testing.T) {
	provider := NewCesiumQuantizedMeshProvider("https://terrain.example.com/{z}/{x}/{y}.terrain")

	mesh1 := &TinMesh{
		Vertices: []vec3d.T{
			{0, 0, 0}, {1, 0, 0}, {0, 1, 0},
		},
		Indices: []uint32{0, 1, 2},
		Bounds: vec2d.Rect{
			Min: vec2d.T{0, 0},
			Max: vec2d.T{1, 1},
		},
		Srs:       geo.NewProj(4326),
		MinHeight: 0,
		MaxHeight: 0,
	}

	mesh2 := &TinMesh{
		Vertices: []vec3d.T{
			{1, 1, 0}, {2, 1, 0}, {1, 2, 0},
		},
		Indices: []uint32{0, 1, 2},
		Bounds: vec2d.Rect{
			Min: vec2d.T{1, 1},
			Max: vec2d.T{2, 2},
		},
		Srs:       geo.NewProj(4326),
		MinHeight: 0,
		MaxHeight: 0,
	}

	merged := provider.mergeMeshes(mesh1, mesh2)

	if merged == nil {
		t.Fatal("mergeMeshes returned nil")
	}

	if len(merged.Vertices) != 6 {
		t.Errorf("Expected 6 vertices, got %d", len(merged.Vertices))
	}

	if len(merged.Indices) != 6 {
		t.Errorf("Expected 6 indices, got %d", len(merged.Indices))
	}

	if merged.Bounds.Min[0] != 0.0 || merged.Bounds.Min[1] != 0.0 {
		t.Errorf("Expected Min [0, 0], got %v", merged.Bounds.Min)
	}

	if merged.Bounds.Max[0] != 2.0 || merged.Bounds.Max[1] != 2.0 {
		t.Errorf("Expected Max [2, 2], got %v", merged.Bounds.Max)
	}
}

func TestCesiumQuantizedMeshProvider_mergeMeshes_NilInputs(t *testing.T) {
	provider := NewCesiumQuantizedMeshProvider("https://terrain.example.com/{z}/{x}/{y}.terrain")

	mesh := &TinMesh{
		Vertices: []vec3d.T{{0, 0, 0}},
		Srs:      geo.NewProj(4326),
	}

	result := provider.mergeMeshes(nil, mesh)
	if result != mesh {
		t.Error("Expected mergeMeshes(nil, mesh) to return mesh")
	}

	result = provider.mergeMeshes(mesh, nil)
	if result != mesh {
		t.Error("Expected mergeMeshes(mesh, nil) to return mesh")
	}
}
