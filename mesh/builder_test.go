package mesh

import (
	"testing"

	"github.com/flywave/go-geo"
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
	mesh := &Mesh{
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
	mesh := &Mesh{
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
	mesh := &Mesh{
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
