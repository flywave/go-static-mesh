package mesh_test

import (
	"image/color"
	"testing"

	"github.com/flywave/go-static-mesh/mesh"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

func TestCloseMeshOptionsDefaults(t *testing.T) {
	options := mesh.NewDefaultCloseMeshOptions()

	if !options.UseMinHeightAsBase {
		t.Error("Expected UseMinHeightAsBase to be true by default")
	}

	if options.Enabled {
		t.Error("Expected Enabled to be false by default")
	}

	if options.Thickness != 5.0 {
		t.Errorf("Expected Thickness to be 5.0, got %f", options.Thickness)
	}

	if options.BottomTextureTilingU != 1.0 {
		t.Errorf("Expected BottomTextureTilingU to be 1.0, got %f", options.BottomTextureTilingU)
	}

	if options.BottomTextureTilingV != 1.0 {
		t.Errorf("Expected BottomTextureTilingV to be 1.0, got %f", options.BottomTextureTilingV)
	}
}

func TestTexturedCloserBasic(t *testing.T) {
	closer := mesh.NewTexturedCloser()

	if closer == nil {
		t.Fatal("Expected non-nil TexturedCloser")
	}

	options := closer.GetOptions()
	if options == nil {
		t.Fatal("Expected non-nil options")
	}
}

func TestTexturedCloserSetOptions(t *testing.T) {
	closer := mesh.NewTexturedCloser()
	options := mesh.NewDefaultCloseMeshOptions()

	options.Enabled = true
	options.Thickness = 10.0
	options.BottomColor = color.RGBA{255, 0, 0, 255}

	closer.SetOptions(options)

	retrieved := closer.GetOptions()
	if retrieved.Thickness != 10.0 {
		t.Errorf("Expected Thickness to be 10.0, got %f", retrieved.Thickness)
	}

	rgb, ok := retrieved.BottomColor.(color.RGBA)
	if !ok || rgb.R != 255 {
		t.Errorf("Expected BottomColor to be red, got %v", retrieved.BottomColor)
	}
}

func TestTexturedCloserCloseDisabled(t *testing.T) {
	closer := mesh.NewTexturedCloser()
	options := mesh.NewDefaultCloseMeshOptions()
	options.Enabled = false

	testData := &testMesh{
		vertices: []vec3d.T{
			{0, 0, 10},
			{10, 0, 10},
			{5, 10, 10},
		},
		indices:   []uint32{0, 1, 2},
		minHeight: 10.0,
		bounds: vec2d.Rect{
			Min: vec2d.T{0, 0},
			Max: vec2d.T{10, 10},
		},
	}

	closedMesh, err := closer.CloseSurfaceMeshWithOptions(testData, options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if closedMesh != nil {
		t.Error("Expected nil mesh when closing is disabled")
	}
}

func TestTexturedCloserCloseEnabled(t *testing.T) {
	closer := mesh.NewTexturedCloser()
	options := mesh.NewDefaultCloseMeshOptions()
	options.Enabled = true
	options.Thickness = 5.0
	options.BottomColor = color.RGBA{139, 119, 101, 255}

	testData := &testMesh{
		vertices: []vec3d.T{
			{0, 0, 10},
			{10, 0, 10},
			{5, 10, 10},
		},
		indices:   []uint32{0, 1, 2},
		minHeight: 10.0,
		bounds: vec2d.Rect{
			Min: vec2d.T{0, 0},
			Max: vec2d.T{10, 10},
		},
	}

	closedMesh, err := closer.CloseSurfaceMeshWithOptions(testData, options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if closedMesh == nil {
		t.Fatal("Expected non-nil mesh when closing is enabled")
	}

	expectedVertices := 6
	if len(closedMesh.Vertices) != expectedVertices {
		t.Errorf("Expected %d vertices, got %d", expectedVertices, len(closedMesh.Vertices))
	}

	expectedTriangles := 2
	if closedMesh.TriangleCount() != expectedTriangles {
		t.Errorf("Expected %d triangles, got %d", expectedTriangles, closedMesh.TriangleCount())
	}

	if len(closedMesh.UVs) != expectedVertices {
		t.Errorf("Expected %d UVs, got %d", expectedVertices, len(closedMesh.UVs))
	}

	if len(closedMesh.Normals) != expectedVertices {
		t.Errorf("Expected %d normals, got %d", expectedVertices, len(closedMesh.Normals))
	}

	if len(closedMesh.Materials) != 1 {
		t.Errorf("Expected 1 material, got %d", len(closedMesh.Materials))
	}

	mat := closedMesh.Materials[0]
	rgb, ok := mat.Diffuse.(color.RGBA)
	if !ok || rgb.R != 139 {
		t.Errorf("Expected material color to match BottomColor, got %v", mat.Diffuse)
	}
}

func TestTexturedCloserMinHeightBase(t *testing.T) {
	closer := mesh.NewTexturedCloser()
	options := mesh.NewDefaultCloseMeshOptions()
	options.Enabled = true
	options.UseMinHeightAsBase = true
	options.Thickness = 5.0

	testData := &testMesh{
		vertices: []vec3d.T{
			{0, 0, 20},
			{10, 0, 15},
			{5, 10, 10},
		},
		indices:   []uint32{0, 1, 2},
		minHeight: 10.0,
		bounds: vec2d.Rect{
			Min: vec2d.T{0, 0},
			Max: vec2d.T{10, 10},
		},
	}

	closedMesh, err := closer.CloseSurfaceMeshWithOptions(testData, options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if closedMesh == nil {
		t.Fatal("Expected non-nil mesh")
	}

	expectedBaseHeight := 5.0
	for i := 3; i < len(closedMesh.Vertices); i++ {
		if closedMesh.Vertices[i][2] != expectedBaseHeight {
			t.Errorf("Vertex %d Z should be %f, got %f", i, expectedBaseHeight, closedMesh.Vertices[i][2])
		}
	}
}

func TestTexturedCloserAbsoluteBase(t *testing.T) {
	closer := mesh.NewTexturedCloser()
	options := mesh.NewDefaultCloseMeshOptions()
	options.Enabled = true
	options.UseMinHeightAsBase = false
	options.Thickness = 5.0

	testData := &testMesh{
		vertices: []vec3d.T{
			{0, 0, 20},
			{10, 0, 15},
			{5, 10, 10},
		},
		indices:   []uint32{0, 1, 2},
		minHeight: 10.0,
		bounds: vec2d.Rect{
			Min: vec2d.T{0, 0},
			Max: vec2d.T{10, 10},
		},
	}

	closedMesh, err := closer.CloseSurfaceMeshWithOptions(testData, options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if closedMesh == nil {
		t.Fatal("Expected non-nil mesh")
	}

	expectedBaseHeight := -5.0
	for i := 3; i < len(closedMesh.Vertices); i++ {
		if closedMesh.Vertices[i][2] != expectedBaseHeight {
			t.Errorf("Vertex %d Z should be %f, got %f", i, expectedBaseHeight, closedMesh.Vertices[i][2])
		}
	}
}

type testMesh struct {
	vertices  []vec3d.T
	indices   []uint32
	minHeight float64
	bounds    vec2d.Rect
}

func (m *testMesh) GetVertices() []vec3d.T {
	return m.vertices
}

func (m *testMesh) GetIndices() []uint32 {
	return m.indices
}

func (m *testMesh) GetMinHeight() float64 {
	return m.minHeight
}

func (m *testMesh) GetBounds() vec2d.Rect {
	return m.bounds
}
