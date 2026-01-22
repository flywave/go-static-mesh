package mesh

import (
	"fmt"
	"math"

	"github.com/flywave/go-static-mesh/draw"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

func (b *Builder) addGeoDataToMesh(mesh *Mesh, terrainMesh interface{}, isPrint bool) error {
	if len(b.geoData) == 0 {
		return nil
	}

	if !b.extrudeGeoData {
		return nil
	}

	pathExtruder := NewPathExtruder()
	areaExtruder := NewAreaExtruder()

	for _, geoObj := range b.geoData {
		if meshObj, ok := geoObj.(MeshObject); ok {
			height := b.geoDataHeight

			if terrainMesh != nil {
				height = b.sampleHeightAtPosition(geoObj, terrainMesh)
			}

			err := meshObj.ExtrudeToMesh(mesh, height)
			if err != nil {
				return fmt.Errorf("failed to extrude geo object: %w", err)
			}
		}

		if path, ok := geoObj.(*draw.Path); ok {
			height := b.geoDataHeight
			if terrainMesh != nil {
				height = b.sampleHeightAtPosition(path, terrainMesh)
			}

			options := &ExtrudeOptions{
				Radius: path.Weight / 2.0,
			}
			err := pathExtruder.ExtrudeToMeshWithResolution(path, mesh, height, options)
			if err != nil {
				return fmt.Errorf("failed to extrude path: %w", err)
			}
		}

		if area, ok := geoObj.(*draw.Area); ok {
			height := b.geoDataHeight
			if terrainMesh != nil {
				height = b.sampleHeightAtPosition(area, terrainMesh)
			}

			err := areaExtruder.ExtrudeToMesh(area, mesh, height)
			if err != nil {
				return fmt.Errorf("failed to extrude area: %w", err)
			}
		}
	}

	return nil
}

func (b *Builder) sampleHeightAtPosition(obj draw.MapObject, terrainMesh interface{}) float64 {
	bounds := obj.Bounds()

	type meshWithElevation interface {
		GetVertices() []vec3d.T
	}

	m, ok := terrainMesh.(meshWithElevation)
	if !ok {
		return b.geoDataHeight
	}

	vertices := m.GetVertices()
	if len(vertices) == 0 {
		return b.geoDataHeight
	}

	centerX := (bounds.Min[0] + bounds.Max[0]) / 2.0
	centerY := (bounds.Min[1] + bounds.Max[1]) / 2.0

	minDist := math.Inf(1)
	sampledHeight := b.geoDataHeight

	for i := 0; i < len(vertices); i += 10 {
		v := vertices[i]
		dist := math.Sqrt(
			math.Pow(v[0]-centerX, 2) +
				math.Pow(v[1]-centerY, 2),
		)

		if dist < minDist {
			minDist = dist
			sampledHeight = v[2]
		}
	}

	return sampledHeight
}
