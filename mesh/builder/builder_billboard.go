package builder

import (
	"fmt"

	"github.com/flywave/go-static-mesh/draw"
	"github.com/flywave/go-static-mesh/mesh"
	"github.com/flywave/go-static-mesh/mesh/extruder"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

func (b *Builder) AddBillboard(billboard *draw.Billboard) {
	b.AddGeoData(billboard)
	b.logger.Debug("Billboard added", "text", billboard.Text, "mode", billboard.Mode)
}

func (b *Builder) AddBillboards(billboards []*draw.Billboard) {
	for _, billboard := range billboards {
		b.AddBillboard(billboard)
	}
	b.logger.Debug("Billboards added", "count", len(billboards))
}

func (b *Builder) processBillboards(terrainMesh *mesh.Mesh) (*mesh.Mesh, error) {
	billboards := b.extractBillboards()

	if len(billboards) == 0 {
		b.logger.Debug("No billboards to process")
		return nil, nil
	}

	b.logger.Info("Processing billboards", "count", len(billboards))

	ext := extruder.NewBillboardExtruder()
	if terrainMesh != nil {
		ext.SetBaseMesh(terrainMesh)
	}

	options := &extruder.BillboardExtrusionOptions{
		TerrainMesh:   terrainMesh,
		SampleRadius:  100.0,
		EnsureContact: true,
	}

	for i, billboard := range billboards {
		b.logger.Debug("Extruding billboard", "index", i, "text", billboard.Text, "mode", billboard.Mode)
		err := ext.ExtrudeBillboardToTerrain(billboard, options)
		if err != nil {
			return nil, fmt.Errorf("failed to extrude billboard '%s': %w", billboard.Text, err)
		}
	}

	result := ext.GetResult()

	if len(result.Vertices) > 0 {
		result.CalculateNormals()
		b.logger.Info("Billboards processed successfully",
			"count", len(billboards),
			"vertices", len(result.Vertices),
			"triangles", result.TriangleCount())
	}

	return result, nil
}

func (b *Builder) extractBillboards() []*draw.Billboard {
	billboards := []*draw.Billboard{}
	nonBillboardGeoData := []draw.MapObject{}

	for _, obj := range b.geoData {
		if billboard, ok := obj.(*draw.Billboard); ok {
			billboards = append(billboards, billboard)
		} else {
			nonBillboardGeoData = append(nonBillboardGeoData, obj)
		}
	}

	b.geoData = nonBillboardGeoData

	if len(billboards) > 0 {
		b.logger.Debug("Extracted billboards from geo data", "count", len(billboards), "remaining objects", len(b.geoData))
	}

	return billboards
}

func (b *Builder) combineMeshes(mesh1, mesh2 *mesh.Mesh) (*mesh.Mesh, error) {
	if mesh1 == nil {
		return mesh2, nil
	}
	if mesh2 == nil {
		return mesh1, nil
	}

	combined := &mesh.Mesh{
		Vertices: make([]vec3d.T, 0, len(mesh1.Vertices)+len(mesh2.Vertices)),
		Indices:  make([]uint32, 0, len(mesh1.Indices)+len(mesh2.Indices)),
	}

	combined.Vertices = append(combined.Vertices, mesh1.Vertices...)
	combined.Vertices = append(combined.Vertices, mesh2.Vertices...)

	combined.Indices = append(combined.Indices, mesh1.Indices...)
	offset := uint32(len(mesh1.Vertices))
	for _, idx := range mesh2.Indices {
		combined.Indices = append(combined.Indices, idx+offset)
	}

	if len(mesh1.UVs) > 0 || len(mesh2.UVs) > 0 {
		combined.UVs = make([]vec2d.T, 0, len(mesh1.UVs)+len(mesh2.UVs))
		combined.UVs = append(combined.UVs, mesh1.UVs...)
		combined.UVs = append(combined.UVs, mesh2.UVs...)
	}

	if len(mesh1.Normals) > 0 || len(mesh2.Normals) > 0 {
		combined.Normals = make([]vec3d.T, 0, len(mesh1.Normals)+len(mesh2.Normals))
		combined.Normals = append(combined.Normals, mesh1.Normals...)
		combined.Normals = append(combined.Normals, mesh2.Normals...)
	}

	if mesh1.Texture != nil {
		combined.Texture = mesh1.Texture
	} else if mesh2.Texture != nil {
		combined.Texture = mesh2.Texture
	}

	if mesh1.Bounds.Area() > 0 {
		combined.Bounds = mesh1.Bounds
		if mesh2.Bounds.Area() > 0 {
			combined.Bounds.Extend(&mesh2.Bounds.Min)
			combined.Bounds.Extend(&mesh2.Bounds.Max)
		}
	} else if mesh2.Bounds.Area() > 0 {
		combined.Bounds = mesh2.Bounds
	}

	if mesh1.Srs != nil {
		combined.Srs = mesh1.Srs
	} else if mesh2.Srs != nil {
		combined.Srs = mesh2.Srs
	}

	return combined, nil
}
