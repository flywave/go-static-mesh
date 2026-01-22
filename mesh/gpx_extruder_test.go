package mesh

import (
	"testing"

	"github.com/flywave/go-static-mesh/draw"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

func TestGPXExtruderPathToTerrainOverlap(t *testing.T) {
	terrainMesh := &Mesh{
		Vertices: []vec3d.T{
			{0, 0, 5}, {10, 0, 5}, {10, 10, 5}, {0, 10, 5},
		},
		Indices: []uint32{0, 1, 2, 0, 2, 3},
	}

	paths := []*draw.Path{
		{
			Positions: []vec2d.T{{2, 2}, {8, 8}},
			Weight:    3.0,
		},
	}

	options := &GPXExtrusionOptions{
		Width:        3.0,
		Height:       8.0,
		Depth:        4.0,
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
		t.Error("Expected vertices in result mesh after subtraction")
	}

	if len(result.Indices) == 0 {
		t.Error("Expected indices in result mesh")
	}
}

func TestGPXExtruderMultiplePaths(t *testing.T) {
	terrainMesh := &Mesh{
		Vertices: []vec3d.T{
			{0, 0, 5}, {10, 0, 5}, {10, 10, 5}, {0, 10, 5},
		},
		Indices: []uint32{0, 1, 2, 0, 2, 3},
	}

	paths := []*draw.Path{
		{
			Positions: []vec2d.T{{2, 2}, {8, 2}},
			Weight:    3.0,
		},
		{
			Positions: []vec2d.T{{2, 5}, {8, 5}},
			Weight:    3.0,
		},
	}

	options := &GPXExtrusionOptions{
		Width:        3.0,
		Height:       8.0,
		Depth:        4.0,
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
		t.Error("Expected vertices in result mesh after multiple paths")
	}

	if len(result.Indices) == 0 {
		t.Error("Expected indices in result mesh")
	}
}

func TestGPXExtruderUnion(t *testing.T) {
	terrainMesh := &Mesh{
		Vertices: []vec3d.T{
			{0, 0, 5}, {10, 0, 5}, {10, 10, 5}, {0, 10, 5},
		},
		Indices: []uint32{0, 1, 2, 0, 2, 3},
	}

	paths := []*draw.Path{
		{
			Positions: []vec2d.T{{2, 2}, {8, 8}},
			Weight:    3.0,
		},
	}

	options := &GPXExtrusionOptions{
		Width:        3.0,
		Height:       9.0,
		Depth:        4.0,
		Operation:    BSPOperationUnion,
		Resolution:   1.0,
		SampleRadius: 100.0,
		TerrainMesh:  terrainMesh,
	}

	result, err := ExtrudeGPXPathsToTerrain(paths, terrainMesh, options)
	if err != nil {
		t.Fatalf("Failed to extrude paths with union: %v", err)
	}

	if len(result.Vertices) == 0 {
		t.Error("Expected vertices in union result")
	}
}

func TestGPXExtruderWithoutTerrain(t *testing.T) {
	paths := []*draw.Path{
		{
			Positions: []vec2d.T{{0, 0}, {10, 0}},
			Weight:    3.0,
		},
	}

	options := &GPXExtrusionOptions{
		Width:     3.0,
		Height:    6.0,
		Depth:     3.0,
		Operation: BSPOperationUnion,
	}

	result, err := ExtrudeGPXPathsToTerrain(paths, nil, options)
	if err != nil {
		t.Fatalf("Failed to extrude paths without terrain: %v", err)
	}

	if len(result.Vertices) == 0 {
		t.Error("Expected vertices in result without terrain")
	}

	if len(result.Indices) == 0 {
		t.Error("Expected indices in result")
	}
}

func TestGPXExtruderWithReset(t *testing.T) {
	terrainMesh := &Mesh{
		Vertices: []vec3d.T{
			{0, 0, 5}, {10, 0, 5}, {10, 10, 5}, {0, 10, 5},
		},
		Indices: []uint32{0, 1, 2, 0, 2, 3},
	}

	paths := []*draw.Path{
		{
			Positions: []vec2d.T{{2, 2}, {8, 8}},
			Weight:    3.0,
		},
	}

	options := &GPXExtrusionOptions{
		Width:        3.0,
		Height:       8.0,
		Depth:        4.0,
		Operation:    BSPOperationSubtraction,
		Resolution:   1.0,
		SampleRadius: 100.0,
		TerrainMesh:  terrainMesh,
	}

	extruder := NewGPXPathExtruder()
	extruder.SetBaseMesh(terrainMesh)

	err := extruder.ExtrudePathToTerrain(paths[0], options)
	if err != nil {
		t.Fatalf("Failed to extrude path: %v", err)
	}

	result1 := extruder.GetResult()
	if len(result1.Vertices) == 0 {
		t.Error("Expected vertices after first extrusion")
	}

	extruder.Reset()
	result2 := extruder.GetResult()

	if len(result2.Vertices) != 0 {
		t.Error("Expected empty mesh after reset")
	}

	if extruder.operation != BSPOperationSubtraction {
		t.Errorf("Expected operation Subtraction after reset, got %v", extruder.operation)
	}
}

func TestGPXExtruderInvalidPath(t *testing.T) {
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

	extruder := NewGPXPathExtruder()
	err := extruder.ExtrudePathToTerrain(path, options)
	if err == nil {
		t.Error("Expected error for invalid path with single point")
	}
}

func TestGPXExtruderNilPath(t *testing.T) {
	options := &GPXExtrusionOptions{
		Width:     2.0,
		Height:    5.0,
		Depth:     3.0,
		Operation: BSPOperationUnion,
	}

	extruder := NewGPXPathExtruder()
	err := extruder.ExtrudePathToTerrain(nil, options)
	if err == nil {
		t.Error("Expected error for nil path")
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

func TestCreateTrenchFromGPX(t *testing.T) {
	terrainMesh := &Mesh{
		Vertices: []vec3d.T{
			{0, 0, 5}, {10, 0, 5}, {10, 10, 5}, {0, 10, 5},
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
		t.Error("Expected vertices in trench result")
	}
}

func TestCreateRaisedPathFromGPX(t *testing.T) {
	terrainMesh := &Mesh{
		Vertices: []vec3d.T{
			{0, 0, 5}, {10, 0, 5}, {10, 10, 5}, {0, 10, 5},
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
		t.Error("Expected vertices in raised path result")
	}
}
