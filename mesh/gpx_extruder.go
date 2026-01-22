package mesh

import (
	"fmt"
	"math"

	"github.com/flywave/go-static-mesh/draw"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

type GPXExtrusionOptions struct {
	Width        float64
	Height       float64
	Depth        float64
	Operation    BSPOperation
	TerrainMesh  *Mesh
	Resolution   float64
	SampleRadius float64
}

type GPXPathExtruder struct {
	baseMesh  *Mesh
	operation BSPOperation
}

func NewGPXPathExtruder() *GPXPathExtruder {
	return &GPXPathExtruder{
		baseMesh:  &Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}},
		operation: BSPOperationSubtraction,
	}
}

func (e *GPXPathExtruder) SetOperation(operation BSPOperation) {
	e.operation = operation
}

func (e *GPXPathExtruder) SetBaseMesh(mesh *Mesh) {
	e.baseMesh = mesh
}

func (e *GPXPathExtruder) ExtrudePathToTerrain(path *draw.Path, options *GPXExtrusionOptions) error {
	if path == nil || len(path.Positions) < 2 {
		return fmt.Errorf("invalid path: must have at least 2 points")
	}

	if options == nil {
		options = &GPXExtrusionOptions{
			Width:        path.Weight,
			Height:       5.0,
			Depth:        5.0,
			Operation:    BSPOperationSubtraction,
			Resolution:   1.0,
			SampleRadius: 100.0,
		}
	}

	tempMesh := &Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}

	err := e.extrudePathWithTerrainSampling(path, tempMesh, options)
	if err != nil {
		return fmt.Errorf("failed to extrude path with terrain sampling: %w", err)
	}

	if len(e.baseMesh.Indices) == 0 {
		e.baseMesh = tempMesh
		return nil
	}

	resultMesh, err := PerformBoolean(e.baseMesh, tempMesh, options.Operation)
	if err != nil {
		return fmt.Errorf("failed to perform boolean operation: %w", err)
	}

	e.baseMesh = resultMesh
	return nil
}

func (e *GPXPathExtruder) extrudePathWithTerrainSampling(path *draw.Path, mesh *Mesh, options *GPXExtrusionOptions) error {
	sampledPositions := make([]vec3d.T, len(path.Positions))

	for i, pos := range path.Positions {
		height := options.Height

		if options.TerrainMesh != nil {
			sampledHeight := e.sampleTerrainHeight(pos, options.TerrainMesh, options.SampleRadius)
			if sampledHeight > 0 {
				height = sampledHeight
			}
		}

		sampledPositions[i] = vec3d.T{pos[0], pos[1], height}
	}

	width := options.Width
	if width <= 0 {
		width = path.Weight
	}
	if width <= 0 {
		width = 2.0
	}

	for i := 0; i < len(sampledPositions)-1; i++ {
		start := sampledPositions[i]
		end := sampledPositions[i+1]

		dx := end[0] - start[0]
		dy := end[1] - start[1]
		length := math.Sqrt(dx*dx + dy*dy)

		if length < 0.0001 {
			continue
		}

		perpX := -dy / length
		perpY := dx / length

		offsetX := (width / 2.0) * perpX
		offsetY := (width / 2.0) * perpY

		v1 := vec3d.T{start[0] + offsetX, start[1] + offsetY, start[2]}
		v2 := vec3d.T{start[0] + offsetX, start[1] + offsetY, start[2] + options.Depth}
		v3 := vec3d.T{end[0] + offsetX, end[1] + offsetY, start[2]}
		v4 := vec3d.T{end[0] + offsetX, end[1] + offsetY, start[2] + options.Depth}

		v5 := vec3d.T{start[0] - offsetX, start[1] - offsetY, start[2]}
		v6 := vec3d.T{start[0] - offsetX, start[1] - offsetY, start[2] + options.Depth}
		v7 := vec3d.T{end[0] - offsetX, end[1] - offsetY, start[2]}
		v8 := vec3d.T{end[0] - offsetX, end[1] - offsetY, start[2] + options.Depth}

		baseIdx := uint32(len(mesh.Vertices))

		mesh.Vertices = append(mesh.Vertices, v1, v2, v3, v4, v5, v6, v7, v8)

		mesh.Indices = append(mesh.Indices,
			baseIdx, baseIdx+1, baseIdx+2,
			baseIdx+2, baseIdx+1, baseIdx+3,
			baseIdx+4, baseIdx+7, baseIdx+6,
			baseIdx+4, baseIdx+5, baseIdx+7,
			baseIdx, baseIdx+4, baseIdx+2,
			baseIdx+2, baseIdx+4, baseIdx+7,
			baseIdx+1, baseIdx+5, baseIdx+3,
			baseIdx+3, baseIdx+5, baseIdx+7,
		)
	}

	return nil
}

func (e *GPXPathExtruder) sampleTerrainHeight(pos vec2d.T, terrainMesh *Mesh, radius float64) float64 {
	if terrainMesh == nil || len(terrainMesh.Vertices) == 0 {
		return 0
	}

	sampleCount := 0
	totalHeight := 0.0

	for i := 0; i < len(terrainMesh.Vertices); i += 10 {
		v := terrainMesh.Vertices[i]
		dist := math.Sqrt(
			math.Pow(v[0]-pos[0], 2) +
				math.Pow(v[1]-pos[1], 2),
		)

		if dist <= radius {
			totalHeight += v[2]
			sampleCount++
		}
	}

	if sampleCount == 0 {
		radius *= 2.0
		for i := 0; i < len(terrainMesh.Vertices); i += 5 {
			v := terrainMesh.Vertices[i]
			dist := math.Sqrt(
				math.Pow(v[0]-pos[0], 2) +
					math.Pow(v[1]-pos[1], 2),
			)

			if dist <= radius {
				totalHeight += v[2]
				sampleCount++
			}
		}
	}

	if sampleCount == 0 {
		minDist := math.Inf(1)
		minHeight := 0.0
		for i := 0; i < len(terrainMesh.Vertices); i += 20 {
			v := terrainMesh.Vertices[i]
			dist := math.Sqrt(
				math.Pow(v[0]-pos[0], 2) +
					math.Pow(v[1]-pos[1], 2),
			)

			if dist < minDist {
				minDist = dist
				minHeight = v[2]
			}
		}
		return minHeight
	}

	return totalHeight / float64(sampleCount)
}

func (e *GPXPathExtruder) GetResult() *Mesh {
	return e.baseMesh
}

func (e *GPXPathExtruder) Reset() {
	e.baseMesh = &Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}
	e.operation = BSPOperationSubtraction
}

func ExtrudeGPXPathsToTerrain(paths []*draw.Path, terrainMesh *Mesh, options *GPXExtrusionOptions) (*Mesh, error) {
	if len(paths) == 0 {
		return terrainMesh, nil
	}

	extruder := NewGPXPathExtruder()
	extruder.SetBaseMesh(terrainMesh)

	if options == nil {
		options = &GPXExtrusionOptions{
			Width:        5.0,
			Height:       5.0,
			Depth:        5.0,
			Operation:    BSPOperationUnion,
			Resolution:   1.0,
			SampleRadius: 100.0,
			TerrainMesh:  terrainMesh,
		}
	}

	for _, path := range paths {
		err := extruder.ExtrudePathToTerrain(path, options)
		if err != nil {
			return nil, fmt.Errorf("failed to extrude GPX path: %w", err)
		}
	}

	result := extruder.GetResult()
	if len(result.Vertices) > 0 {
		result.CalculateNormals()
	}

	return result, nil
}

func ExtrudeGPXPathsToTerrainNilTerrain(paths []*draw.Path, terrainMesh *Mesh, options *GPXExtrusionOptions) (*Mesh, error) {
	if len(paths) == 0 {
		return terrainMesh, nil
	}

	extruder := NewGPXPathExtruder()
	extruder.SetBaseMesh(terrainMesh)

	if options == nil {
		options = &GPXExtrusionOptions{
			Width:        5.0,
			Height:       5.0,
			Depth:        5.0,
			Operation:    BSPOperationSubtraction,
			Resolution:   1.0,
			SampleRadius: 100.0,
			TerrainMesh:  terrainMesh,
		}
	}

	for _, path := range paths {
		err := extruder.ExtrudePathToTerrain(path, options)
		if err != nil {
			return nil, fmt.Errorf("failed to extrude GPX path: %w", err)
		}
	}

	result := extruder.GetResult()
	if len(result.Vertices) > 0 {
		result.CalculateNormals()
	}

	return result, nil
}

func ExtrudeGPXPathsWithUnion(paths []*draw.Path, terrainMesh *Mesh, options *GPXExtrusionOptions) (*Mesh, error) {
	if len(paths) == 0 {
		return terrainMesh, nil
	}

	if options == nil {
		options = &GPXExtrusionOptions{
			Width:        5.0,
			Height:       5.0,
			Depth:        5.0,
			Operation:    BSPOperationSubtraction,
			Resolution:   1.0,
			SampleRadius: 100.0,
			TerrainMesh:  terrainMesh,
		}
	}

	tempMesh := &Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}

	for _, path := range paths {
		extruder := NewGPXPathExtruder()
		pathMesh := &Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}

		optionsCopy := *options
		optionsCopy.Operation = BSPOperationUnion

		err := extruder.extrudePathWithTerrainSampling(path, pathMesh, &optionsCopy)
		if err != nil {
			return nil, fmt.Errorf("failed to extrude path: %w", err)
		}

		if len(tempMesh.Indices) == 0 {
			tempMesh = pathMesh
		} else {
			combined, err := PerformBoolean(tempMesh, pathMesh, BSPOperationUnion)
			if err != nil {
				return nil, fmt.Errorf("failed to union paths: %w", err)
			}
			tempMesh = combined
		}
	}

	result, err := PerformBoolean(terrainMesh, tempMesh, options.Operation)
	if err != nil {
		return nil, fmt.Errorf("failed to apply to terrain: %w", err)
	}

	if len(result.Vertices) > 0 {
		result.CalculateNormals()
	}

	return result, nil
}

func CreateTrenchFromGPX(paths []*draw.Path, terrainMesh *Mesh, width, depth float64) (*Mesh, error) {
	options := &GPXExtrusionOptions{
		Width:        width,
		Height:       0,
		Depth:        depth,
		Operation:    BSPOperationSubtraction,
		Resolution:   1.0,
		SampleRadius: 100.0,
		TerrainMesh:  terrainMesh,
	}

	return ExtrudeGPXPathsToTerrain(paths, terrainMesh, options)
}

func CreateRaisedPathFromGPX(paths []*draw.Path, terrainMesh *Mesh, width, height float64) (*Mesh, error) {
	options := &GPXExtrusionOptions{
		Width:        width,
		Height:       height,
		Depth:        0,
		Operation:    BSPOperationUnion,
		Resolution:   1.0,
		SampleRadius: 100.0,
		TerrainMesh:  terrainMesh,
	}

	return ExtrudeGPXPathsToTerrain(paths, terrainMesh, options)
}
