package extruder

import (
	"fmt"
	"math"

	"github.com/flywave/go-static-mesh/draw"
	"github.com/flywave/go-static-mesh/mesh"
	bsp "github.com/flywave/go-static-mesh/mesh/bsp"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

type BillboardExtrusionOptions struct {
	TerrainMesh   *mesh.Mesh
	SampleRadius  float64
	EnsureContact bool
}

type BillboardExtruder struct {
	baseMesh  *mesh.Mesh
	operation bsp.BSPOperation
}

func NewBillboardExtruder() *BillboardExtruder {
	return &BillboardExtruder{
		baseMesh:  &mesh.Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}},
		operation: bsp.BSPOperationUnion,
	}
}

func (e *BillboardExtruder) SetOperation(operation bsp.BSPOperation) {
	e.operation = operation
}

func (e *BillboardExtruder) SetBaseMesh(m *mesh.Mesh) {
	e.baseMesh = m
}

func (e *BillboardExtruder) ExtrudeBillboardToTerrain(billboard *draw.Billboard, options *BillboardExtrusionOptions) error {
	if billboard == nil {
		return fmt.Errorf("billboard is nil")
	}

	if options == nil {
		options = &BillboardExtrusionOptions{
			SampleRadius:  100.0,
			EnsureContact: true,
		}
	}

	signMesh := e.createBillboardBox(billboard, options)

	if billboard.Mode == draw.BillboardModeDisplay {
		texture, err := billboard.CreateDisplayTexture()
		if err == nil && texture != nil {
			signMesh.Texture = texture
			e.generateBillboardUVs(signMesh)
		}
	} else if billboard.Mode == draw.BillboardModePrint && billboard.FontParser != nil && billboard.Text != "" {
		textMesh, err := e.createRaisedText(billboard)
		if err != nil {
			return fmt.Errorf("failed to create raised text: %w", err)
		}
		if textMesh != nil && len(textMesh.Vertices) > 0 {
			combined, err := bsp.PerformBoolean(signMesh, textMesh, bsp.BSPOperationUnion)
			if err == nil {
				signMesh = combined
			}
		}
	}

	if e.baseMesh == nil || len(e.baseMesh.Indices) == 0 {
		e.baseMesh = signMesh
		return nil
	}

	resultMesh, err := bsp.PerformBoolean(e.baseMesh, signMesh, e.operation)
	if err != nil {
		return fmt.Errorf("failed to perform boolean operation: %w", err)
	}

	e.baseMesh = resultMesh
	return nil
}

func (e *BillboardExtruder) createBillboardBox(billboard *draw.Billboard, options *BillboardExtrusionOptions) *mesh.Mesh {
	resultMesh := &mesh.Mesh{
		Vertices: []vec3d.T{},
		Indices:  []uint32{},
	}

	corners := billboard.GetRotatedCorners()

	baseHeight := 0.0
	if options.TerrainMesh != nil && options.EnsureContact {
		baseHeight = e.sampleMinTerrainHeight(corners, options.TerrainMesh, options.SampleRadius)
	}

	rotated3D := billboard.GetRotated3DCorners(0, billboard.Thickness)

	minRotatedZ := rotated3D[0][2]
	for i := 1; i < 8; i++ {
		if rotated3D[i][2] < minRotatedZ {
			minRotatedZ = rotated3D[i][2]
		}
	}

	vertexIndices := make([]uint32, 8)
	for i := 0; i < 8; i++ {
		z := rotated3D[i][2] - minRotatedZ + baseHeight
		vertexIndices[i] = uint32(len(resultMesh.Vertices))
		resultMesh.Vertices = append(resultMesh.Vertices, vec3d.T{rotated3D[i][0], rotated3D[i][1], z})
	}

	resultMesh.Indices = append(resultMesh.Indices,
		vertexIndices[0], vertexIndices[1], vertexIndices[2],
		vertexIndices[0], vertexIndices[2], vertexIndices[3],
	)

	resultMesh.Indices = append(resultMesh.Indices,
		vertexIndices[6], vertexIndices[5], vertexIndices[4],
		vertexIndices[7], vertexIndices[6], vertexIndices[4],
	)

	resultMesh.Indices = append(resultMesh.Indices,
		vertexIndices[0], vertexIndices[4], vertexIndices[1],
		vertexIndices[1], vertexIndices[4], vertexIndices[5],
	)

	resultMesh.Indices = append(resultMesh.Indices,
		vertexIndices[1], vertexIndices[5], vertexIndices[2],
		vertexIndices[2], vertexIndices[5], vertexIndices[6],
	)

	resultMesh.Indices = append(resultMesh.Indices,
		vertexIndices[2], vertexIndices[6], vertexIndices[3],
		vertexIndices[3], vertexIndices[6], vertexIndices[7],
	)

	resultMesh.Indices = append(resultMesh.Indices,
		vertexIndices[3], vertexIndices[7], vertexIndices[0],
		vertexIndices[0], vertexIndices[7], vertexIndices[4],
	)

	return resultMesh
}

func (e *BillboardExtruder) generateBillboardUVs(m *mesh.Mesh) {
	if len(m.Vertices) == 0 {
		return
	}

	minX := math.Inf(1)
	maxX := math.Inf(-1)
	minY := math.Inf(1)
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

	width := maxX - minX
	height := maxY - minY
	if width == 0 || height == 0 {
		return
	}

	m.UVs = make([]vec2d.T, len(m.Vertices))
	for i, v := range m.Vertices {
		u := (v[0] - minX) / width
		vCoord := (v[1] - minY) / height
		m.UVs[i] = vec2d.T{u, vCoord}
	}
}

func (e *BillboardExtruder) sampleMinTerrainHeight(corners [4]vec2d.T, terrainMesh *mesh.Mesh, radius float64) float64 {
	if terrainMesh == nil || len(terrainMesh.Vertices) == 0 {
		return 0
	}

	minHeight := math.Inf(1)
	for _, corner := range corners {
		h := e.sampleTerrainHeight(corner, terrainMesh, radius)
		if h < minHeight {
			minHeight = h
		}
	}

	if math.IsInf(minHeight, 1) {
		return 0
	}

	return minHeight
}

func (e *BillboardExtruder) sampleTerrainHeight(pos vec2d.T, terrainMesh *mesh.Mesh, radius float64) float64 {
	if terrainMesh == nil || len(terrainMesh.Vertices) == 0 {
		return 0
	}

	sampleCount := 0
	totalHeight := 0.0
	dynamicRadius := radius

	sampleStep := int(math.Max(1, float64(len(terrainMesh.Vertices))/1000.0))

	for i := 0; i < len(terrainMesh.Vertices); i += sampleStep {
		v := terrainMesh.Vertices[i]
		dist := math.Sqrt(math.Pow(v[0]-pos[0], 2) + math.Pow(v[1]-pos[1], 2))

		if dist <= dynamicRadius {
			totalHeight += v[2]
			sampleCount++
		}
	}

	if sampleCount == 0 {
		dynamicRadius *= 2.0
		for i := 0; i < len(terrainMesh.Vertices); i += sampleStep {
			v := terrainMesh.Vertices[i]
			dist := math.Sqrt(math.Pow(v[0]-pos[0], 2) + math.Pow(v[1]-pos[1], 2))

			if dist <= dynamicRadius {
				totalHeight += v[2]
				sampleCount++
			}
		}
	}

	if sampleCount == 0 {
		minDist := math.Inf(1)
		minHeight := 0.0
		for i := 0; i < len(terrainMesh.Vertices); i += sampleStep {
			v := terrainMesh.Vertices[i]
			dist := math.Sqrt(math.Pow(v[0]-pos[0], 2) + math.Pow(v[1]-pos[1], 2))

			if dist < minDist {
				minDist = dist
				minHeight = v[2]
			}
		}
		return minHeight
	}

	return totalHeight / float64(sampleCount)
}

func (e *BillboardExtruder) createRaisedText(billboard *draw.Billboard) (*mesh.Mesh, error) {
	if billboard.FontParser == nil || billboard.Text == "" {
		return nil, nil
	}

	paths, totalWidth, err := billboard.FontParser.GetTextPaths(billboard.Text, billboard.FontSize)
	if err != nil || len(paths) == 0 {
		return nil, err
	}

	scale := billboard.Width * 0.8 / totalWidth
	if scale <= 0 {
		scale = 1.0
	}

	textMesh := &mesh.Mesh{
		Vertices: []vec3d.T{},
		Indices:  []uint32{},
	}

	for _, path := range paths {
		if len(path) < 3 {
			continue
		}

		scaledPath := make([]vec2d.T, len(path))
		for i, p := range path {
			scaledPath[i] = vec2d.T{p[0] * scale, p[1] * scale}
		}

		e.extrudeTextPath(scaledPath, textMesh, billboard.TextDepth)
	}

	e.positionTextOnSign(textMesh, billboard)

	return textMesh, nil
}

func (e *BillboardExtruder) extrudeTextPath(path []vec2d.T, mesh *mesh.Mesh, depth float64) {
	if len(path) < 3 {
		return
	}

	baseIdx := uint32(len(mesh.Vertices))

	for _, p := range path {
		mesh.Vertices = append(mesh.Vertices, vec3d.T{p[0], p[1], 0})
	}

	for _, p := range path {
		mesh.Vertices = append(mesh.Vertices, vec3d.T{p[0], p[1], depth})
	}

	n := len(path)
	for i := 0; i < n-2; i++ {
		mesh.Indices = append(mesh.Indices, baseIdx, baseIdx+uint32(i+1), baseIdx+uint32(i+2))
		mesh.Indices = append(mesh.Indices, baseIdx+uint32(n), baseIdx+uint32(i+2)+uint32(n), baseIdx+uint32(i+1)+uint32(n))
	}

	for i := 0; i < n; i++ {
		next := (i + 1) % n
		bottom := baseIdx + uint32(i)
		bottomNext := baseIdx + uint32(next)
		top := baseIdx + uint32(n) + uint32(i)
		topNext := baseIdx + uint32(n) + uint32(next)

		mesh.Indices = append(mesh.Indices, bottom, bottomNext, top)
		mesh.Indices = append(mesh.Indices, bottomNext, topNext, top)
	}
}

func (e *BillboardExtruder) positionTextOnSign(textMesh *mesh.Mesh, billboard *draw.Billboard) {
	if len(textMesh.Vertices) == 0 {
		return
	}

	minX := math.Inf(1)
	maxX := math.Inf(-1)
	minY := math.Inf(1)
	maxY := math.Inf(-1)

	for _, v := range textMesh.Vertices {
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

	textWidth := maxX - minX
	textHeight := maxY - minY

	offsetX := -minX - textWidth/2
	offsetY := -minY - textHeight/2

	rad := billboard.Rotation * math.Pi / 180.0
	cosR := math.Cos(rad)
	sinR := math.Sin(rad)

	surfaceZ := billboard.Thickness

	for i := range textMesh.Vertices {
		x := textMesh.Vertices[i][0] + offsetX
		y := textMesh.Vertices[i][1] + offsetY
		z := textMesh.Vertices[i][2] + surfaceZ

		rotY := y*cosR - z*sinR
		rotZ := y*sinR + z*cosR

		textMesh.Vertices[i][0] = billboard.Position[0] + x
		textMesh.Vertices[i][1] = billboard.Position[1] + rotY
		textMesh.Vertices[i][2] = rotZ
	}
}

func (e *BillboardExtruder) GetResult() *mesh.Mesh {
	return e.baseMesh
}

func (e *BillboardExtruder) Reset() {
	e.baseMesh = &mesh.Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}
	e.operation = bsp.BSPOperationUnion
}

func ExtrudeBillboardsToTerrain(billboards []*draw.Billboard, terrainMesh *mesh.Mesh, options *BillboardExtrusionOptions) (*mesh.Mesh, error) {
	if len(billboards) == 0 {
		return terrainMesh, nil
	}

	extruder := NewBillboardExtruder()
	if terrainMesh != nil {
		extruder.SetBaseMesh(terrainMesh)
	}

	if options == nil {
		options = &BillboardExtrusionOptions{
			SampleRadius:  100.0,
			EnsureContact: true,
			TerrainMesh:   terrainMesh,
		}
	}

	for _, billboard := range billboards {
		err := extruder.ExtrudeBillboardToTerrain(billboard, options)
		if err != nil {
			return nil, fmt.Errorf("failed to extrude billboard: %w", err)
		}
	}

	result := extruder.GetResult()
	if len(result.Vertices) > 0 {
		result.CalculateNormals()
	}

	return result, nil
}
