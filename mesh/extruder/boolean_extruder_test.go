package extruder

import (
	"testing"

	"github.com/flywave/go-static-mesh/draw"
	"github.com/flywave/go-static-mesh/mesh"
	bsp "github.com/flywave/go-static-mesh/mesh/bsp"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

func TestNewBooleanExtruder(t *testing.T) {
	extruder := NewBooleanExtruder()

	if extruder.baseMesh == nil {
		t.Error("Expected base mesh to be initialized")
	}

	if extruder.operation != bsp.BSPOperationUnion {
		t.Errorf("Expected default operation Union, got %v", extruder.operation)
	}
}

func TestBooleanExtruderSetOperation(t *testing.T) {
	extruder := NewBooleanExtruder()

	extruder.SetOperation(bsp.BSPOperationIntersection)

	if extruder.operation != bsp.BSPOperationIntersection {
		t.Errorf("Expected Intersection, got %v", extruder.operation)
	}

	extruder.SetOperation(bsp.BSPOperationSubtraction)

	if extruder.operation != bsp.BSPOperationSubtraction {
		t.Errorf("Expected Subtraction, got %v", extruder.operation)
	}
}

func TestBooleanExtruderSetBaseMesh(t *testing.T) {
	extruder := NewBooleanExtruder()

	mesh := &mesh.Mesh{
		Vertices: []vec3d.T{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}},
		Indices:  []uint32{0, 1, 2},
	}

	extruder.SetBaseMesh(mesh)

	if extruder.baseMesh != mesh {
		t.Error("Expected base mesh to be set")
	}
}

func TestBooleanExtruderExtrudeArea(t *testing.T) {
	extruder := NewBooleanExtruder()

	area := &draw.Area{
		Positions: []vec2d.T{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
	}

	err := extruder.ExtrudeAreaAndApply(area, 5.0)
	if err != nil {
		t.Fatalf("Failed to extrude area: %v", err)
	}

	result := extruder.GetResult()
	if len(result.Vertices) == 0 {
		t.Error("Expected vertices in result mesh")
	}
}

func TestBooleanExtruderMultipleExtrusions(t *testing.T) {
	extruder := NewBooleanExtruder()
	extruder.SetOperation(bsp.BSPOperationUnion)

	area1 := &draw.Area{
		Positions: []vec2d.T{{0, 0}, {5, 0}, {5, 5}, {0, 5}},
	}

	area2 := &draw.Area{
		Positions: []vec2d.T{{3, 3}, {8, 3}, {8, 8}, {3, 8}},
	}

	err := extruder.ExtrudeAreaAndApply(area1, 5.0)
	if err != nil {
		t.Fatalf("Failed to extrude first area: %v", err)
	}

	err = extruder.ExtrudeAreaAndApply(area2, 5.0)
	if err != nil {
		t.Fatalf("Failed to extrude second area: %v", err)
	}

	result := extruder.GetResult()
	if len(result.Vertices) == 0 {
		t.Error("Expected vertices in result mesh after union")
	}
}

func TestBooleanExtruderReset(t *testing.T) {
	extruder := NewBooleanExtruder()

	area := &draw.Area{
		Positions: []vec2d.T{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
	}

	err := extruder.ExtrudeAreaAndApply(area, 5.0)
	if err != nil {
		t.Fatalf("Failed to extrude area: %v", err)
	}

	result := extruder.GetResult()
	if len(result.Vertices) == 0 {
		t.Error("Expected vertices before reset")
	}

	extruder.Reset()

	result = extruder.GetResult()
	if len(result.Vertices) != 0 {
		t.Error("Expected empty mesh after reset")
	}

	if extruder.operation != bsp.BSPOperationUnion {
		t.Errorf("Expected operation to be Union after reset, got %v", extruder.operation)
	}
}

func TestExtrudeMultipleWithBoolean(t *testing.T) {
	areas := []draw.MapObject{
		&draw.Area{
			Positions: []vec2d.T{{0, 0}, {5, 0}, {5, 5}, {0, 5}},
		},
		&draw.Area{
			Positions: []vec2d.T{{3, 3}, {8, 3}, {8, 8}, {3, 8}},
		},
	}

	options := &MultiExtrusionOptions{
		Operation:  bsp.BSPOperationUnion,
		BaseMesh:   &mesh.Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}},
		Height:     5.0,
		Resolution: 1.0,
	}

	result, err := ExtrudeMultipleWithBoolean(areas, options)
	if err != nil {
		t.Fatalf("Failed to extrude multiple: %v", err)
	}

	if len(result.Vertices) == 0 {
		t.Error("Expected vertices in result mesh")
	}
}

func TestExtrudeMultipleWithBooleanNilOptions(t *testing.T) {
	areas := []draw.MapObject{
		&draw.Area{
			Positions: []vec2d.T{{0, 0}, {5, 0}, {5, 5}, {0, 5}},
		},
	}

	result, err := ExtrudeMultipleWithBoolean(areas, nil)
	if err != nil {
		t.Fatalf("Failed to extrude multiple with nil options: %v", err)
	}

	if len(result.Vertices) == 0 {
		t.Error("Expected vertices in result mesh with default options")
	}
}

func TestExtrudeAndSubtract(t *testing.T) {
	baseMesh := &mesh.Mesh{
		Vertices: []vec3d.T{
			{0, 0, 0}, {10, 0, 0}, {10, 10, 0}, {0, 10, 0},
			{0, 0, 5}, {10, 0, 5}, {10, 10, 5}, {0, 10, 5},
		},
		Indices: []uint32{
			0, 1, 2, 0, 2, 3,
			4, 7, 6, 4, 6, 5,
			0, 4, 5, 0, 5, 1,
			1, 5, 6, 1, 6, 2,
			2, 6, 7, 2, 7, 3,
			3, 7, 4, 3, 4, 0,
		},
	}

	cutouts := []draw.MapObject{
		&draw.Area{
			Positions: []vec2d.T{{2, 2}, {8, 2}, {8, 8}, {2, 8}},
		},
	}

	result, err := ExtrudeAndSubtract(baseMesh, cutouts, 5.0)
	if err != nil {
		t.Fatalf("Failed to subtract: %v", err)
	}

	if len(result.Vertices) == 0 {
		t.Error("Expected vertices in result mesh after subtraction")
	}
}

func TestExtrudeAndSubtractEmptyCutouts(t *testing.T) {
	baseMesh := &mesh.Mesh{
		Vertices: []vec3d.T{
			{0, 0, 0}, {10, 0, 0}, {10, 10, 0}, {0, 10, 0},
		},
		Indices: []uint32{0, 1, 2, 0, 2, 3},
	}

	cutouts := []draw.MapObject{}

	result, err := ExtrudeAndSubtract(baseMesh, cutouts, 5.0)
	if err != nil {
		t.Fatalf("Failed to subtract empty cutouts: %v", err)
	}

	if len(result.Vertices) == 0 {
		t.Error("Expected vertices in result mesh")
	}
}

func TestExtrudeAndIntersect(t *testing.T) {
	meshA := &mesh.Mesh{
		Vertices: []vec3d.T{
			{0, 0, 0}, {5, 0, 0}, {5, 5, 0}, {0, 5, 0},
		},
		Indices: []uint32{0, 1, 2, 0, 2, 3},
	}

	meshB := &mesh.Mesh{
		Vertices: []vec3d.T{
			{3, 3, 0}, {8, 3, 0}, {8, 8, 0}, {3, 8, 0},
		},
		Indices: []uint32{0, 1, 2, 0, 2, 3},
	}

	result, err := ExtrudeAndIntersect(meshA, meshB, 5.0)
	if err != nil {
		t.Fatalf("Failed to intersect: %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result mesh")
	}
}

func TestExtrudeAndIntersectEmptyMeshes(t *testing.T) {
	meshA := &mesh.Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}
	meshB := &mesh.Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}

	result, err := ExtrudeAndIntersect(meshA, meshB, 5.0)
	if err != nil {
		t.Fatalf("Failed to intersect empty meshes: %v", err)
	}

	if len(result.Vertices) != 0 {
		t.Error("Expected empty result mesh for empty inputs")
	}
}

func TestBooleanOperationTypes(t *testing.T) {
	operations := []bsp.BSPOperation{
		bsp.BSPOperationUnion,
		bsp.BSPOperationIntersection,
		bsp.BSPOperationSubtraction,
	}

	expectedValues := []int{0, 1, 2}

	for i, op := range operations {
		if int(op) != expectedValues[i] {
			t.Errorf("Expected operation value %d, got %d", expectedValues[i], op)
		}
	}
}

func TestBooleanExtruderGetResult(t *testing.T) {
	extruder := NewBooleanExtruder()

	mesh := &mesh.Mesh{
		Vertices: []vec3d.T{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}},
		Indices:  []uint32{0, 1, 2},
	}

	extruder.SetBaseMesh(mesh)

	result := extruder.GetResult()

	if result != mesh {
		t.Error("Expected GetResult to return the base mesh")
	}

	if len(result.Vertices) != 3 {
		t.Errorf("Expected 3 vertices, got %d", len(result.Vertices))
	}
}
