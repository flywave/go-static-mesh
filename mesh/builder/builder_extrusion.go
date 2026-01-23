package builder

import (
	"fmt"
	"math"

	"github.com/flywave/go-static-mesh/draw"
	"github.com/flywave/go-static-mesh/mesh"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

func (b *Builder) addGeoDataToMesh(mesh *mesh.Mesh, terrainMesh interface{}, isPrint bool) error {
	if len(b.geoData) == 0 {
		b.logger.Debug("No geo data to add")
		return nil
	}

	if !b.extrudeGeoData {
		b.logger.Debug("Geo data extrusion disabled")
		return nil
	}

	b.logger.Info("Adding geo data to mesh", "objects", len(b.geoData))

	for i, geoObj := range b.geoData {
		height := b.geoDataHeight

		if terrainMesh != nil {
			height = b.sampleHeightAtPosition(geoObj, terrainMesh)
		}

		b.logger.Debug("Extruding geo object", "index", i, "height", height)

		err := geoObj.ExtrudeToMesh(mesh, height)
		if err != nil {
			b.logger.Error("Failed to extrude geo object", "index", i, "error", err)
			return fmt.Errorf("failed to extrude geo object: %w", err)
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
