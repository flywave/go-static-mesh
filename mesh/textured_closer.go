package mesh

import (
	"image"
	"image/color"

	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

type edgeKey struct {
	a, b uint32
}

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

func (c *TexturedCloser) SetOptions(opts *CloseMeshOptions) {
	if opts == nil {
		opts = NewDefaultCloseMeshOptions()
	}
	c.options = opts
}

func (c *TexturedCloser) buildSideWalls(indices []uint32, offset uint32) []uint32 {
	edgeCount := make(map[edgeKey]int)

	for i := 0; i < len(indices); i += 3 {
		tri := [...]uint32{indices[i], indices[i+1], indices[i+2]}
		for j := 0; j < 3; j++ {
			a, b := tri[j], tri[(j+1)%3]
			if a > b {
				a, b = b, a
			}
			edgeCount[edgeKey{a, b}]++
		}
	}

	sideIndices := make([]uint32, 0, len(edgeCount)*6)
	for i := 0; i < len(indices); i += 3 {
		tri := [...]uint32{indices[i], indices[i+1], indices[i+2]}
		for j := 0; j < 3; j++ {
			a, b := tri[j], tri[(j+1)%3]
			ea, eb := a, b
			if ea > eb {
				ea, eb = eb, ea
			}
			if edgeCount[edgeKey{ea, eb}] != 1 {
				continue
			}
			edgeCount[edgeKey{ea, eb}] = 0
			sideIndices = append(sideIndices,
				a, a+offset, b,
				a+offset, b+offset, b,
			)
		}
	}

	return sideIndices
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

// CloseUnifiedMesh 由顶面复制出底面并沿边界补齐侧墙，得到封闭实体。
// 底面顶点是独立副本，因此底部可以有自己的纹理与 UV；
// 侧墙复用顶/底面顶点（每个顶点只有一组 UV），因此侧面只能表达颜色，不能表达独立纹理。
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

	vertexCount := len(vertices)
	bottomVertices := make([]vec3d.T, vertexCount)
	for i := 0; i < vertexCount; i++ {
		bottomVertices[i] = vec3d.T{
			vertices[i][0],
			vertices[i][1],
			adjustedBaseHeight,
		}
	}

	newVertices := make([]vec3d.T, 0, vertexCount*2)
	newVertices = append(newVertices, vertices...)
	newVertices = append(newVertices, bottomVertices...)

	uvs := make([]vec2d.T, 0, vertexCount*2)
	if len(mesh.UVs) == vertexCount {
		uvs = append(uvs, mesh.UVs...)
	} else {
		uvs = append(uvs, make([]vec2d.T, vertexCount)...)
	}
	uvs = append(uvs, c.calculateBottomUVs(bottomVertices, mesh.Bounds,
		c.options.BottomTextureTilingU, c.options.BottomTextureTilingV)...)

	newIndices := make([]uint32, 0, len(indices)*2)
	materialIndices := make([]uint32, 0, len(indices)/3*2)

	// 材质槽：顶面固定 0，底面/侧面按需追加
	materials := []Material{*NewMaterial()}
	if mesh.Texture != nil {
		materials[0].Diffuse = color.RGBA{255, 255, 255, 255}
	}

	bottomMaterial := uint32(0)
	if c.options.ApplyToBottom {
		material := *NewMaterial()
		material.Name = "bottom"
		if c.options.BottomColor != nil {
			material.Diffuse = c.options.BottomColor
		}
		material.Texture = c.options.BottomTextureImage
		materials = append(materials, material)
		bottomMaterial = uint32(len(materials) - 1)
	}

	sideMaterial := uint32(0)
	if c.options.ApplyToSides {
		sideColor := color.RGBA{160, 130, 90, 255}
		if c.options.SideColor != nil {
			if c, ok := c.options.SideColor.(color.RGBA); ok {
				sideColor = c
			}
		}
		materials = append(materials, *NewPBRMaterial("side", sideColor, 0.6, 0.8))
		sideMaterial = uint32(len(materials) - 1)
	}

	// 顶面沿用原始索引
	newIndices = append(newIndices, indices...)
	materialIndices = append(materialIndices, make([]uint32, len(indices)/3)...)

	// 底面：镜像顶面并反转绕序
	offset := uint32(vertexCount)
	for i := 0; i < len(indices); i += 3 {
		newIndices = append(newIndices,
			indices[i+2]+offset,
			indices[i+1]+offset,
			indices[i+0]+offset)
		materialIndices = append(materialIndices, bottomMaterial)
	}

	sideIndices := c.buildSideWalls(indices, offset)
	newIndices = append(newIndices, sideIndices...)
	for i := 0; i < len(sideIndices)/3; i++ {
		materialIndices = append(materialIndices, sideMaterial)
	}

	result := &Mesh{
		Vertices:        newVertices,
		Indices:         newIndices,
		UVs:             uvs,
		Materials:       materials,
		MaterialIndices: materialIndices,
		Bounds:          mesh.Bounds,
		Srs:             mesh.Srs,
		Texture:         mesh.Texture,
	}

	result.CalculateNormals()

	return result, nil
}
