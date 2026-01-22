package writer

import (
	"fmt"
	"io"
	"os"

	"github.com/flywave/go-static-mesh/mesh"
)

type OBJWriter struct {
	IncludeNormals bool
	IncludeUVs     bool
	SeparateMtl    bool
}

func NewOBJWriter(includeNormals, includeUVs bool) *OBJWriter {
	return &OBJWriter{
		IncludeNormals: includeNormals,
		IncludeUVs:     includeUVs,
	}
}

func NewObjWriter() *OBJWriter {
	return &OBJWriter{
		IncludeNormals: true,
		IncludeUVs:     true,
	}
}

func (w *OBJWriter) Write(mesh *mesh.Mesh, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return w.WriteTo(mesh, file)
}

func (w *OBJWriter) WriteTo(m *mesh.Mesh, writer io.Writer) error {
	_, err := io.WriteString(writer, "# Exported using go-static-mesh\n")
	if err != nil {
		return err
	}

	_, err = io.WriteString(writer, fmt.Sprintf("# %d vertices, %d faces\n", len(m.Vertices), len(m.Indices)/3))
	if err != nil {
		return err
	}

	for _, v := range m.Vertices {
		_, err = io.WriteString(writer, fmt.Sprintf("v %g %g %g\n", v[0], v[1], v[2]))
		if err != nil {
			return err
		}
	}

	if w.IncludeNormals && len(m.Normals) > 0 {
		for _, n := range m.Normals {
			_, err = io.WriteString(writer, fmt.Sprintf("vn %g %g %g\n", n[0], n[1], n[2]))
			if err != nil {
				return err
			}
		}
	}

	if w.IncludeUVs && len(m.UVs) > 0 {
		for _, uv := range m.UVs {
			_, err = io.WriteString(writer, fmt.Sprintf("vt %g %g\n", uv[0], uv[1]))
			if err != nil {
				return err
			}
		}
	}

	_, err = io.WriteString(writer, "g mesh\n")
	if err != nil {
		return err
	}

	for i := 0; i < len(m.Indices); i += 3 {
		line := "f"
		if w.IncludeNormals && len(m.Normals) > 0 {
			if w.IncludeUVs && len(m.UVs) > 0 {
				line += fmt.Sprintf(" %d/%d/%d %d/%d/%d %d/%d/%d",
					m.Indices[i]+1, m.Indices[i]+1, m.Indices[i]+1,
					m.Indices[i+1]+1, m.Indices[i+1]+1, m.Indices[i+1]+1,
					m.Indices[i+2]+1, m.Indices[i+2]+1, m.Indices[i+2]+1)
			} else {
				line += fmt.Sprintf(" %d//%d %d//%d %d//%d",
					m.Indices[i]+1, m.Indices[i]+1,
					m.Indices[i+1]+1, m.Indices[i+1]+1,
					m.Indices[i+2]+1, m.Indices[i+2]+1)
			}
		} else if w.IncludeUVs && len(m.UVs) > 0 {
			line += fmt.Sprintf(" %d/%d %d/%d %d/%d",
				m.Indices[i]+1, m.Indices[i]+1,
				m.Indices[i+1]+1, m.Indices[i+1]+1,
				m.Indices[i+2]+1, m.Indices[i+2]+1)
		} else {
			line += fmt.Sprintf(" %d %d %d",
				m.Indices[i]+1,
				m.Indices[i+1]+1,
				m.Indices[i+2]+1)
		}
		_, err = io.WriteString(writer, line+"\n")
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *OBJWriter) WriteFile(mesh *mesh.Mesh, path string) error {
	return w.Write(mesh, path)
}

func (w *OBJWriter) SetSeparateMtl(separate bool) {
	w.SeparateMtl = separate
}
