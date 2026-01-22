package mesh

import (
	"fmt"
	"io"
	"math"
	"os"

	stl "github.com/flywave/go-stl"
	vec3d "github.com/flywave/go3d/float64/vec3"
	vec3 "github.com/flywave/go3d/vec3"
)

type STLWriter struct {
	ASCII     bool
	MergeMesh bool
}

func NewSTLWriter(ascii bool) *STLWriter {
	return &STLWriter{ASCII: ascii}
}

func NewStlWriter() *STLWriter {
	return &STLWriter{ASCII: false}
}

func (w *STLWriter) SetBinary(binary bool) {
	w.ASCII = !binary
}

func (w *STLWriter) SetMergeMesh(merge bool) {
	w.MergeMesh = merge
}

func (w *STLWriter) WriteFile(mesh *Mesh, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return w.WriteTo(mesh, file)
}

func (w *STLWriter) Write(mesh *Mesh, path string) error {
	return w.WriteFile(mesh, path)
}

func (w *STLWriter) WriteTo(mesh *Mesh, writer io.Writer) error {
	solid := &stl.Solid{
		Name:    "mesh",
		IsAscii: w.ASCII,
	}

	triangles := make([]stl.Triangle, len(mesh.Indices)/3)

	for i := 0; i < len(mesh.Indices); i += 3 {
		v0 := mesh.Vertices[mesh.Indices[i]]
		v1 := mesh.Vertices[mesh.Indices[i+1]]
		v2 := mesh.Vertices[mesh.Indices[i+2]]

		normal := w.calculateNormal(v0, v1, v2)

		triangles[i/3] = stl.Triangle{
			Normal: normal,
			Vertices: [3]vec3.T{
				{float32(v0[0]), float32(v0[1]), float32(v0[2])},
				{float32(v1[0]), float32(v1[1]), float32(v1[2])},
				{float32(v2[0]), float32(v2[1]), float32(v2[2])},
			},
		}
	}

	solid.Triangles = triangles

	return solid.WriteAll(writer)
}

func (w *STLWriter) calculateNormal(v0, v1, v2 vec3d.T) vec3.T {
	edge1 := vec3d.T{v1[0] - v0[0], v1[1] - v0[1], v1[2] - v0[2]}
	edge2 := vec3d.T{v2[0] - v0[0], v2[1] - v0[1], v2[2] - v0[2]}

	cross := vec3d.T{
		edge1[1]*edge2[2] - edge1[2]*edge2[1],
		edge1[2]*edge2[0] - edge1[0]*edge2[2],
		edge1[0]*edge2[1] - edge1[1]*edge2[0],
	}

	length := w.vectorLength(cross)

	if length > 0 {
		return vec3.T{
			float32(cross[0] / length),
			float32(cross[1] / length),
			float32(cross[2] / length),
		}
	}
	return vec3.T{0, 0, 1}
}

func (w *STLWriter) vectorLength(v vec3d.T) float64 {
	return math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
}
