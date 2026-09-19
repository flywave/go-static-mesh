package mesh

import (
	"fmt"
	"math"
	"reflect"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-geoid"

	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"

	tin "github.com/flywave/go-tin"
)

type TINGenerator interface {
	GenerateFromRaster(grid interface{}) (interface{}, error)
	GenerateFromPoints(points []vec2d.T, elevations []float64) (interface{}, error)
	Simplify(mesh interface{}, tolerance float64) (interface{}, error)
	Optimize(mesh interface{}) (interface{}, error)
}

type MeshWrapper struct {
	Vertices    []vec3d.T
	Indices     []uint32
	EdgeIndices []uint32
	MinHeight   float64
	Bounds      vec2d.Rect
	Srs         geo.Proj
}

func (m *MeshWrapper) GetVertices() []vec3d.T {
	return m.Vertices
}

func (m *MeshWrapper) GetIndices() []uint32 {
	return m.Indices
}

func (m *MeshWrapper) GetEdgeIndices() []uint32 {
	return m.EdgeIndices
}

func (m *MeshWrapper) GetMinHeight() float64 {
	return m.MinHeight
}

func (m *MeshWrapper) GetBounds() vec2d.Rect {
	return m.Bounds
}

func (m *MeshWrapper) GetSrs() geo.Proj {
	return m.Srs
}

type Tingenerator struct {
	maxError      float64
	srcProj       geo.Proj
	datum         geoid.VerticalDatum
	offset        float64
	vertexLimit   int
	triangleLimit int
	useCache      bool
}

func NewTINGenerator() *Tingenerator {
	return &Tingenerator{
		maxError:      1.0,
		vertexLimit:   100000,
		triangleLimit: 200000,
		useCache:      true,
	}
}

// srcProjOrNil 过滤掉 geo.NewProj 失败时产生的"类型非空、值为空"接口：
// 传给底层坐标转换会在取字段时 panic。
func srcProjOrNil(p geo.Proj) geo.Proj {
	if p == nil {
		return nil
	}
	v := reflect.ValueOf(p)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}
	return p
}

func (g *Tingenerator) SetMaxError(error float64) {
	g.maxError = error
}

func (g *Tingenerator) SetSrcProj(proj geo.Proj) {
	g.srcProj = proj
}

func (g *Tingenerator) SetDatum(datum geoid.VerticalDatum) {
	g.datum = datum
}

func (g *Tingenerator) SetOffset(offset float64) {
	g.offset = offset
}

func (g *Tingenerator) SetVertexLimit(limit int) {
	g.vertexLimit = limit
}

func (g *Tingenerator) SetTriangleLimit(limit int) {
	g.triangleLimit = limit
}

func (g *Tingenerator) GenerateFromRaster(grid interface{}) (interface{}, error) {
	type gridWithElevation interface {
		GetWidth() int
		GetHeight() int
		GetData() []float64
		GetMinX() float64
		GetMinY() float64
		GetCellSize() float64
		GetNoData() float64
	}

	eg, ok := grid.(gridWithElevation)
	if !ok {
		return nil, fmt.Errorf("grid does not implement gridWithElevation")
	}

	width := eg.GetWidth()
	height := eg.GetHeight()
	data := eg.GetData()
	noData := eg.GetNoData()

	if len(data) != width*height {
		return nil, fmt.Errorf("data size mismatch: expected %d, got %d", width*height, len(data))
	}

	r := tin.NewRasterDoubleWithData(height, width, data)
	r.NoData = noData

	r.SetXYPos(eg.GetMinX(), eg.GetMinY(), eg.GetCellSize())

	config := &tin.GeoConfig{
		SrcProj: srcProjOrNil(g.srcProj),
		Datum:   g.datum,
		Offset:  g.offset,
	}

	zmesh, tmesh := tin.GenerateTinMesh(r, g.maxError, config)

	tinMesh, err := g.convertTinMesh(zmesh, tmesh, eg)
	if err != nil {
		return nil, fmt.Errorf("failed to convert TIN mesh: %w", err)
	}

	return tinMesh, nil
}

func (g *Tingenerator) GenerateFromPoints(points []vec2d.T, elevations []float64) (interface{}, error) {
	if len(points) != len(elevations) {
		return nil, fmt.Errorf("points and elevations length mismatch")
	}

	if len(points) < 3 {
		return nil, fmt.Errorf("at least 3 points required")
	}

	if len(points) > g.vertexLimit {
		return nil, fmt.Errorf("vertex limit exceeded: %d > %d", len(points), g.vertexLimit)
	}

	return g.generateFromDelaunay(points, elevations)
}

func (g *Tingenerator) Simplify(mesh interface{}, tolerance float64) (interface{}, error) {
	type meshWithVertices interface {
		GetVertices() []vec3d.T
		GetIndices() []uint32
		GetMinHeight() float64
	}

	m, ok := mesh.(meshWithVertices)
	if !ok {
		return nil, nil
	}

	vertices := m.GetVertices()
	indices := m.GetIndices()

	if g.triangleLimit > 0 && len(indices)/3 > g.triangleLimit {
		return g.simplifyByTriCount(vertices, indices, g.triangleLimit)
	}

	return g.simplifyByTolerance(vertices, indices, tolerance)
}

func (g *Tingenerator) Optimize(mesh interface{}) (interface{}, error) {
	return g.Simplify(mesh, 0.01)
}

func (g *Tingenerator) convertTinMesh(zmesh *tin.ZemlyaMesh, tmesh *tin.Mesh, grid interface{}) (interface{}, error) {
	if zmesh == nil || tmesh == nil {
		return nil, nil
	}

	type gridWithFullInfo interface {
		GetBounds() vec2d.Rect
		GetMinX() float64
		GetMinY() float64
		GetCellSize() float64
	}

	eg, ok := grid.(gridWithFullInfo)
	if !ok {
		return nil, nil
	}

	bounds := eg.GetBounds()

	centerLon := (bounds.Min[0] + bounds.Max[0]) / 2.0
	centerLat := (bounds.Min[1] + bounds.Max[1]) / 2.0

	// 局部坐标的单位换算：经纬度网格要乘每度米数，投影网格本身已是米
	srs := srcProjOrNil(g.srcProj)
	type gridWithSrs interface {
		GetSrs() geo.Proj
	}
	if gs, ok := grid.(gridWithSrs); ok {
		if gridSrs := srcProjOrNil(gs.GetSrs()); gridSrs != nil {
			srs = gridSrs
		}
	}

	scale := 1.0
	if srs == nil || srs.IsLatLong() {
		scale = 111319.5
	}

	vertices := make([]vec3d.T, len(tmesh.Vertices))

	for i, v := range tmesh.Vertices {
		geoX := v[0]
		geoY := v[1]

		localX := (geoX - centerLon) * scale
		localY := (geoY - centerLat) * scale
		vertices[i] = vec3d.T{localX, localY, v[2]}
	}

	indices := make([]uint32, len(tmesh.Faces)*3)
	for i, f := range tmesh.Faces {
		indices[i*3+0] = uint32(f[0])
		indices[i*3+1] = uint32(f[1])
		indices[i*3+2] = uint32(f[2])
	}

	minHeight := math.Inf(1)
	maxHeight := math.Inf(-1)

	for _, v := range vertices {
		if v[2] < minHeight {
			minHeight = v[2]
		}
		if v[2] > maxHeight {
			maxHeight = v[2]
		}
	}

	wrapper := &MeshWrapper{
		Vertices:    vertices,
		Indices:     indices,
		EdgeIndices: nil,
		MinHeight:   minHeight,
		Bounds:      bounds,
		Srs:         srs,
	}

	return wrapper, nil
}

func (g *Tingenerator) generateFromDelaunay(points []vec2d.T, elevations []float64) (interface{}, error) {
	vertices := make([]vec3d.T, len(points))
	for i, p := range points {
		vertices[i] = vec3d.T{p[0], p[1], elevations[i]}
	}

	if len(vertices) < 3 {
		return nil, fmt.Errorf("not enough vertices for TIN generation")
	}

	indices := make([]uint32, 0, (len(vertices)-2)*3)
	for i := 1; i < len(vertices)-1; i++ {
		indices = append(indices, 0, uint32(i), uint32(i+1))
	}

	minHeight := math.Inf(1)
	for _, v := range vertices {
		if v[2] < minHeight {
			minHeight = v[2]
		}
	}

	wrapper := &MeshWrapper{
		Vertices:    vertices,
		Indices:     indices,
		EdgeIndices: nil,
		MinHeight:   minHeight,
		Bounds:      vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{0, 0}},
		Srs:         g.srcProj,
	}

	return wrapper, nil
}

func (g *Tingenerator) simplifyByTolerance(vertices []vec3d.T, indices []uint32, tolerance float64) (interface{}, error) {
	if tolerance <= 0 {
		minHeight := math.Inf(1)
		for _, v := range vertices {
			if v[2] < minHeight {
				minHeight = v[2]
			}
		}

		wrapper := &MeshWrapper{
			Vertices:    vertices,
			Indices:     indices,
			EdgeIndices: nil,
			MinHeight:   minHeight,
			Bounds:      vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{0, 0}},
		}
		return wrapper, nil
	}

	newVertices := make([]vec3d.T, 0, len(vertices))
	newIndices := make([]uint32, 0, len(indices))

	for i := 0; i < len(indices); i += 3 {
		v0 := vertices[indices[i]]
		v1 := vertices[indices[i+1]]
		v2 := vertices[indices[i+2]]

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

		area := math.Abs(edge1[0]*edge2[1]-edge1[1]*edge2[0]) / 2.0

		if area >= tolerance {
			newVertices = append(newVertices, v0, v1, v2)
			newIndices = append(newIndices,
				uint32(len(newVertices)-3),
				uint32(len(newVertices)-2),
				uint32(len(newVertices)-1))
		}
	}

	if len(newVertices) == 0 {
		minHeight := math.Inf(1)
		for _, v := range vertices {
			if v[2] < minHeight {
				minHeight = v[2]
			}
		}

		wrapper := &MeshWrapper{
			Vertices:    vertices,
			Indices:     indices,
			EdgeIndices: nil,
			MinHeight:   minHeight,
			Bounds:      vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{0, 0}},
		}
		return wrapper, nil
	}

	minHeight := math.Inf(1)
	for _, v := range newVertices {
		if v[2] < minHeight {
			minHeight = v[2]
		}
	}

	wrapper := &MeshWrapper{
		Vertices:    newVertices,
		Indices:     newIndices,
		EdgeIndices: nil,
		MinHeight:   minHeight,
		Bounds:      vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{0, 0}},
	}

	return wrapper, nil
}

func (g *Tingenerator) simplifyByTriCount(vertices []vec3d.T, indices []uint32, maxTriangles int) (interface{}, error) {
	if maxTriangles <= 0 {
		maxTriangles = int(float64(len(indices)) / 3)
	}

	triCount := len(indices) / 3
	if triCount <= maxTriangles {
		minHeight := math.Inf(1)
		for _, v := range vertices {
			if v[2] < minHeight {
				minHeight = v[2]
			}
		}

		wrapper := &MeshWrapper{
			Vertices:    vertices,
			Indices:     indices,
			EdgeIndices: nil,
			MinHeight:   minHeight,
			Bounds:      vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{0, 0}},
		}
		return wrapper, nil
	}

	targetCount := maxTriangles * 3
	step := triCount / maxTriangles
	if step < 1 {
		step = 1
	}

	newIndices := make([]uint32, 0, targetCount)
	for i := 0; i < len(indices); i += 3 * step {
		end := i + 3*step
		if end > len(indices) {
			end = len(indices)
		}
		newIndices = append(newIndices, indices[i:end]...)
	}

	minHeight := math.Inf(1)
	for _, v := range vertices {
		if v[2] < minHeight {
			minHeight = v[2]
		}
	}

	wrapper := &MeshWrapper{
		Vertices:    vertices,
		Indices:     newIndices,
		EdgeIndices: nil,
		MinHeight:   minHeight,
		Bounds:      vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{0, 0}},
	}

	return wrapper, nil
}
