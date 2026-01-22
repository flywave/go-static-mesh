package extruder

import (
	"testing"

	"github.com/flywave/go-static-mesh/draw"
	"github.com/flywave/go-static-mesh/mesh"
	bsp "github.com/flywave/go-static-mesh/mesh/bsp"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

func TestGPXExtruderMultiplePaths(t *testing.T) {
	terrainMesh := &mesh.Mesh{
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
		Operation:    bsp.BSPOperationSubtraction,
		Resolution:   1.0,
		SampleRadius: 100.0,
		TerrainMesh:  terrainMesh,
	}

	result, err := ExtrudeGPXPathsToTerrain(paths, terrainMesh, options)
	if err != nil {
		t.Fatalf("Failed to extrude paths: %v", err)
	}

	if len(result.Vertices) == 0 {
		t.Error("Expected vertices in result")
	}

	if len(result.Vertices) == 0 {
		t.Error("Expected vertices in result mesh after multiple paths")
	}

	if len(result.Indices) == 0 {
		t.Error("Expected indices in result mesh")
	}
}

func TestGPXExtruderUnion(t *testing.T) {
	terrainMesh := &mesh.Mesh{
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
		Operation:    bsp.BSPOperationUnion,
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
		Operation: bsp.BSPOperationUnion,
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

func TestGPXExtruderInvalidPath(t *testing.T) {
	path := &draw.Path{
		Positions: []vec2d.T{{0, 0}},
		Weight:    2.0,
	}

	options := &GPXExtrusionOptions{
		Width:     2.0,
		Height:    5.0,
		Depth:     3.0,
		Operation: bsp.BSPOperationUnion,
	}

	extruder := NewGPXPathExtruder()
	err := extruder.ExtrudePathToTerrain(path, options)
	if err == nil {
		t.Error("Expected error for invalid path with single point")
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

	if options.Operation != bsp.BSPOperationUnion {
		t.Errorf("Expected default Operation Union (0), got %v", options.Operation)
	}

	if options.SampleRadius != 0 {
		t.Errorf("Expected default SampleRadius 0, got %v", options.SampleRadius)
	}
}

func TestCreateTrenchFromGPX(t *testing.T) {
	terrainMesh := &mesh.Mesh{
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
	terrainMesh := &mesh.Mesh{
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
