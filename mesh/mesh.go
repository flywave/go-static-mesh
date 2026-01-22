package mesh

import (
	"math"

	"image"
	"image/color"

	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
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
	if len(m.Indices) == 0 {
		return nil
	}

	verticesToRemove := make(map[uint32]bool)
	indicesToRemove := make(map[uint32]bool)

	merged := make(map[[3]uint32]uint32)
	for i := 0; i < len(m.Indices); i += 3 {
		if i+2 >= len(m.Indices) {
			break
		}

		v0 := m.Indices[i]
		v1 := m.Indices[i+1]
		v2 := m.Indices[i+2]

		if v0 >= uint32(len(m.Vertices)) || v1 >= uint32(len(m.Vertices)) || v2 >= uint32(len(m.Vertices)) {
			continue
		}

		if verticesToRemove[v0] || verticesToRemove[v1] || verticesToRemove[v2] {
			continue
		}

		tri := [3]uint32{v0, v1, v2}
		if existing, ok := merged[tri]; ok {
			indicesToRemove[existing] = true
			continue
		}

		merged[tri] = uint32(i / 3)
	}

	newIndices := make([]uint32, 0, len(m.Indices))
	for i := 0; i < len(m.Indices); i++ {
		if !indicesToRemove[uint32(i)] {
			newIndices = append(newIndices, m.Indices[i])
		}
	}

	newVertices := make([]vec3d.T, 0, len(m.Vertices))
	vertexMap := make(map[uint32]uint32)

	for i := 0; i < len(newIndices); i++ {
		oldIdx := newIndices[i]
		if _, ok := vertexMap[oldIdx]; !ok {
			vertexMap[oldIdx] = uint32(len(newVertices))
			newVertices = append(newVertices, m.Vertices[oldIdx])
		}
		newIndices[i] = vertexMap[oldIdx]
	}

	m.Vertices = newVertices
	m.Indices = newIndices

	if len(m.Normals) > 0 {
		m.CalculateNormals()
	}

	if len(m.UVs) > 0 {
		m.CalculateUVs(m.Bounds)
	}

	return nil
}

func (m *Mesh) Merge(other *Mesh) error {
	if other == nil {
		return nil
	}

	baseVertexIndex := uint32(len(m.Vertices))

	m.Vertices = append(m.Vertices, other.Vertices...)
	m.Normals = append(m.Normals, other.Normals...)

	for i := range other.UVs {
		m.UVs = append(m.UVs, other.UVs[i])
	}

	for _, idx := range other.Indices {
		m.Indices = append(m.Indices, baseVertexIndex+idx)
	}

	for _, mat := range other.Materials {
		m.Materials = append(m.Materials, mat)
	}

	if m.Bounds.Min[0] > other.Bounds.Min[0] || m.Bounds.Min[1] > other.Bounds.Min[1] {
		m.Bounds.Min[0] = math.Min(m.Bounds.Min[0], other.Bounds.Min[0])
		m.Bounds.Min[1] = math.Min(m.Bounds.Min[1], other.Bounds.Min[1])
	}
	if m.Bounds.Max[0] < other.Bounds.Max[0] || m.Bounds.Max[1] < other.Bounds.Max[1] {
		m.Bounds.Max[0] = math.Max(m.Bounds.Max[0], other.Bounds.Max[0])
		m.Bounds.Max[1] = math.Max(m.Bounds.Max[1], other.Bounds.Max[1])
	}

	for _, idx := range other.Indices {
		m.Indices = append(m.Indices, baseVertexIndex+idx)
	}

	for _, mat := range other.Materials {
		m.Materials = append(m.Materials, mat)
	}

	if m.Bounds.Min[0] > other.Bounds.Min[0] || m.Bounds.Min[1] > other.Bounds.Min[1] {
		m.Bounds.Min[0] = math.Min(m.Bounds.Min[0], other.Bounds.Min[0])
		m.Bounds.Min[1] = math.Min(m.Bounds.Min[1], other.Bounds.Min[1])
	}
	if m.Bounds.Max[0] < other.Bounds.Max[0] || m.Bounds.Max[1] < other.Bounds.Max[1] {
		m.Bounds.Max[0] = math.Max(m.Bounds.Max[0], other.Bounds.Max[0])
		m.Bounds.Max[1] = math.Max(m.Bounds.Max[1], other.Bounds.Max[1])
	}

	if m.TinMesh != nil && other.TinMesh != nil {
		type meshWithVertices interface {
			GetVertices() []vec3d.T
			GetIndices() []uint32
		}

		tin1, ok1 := m.TinMesh.(meshWithVertices)
		tin2, ok2 := other.TinMesh.(meshWithVertices)
		if ok1 && ok2 {
			mergedTin := &TinMesh{
				Vertices:  append(tin1.GetVertices(), tin2.GetVertices()...),
				Indices:   append(tin1.GetIndices(), tin2.GetIndices()...),
				MinHeight: math.Min(tin1.GetMinHeight(), tin2.GetMinHeight()),
				MaxHeight: math.Max(tin1.GetMaxHeight(), tin2.GetMaxHeight()),
			}
			m.TinMesh = mergedTin
		}
	}

	return nil
}
