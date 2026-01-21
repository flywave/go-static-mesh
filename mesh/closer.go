package mesh

import (
	vec3d "github.com/flywave/go3d/float64/vec3"
)

type MeshCloser interface {
	CloseSurfaceMesh(mesh interface{}, thickness float64) (*Mesh, error)
	CloseWithBase(mesh interface{}, baseHeight float64) (*Mesh, error)
}

type SimpleCloser struct{}

func (c *SimpleCloser) CloseSurfaceMesh(mesh interface{}, thickness float64) (*Mesh, error) {
	if mesh == nil {
		return nil, nil
	}

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
	minHeight := m.GetMinHeight()

	if len(vertices) == 0 {
		return nil, nil
	}

	baseHeight := minHeight - thickness

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

	result.CalculateNormals()

	return result, nil
}

func (c *SimpleCloser) CloseWithBase(mesh interface{}, baseHeight float64) (*Mesh, error) {
	type meshWithMinHeight interface {
		GetMinHeight() float64
	}

	m, ok := mesh.(meshWithMinHeight)
	if !ok {
		return nil, nil
	}

	return c.CloseSurfaceMesh(mesh, m.GetMinHeight()-baseHeight)
}
