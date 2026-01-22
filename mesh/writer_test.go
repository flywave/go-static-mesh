package mesh

import (
	"bytes"
	"os"
	"testing"

	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

func createTestMesh() *Mesh {
	vertices := []vec3d.T{
		{0, 0, 0},
		{1, 0, 0},
		{1, 1, 0},
		{0, 1, 0},
	}

	indices := []uint32{0, 1, 2, 0, 2, 3}

	normals := []vec3d.T{
		{0, 0, 1},
		{0, 0, 1},
		{0, 0, 1},
		{0, 0, 1},
	}

	return &Mesh{
		Vertices: vertices,
		Indices:  indices,
		Normals:  normals,
		UVs:      []vec2d.T{{0, 0}, {1, 0}, {1, 1}, {0, 1}},
		Bounds:   vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{1, 1}},
		Srs:      geo.NewProj(4326),
	}
}

func TestSTLWriter(t *testing.T) {
	mesh := createTestMesh()
	writer := NewSTLWriter(false)

	tmpFile, err := os.CreateTemp("", "test-*.stl")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	err = writer.WriteFile(mesh, tmpFile.Name())
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	info, err := os.Stat(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Size() == 0 {
		t.Error("File is empty")
	}

	t.Logf("STL file written: %d bytes", info.Size())
}

func TestSTLWriterBinary(t *testing.T) {
	mesh := createTestMesh()
	writer := NewSTLWriter(true)
	writer.SetBinary(true)

	tmpFile, err := os.CreateTemp("", "test-*.stl")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	err = writer.WriteFile(mesh, tmpFile.Name())
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	info, err := os.Stat(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Size() == 0 {
		t.Error("File is empty")
	}

	t.Logf("Binary STL file written: %d bytes", info.Size())
}

func TestSTLWriterToMemory(t *testing.T) {
	mesh := createTestMesh()
	writer := NewSTLWriter(false)

	var buf bytes.Buffer
	err := writer.WriteTo(mesh, &buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Buffer is empty")
	}

	t.Logf("STL data written: %d bytes", buf.Len())
}

func TestGLTFWriter(t *testing.T) {
	mesh := createTestMesh()
	writer := NewGLTFWriter(false, true)

	tmpFile, err := os.CreateTemp("", "test-*.gltf")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	err = writer.WriteFile(mesh, tmpFile.Name())
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	info, err := os.Stat(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Size() == 0 {
		t.Error("File is empty")
	}

	t.Logf("GLTF file written: %d bytes", info.Size())
}

func TestGLBWriter(t *testing.T) {
	mesh := createTestMesh()
	writer := NewGLTFWriter(true, true)
	writer.SetBinary(true)

	tmpFile, err := os.CreateTemp("", "test-*.glb")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	err = writer.WriteFile(mesh, tmpFile.Name())
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	info, err := os.Stat(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Size() == 0 {
		t.Error("File is empty")
	}

	t.Logf("GLB file written: %d bytes", info.Size())
}

func TestGLTFWriterWithoutUVs(t *testing.T) {
	mesh := createTestMesh()
	writer := NewGLTFWriter(false, false)
	writer.SetIncludeUVs(false)

	tmpFile, err := os.CreateTemp("", "test-*.gltf")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	err = writer.WriteFile(mesh, tmpFile.Name())
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	info, err := os.Stat(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Size() == 0 {
		t.Error("File is empty")
	}

	t.Logf("GLTF file (no UVs) written: %d bytes", info.Size())
}

func TestOBJWriter(t *testing.T) {
	mesh := createTestMesh()
	writer := NewOBJWriter(true, true)

	tmpFile, err := os.CreateTemp("", "test-*.obj")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	err = writer.WriteFile(mesh, tmpFile.Name())
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	info, err := os.Stat(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Size() == 0 {
		t.Error("File is empty")
	}

	t.Logf("OBJ file written: %d bytes", info.Size())
}

func TestOBJWriterWithoutNormals(t *testing.T) {
	mesh := createTestMesh()
	writer := NewOBJWriter(false, true)

	tmpFile, err := os.CreateTemp("", "test-*.obj")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	err = writer.WriteFile(mesh, tmpFile.Name())
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	info, err := os.Stat(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Size() == 0 {
		t.Error("File is empty")
	}

	t.Logf("OBJ file (no normals) written: %d bytes", info.Size())
}

func TestOBJWriterToMemory(t *testing.T) {
	mesh := createTestMesh()
	writer := NewOBJWriter(true, false)

	var buf bytes.Buffer
	err := writer.WriteTo(mesh, &buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Buffer is empty")
	}

	t.Logf("OBJ data written: %d bytes", buf.Len())
}
