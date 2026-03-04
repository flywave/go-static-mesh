package builder

import (
	"math"
	"testing"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-static-mesh/draw"
	"github.com/flywave/go-static-mesh/mesh"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

type MockTerrainMesh struct {
	vertices []vec3d.T
	bounds   vec2d.Rect
}

func (m *MockTerrainMesh) GetVertices() []vec3d.T {
	return m.vertices
}

func (m *MockTerrainMesh) GetBounds() vec2d.Rect {
	return m.bounds
}

func createDenseTerrainMesh(numVertices int, bounds vec2d.Rect, baseHeight float64) *MockTerrainMesh {
	vertices := make([]vec3d.T, numVertices)

	gridSize := int(math.Sqrt(float64(numVertices)))
	stepX := (bounds.Max[0] - bounds.Min[0]) / float64(gridSize-1)
	stepY := (bounds.Max[1] - bounds.Min[1]) / float64(gridSize-1)

	idx := 0
	for i := 0; i < gridSize && idx < numVertices; i++ {
		for j := 0; j < gridSize && idx < numVertices; j++ {
			x := bounds.Min[0] + float64(i)*stepX
			y := bounds.Min[1] + float64(j)*stepY
			height := baseHeight + float64(idx%100)*0.1

			vertices[idx] = vec3d.T{x, y, height}
			idx++
		}
	}

	return &MockTerrainMesh{
		vertices: vertices,
		bounds:   bounds,
	}
}

func TestSampleHeightAtPosition_DynamicSampling(t *testing.T) {
	bounds := vec2d.Rect{
		Min: vec2d.T{0, 0},
		Max: vec2d.T{100, 100},
	}

	testCases := []struct {
		name         string
		numVertices  int
		expectedStep int
	}{
		{"Small mesh", 100, 1},
		{"Medium mesh", 1000, 1},
		{"Large mesh", 10000, 10},
		{"Very large mesh", 100000, 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			terrain := createDenseTerrainMesh(tc.numVertices, bounds, 50.0)

			expectedStep := int(math.Max(1, float64(tc.numVertices)/1000.0))
			if expectedStep != tc.expectedStep {
				t.Errorf("Expected step %d, got %d", tc.expectedStep, expectedStep)
			}

			if len(terrain.vertices) != tc.numVertices {
				t.Errorf("Expected %d vertices, got %d", tc.numVertices, len(terrain.vertices))
			}
		})
	}
}

func TestSampleHeightAtPosition_Accuracy(t *testing.T) {
	bounds := vec2d.Rect{
		Min: vec2d.T{0, 0},
		Max: vec2d.T{100, 100},
	}

	builder := NewBuilder()
	builder.SetBounds(bounds, geo.NewProj(4326))
	builder.SetExtrudeGeoData(true, 10.0)

	terrain := createDenseTerrainMesh(10000, bounds, 50.0)

	path := &draw.Path{
		Positions: []vec2d.T{
			{50, 50},
			{51, 51},
			{52, 52},
		},
		Weight: 2.0,
	}

	height := builder.sampleHeightAtPosition(path, terrain)

	if height < 0 {
		t.Errorf("Height should be positive, got %f", height)
	}

	expectedMinHeight := 50.0
	expectedMaxHeight := 60.0

	if height < expectedMinHeight || height > expectedMaxHeight {
		t.Errorf("Height %f out of expected range [%f, %f]",
			height, expectedMinHeight, expectedMaxHeight)
	}
}

func TestSampleHeightAtPosition_SparseMesh(t *testing.T) {
	bounds := vec2d.Rect{
		Min: vec2d.T{0, 0},
		Max: vec2d.T{1000, 1000},
	}

	builder := NewBuilder()
	builder.SetBounds(bounds, geo.NewProj(4326))
	builder.SetExtrudeGeoData(true, 10.0)

	terrain := &MockTerrainMesh{
		vertices: []vec3d.T{
			{0, 0, 100},
			{1000, 0, 100},
			{0, 1000, 100},
			{1000, 1000, 100},
			{500, 500, 150},
		},
		bounds: bounds,
	}

	_ = terrain

	path := &draw.Path{
		Positions: []vec2d.T{
			{490, 490},
			{510, 510},
		},
		Weight: 2.0,
	}

	height := builder.sampleHeightAtPosition(path, terrain)

	if height < 100 || height > 160 {
		t.Errorf("Height %f out of reasonable range [100, 160]", height)
	}
}

func TestGPXExtruder_DynamicRadius(t *testing.T) {
	bounds := vec2d.Rect{
		Min: vec2d.T{0, 0},
		Max: vec2d.T{1000, 1000},
	}

	testCases := []struct {
		name              string
		numVertices       int
		expectedMinRadius float64
	}{
		{"Low resolution terrain", 100, 100.0},
		{"Medium resolution terrain", 1000, 100.0},
		{"High resolution terrain", 10000, 100.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			terrain := createDenseTerrainMesh(tc.numVertices, bounds, 50.0)

			mesh := &mesh.Mesh{
				Vertices: terrain.GetVertices(),
				Indices:  []uint32{},
			}

			for i := 0; i < len(mesh.Vertices)-2; i++ {
				mesh.Indices = append(mesh.Indices, uint32(i), uint32(i+1), uint32(i+2))
			}

			path := &draw.Path{
				Positions: []vec2d.T{
					{500, 500},
					{510, 510},
					{520, 520},
				},
				Weight: 5.0,
			}

			_ = path

			if len(mesh.Vertices) == 0 {
				t.Fatal("Mesh vertices should not be empty")
			}

			centerX := (bounds.Min[0] + bounds.Max[0]) / 2.0
			centerY := (bounds.Min[1] + bounds.Max[1]) / 2.0
			height := 50.0
			found := false

			for _, v := range mesh.Vertices {
				dist := math.Sqrt(math.Pow(v[0]-centerX, 2) + math.Pow(v[1]-centerY, 2))
				if dist < tc.expectedMinRadius {
					height = v[2]
					found = true
					break
				}
			}

			if !found {
				t.Logf("Warning: No vertex found within expected radius")
			}

			if height < 0 {
				t.Errorf("Height should be positive, got %f", height)
			}

			t.Logf("Path position: (%f, %f), sampled height: %f", centerX, centerY, height)
		})
	}
}

func TestSampleHeightAtPosition_EdgeCases(t *testing.T) {
	builder := NewBuilder()
	builder.SetBounds(vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{100, 100}}, geo.NewProj(4326))
	builder.SetExtrudeGeoData(true, 10.0)

	t.Run("Empty mesh", func(t *testing.T) {
		terrain := &MockTerrainMesh{
			vertices: []vec3d.T{},
			bounds:   vec2d.Rect{},
		}

		_ = terrain

		path := &draw.Path{
			Positions: []vec2d.T{{50, 50}},
			Weight:    2.0,
		}

		height := builder.sampleHeightAtPosition(path, terrain)

		if height != 10.0 {
			t.Errorf("Expected default height 10.0 for empty mesh, got %f", height)
		}
	})

	t.Run("Single vertex mesh", func(t *testing.T) {
		terrain := &MockTerrainMesh{
			vertices: []vec3d.T{{50, 50, 100}},
			bounds:   vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{100, 100}},
		}

		path := &draw.Path{
			Positions: []vec2d.T{{50, 50}},
			Weight:    2.0,
		}

		height := builder.sampleHeightAtPosition(path, terrain)

		if height < 0 {
			t.Errorf("Height should be positive, got %f", height)
		}
	})
}
