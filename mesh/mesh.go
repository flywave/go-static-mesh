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
}

type Material struct {
	Name      string
	Diffuse   color.Color
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

func (m *Mesh) AppendVertex(x, y, z float64) uint32 {
	m.Vertices = append(m.Vertices, vec3d.T{x, y, z})
	return uint32(len(m.Vertices) - 1)
}

func (m *Mesh) AppendTriangle(a, b, c uint32) {
	m.Indices = append(m.Indices, a, b, c)
}

func (m *Mesh) GetVertices() interface{} {
	return m.Vertices
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

func (m *Mesh) CalculateNormalsSmooth(iterations int) {
	m.CalculateNormals()

	if iterations <= 0 {
		return
	}

	for iter := 0; iter < iterations; iter++ {
		smoothedNormals := make([]vec3d.T, len(m.Normals))

		for i := 0; i < len(m.Indices); i += 3 {
			if i+2 >= len(m.Indices) {
				break
			}

			idx0, idx1, idx2 := m.Indices[i], m.Indices[i+1], m.Indices[i+2]

			for _, idx := range []uint32{idx0, idx1, idx2} {
				if idx >= uint32(len(m.Normals)) {
					continue
				}

				avgNormal := vec3d.T{0, 0, 0}
				for _, nIdx := range []uint32{idx0, idx1, idx2} {
					if nIdx >= uint32(len(m.Normals)) {
						continue
					}
					avgNormal[0] += m.Normals[nIdx][0]
					avgNormal[1] += m.Normals[nIdx][1]
					avgNormal[2] += m.Normals[nIdx][2]
				}

				length := vec3d.T(avgNormal).Length()
				if length > 0 {
					avgNormal[0] /= length
					avgNormal[1] /= length
					avgNormal[2] /= length
				}

				smoothedNormals[idx][0] = (m.Normals[idx][0] + avgNormal[0]) / 2
				smoothedNormals[idx][1] = (m.Normals[idx][1] + avgNormal[1]) / 2
				smoothedNormals[idx][2] = (m.Normals[idx][2] + avgNormal[2]) / 2
			}
		}

		for i := range smoothedNormals {
			length := vec3d.T(smoothedNormals[i]).Length()
			if length > 0 {
				m.Normals[i][0] = smoothedNormals[i][0] / length
				m.Normals[i][1] = smoothedNormals[i][1] / length
				m.Normals[i][2] = smoothedNormals[i][2] / length
			}
		}
	}
}

func (m *Mesh) CalculateUVs(bounds vec2d.Rect) {
	m.CalculateUVsAdvanced(bounds, "planar", 1.0, false)
}

func (m *Mesh) CalculateUVsAdvanced(bounds vec2d.Rect, method string, scale float64, preserveAspect bool) {
	if len(m.UVs) != len(m.Vertices) {
		m.UVs = make([]vec2d.T, len(m.Vertices))
	}

	switch method {
	case "planar":
		m.calculatePlanarUVs(bounds, scale, preserveAspect)
	case "cylindrical":
		m.calculateCylindricalUVs(bounds, scale)
	case "spherical":
		m.calculateSphericalUVs(bounds, scale)
	case "triplanar":
		m.calculateTriplanarUVs(bounds, scale)
	case "terrain":
		m.calculateTerrainUVs(bounds, scale)
	default:
		m.calculatePlanarUVs(bounds, scale, preserveAspect)
	}
}

func (m *Mesh) calculatePlanarUVs(bounds vec2d.Rect, scale float64, preserveAspect bool) {
	boundsW := bounds.Max[0] - bounds.Min[0]
	boundsH := bounds.Max[1] - bounds.Min[1]

	aspectRatio := boundsW / boundsH
	if boundsH == 0 {
		aspectRatio = 1.0
	}

	for i, v := range m.Vertices {
		u := (v[0] - bounds.Min[0]) / boundsW
		vCoord := (v[1] - bounds.Min[1]) / boundsH

		if preserveAspect && aspectRatio > 1 {
			vCoord = vCoord * aspectRatio
		} else if preserveAspect && aspectRatio < 1 {
			u = u / aspectRatio
		}

		u *= scale
		vCoord *= scale

		m.UVs[i] = vec2d.T{u, vCoord}
	}
}

func (m *Mesh) calculateCylindricalUVs(bounds vec2d.Rect, scale float64) {
	if len(m.Vertices) == 0 {
		return
	}

	centerX := (bounds.Min[0] + bounds.Max[0]) / 2
	centerY := (bounds.Min[1] + bounds.Max[1]) / 2

	radius := bounds.Max[0] - bounds.Min[0]
	if radius == 0 {
		radius = 1
	}

	for i, v := range m.Vertices {
		dx := v[0] - centerX
		dy := v[1] - centerY

		angle := math.Atan2(dy, dx)
		u := (angle + math.Pi) / (2 * math.Pi) * scale

		height := v[2]
		minZ := m.Vertices[0][2]
		maxZ := m.Vertices[0][2]
		for _, vert := range m.Vertices {
			if vert[2] < minZ {
				minZ = vert[2]
			}
			if vert[2] > maxZ {
				maxZ = vert[2]
			}
		}

		vCoord := 0.0
		if maxZ != minZ {
			vCoord = (height - minZ) / (maxZ - minZ) * scale
		}

		m.UVs[i] = vec2d.T{u, vCoord}
	}
}

func (m *Mesh) calculateSphericalUVs(bounds vec2d.Rect, scale float64) {
	if len(m.Vertices) == 0 {
		return
	}

	centerX := (bounds.Min[0] + bounds.Max[0]) / 2
	centerY := (bounds.Min[1] + bounds.Max[1]) / 2
	centerZ := 0.0

	minZ := m.Vertices[0][2]
	maxZ := m.Vertices[0][2]
	for _, vert := range m.Vertices {
		if vert[2] < minZ {
			minZ = vert[2]
		}
		if vert[2] > maxZ {
			maxZ = vert[2]
		}
	}
	centerZ = (minZ + maxZ) / 2

	for i, v := range m.Vertices {
		dx := v[0] - centerX
		dy := v[1] - centerY
		dz := v[2] - centerZ

		u := math.Atan2(dx, dz) / (2 * math.Pi) * scale
		vCoord := (math.Asin(dy/math.Sqrt(dx*dx+dy*dy+dz*dz)) + math.Pi/2) / math.Pi * scale

		m.UVs[i] = vec2d.T{u, vCoord}
	}
}

func (m *Mesh) calculateTriplanarUVs(bounds vec2d.Rect, scale float64) {
	for i, v := range m.Vertices {
		u1 := (v[0] - bounds.Min[0]) / (bounds.Max[0] - bounds.Min[0]) * scale
		v1 := (v[1] - bounds.Min[1]) / (bounds.Max[1] - bounds.Min[1]) * scale

		u2 := (v[0] - bounds.Min[0]) / (bounds.Max[0] - bounds.Min[0]) * scale
		v2 := (v[2] - bounds.Min[0]) / (bounds.Max[1] - bounds.Min[0]) * scale

		u3 := (v[1] - bounds.Min[1]) / (bounds.Max[1] - bounds.Min[1]) * scale
		v3 := (v[2] - bounds.Min[0]) / (bounds.Max[1] - bounds.Min[0]) * scale

		u := (u1 + u2 + u3) / 3
		vCoord := (v1 + v2 + v3) / 3

		m.UVs[i] = vec2d.T{u, vCoord}
	}
}

func (m *Mesh) calculateTerrainUVs(bounds vec2d.Rect, scale float64) {
	if len(m.Normals) == 0 {
		m.CalculateNormals()
	}

	for i, v := range m.Vertices {
		normal := m.Normals[i]

		weightX := math.Abs(normal[0])
		weightY := math.Abs(normal[1])
		weightZ := math.Abs(normal[2])

		totalWeight := weightX + weightY + weightZ
		if totalWeight == 0 {
			totalWeight = 1
		}

		weightX /= totalWeight
		weightY /= totalWeight
		weightZ /= totalWeight

		uX := (v[0] - bounds.Min[0]) / (bounds.Max[0] - bounds.Min[0]) * scale
		vX := (v[1] - bounds.Min[1]) / (bounds.Max[1] - bounds.Min[1]) * scale

		uY := (v[0] - bounds.Min[0]) / (bounds.Max[0] - bounds.Min[0]) * scale
		vY := (v[2] - bounds.Min[0]) / (bounds.Max[1] - bounds.Min[0]) * scale

		uZ := (v[1] - bounds.Min[1]) / (bounds.Max[1] - bounds.Min[1]) * scale
		vZ := (v[2] - bounds.Min[0]) / (bounds.Max[1] - bounds.Min[0]) * scale

		u := weightX*uX + weightY*uY + weightZ*uZ
		vCoord := weightX*vX + weightY*vY + weightZ*vZ

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
