package writer

import (
	"io"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-static-mesh/mesh"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

type Writer interface {
	Write(mesh *mesh.Mesh, writer io.Writer) error
}

type WriterOptions struct {
	Bounds   vec2d.Rect
	Srs      geo.Proj
	Scale    float64
	Offset   vec3d.T
	FlipAxis bool
	Format   string
}
