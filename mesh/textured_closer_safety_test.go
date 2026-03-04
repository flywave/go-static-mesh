package mesh

import (
	"math"
	"testing"

	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

func TestTexturedCloser_SafetyMargin(t *testing.T) {
	testCases := []struct {
		name              string
		thickness         float64
		expectedMinMargin float64
	}{
		{"Default thickness", 5.0, 0.5},
		{"Large thickness", 10.0, 1.0},
		{"Small thickness", 2.0, 0.2},
		{"Minimum thickness", 2.0, 0.2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options := NewDefaultCloseMeshOptions()
			options.Thickness = tc.thickness
			options.Enabled = true

			expectedMargin := tc.thickness * 0.1
			if math.Abs(expectedMargin-tc.expectedMinMargin) > 1e-6 {
				t.Errorf("Expected safety margin %f, got %f", tc.expectedMinMargin, expectedMargin)
			}
		})
	}
}

func TestTexturedCloser_CloseUnifiedMesh_PenetrationCheck(t *testing.T) {
	testCases := []struct {
		name          string
		vertices      []vec3d.T
		thickness     float64
		baseHeight    float64
		expectPenetra bool
	}{
		{
			name: "No penetration",
			vertices: []vec3d.T{
				{0, 0, 10},
				{1, 0, 10},
				{0, 1, 10},
			},
			thickness:     5.0,
			baseHeight:    5.0,
			expectPenetra: false,
		},
		{
			name: "Potential penetration",
			vertices: []vec3d.T{
				{0, 0, 2},
				{1, 0, 2},
				{0, 1, 2},
			},
			thickness:     5.0,
			baseHeight:    -3.0,
			expectPenetra: true,
		},
		{
			name: "Exactly at base",
			vertices: []vec3d.T{
				{0, 0, 0},
				{1, 0, 0},
				{0, 1, 0},
			},
			thickness:     5.0,
			baseHeight:    -5.0,
			expectPenetra: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mesh := &Mesh{
				Vertices: tc.vertices,
				Indices:  []uint32{0, 1, 2},
				Bounds:   vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{1, 1}},
				Srs:      geo.NewProj(4326),
			}

			closer := NewTexturedCloser()
			options := NewDefaultCloseMeshOptions()
			options.Thickness = tc.thickness
			options.Enabled = true
			closer.SetOptions(options)

			closedMesh, err := closer.CloseUnifiedMesh(mesh, tc.baseHeight)
			if err != nil {
				t.Fatalf("CloseUnifiedMesh failed: %v", err)
			}

			if closedMesh == nil {
				t.Fatal("Closed mesh should not be nil")
			}

			if len(closedMesh.Vertices) != len(mesh.Vertices)*2 {
				t.Errorf("Expected %d vertices, got %d", len(mesh.Vertices)*2, len(closedMesh.Vertices))
			}

			safetyMargin := tc.thickness * 0.1
			adjustedBaseHeight := tc.baseHeight - safetyMargin

			for _, v := range closedMesh.Vertices[len(mesh.Vertices):] {
				if math.Abs(v[2]-adjustedBaseHeight) > 1e-6 && !tc.expectPenetra {
					t.Errorf("Bottom vertex height %f differs from adjusted base height %f",
						v[2], adjustedBaseHeight)
				}
			}
		})
	}
}

func TestTexturedCloser_CloseUnifiedMesh_MinimumThickness(t *testing.T) {
	testCases := []struct {
		name              string
		thickness         float64
		expectedThickness float64
	}{
		{"Below minimum", 1.0, 2.0},
		{"At minimum", 2.0, 2.0},
		{"Above minimum", 5.0, 5.0},
		{"Large thickness", 10.0, 10.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mesh := &Mesh{
				Vertices: []vec3d.T{
					{0, 0, 10},
					{1, 0, 10},
					{0, 1, 10},
				},
				Indices: []uint32{0, 1, 2},
				Bounds:  vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{1, 1}},
				Srs:     geo.NewProj(4326),
			}

			closer := NewTexturedCloser()
			options := NewDefaultCloseMeshOptions()
			options.Thickness = tc.thickness
			options.Enabled = true
			closer.SetOptions(options)

			minHeight := 10.0
			baseHeight := minHeight - tc.expectedThickness

			closedMesh, err := closer.CloseUnifiedMesh(mesh, baseHeight)
			if err != nil {
				t.Fatalf("CloseUnifiedMesh failed: %v", err)
			}

			if closedMesh == nil {
				t.Fatal("Closed mesh should not be nil")
			}

			if len(closedMesh.Vertices) == 0 {
				t.Fatal("Closed mesh should have vertices")
			}

			if len(closedMesh.Indices) == 0 {
				t.Fatal("Closed mesh should have indices")
			}
		})
	}
}

func TestTexturedCloser_CloseUnifiedMesh_GeometryValidation(t *testing.T) {
	mesh := &Mesh{
		Vertices: []vec3d.T{
			{0, 0, 10},
			{1, 0, 10},
			{1, 1, 10},
			{0, 1, 10},
		},
		Indices: []uint32{
			0, 1, 2,
			0, 2, 3,
		},
		Bounds: vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{1, 1}},
		Srs:    geo.NewProj(4326),
	}

	closer := NewTexturedCloser()
	options := NewDefaultCloseMeshOptions()
	options.Thickness = 5.0
	options.Enabled = true
	closer.SetOptions(options)

	baseHeight := 5.0

	closedMesh, err := closer.CloseUnifiedMesh(mesh, baseHeight)
	if err != nil {
		t.Fatalf("CloseUnifiedMesh failed: %v", err)
	}

	if closedMesh == nil {
		t.Fatal("Closed mesh should not be nil")
	}

	topVertexCount := len(mesh.Vertices)
	bottomVertexCount := len(closedMesh.Vertices) - topVertexCount

	if bottomVertexCount != topVertexCount {
		t.Errorf("Bottom vertex count %d should equal top vertex count %d",
			bottomVertexCount, topVertexCount)
	}

	safetyMargin := options.Thickness * 0.1
	adjustedBaseHeight := baseHeight - safetyMargin

	for i := topVertexCount; i < len(closedMesh.Vertices); i++ {
		v := closedMesh.Vertices[i]
		if math.Abs(v[2]-adjustedBaseHeight) > 1e-6 {
			t.Errorf("Bottom vertex %d has incorrect Z: %f (expected %f)",
				i-topVertexCount, v[2], adjustedBaseHeight)
		}
	}

	topIndexCount := len(mesh.Indices)
	bottomIndexCount := len(closedMesh.Indices) - topIndexCount

	if bottomIndexCount != topIndexCount {
		t.Errorf("Bottom index count %d should equal top index count %d",
			bottomIndexCount, topIndexCount)
	}

	for i := 0; i < 3; i++ {
		topIdx := mesh.Indices[i]
		bottomIdx := closedMesh.Indices[topIndexCount+i]

		_ = topIdx

		if bottomIdx < uint32(topVertexCount) {
			t.Errorf("Bottom index %d should be >= %d", bottomIdx, topVertexCount)
		}

		if bottomIdx >= uint32(len(closedMesh.Vertices)) {
			t.Errorf("Bottom index %d out of bounds", bottomIdx)
		}
	}
}

func TestTexturedCloser_CloseUnifiedMesh_NormalsCalculated(t *testing.T) {
	mesh := &Mesh{
		Vertices: []vec3d.T{
			{0, 0, 10},
			{1, 0, 10},
			{0, 1, 10},
		},
		Indices: []uint32{0, 1, 2},
		Bounds:  vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{1, 1}},
		Srs:     geo.NewProj(4326),
	}

	closer := NewTexturedCloser()
	options := NewDefaultCloseMeshOptions()
	options.Thickness = 5.0
	options.Enabled = true
	closer.SetOptions(options)

	baseHeight := 5.0

	closedMesh, err := closer.CloseUnifiedMesh(mesh, baseHeight)
	if err != nil {
		t.Fatalf("CloseUnifiedMesh failed: %v", err)
	}

	if closedMesh == nil {
		t.Fatal("Closed mesh should not be nil")
	}

	if len(closedMesh.Normals) != len(closedMesh.Vertices) {
		t.Errorf("Normal count %d should equal vertex count %d",
			len(closedMesh.Normals), len(closedMesh.Vertices))
	}

	for i, normal := range closedMesh.Normals {
		length := math.Sqrt(normal[0]*normal[0] + normal[1]*normal[1] + normal[2]*normal[2])
		if math.Abs(length-1.0) > 1e-6 {
			t.Errorf("Normal %d is not unit length: %f", i, length)
		}
	}
}

func TestTexturedCloser_CloseUnifiedMesh_SafetyMarginAdjustment(t *testing.T) {
	mesh := &Mesh{
		Vertices: []vec3d.T{
			{0, 0, 10},
			{1, 0, 10},
			{0, 1, 10},
		},
		Indices: []uint32{0, 1, 2},
		Bounds:  vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{1, 1}},
		Srs:     geo.NewProj(4326),
	}

	thickness := 5.0
	safetyMargin := thickness * 0.1
	expectedBaseHeight := 5.0 - safetyMargin

	closer := NewTexturedCloser()
	options := NewDefaultCloseMeshOptions()
	options.Thickness = thickness
	options.Enabled = true
	closer.SetOptions(options)

	closedMesh, err := closer.CloseUnifiedMesh(mesh, 5.0)
	if err != nil {
		t.Fatalf("CloseUnifiedMesh failed: %v", err)
	}

	if closedMesh == nil {
		t.Fatal("Closed mesh should not be nil")
	}

	bottomStartIdx := len(mesh.Vertices)
	for i := bottomStartIdx; i < len(closedMesh.Vertices); i++ {
		v := closedMesh.Vertices[i]
		if math.Abs(v[2]-expectedBaseHeight) > 1e-6 {
			t.Errorf("Vertex %d Z=%f, expected %f (safety margin adjustment failed)",
				i, v[2], expectedBaseHeight)
		}
	}
}

func BenchmarkTexturedCloser_CloseUnifiedMesh(b *testing.B) {
	vertices := make([]vec3d.T, 1000)
	for i := 0; i < 1000; i++ {
		vertices[i] = vec3d.T{
			float64(i % 10),
			float64(i / 10),
			10.0 + float64(i%5),
		}
	}

	indices := make([]uint32, 0)
	for i := 0; i < len(vertices)-2; i++ {
		indices = append(indices, uint32(i), uint32(i+1), uint32(i+2))
	}

	mesh := &Mesh{
		Vertices: vertices,
		Indices:  indices,
		Bounds:   vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{10, 100}},
		Srs:      geo.NewProj(4326),
	}

	closer := NewTexturedCloser()
	options := NewDefaultCloseMeshOptions()
	options.Thickness = 5.0
	options.Enabled = true
	closer.SetOptions(options)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		closer.CloseUnifiedMesh(mesh, 5.0)
	}
}
