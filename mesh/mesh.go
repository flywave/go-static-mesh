package mesh

import (
	"math"

	"image"
	"image/color"

	"github.com/flywave/go-geo"
	tile "github.com/flywave/go-static-mesh/tile"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

func (m *Mesh) CalculateUVsFromExtent() {
	if len(m.Vertices) == 0 {
		return
	}

	minX := math.Inf(1)
	minY := math.Inf(1)
	maxX := math.Inf(-1)
	maxY := math.Inf(-1)

	for _, v := range m.Vertices {
		if v[0] < minX {
			minX = v[0]
		}
		if v[0] > maxX {
			maxX = v[0]
		}
		if v[1] < minY {
			minY = v[1]
		}
		if v[1] > maxY {
			maxY = v[1]
		}
	}

	bounds := vec2d.Rect{Min: vec2d.T{minX, minY}, Max: vec2d.T{maxX, maxY}}
	m.CalculateUVs(bounds)
}

type Tile struct {
	Coord     [3]int
	Zoom      int
	Data      []float64
	Elevation *float64
	Bounds    vec2d.Rect
	Width     int
	Height    int
}

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

	MaterialIndices []uint32
}

type Material struct {
	Name      string
	Diffuse   color.Color
	Texture   image.Image
	Specular  color.Color
	Shininess float32
	Alpha     float32

	Metalness         float32
	Roughness         float32
	Ao                float32
	Emissive          color.Color
	EmissiveIntensity float32
	NormalStrength    float32
	Displacement      float32
}

func NewMaterial() *Material {
	return &Material{
		Name:              "default",
		Diffuse:           color.RGBA{200, 200, 200, 255},
		Specular:          color.RGBA{255, 255, 255, 255},
		Shininess:         32,
		Alpha:             1.0,
		Metalness:         0.0,
		Roughness:         0.5,
		Ao:                1.0,
		Emissive:          color.RGBA{0, 0, 0, 255},
		EmissiveIntensity: 1.0,
		NormalStrength:    1.0,
		Displacement:      0.0,
	}
}

func NewPBRMaterial(name string, baseColor color.Color, metalness, roughness float32) *Material {
	return &Material{
		Name:              name,
		Diffuse:           baseColor,
		Specular:          color.RGBA{255, 255, 255, 255},
		Shininess:         32,
		Alpha:             1.0,
		Metalness:         metalness,
		Roughness:         roughness,
		Ao:                1.0,
		Emissive:          color.RGBA{0, 0, 0, 255},
		EmissiveIntensity: 1.0,
		NormalStrength:    1.0,
		Displacement:      0.0,
	}
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

	vertexNormalAccum := make([]vec3d.T, len(m.Vertices))
	vertexFaceCount := make([]int, len(m.Vertices))
	vertexFaceArea := make([]float64, len(m.Vertices))

	for i := 0; i < len(m.Indices); i += 3 {
		if i+2 >= len(m.Indices) {
			break
		}

		idx0, idx1, idx2 := m.Indices[i], m.Indices[i+1], m.Indices[i+2]

		if idx0 >= uint32(len(m.Vertices)) || idx1 >= uint32(len(m.Vertices)) || idx2 >= uint32(len(m.Vertices)) {
			continue
		}

		v0 := m.Vertices[idx0]
		v1 := m.Vertices[idx1]
		v2 := m.Vertices[idx2]

		edge1 := [3]float64{v1[0] - v0[0], v1[1] - v0[1], v1[2] - v0[2]}
		edge2 := [3]float64{v2[0] - v0[0], v2[1] - v0[1], v2[2] - v0[2]}

		normal := [3]float64{
			edge1[1]*edge2[2] - edge1[2]*edge2[1],
			edge1[2]*edge2[0] - edge1[0]*edge2[2],
			edge1[0]*edge2[1] - edge1[1]*edge2[0],
		}

		area := 0.5 * vec3d.T(normal).Length()

		normalLength := vec3d.T(normal).Length()
		if normalLength > 0 {
			normal[0] /= normalLength
			normal[1] /= normalLength
			normal[2] /= normalLength
		}

		for _, idx := range []uint32{idx0, idx1, idx2} {
			weight := area
			vertexNormalAccum[idx][0] += normal[0] * weight
			vertexNormalAccum[idx][1] += normal[1] * weight
			vertexNormalAccum[idx][2] += normal[2] * weight
			vertexFaceCount[idx]++
			vertexFaceArea[idx] += area
		}
	}

	for i := range m.Normals {
		if vertexFaceCount[i] > 0 {
			m.Normals[i][0] = vertexNormalAccum[i][0]
			m.Normals[i][1] = vertexNormalAccum[i][1]
			m.Normals[i][2] = vertexNormalAccum[i][2]

			length := vec3d.T(m.Normals[i]).Length()
			if length > 0 {
				m.Normals[i][0] /= length
				m.Normals[i][1] /= length
				m.Normals[i][2] /= length
			}
		} else {
			m.Normals[i] = vec3d.T{0, 0, 1}
		}
	}
}

func (m *Mesh) CalculateUVs(bounds vec2d.Rect) {
	if len(m.UVs) != len(m.Vertices) {
		m.UVs = make([]vec2d.T, len(m.Vertices))
	}

	boundsW := bounds.Max[0] - bounds.Min[0]
	boundsH := bounds.Max[1] - bounds.Min[1]

	for i, v := range m.Vertices {
		u := (v[0] - bounds.Min[0]) / boundsW
		vCoord := (v[1] - bounds.Min[1]) / boundsH
		m.UVs[i] = vec2d.T{u, vCoord}
	}
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

	if m.TinMesh != nil && other.TinMesh != nil {
		tin1, ok1 := m.TinMesh.(interface {
			GetVertices() []vec3d.T
			GetIndices() []uint32
		})
		tin2, ok2 := other.TinMesh.(interface {
			GetVertices() []vec3d.T
			GetIndices() []uint32
		})

		if ok1 && ok2 {
			mergedTin := &tile.TinMesh{
				Vertices: append(tin1.GetVertices(), tin2.GetVertices()...),
				Indices:  append(tin1.GetIndices(), tin2.GetIndices()...),
			}

			minHeight := 0.0
			maxHeight := 0.0
			for _, v := range mergedTin.Vertices {
				if v[2] < minHeight {
					minHeight = v[2]
				}
				if v[2] > maxHeight {
					maxHeight = v[2]
				}
			}

			mergedTin.MinHeight = minHeight
			mergedTin.MaxHeight = maxHeight

			m.TinMesh = mergedTin
		}
	}

	return nil
}
