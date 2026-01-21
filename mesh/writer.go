package mesh

import (
	"io"
)

type Writer interface {
	Write(mesh *Mesh, path string) error
	WriteTo(mesh *Mesh, w io.Writer) error
}
