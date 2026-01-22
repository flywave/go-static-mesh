package extruder

import (
	"fmt"
	"math"

	"github.com/flywave/go-static-mesh/draw"
	"github.com/flywave/go-static-mesh/mesh"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

type MeshObject interface {
	draw.MapObject

	ExtrudeToMesh(mesh *mesh.Mesh, height float64) error
	ExtrudeToMeshWithResolution(mesh *mesh.Mesh, height, resolution float64) error
}

type ExtrudeOptions struct {
	Resolution float64
	Radius     float64
	Sides      bool
	Bottom     bool
	Segments   int
}

type PathExtruder struct{}

func NewPathExtruder() *PathExtruder {
	return &PathExtruder{}
}

func (e *PathExtruder) ExtrudeToMesh(path *draw.Path, mesh *mesh.Mesh, height float64) error {
	options := &ExtrudeOptions{
		Resolution: 1.0,
		Radius:     path.Weight / 2.0,
		Segments:   16,
	}
	return e.ExtrudeToMeshWithResolution(path, mesh, height, options)
}

func (e *PathExtruder) ExtrudeToMeshWithResolution(path *draw.Path, mesh *mesh.Mesh, height float64, options *ExtrudeOptions) error {
	if len(path.Positions) < 2 {
		return nil
	}

	segmentCount := len(path.Positions) - 1

	for seg := 0; seg < segmentCount; seg++ {
		start := path.Positions[seg]
		end := path.Positions[seg+1]

		x1, y1 := start[0], start[1]
		x2, y2 := end[0], end[1]

		dx := x2 - x1
		dy := y2 - y1
		length := math.Sqrt(dx*dx + dy*dy)

		if length < 0.0001 {
			continue
		}

		perpX := -dy / length
		perpY := dx / length

		offsetX := options.Radius * perpX
		offsetY := options.Radius * perpY

		v1 := vec3d.T{x1 + offsetX, y1 + offsetY, 0}
		v2 := vec3d.T{x1 + offsetX, y1 + offsetY, height}
		v3 := vec3d.T{x2 + offsetX, y2 + offsetY, 0}
		v4 := vec3d.T{x2 + offsetX, y2 + offsetY, height}

		baseIdx := uint32(len(mesh.Vertices))

		mesh.Vertices = append(mesh.Vertices, v1, v2, v3, v4)

		mesh.Indices = append(mesh.Indices,
			baseIdx, baseIdx+1, baseIdx+2,
			baseIdx+2, baseIdx+1, baseIdx+3,
		)
	}

	return nil
}

type AreaExtruder struct {
	triangulator Triangulator
}

func NewAreaExtruder() *AreaExtruder {
	return &AreaExtruder{
		triangulator: NewEarClippingTriangulator(),
	}
}

func (e *AreaExtruder) ExtrudeToMesh(area *draw.Area, mesh *mesh.Mesh, height float64) error {
	return e.ExtrudeToMeshWithResolution(area, mesh, height, 1.0)
}

func (e *AreaExtruder) ExtrudeToMeshWithResolution(area *draw.Area, mesh *mesh.Mesh, height, resolution float64) error {
	if len(area.Positions) < 3 {
		return nil
	}

	triangles, err := e.triangulator.Triangulate(area.Positions, resolution)
	if err != nil {
		return fmt.Errorf("failed to triangulate area: %w", err)
	}

	topVertexStart := uint32(len(mesh.Vertices))
	bottomVertexStart := topVertexStart + uint32(len(triangles)*3)

	for _, tri := range triangles {
		v1 := area.Positions[tri[0]]
		v2 := area.Positions[tri[1]]
		v3 := area.Positions[tri[2]]

		mesh.Vertices = append(mesh.Vertices, vec3d.T{v1[0], v1[1], height})
		mesh.Vertices = append(mesh.Vertices, vec3d.T{v2[0], v2[1], height})
		mesh.Vertices = append(mesh.Vertices, vec3d.T{v3[0], v3[1], height})
		mesh.Vertices = append(mesh.Vertices, vec3d.T{v1[0], v1[1], 0})
		mesh.Vertices = append(mesh.Vertices, vec3d.T{v2[0], v2[1], 0})
		mesh.Vertices = append(mesh.Vertices, vec3d.T{v3[0], v3[1], 0})
	}

	numTriangles := uint32(len(triangles))
	for i := uint32(0); i < numTriangles; i++ {
		baseTop := topVertexStart + i*3
		baseBottom := bottomVertexStart + i*3

		mesh.Indices = append(mesh.Indices, baseTop, baseTop+1, baseTop+2)
		mesh.Indices = append(mesh.Indices, baseBottom+2, baseBottom+1, baseBottom)

		mesh.Indices = append(mesh.Indices,
			baseTop, baseBottom, baseTop+1,
		)
		mesh.Indices = append(mesh.Indices,
			baseTop+1, baseBottom, baseBottom+1,
		)

		mesh.Indices = append(mesh.Indices,
			baseTop+1, baseBottom+1, baseTop+2,
		)
		mesh.Indices = append(mesh.Indices,
			baseTop+2, baseBottom+1, baseBottom+2,
		)

		mesh.Indices = append(mesh.Indices,
			baseTop+2, baseBottom+2, baseTop,
		)
		mesh.Indices = append(mesh.Indices,
			baseTop, baseBottom+2, baseBottom,
		)
	}

	return nil
}

type Triangulator interface {
	Triangulate(points []vec2d.T, resolution float64) ([][3]int, error)
}

type EarClippingTriangulator struct{}

func NewEarClippingTriangulator() *EarClippingTriangulator {
	return &EarClippingTriangulator{}
}

func (t *EarClippingTriangulator) Triangulate(points []vec2d.T, resolution float64) ([][3]int, error) {
	if len(points) < 3 {
		return [][3]int{{0, 1, 2}}, nil
	}

	if len(points) == 3 {
		return [][3]int{{0, 1, 2}}, nil
	}

	if len(points) == 4 {
		return [][3]int{
			{0, 1, 2},
			{0, 2, 3},
		}, nil
	}

	indices := make([]int, len(points))
	for i := 0; i < len(points); i++ {
		indices[i] = i
	}

	var triangles [][3]int
	maxIterations := len(points) * len(points)

	for len(indices) > 3 {
		if maxIterations <= 0 {
			break
		}
		maxIterations--

		triangle := t.findEarAndClip(indices, points)
		if len(triangle) == 0 {
			break
		}
		triangles = append(triangles, triangle)
	}

	if len(indices) == 3 {
		triangles = append(triangles, [3]int{indices[0], indices[1], indices[2]})
	}

	return triangles, nil
}

func (t *EarClippingTriangulator) findEarAndClip(indices []int, points []vec2d.T) [3]int {
	maxArea := -1.0
	earIndex := -1

	for i := 0; i < len(indices); i++ {
		if t.isReflex(indices, points, i) {
			continue
		}

		area := t.calculateArea(indices, points, i)
		if area > maxArea {
			maxArea = area
			earIndex = i
		}
	}

	if earIndex == -1 {
		return [3]int{}
	}

	prev := (earIndex - 1 + len(indices)) % len(indices)
	next := (earIndex + 1) % len(indices)

	triangle := [3]int{indices[prev], indices[earIndex], indices[next]}

	indices = append(indices[:earIndex], indices[earIndex+1:]...)

	return triangle
}

func (t *EarClippingTriangulator) isReflex(indices []int, points []vec2d.T, vertex int) bool {
	if len(indices) < 3 {
		return false
	}

	prev := (vertex - 1 + len(indices)) % len(indices)
	next := (vertex + 1) % len(indices)

	p := points[indices[prev]]
	c := points[indices[vertex]]
	n := points[indices[next]]

	v1 := vec2d.T{c[0] - p[0], c[1] - p[1]}
	v2 := vec2d.T{n[0] - c[0], n[1] - c[1]}

	cross := v1[0]*v2[1] - v1[1]*v2[0]
	return cross < 0
}

func (t *EarClippingTriangulator) calculateArea(indices []int, points []vec2d.T, vertex int) float64 {
	prev := (vertex - 1 + len(indices)) % len(indices)
	next := (vertex + 1) % len(indices)

	pi := points[indices[prev]]
	pj := points[indices[vertex]]
	pk := points[indices[next]]

	area := math.Abs((pj[0]-pi[0])*(pk[1]-pi[1])-(pi[0]-pk[0])*(pj[1]-pi[1])) / 2.0
	return area
}
