package mesh

import (
	"github.com/flywave/go-geo"
	"github.com/flywave/go-static-mesh/draw"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
	"image/color"
	"testing"
)

func TestPathExtruder_ExtrudeToMesh(t *testing.T) {
	extruder := NewPathExtruder()
	path := draw.NewPath(
		[]vec2d.T{{0, 0}, {10, 0}, {10, 10}},
		geo.NewProj(4326),
		color.RGBA{255, 0, 0, 255},
		2.0,
	)

	mesh := &Mesh{
		Vertices: []vec3d.T{},
		Indices:  []uint32{},
	}

	err := extruder.ExtrudeToMesh(path, mesh, 5.0)
	if err != nil {
		t.Fatalf("ExtrudeToMesh failed: %v", err)
	}

	if len(mesh.Vertices) == 0 {
		t.Error("No vertices generated")
	}

	if len(mesh.Indices) == 0 {
		t.Error("No indices generated")
	}

	if len(mesh.Indices)%6 != 0 {
		t.Errorf("Indices should be in multiples of 6 (quads), got %d", len(mesh.Indices))
	}

	for _, v := range mesh.Vertices {
		if v[2] != 0 && v[2] != 5.0 {
			t.Errorf("Unexpected Z value: %f (expected 0 or 5.0)", v[2])
		}
	}
}

func TestPathExtruder_Resolution(t *testing.T) {
	extruder := NewPathExtruder()
	path := draw.NewPath(
		[]vec2d.T{{0, 0}, {100, 0}},
		geo.NewProj(4326),
		color.RGBA{255, 0, 0, 255},
		2.0,
	)

	options := &ExtrudeOptions{
		Resolution: 0.5,
		Radius:     1.0,
	}

	mesh := &Mesh{
		Vertices: []vec3d.T{},
		Indices:  []uint32{},
	}

	err := extruder.ExtrudeToMeshWithResolution(path, mesh, 10.0, options)
	if err != nil {
		t.Fatalf("ExtrudeToMeshWithResolution failed: %v", err)
	}

	if len(mesh.Vertices) == 0 {
		t.Error("No vertices generated with high resolution")
	}
}

func TestAreaExtruder_ExtrudeToMesh(t *testing.T) {
	extruder := NewAreaExtruder()
	area := draw.NewArea(
		[]vec2d.T{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
		geo.NewProj(4326),
		color.RGBA{255, 0, 0, 255},
		color.RGBA{200, 200, 200, 255},
		1.0,
	)

	mesh := &Mesh{
		Vertices: []vec3d.T{},
		Indices:  []uint32{},
	}

	err := extruder.ExtrudeToMesh(area, mesh, 5.0)
	if err != nil {
		t.Fatalf("ExtrudeToMesh failed: %v", err)
	}

	if len(mesh.Vertices) == 0 {
		t.Error("No vertices generated")
	}

	if len(mesh.Indices) == 0 {
		t.Error("No indices generated")
	}

	if len(mesh.Indices)%3 != 0 {
		t.Errorf("Indices should be in multiples of 3 (triangles), got %d", len(mesh.Indices))
	}

	hasTop := false
	hasBottom := false
	for _, v := range mesh.Vertices {
		if v[2] == 5.0 {
			hasTop = true
		}
		if v[2] == 0 {
			hasBottom = true
		}
	}

	if !hasTop {
		t.Error("No top vertices at Z=5.0")
	}
	if !hasBottom {
		t.Error("No bottom vertices at Z=0")
	}
}

func TestAreaExtruder_Triangle(t *testing.T) {
	extruder := NewAreaExtruder()
	area := draw.NewArea(
		[]vec2d.T{{0, 0}, {10, 0}, {5, 10}},
		geo.NewProj(4326),
		color.RGBA{255, 0, 0, 255},
		color.RGBA{200, 200, 200, 255},
		1.0,
	)

	mesh := &Mesh{
		Vertices: []vec3d.T{},
		Indices:  []uint32{},
	}

	err := extruder.ExtrudeToMesh(area, mesh, 3.0)
	if err != nil {
		t.Fatalf("ExtrudeToMesh failed: %v", err)
	}

	if len(mesh.Vertices) < 6 {
		t.Errorf("Expected at least 6 vertices (3 top + 3 bottom), got %d", len(mesh.Vertices))
	}
}

func TestEarClippingTriangulator_Triangle(t *testing.T) {
	triangulator := NewEarClippingTriangulator()

	triangles := []vec2d.T{{0, 0}, {10, 0}, {5, 10}}
	result, err := triangulator.Triangulate(triangles, 1.0)
	if err != nil {
		t.Fatalf("Triangulate failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 triangle, got %d", len(result))
	}

	if result[0][0] != 0 || result[0][1] != 1 || result[0][2] != 2 {
		t.Errorf("Unexpected triangle indices: %v", result[0])
	}
}

func TestEarClippingTriangulator_Quad(t *testing.T) {
	triangulator := NewEarClippingTriangulator()

	quad := []vec2d.T{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	result, err := triangulator.Triangulate(quad, 1.0)
	if err != nil {
		t.Fatalf("Triangulate failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 triangles for quad, got %d", len(result))
	}
}

func TestEarClippingTriangulator_Pentagon(t *testing.T) {
	triangulator := NewEarClippingTriangulator()

	pentagon := []vec2d.T{{0, 0}, {10, 0}, {12, 5}, {10, 10}, {0, 10}}
	result, err := triangulator.Triangulate(pentagon, 1.0)
	if err != nil {
		t.Fatalf("Triangulate failed: %v", err)
	}

	if len(result) < 3 {
		t.Errorf("Expected at least 3 triangles for pentagon, got %d", len(result))
	}
}

func TestEarClippingTriangulator_Concave(t *testing.T) {
	triangulator := NewEarClippingTriangulator()

	concave := []vec2d.T{{0, 0}, {10, 0}, {10, 10}, {5, 5}, {0, 10}}
	result, err := triangulator.Triangulate(concave, 1.0)
	if err != nil {
		t.Fatalf("Triangulate failed: %v", err)
	}

	if len(result) < 2 {
		t.Errorf("Expected at least 2 triangles for concave polygon, got %d", len(result))
	}
}

func TestExtrudeOptions(t *testing.T) {
	options := &ExtrudeOptions{
		Resolution: 2.0,
		Radius:     5.0,
		Sides:      true,
		Bottom:     true,
		Segments:   32,
	}

	if options.Resolution != 2.0 {
		t.Errorf("Expected Resolution 2.0, got %f", options.Resolution)
	}
	if options.Radius != 5.0 {
		t.Errorf("Expected Radius 5.0, got %f", options.Radius)
	}
	if options.Segments != 32 {
		t.Errorf("Expected Segments 32, got %d", options.Segments)
	}
}
