package mesh

import (
	"image/color"
	"testing"

	vec3d "github.com/flywave/go3d/float64/vec3"
)

func TestNewPBRMaterial(t *testing.T) {
	baseColor := color.RGBA{255, 0, 0, 255}
	material := NewPBRMaterial("test", baseColor, 0.5, 0.3)

	if material == nil {
		t.Fatal("NewPBRMaterial returned nil")
	}

	if material.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", material.Name)
	}

	if material.Metalness != 0.5 {
		t.Errorf("expected metalness 0.5, got %f", material.Metalness)
	}

	if material.Roughness != 0.3 {
		t.Errorf("expected roughness 0.3, got %f", material.Roughness)
	}
}

func TestMeshVertexCount(t *testing.T) {
	mesh := &Mesh{
		Vertices: []vec3d.T{
			{0, 0, 0},
			{1, 0, 0},
			{0, 1, 0},
		},
	}

	count := mesh.VertexCount()
	if count != 3 {
		t.Errorf("expected vertex count 3, got %d", count)
	}
}

func TestMeshAppendVertex(t *testing.T) {
	mesh := &Mesh{
		Vertices: []vec3d.T{},
	}

	index := mesh.AppendVertex(1.0, 2.0, 3.0)

	if index != 0 {
		t.Errorf("expected index 0, got %d", index)
	}

	if len(mesh.Vertices) != 1 {
		t.Errorf("expected 1 vertex, got %d", len(mesh.Vertices))
	}

	if mesh.Vertices[0] != (vec3d.T{1.0, 2.0, 3.0}) {
		t.Errorf("vertex not added correctly")
	}
}

func TestMeshAppendTriangle(t *testing.T) {
	mesh := &Mesh{
		Indices: []uint32{},
	}

	mesh.AppendTriangle(0, 1, 2)

	if len(mesh.Indices) != 3 {
		t.Errorf("expected 3 indices, got %d", len(mesh.Indices))
	}

	if mesh.Indices[0] != 0 || mesh.Indices[1] != 1 || mesh.Indices[2] != 2 {
		t.Errorf("triangle indices not added correctly")
	}
}

func TestMeshGetVertices(t *testing.T) {
	vertices := []vec3d.T{
		{0, 0, 0},
		{1, 0, 0},
		{0, 1, 0},
	}

	mesh := &Mesh{
		Vertices: vertices,
	}

	retrieved := mesh.GetVertices()
	retrievedSlice, ok := retrieved.([]vec3d.T)
	if !ok {
		t.Fatal("GetVertices did not return []vec3d.T")
	}

	if len(retrievedSlice) != len(vertices) {
		t.Errorf("expected %d vertices, got %d", len(vertices), len(retrievedSlice))
	}

	for i, v := range retrievedSlice {
		if v != vertices[i] {
			t.Errorf("vertex %d mismatch", i)
		}
	}
}

func TestNewCoordinateError(t *testing.T) {
	err := NewCoordinateError("test op", nil)

	if err == nil {
		t.Fatal("NewCoordinateError returned nil")
	}

	if err.Type != ErrTypeCoordinate {
		t.Errorf("expected error type ErrTypeCoordinate, got %d", err.Type)
	}

	if err.Operation != "test op" {
		t.Errorf("expected operation 'test op', got '%s'", err.Operation)
	}
}

func TestNewMemoryError(t *testing.T) {
	err := NewMemoryError("test op", nil)

	if err == nil {
		t.Fatal("NewMemoryError returned nil")
	}

	if err.Type != ErrTypeMemory {
		t.Errorf("expected error type ErrTypeMemory, got %d", err.Type)
	}

	if err.Operation != "test op" {
		t.Errorf("expected operation 'test op', got '%s'", err.Operation)
	}
}

func TestBuildError(t *testing.T) {
	buildErr := &BuildError{
		Stage: "test_stage",
		Step:  "test_step",
		Err:   nil,
	}

	errMsg := buildErr.Error()
	if errMsg != "test_stage" {
		t.Errorf("expected error message 'test_stage', got '%s'", errMsg)
	}
}

func TestBuildErrorWithStep(t *testing.T) {
	buildErr := &BuildError{
		Stage: "test_stage",
		Step:  "test_step",
		Err:   NewTINError("test", nil),
	}

	errMsg := buildErr.Error()
	expectedPrefix := "test_stage: test_step:"
	if len(errMsg) < len(expectedPrefix) || errMsg[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("expected error message to start with '%s', got '%s'", expectedPrefix, errMsg)
	}
}

func TestBuildErrorUnwrap(t *testing.T) {
	underlyingErr := NewTINError("test", nil)
	buildErr := &BuildError{
		Stage: "test_stage",
		Err:   underlyingErr,
	}

	unwrapped := buildErr.Unwrap()
	if unwrapped != underlyingErr {
		t.Error("Unwrap did not return the underlying error")
	}
}
