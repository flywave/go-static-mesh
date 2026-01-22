package extruder

import (
	"fmt"
	"math"

	"github.com/flywave/go-static-mesh/mesh"
	"github.com/flywave/go-static-mesh/tile"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

type Model3DExtruder struct {
	transformer *ModelTransformer
}

func NewModel3DExtruder() *Model3DExtruder {
	return &Model3DExtruder{
		transformer: &ModelTransformer{},
	}
}

func (e *Model3DExtruder) ExtrudeModelToMesh(mesh *mesh.Mesh, model *tile.Model3D, terrainMesh interface{}) error {
	if model.Mesh == nil {
		return fmt.Errorf("model mesh is nil")
	}

	sourceVertices := model.Mesh.Vertices
	sourceIndices := model.Mesh.Indices

	baseIndex := uint32(len(mesh.Vertices))

	transformedVertices := e.transformModelVertices(sourceVertices, model)
	mesh.Vertices = append(mesh.Vertices, transformedVertices...)

	for _, idx := range sourceIndices {
		mesh.Indices = append(mesh.Indices, baseIndex+idx)
	}

	return nil
}

func (e *Model3DExtruder) ExtrudeModelsToMesh(mesh *mesh.Mesh, models []tile.Model3D, terrainMesh interface{}) error {
	for i := range models {
		err := e.ExtrudeModelToMesh(mesh, &models[i], terrainMesh)
		if err != nil {
			return fmt.Errorf("failed to extrude model %d: %w", i, err)
		}
	}
	return nil
}

func (e *Model3DExtruder) ExtrudeModelAtElevation(mesh *mesh.Mesh, model *tile.Model3D, terrainMesh interface{}, elevation float64) error {
	if model.Mesh == nil {
		return fmt.Errorf("model mesh is nil")
	}

	elevationDelta := elevation
	if model.Elevation > 0 {
		elevationDelta = model.Elevation
	}

	if terrainMesh != nil {
		terrainHeight := e.sampleTerrainHeight(model.Position, terrainMesh)
		if terrainHeight > 0 {
			elevationDelta = terrainHeight + model.Elevation
		}
	}

	sourceVertices := model.Mesh.Vertices
	sourceIndices := model.Mesh.Indices

	baseIndex := uint32(len(mesh.Vertices))

	for _, v := range sourceVertices {
		mesh.Vertices = append(mesh.Vertices, vec3d.T{
			v[0] + model.Position[0],
			v[1] + model.Position[1],
			v[2] + elevationDelta,
		})
	}

	for _, idx := range sourceIndices {
		mesh.Indices = append(mesh.Indices, baseIndex+idx)
	}

	return nil
}

func (e *Model3DExtruder) SampleHeightAtPosition(pos vec2d.T, terrainMesh interface{}) float64 {
	if terrainMesh == nil {
		return 0
	}

	return e.sampleTerrainHeight(pos, terrainMesh)
}

func (e *Model3DExtruder) sampleTerrainHeight(pos vec2d.T, terrainMesh interface{}) float64 {
	type meshWithElevation interface {
		GetVertices() []vec3d.T
		GetBounds() vec2d.Rect
	}

	m, ok := terrainMesh.(meshWithElevation)
	if !ok {
		return 0
	}

	vertices := m.GetVertices()
	if len(vertices) == 0 {
		return 0
	}

	minDist := math.Inf(1)
	sampledHeight := 0.0

	for i := 0; i < len(vertices); i += 10 {
		v := vertices[i]
		dist := math.Sqrt(
			math.Pow(v[0]-pos[0], 2) +
				math.Pow(v[1]-pos[1], 2),
		)

		if dist < minDist {
			minDist = dist
			sampledHeight = v[2]
		}
	}

	return sampledHeight
}

func (e *Model3DExtruder) transformModelVertices(vertices []vec3d.T, model *tile.Model3D) []vec3d.T {
	transformed := make([]vec3d.T, len(vertices))

	for i, v := range vertices {
		transformed[i] = vec3d.T{
			v[0] * model.Scale[0],
			v[1] * model.Scale[1],
			v[2] * model.Scale[2],
		}

		transformed[i] = e.applyRotation(transformed[i], model.Rotation)

		transformed[i][0] += model.Position[0]
		transformed[i][1] += model.Position[1]
		transformed[i][2] += model.Elevation
	}

	return transformed
}

func (e *Model3DExtruder) applyRotation(v vec3d.T, rotation vec3d.T) vec3d.T {
	result := v

	if rotation[0] != 0 {
		cosY := math.Cos(rotation[0])
		sinY := math.Sin(rotation[0])
		result[1] = v[1]*cosY - v[2]*sinY
		result[2] = v[1]*sinY + v[2]*cosY
	}

	if rotation[1] != 0 {
		cosX := math.Cos(rotation[1])
		sinX := math.Sin(rotation[1])
		result[0] = v[0]*cosX + v[2]*sinX
		result[2] = -v[0]*sinX + v[2]*cosX
	}

	if rotation[2] != 0 {
		cosZ := math.Cos(rotation[2])
		sinZ := math.Sin(rotation[2])
		result[0] = v[0]*cosZ - v[1]*sinZ
		result[1] = v[0]*sinZ + v[1]*cosZ
	}

	return result
}

type ModelTransformer struct {
	scale       vec3d.T
	rotation    vec3d.T
	translation vec3d.T
}

func (t *ModelTransformer) Transform(v vec3d.T) vec3d.T {
	result := vec3d.T{
		v[0] * t.scale[0],
		v[1] * t.scale[1],
		v[2] * t.scale[2],
	}

	result = t.applyRotation(result)
	result[0] += t.translation[0]
	result[1] += t.translation[1]
	result[2] += t.translation[2]

	return result
}

func (t *ModelTransformer) applyRotation(v vec3d.T) vec3d.T {
	result := v

	if t.rotation[0] != 0 {
		cosY := math.Cos(t.rotation[0])
		sinY := math.Sin(t.rotation[0])
		result[1] = v[1]*cosY - v[2]*sinY
		result[2] = v[1]*sinY + v[2]*cosY
	}

	if t.rotation[1] != 0 {
		cosX := math.Cos(t.rotation[1])
		sinX := math.Sin(t.rotation[1])
		result[0] = v[0]*cosX + v[2]*sinX
		result[2] = -v[0]*sinX + v[2]*cosX
	}

	if t.rotation[2] != 0 {
		cosZ := math.Cos(t.rotation[2])
		sinZ := math.Sin(t.rotation[2])
		result[0] = v[0]*cosZ - v[1]*sinZ
		result[1] = v[0]*sinZ + v[1]*cosZ
	}

	return result
}

func (t *ModelTransformer) SetScale(scale vec3d.T) {
	t.scale = scale
}

func (t *ModelTransformer) SetRotation(rotation vec3d.T) {
	t.rotation = rotation
}

func (t *ModelTransformer) SetTranslation(translation vec3d.T) {
	t.translation = translation
}
