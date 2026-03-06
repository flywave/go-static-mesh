package mesh

import (
	"fmt"
	"image"
	"image/color"

	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

type CloseMeshOptions struct {
	Thickness          float64
	UseMinHeightAsBase bool
	Enabled            bool

	BottomTextureImage   image.Image
	BottomTextureTilingU float64
	BottomTextureTilingV float64
	BottomColor          color.Color

	SideTextureImage   image.Image
	SideTextureTilingU float64
	SideTextureTilingV float64
	SideColor          color.Color

	ApplyToSides  bool
	ApplyToBottom bool
}

func NewDefaultCloseMeshOptions() *CloseMeshOptions {
	return &CloseMeshOptions{
		Thickness:          5.0,
		UseMinHeightAsBase: true,
		Enabled:            false,

		BottomTextureTilingU: 1.0,
		BottomTextureTilingV: 1.0,
		BottomColor:          color.RGBA{139, 119, 101, 255},

		SideTextureTilingU: 1.0,
		SideTextureTilingV: 1.0,
		SideColor:          color.RGBA{139, 119, 101, 255},

		ApplyToSides:  true,
		ApplyToBottom: true,
	}
}

type TexturedCloser struct {
	options *CloseMeshOptions
}

func NewTexturedCloser() *TexturedCloser {
	return &TexturedCloser{
		options: NewDefaultCloseMeshOptions(),
	}
}

func (c *TexturedCloser) SetOptions(options *CloseMeshOptions) {
	c.options = options
}

func (c *TexturedCloser) GetOptions() *CloseMeshOptions {
	return c.options
}

func (c *TexturedCloser) CloseSurfaceMesh(mesh interface{}, thickness float64) (*Mesh, error) {
	if !c.options.Enabled {
		return nil, nil
	}

	c.options.Thickness = thickness
	return c.closeMeshWithTexture(mesh)
}

func (c *TexturedCloser) CloseSurfaceMeshWithOptions(mesh interface{}, options *CloseMeshOptions) (*Mesh, error) {
	if !options.Enabled {
		return nil, nil
	}

	c.options = options
	return c.closeMeshWithTexture(mesh)
}

func (c *TexturedCloser) closeMeshWithTexture(mesh interface{}) (*Mesh, error) {
	if mesh == nil {
		return nil, nil
	}

	type meshWithVertices interface {
		GetVertices() []vec3d.T
		GetIndices() []uint32
		GetMinHeight() float64
		GetBounds() vec2d.Rect
	}

	m, ok := mesh.(meshWithVertices)
	if !ok {
		return nil, fmt.Errorf("mesh does not implement required interface")
	}

	vertices := m.GetVertices()
	indices := m.GetIndices()
	minHeight := m.GetMinHeight()
	bounds := m.GetBounds()

	if len(vertices) == 0 {
		return nil, nil
	}

	var baseHeight float64
	if c.options.UseMinHeightAsBase {
		baseHeight = minHeight - c.options.Thickness
	} else {
		baseHeight = -c.options.Thickness
	}

	newVertices := make([]vec3d.T, len(vertices)*2)
	copy(newVertices, vertices)

	for i := 0; i < len(vertices); i++ {
		newVertices[len(vertices)+i] = vec3d.T{
			vertices[i][0],
			vertices[i][1],
			baseHeight,
		}
	}

	newIndices := make([]uint32, len(indices)*2)
	copy(newIndices, indices)

	offset := uint32(len(vertices))

	for i := 0; i < len(indices); i += 3 {
		idx := len(indices) + i
		newIndices[idx+0] = indices[i+2] + offset
		newIndices[idx+1] = indices[i+1] + offset
		newIndices[idx+2] = indices[i+0] + offset
	}

	result := &Mesh{
		Vertices: newVertices,
		Indices:  newIndices,
	}

	if c.options.BottomTextureImage != nil || c.options.BottomColor != nil {
		result.Texture = c.options.BottomTextureImage
		result.UVs = c.calculateBottomUVs(newVertices, bounds, c.options.BottomTextureTilingU, c.options.BottomTextureTilingV)
		result.Materials = []Material{*NewMaterial()}
		if c.options.BottomColor != nil {
			result.Materials[0].Diffuse = c.options.BottomColor
		}
	}

	result.CalculateNormals()

	return result, nil
}

func (c *TexturedCloser) calculateBottomUVs(vertices []vec3d.T, bounds vec2d.Rect, tilingU, tilingV float64) []vec2d.T {
	if len(vertices) == 0 {
		return nil
	}

	uvs := make([]vec2d.T, len(vertices))
	minX := bounds.Min[0]
	minY := bounds.Min[1]
	width := bounds.Max[0] - bounds.Min[0]
	depth := bounds.Max[1] - bounds.Min[1]

	if width == 0 || depth == 0 {
		for i := range uvs {
			uvs[i] = vec2d.T{0.5, 0.5}
		}
		return uvs
	}

	for i, v := range vertices {
		u := (v[0] - minX) / width * tilingU
		vCoord := (v[1] - minY) / depth * tilingV
		uvs[i] = vec2d.T{u, vCoord}
	}

	return uvs
}

func (c *TexturedCloser) CloseUnifiedMesh(mesh *Mesh, baseHeight float64) (*Mesh, error) {
	if mesh == nil {
		return nil, nil
	}

	vertices := mesh.Vertices
	indices := mesh.Indices

	if len(vertices) == 0 {
		return nil, nil
	}

	if c.options == nil {
		c.options = NewDefaultCloseMeshOptions()
	}

	safetyMargin := c.options.Thickness * 0.1
	adjustedBaseHeight := baseHeight - safetyMargin

	for _, v := range vertices {
		if v[2] < adjustedBaseHeight+safetyMargin {
			penetration := adjustedBaseHeight + safetyMargin - v[2]
			adjustedBaseHeight -= penetration + safetyMargin
		}
	}

	bottomVertices := make([]vec3d.T, len(vertices))
	for i := 0; i < len(vertices); i++ {
		bottomVertices[i] = vec3d.T{
			vertices[i][0],
			vertices[i][1],
			adjustedBaseHeight,
		}
	}

	newVertices := make([]vec3d.T, len(vertices)*2)
	copy(newVertices, vertices)
	copy(newVertices[len(vertices):], bottomVertices)

	newIndices := make([]uint32, len(indices)*2)
	copy(newIndices, indices)

	offset := uint32(len(vertices))

	for i := 0; i < len(indices); i += 3 {
		idx := len(indices) + i
		newIndices[idx+0] = indices[i+2] + offset
		newIndices[idx+1] = indices[i+1] + offset
		newIndices[idx+2] = indices[i+0] + offset
	}

	result := &Mesh{
		Vertices: newVertices,
		Indices:  newIndices,
		Bounds:   mesh.Bounds,
		Srs:      mesh.Srs,
		Texture:  mesh.Texture,
	}

	if len(mesh.UVs) > 0 {
		result.UVs = make([]vec2d.T, len(newVertices))
		copy(result.UVs, mesh.UVs)
		bottomUVs := c.calculateBottomUVs(bottomVertices, mesh.Bounds,
			c.options.BottomTextureTilingU,
			c.options.BottomTextureTilingV)
		copy(result.UVs[len(vertices):], bottomUVs)
	} else if c.options.BottomTextureImage != nil || c.options.BottomColor != nil {
		result.Texture = c.options.BottomTextureImage
		result.UVs = c.calculateBottomUVs(newVertices, mesh.Bounds,
			c.options.BottomTextureTilingU,
			c.options.BottomTextureTilingV)
		result.Materials = []Material{*NewMaterial()}
		if c.options.BottomColor != nil {
			result.Materials[0].Diffuse = c.options.BottomColor
		}
	}

	result.CalculateNormals()

	return result, nil
}
