package mesh

import (
	"testing"

	"github.com/flywave/go-static-mesh/draw"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

func TestNewGPXPathExtruder(t *testing.T) {
	extruder := NewGPXPathExtruder()

	if extruder.baseMesh == nil {
		t.Error("Expected base mesh to be initialized")
	}

	if extruder.operation != BSPOperationSubtraction {
		t.Errorf("Expected default operation Subtraction, got %v", extruder.operation)
	}
}

func TestGPXPathExtruderSetOperation(t *testing.T) {
	extruder := NewGPXPathExtruder()

	extruder.SetOperation(BSPOperationUnion)

	if extruder.operation != BSPOperationUnion {
		t.Errorf("Expected Union, got %v", extruder.operation)
	}

	extruder.SetOperation(BSPOperationIntersection)

	if extruder.operation != BSPOperationIntersection {
		t.Errorf("Expected Intersection, got %v", extruder.operation)
	}
}

func TestGPXPathExtruderSetBaseMesh(t *testing.T) {
	extruder := NewGPXPathExtruder()

	mesh := &Mesh{
		Vertices: []vec3d.T{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}},
		Indices:  []uint32{0, 1, 2},
	}

	extruder.SetBaseMesh(mesh)

	if extruder.baseMesh != mesh {
		t.Error("Expected base mesh to be set")
	}
}

func TestGPXPathExtruderExtrudePathToTerrain(t *testing.T) {
	extruder := NewGPXPathExtruder()

	terrainMesh := &Mesh{
		Vertices: []vec3d.T{
			{0, 0, 10}, {10, 0, 10}, {10, 10, 10}, {0, 10, 10},
		},
		Indices: []uint32{0, 1, 2, 0, 2, 3},
	}

	path := &draw.Path{
		Positions: []vec2d.T{{2, 2}, {8, 8}},
		Weight:    2.0,
	}

	options := &GPXExtrusionOptions{
		Width:        2.0,
		Height:       5.0,
		Depth:        3.0,
		Operation:    BSPOperationSubtraction,
		Resolution:   1.0,
		SampleRadius: 100.0,
		TerrainMesh:  terrainMesh,
	}

	err := extruder.ExtrudePathToTerrain(path, options)
	if err != nil {
		t.Fatalf("Failed to extrude path: %v", err)
	}

	result := extruder.GetResult()
	if len(result.Vertices) == 0 {
		t.Error("Expected vertices in result mesh")
	}
}

func TestGPXPathExtruderExtrudePathWithoutTerrain(t *testing.T) {
	extruder := NewGPXPathExtruder()

	path := &draw.Path{
		Positions: []vec2d.T{{0, 0}, {10, 0}},
		Weight:    2.0,
	}

	options := &GPXExtrusionOptions{
		Width:     2.0,
		Height:    5.0,
		Depth:     3.0,
		Operation: BSPOperationUnion,
	}

	err := extruder.ExtrudePathToTerrain(path, options)
	if err != nil {
		t.Fatalf("Failed to extrude path: %v", err)
	}

	result := extruder.GetResult()
	if len(result.Vertices) == 0 {
		t.Error("Expected vertices in result mesh")
	}
}

func TestGPXPathExtruderExtrudePathNilOptions(t *testing.T) {
	extruder := NewGPXPathExtruder()

	path := &draw.Path{
		Positions: []vec2d.T{{0, 0}, {10, 0}},
		Weight:    5.0,
	}

	err := extruder.ExtrudePathToTerrain(path, nil)
	if err != nil {
		t.Fatalf("Failed to extrude path with nil options: %v", err)
	}

	result := extruder.GetResult()
	if len(result.Vertices) == 0 {
		t.Error("Expected vertices in result mesh with default options")
	}
}

func TestGPXPathExtruderReset(t *testing.T) {
	extruder := NewGPXPathExtruder()

	terrainMesh := &Mesh{
		Vertices: []vec3d.T{{0, 0, 10}, {10, 0, 10}, {10, 10, 10}, {0, 10, 10}},
		Indices:  []uint32{0, 1, 2, 0, 2, 3},
	}

	path := &draw.Path{
		Positions: []vec2d.T{{2, 2}, {8, 8}},
		Weight:    2.0,
	}

	options := &GPXExtrusionOptions{
		Width:        2.0,
		Height:       5.0,
		Depth:        3.0,
		Operation:    BSPOperationSubtraction,
		SampleRadius: 100.0,
		TerrainMesh:  terrainMesh,
	}

	err := extruder.ExtrudePathToTerrain(path, options)
	if err != nil {
		t.Fatalf("Failed to extrude path: %v", err)
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

	if extruder.operation != BSPOperationSubtraction {
		t.Errorf("Expected operation to be Subtraction after reset, got %v", extruder.operation)
	}
}

func TestExtrudeGPXPathsToTerrain(t *testing.T) {
	terrainMesh := &Mesh{
		Vertices: []vec3d.T{
			{0, 0, 10}, {10, 0, 10}, {10, 10, 10}, {0, 10, 10},
		},
		Indices: []uint32{0, 1, 2, 0, 2, 3},
	}

	paths := []*draw.Path{
		{
			Positions: []vec2d.T{{2, 2}, {8, 8}},
			Weight:    2.0,
		},
	}

	options := &GPXExtrusionOptions{
		Width:        2.0,
		Height:       5.0,
		Depth:        3.0,
		Operation:    BSPOperationSubtraction,
		Resolution:   1.0,
		SampleRadius: 100.0,
		TerrainMesh:  terrainMesh,
	}

	result, err := ExtrudeGPXPathsToTerrain(paths, terrainMesh, options)
	if err != nil {
		t.Fatalf("Failed to extrude GPX paths: %v", err)
	}

	if len(result.Vertices) == 0 {
		t.Error("Expected vertices in result mesh")
	}
}

func TestExtrudeGPXPathsToTerrainNilOptions(t *testing.T) {
	terrainMesh := &Mesh{
		Vertices: []vec3d.T{
			{0, 0, 10}, {10, 0, 10}, {10, 10, 10}, {0, 10, 10},
		},
		Indices: []uint32{0, 1, 2, 0, 2, 3},
	}

	paths := []*draw.Path{
		{
			Positions: []vec2d.T{{2, 2}, {8, 8}},
			Weight:    2.0,
		},
	}

	result, err := ExtrudeGPXPathsToTerrain(paths, terrainMesh, nil)
	if err != nil {
		t.Fatalf("Failed to extrude GPX paths with nil options: %v", err)
	}

	if len(result.Vertices) == 0 {
		t.Error("Expected vertices in result mesh with default options")
	}
}

func TestExtrudeGPXPathsToTerrainEmptyPaths(t *testing.T) {
	terrainMesh := &Mesh{
		Vertices: []vec3d.T{{0, 0, 10}, {10, 0, 10}, {10, 10, 10}, {0, 10, 10}},
		Indices:  []uint32{0, 1, 2, 0, 2, 3},
	}

	paths := []*draw.Path{}

	result, err := ExtrudeGPXPathsToTerrain(paths, terrainMesh, nil)
	if err != nil {
		t.Fatalf("Failed to extrude empty paths: %v", err)
	}

	if result != terrainMesh {
		t.Error("Expected original terrain mesh for empty paths")
	}
}

func TestExtrudeGPXPathsToTerrainNilTerrain(t *testing.T) {
	paths := []*draw.Path{
		{
			Positions: []vec2d.T{{0, 0}, {10, 0}},
			Weight:    2.0,
		},
	}

	options := &GPXExtrusionOptions{
		Width:     2.0,
		Height:    5.0,
		Depth:     3.0,
		Operation: BSPOperationUnion,
	}

	result, err := ExtrudeGPXPathsToTerrain(paths, nil, options)
	if err != nil {
		t.Fatalf("Failed to extrude paths with nil terrain: %v", err)
	}

	if len(result.Vertices) == 0 {
		t.Error("Expected vertices in result mesh without terrain")
	}
}

func TestExtrudeGPXPathsToTerrainUnion(t *testing.T) {
	terrainMesh := &Mesh{
		Vertices: []vec3d.T{
			{0, 0, 10}, {10, 0, 10}, {10, 10, 10}, {0, 10, 10},
		},
		Indices: []uint32{0, 1, 2, 0, 2, 3},
	}

	paths := []*draw.Path{
		{
			Positions: []vec2d.T{{2, 2}, {8, 8}},
			Weight:    2.0,
		},
	}

	result, err := ExtrudeGPXPathsToTerrain(paths, terrainMesh, &GPXExtrusionOptions{
		Width:        2.0,
		Height:       5.0,
		Depth:        3.0,
		Operation:    BSPOperationSubtraction,
		Resolution:   1.0,
		SampleRadius: 100.0,
		TerrainMesh:  terrainMesh,
	})
	if err != nil {
		t.Fatalf("Failed to extrude paths to terrain: %v", err)
	}

	if len(result.Vertices) == 0 {
		t.Error("Expected vertices in result mesh")
	}
}

func TestGPXPathExtruderInvalidPath(t *testing.T) {
	extruder := NewGPXPathExtruder()

	path := &draw.Path{
		Positions: []vec2d.T{{0, 0}},
		Weight:    2.0,
	}

	options := &GPXExtrusionOptions{
		Width:     2.0,
		Height:    5.0,
		Depth:     3.0,
		Operation: BSPOperationUnion,
	}

	err := extruder.ExtrudePathToTerrain(path, options)
	if err == nil {
		t.Error("Expected error for invalid path with single point")
	}
}

func TestGPXPathExtruderNilPath(t *testing.T) {
	extruder := NewGPXPathExtruder()

	options := &GPXExtrusionOptions{
		Width:     2.0,
		Height:    5.0,
		Depth:     3.0,
		Operation: BSPOperationUnion,
	}

	err := extruder.ExtrudePathToTerrain(nil, options)
	if err == nil {
		t.Error("Expected error for nil path")
	}
}

func TestCreateTrenchFromGPX(t *testing.T) {
	terrainMesh := &Mesh{
		Vertices: []vec3d.T{
			{0, 0, 10}, {10, 0, 10}, {10, 10, 10}, {0, 10, 10},
		},
		Indices: []uint32{0, 1, 2, 0, 2, 3},
	}

	paths := []*draw.Path{
		{
			Positions: []vec2d.T{{2, 2}, {8, 8}},
			Weight:    3.0,
		},
	}

	result, err := CreateTrenchFromGPX(paths, terrainMesh, 3.0, 5.0)
	if err != nil {
		t.Fatalf("Failed to create trench: %v", err)
	}

	if len(result.Vertices) == 0 {
		t.Error("Expected vertices in result mesh")
	}
}

func TestCreateRaisedPathFromGPX(t *testing.T) {
	terrainMesh := &Mesh{
		Vertices: []vec3d.T{
			{0, 0, 10}, {10, 0, 10}, {10, 10, 10}, {0, 10, 10},
		},
		Indices: []uint32{0, 1, 2, 0, 2, 3},
	}

	paths := []*draw.Path{
		{
			Positions: []vec2d.T{{2, 2}, {8, 8}},
			Weight:    2.0,
		},
	}

	result, err := CreateRaisedPathFromGPX(paths, terrainMesh, 2.0, 5.0)
	if err != nil {
		t.Fatalf("Failed to create raised path: %v", err)
	}

	if len(result.Vertices) == 0 {
		t.Error("Expected vertices in result mesh")
	}
}

func TestGPXExtrusionOptionsDefaults(t *testing.T) {
	options := &GPXExtrusionOptions{}

	if options.Width != 0 {
		t.Errorf("Expected default Width 0, got %v", options.Width)
	}

	if options.Height != 0 {
		t.Errorf("Expected default Height 0, got %v", options.Height)
	}

	if options.Depth != 0 {
		t.Errorf("Expected default Depth 0, got %v", options.Depth)
	}

	if options.Operation != BSPOperationUnion {
		t.Errorf("Expected default Operation Union (0), got %v", options.Operation)
	}

	if options.SampleRadius != 0 {
		t.Errorf("Expected default SampleRadius 0, got %v", options.SampleRadius)
	}
}
