package builder

import (
	"testing"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-static-mesh/mesh"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

func TestNewBuilder(t *testing.T) {
	builder := NewBuilder()
	if builder == nil {
		t.Fatal("NewBuilder returned nil")
	}
}

func TestBuilderSetBounds(t *testing.T) {
	builder := NewBuilder()
	bounds := vec2d.Rect{Min: vec2d.T{30.0, 120.0}, Max: vec2d.T{31.0, 121.0}}
	srs := geo.NewProj(4326)

	builder.SetBounds(bounds, srs)

	if builder.bounds != bounds {
		t.Errorf("bounds not set correctly")
	}
}

func TestBuilderSetZoom(t *testing.T) {
	builder := NewBuilder()
	zoom := 15

	builder.SetZoom(zoom)

	if builder.zoom == nil || *builder.zoom != zoom {
		t.Errorf("zoom not set correctly")
	}
}

func TestBuilderSetAutoZoomRange(t *testing.T) {
	builder := NewBuilder()

	builder.SetAutoZoomRange(10, 18)

	if builder.autoZoomMin != 10 || builder.autoZoomMax != 18 {
		t.Errorf("auto zoom range not set correctly")
	}
}

func TestBuilderSetCloseMesh(t *testing.T) {
	builder := NewBuilder()

	builder.SetCloseMesh(true, 5.0)

	if !builder.closeMesh || builder.baseThickness != 5.0 {
		t.Errorf("close mesh parameters not set correctly")
	}
}

func TestTileErrorHandler(t *testing.T) {
	handler := TileErrorHandler{
		SkipMissing: true,
		UseNoData:   false,
		NoDataValue: -9999.0,
		MaxRetries:  3,
	}

	if !handler.SkipMissing {
		t.Errorf("SkipMissing not set")
	}

	if handler.MaxRetries != 3 {
		t.Errorf("MaxRetries not set")
	}
}

func TestMeshCreate(t *testing.T) {
	mesh := &mesh.Mesh{
		Vertices: []vec3d.T{
			{0, 0, 0},
			{1, 0, 0},
			{0, 1, 0},
		},
		Indices: []uint32{0, 1, 2},
	}

	if mesh.VertexCount() != 3 {
		t.Errorf("VertexCount incorrect")
	}

	if mesh.TriangleCount() != 1 {
		t.Errorf("TriangleCount incorrect")
	}
}

func TestMeshCalculateNormals(t *testing.T) {
	mesh := &mesh.Mesh{
		Vertices: []vec3d.T{
			{0, 0, 0},
			{1, 0, 0},
			{0, 1, 0},
		},
		Indices: []uint32{0, 1, 2},
	}

	mesh.CalculateNormals()

	if len(mesh.Normals) != 3 {
		t.Errorf("Normals not calculated")
	}
}

func TestMeshCalculateUVs(t *testing.T) {
	mesh := &mesh.Mesh{
		Vertices: []vec3d.T{
			{0, 0, 0},
			{1, 0, 0},
			{0, 1, 0},
		},
		Indices: []uint32{0, 1, 2},
	}

	bounds := vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{1, 1}}
	mesh.CalculateUVs(bounds)

	if len(mesh.UVs) != 3 {
		t.Errorf("UVs not calculated")
	}
}

func TestDetermineZoom(t *testing.T) {
	builder := NewBuilder()

	bounds := vec2d.Rect{Min: vec2d.T{30.0, 120.0}, Max: vec2d.T{31.0, 121.0}}
	srs := geo.NewProj(4326)
	builder.SetBounds(bounds, srs)

	zoom, err := builder.determineZoom()
	if err != nil {
		t.Fatal(err)
	}

	if zoom < builder.autoZoomMin || zoom > builder.autoZoomMax {
		t.Errorf("zoom out of range: %d", zoom)
	}
}

func TestBuilderSetTinMeshProvider(t *testing.T) {
	builder := NewBuilder()

	mockProvider := struct{}{}
	builder.SetTinMeshProvider(mockProvider)

	if builder.tinMeshProvider != mockProvider {
		t.Error("TIN mesh provider not set correctly")
	}
	if builder.rasterProvider != nil {
		t.Error("raster provider should be nil when TIN mesh provider is set")
	}
}

func TestBuilderAddImageryProvider(t *testing.T) {
	builder := NewBuilder()

	mockProvider := struct{}{}
	builder.AddImageryProvider(mockProvider)

	if builder.imageryProvider != mockProvider {
		t.Error("imagery provider not added correctly")
	}
}

func TestBuilderSetVerticalExaggeration(t *testing.T) {
	builder := NewBuilder()
	scale := 2.5

	builder.SetVerticalExaggeration(scale)

	if builder.verticalExaggeration != scale {
		t.Errorf("vertical exaggeration not set correctly, got %f, want %f", builder.verticalExaggeration, scale)
	}
}

func TestBuilderSetBaseElevation(t *testing.T) {
	builder := NewBuilder()
	elevation := 100.0

	builder.SetBaseElevation(elevation)

	if builder.baseElevation != elevation {
		t.Errorf("base elevation not set correctly, got %f, want %f", builder.baseElevation, elevation)
	}
}

func TestBuilderSetTileErrorHandler(t *testing.T) {
	builder := NewBuilder()
	handler := TileErrorHandler{
		SkipMissing: false,
		UseNoData:   true,
		MaxRetries:  5,
	}

	builder.SetTileErrorHandler(handler)

	if builder.tileErrorHandler.SkipMissing != false {
		t.Error("tile error handler not set correctly")
	}
}

func TestBuilderSetTileCache(t *testing.T) {
	builder := NewBuilder()
	cache := mesh.NewMemoryTileCache(100, 0)

	builder.SetTileCache(cache)

	if builder.tileCache != cache {
		t.Error("tile cache not set correctly")
	}
}

func TestBuilderSetResolution(t *testing.T) {
	builder := NewBuilder()
	resolution := 2.0

	builder.SetResolution(resolution)

	if builder.resolution != resolution {
		t.Errorf("resolution not set correctly, got %f, want %f", builder.resolution, resolution)
	}
}

func TestBuilderSetCloseMeshOptions(t *testing.T) {
	builder := NewBuilder()
	options := &mesh.CloseMeshOptions{
		Enabled:       true,
		Thickness:     10.0,
		ApplyToBottom: true,
	}

	builder.SetCloseMeshOptions(options)

	if !builder.closeMesh || builder.closeMeshOptions.Thickness != 10.0 {
		t.Error("close mesh options not set correctly")
	}
}

func TestBuilderGetCloseMeshOptions(t *testing.T) {
	builder := NewBuilder()
	options := &mesh.CloseMeshOptions{
		Enabled:   true,
		Thickness: 15.0,
	}

	builder.SetCloseMeshOptions(options)

	retrieved := builder.GetCloseMeshOptions()
	if retrieved == nil || retrieved.Thickness != 15.0 {
		t.Error("failed to get close mesh options")
	}
}

func TestBuilderSetLogger(t *testing.T) {
	builder := NewBuilder()
	logger := &mesh.NoOpLogger{}

	builder.SetLogger(logger)

	if builder.logger != logger {
		t.Error("logger not set correctly")
	}
}

func TestBuilderSetProgressCallback(t *testing.T) {
	builder := NewBuilder()
	var callback ProgressCallback = &mockProgressCallback{}

	builder.SetProgressCallback(callback)

	if builder.progressCallback == nil {
		t.Error("progress callback not set correctly")
	}
}

type mockProgressCallback struct{}

func (m *mockProgressCallback) OnStageStart(stage string, totalSteps uint64) {}
func (m *mockProgressCallback) OnProgress(step, total uint64)                {}
func (m *mockProgressCallback) OnStageComplete(stage string)                 {}
func (m *mockProgressCallback) OnProgressError(err error)                    {}
