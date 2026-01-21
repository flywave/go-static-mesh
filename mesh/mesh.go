package mesh

import (
	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
	"image"
	"image/color"
)

type Mesh struct {
	Vertices  []vec3d.T
	Normals   []vec3d.T
	UVs       []vec2d.T
	Indices   []uint32
	Materials []Material
	Texture   image.Image
	Bounds    vec2d.Rect
	Srs       geo.Proj
	TinMesh   interface{}
}

type Material struct {
	Name      string
	Diffuse   color.Color
	Specular  color.Color
	Shininess float32
	Alpha     float32
}

func (m *Mesh) VertexCount() int {
	return len(m.Vertices)
}

func (m *Mesh) TriangleCount() int {
	return len(m.Indices) / 3
}

func (m *Mesh) CalculateNormals() {
	if len(m.Normals) != len(m.Vertices) {
		m.Normals = make([]vec3d.T, len(m.Vertices))
	}

	for i := range m.Normals {
		m.Normals[i] = vec3d.T{0, 0, 0}
	}

	for i := 0; i < len(m.Indices); i += 3 {
		v0 := m.Vertices[m.Indices[i]]
		v1 := m.Vertices[m.Indices[i+1]]
		v2 := m.Vertices[m.Indices[i+2]]

		edge1 := [3]float64{
			v1[0] - v0[0],
			v1[1] - v0[1],
			v1[2] - v0[2],
		}
		edge2 := [3]float64{
			v2[0] - v0[0],
			v2[1] - v0[1],
			v2[2] - v0[2],
		}

		normal := [3]float64{
			edge1[1]*edge2[2] - edge1[2]*edge2[1],
			edge1[2]*edge2[0] - edge1[0]*edge2[2],
			edge1[0]*edge2[1] - edge1[1]*edge2[0],
		}

		length := vec3d.T(normal).Length()
		if length > 0 {
			normal[0] /= length
			normal[1] /= length
			normal[2] /= length
		}

		m.Normals[m.Indices[i]] = [3]float64{
			m.Normals[m.Indices[i]][0] + normal[0],
			m.Normals[m.Indices[i]][1] + normal[1],
			m.Normals[m.Indices[i]][2] + normal[2],
		}
		m.Normals[m.Indices[i+1]] = [3]float64{
			m.Normals[m.Indices[i+1]][0] + normal[0],
			m.Normals[m.Indices[i+1]][1] + normal[1],
			m.Normals[m.Indices[i+1]][2] + normal[2],
		}
		m.Normals[m.Indices[i+2]] = [3]float64{
			m.Normals[m.Indices[i+2]][0] + normal[0],
			m.Normals[m.Indices[i+2]][1] + normal[1],
			m.Normals[m.Indices[i+2]][2] + normal[2],
		}
	}

	for i := range m.Normals {
		length := vec3d.T(m.Normals[i]).Length()
		if length > 0 {
			m.Normals[i][0] /= length
			m.Normals[i][1] /= length
			m.Normals[i][2] /= length
		}
	}
}

func (m *Mesh) CalculateUVs(bounds vec2d.Rect) {
	if len(m.UVs) != len(m.Vertices) {
		m.UVs = make([]vec2d.T, len(m.Vertices))
	}

	for i, v := range m.Vertices {
		u := (v[0] - bounds.Min[0]) / (bounds.Max[0] - bounds.Min[0])
		vCoord := (v[1] - bounds.Min[1]) / (bounds.Max[1] - bounds.Min[1])
		m.UVs[i] = vec2d.T{u, vCoord}
	}
}

func (m *Mesh) Optimize() error {
	return nil
}

func (m *Mesh) Merge(other *Mesh) error {
	return nil
}
