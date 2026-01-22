package mesh

import (
	"testing"

	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

func TestNewBSPPlane(t *testing.T) {
	normal := vec3d.T{0, 0, 1}
	distance := 10.0

	plane := NewBSPPlane(normal, distance)

	if plane.Normal[0] != 0 || plane.Normal[1] != 0 || plane.Normal[2] != 1 {
		t.Errorf("Expected normal [0, 0, 1], got %v", plane.Normal)
	}
	if plane.Distance != 10.0 {
		t.Errorf("Expected distance 10.0, got %v", plane.Distance)
	}
}

func TestBSPPlaneDistanceToPoint(t *testing.T) {
	plane := NewBSPPlane(vec3d.T{0, 0, 1}, -10.0)

	tests := []struct {
		name     string
		point    vec3d.T
		expected float64
	}{
		{"Point on plane", vec3d.T{0, 0, 10}, 0},
		{"Point above plane", vec3d.T{0, 0, 15}, 5},
		{"Point below plane", vec3d.T{0, 0, 5}, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := plane.DistanceToPoint(tt.point)
			if dist != tt.expected {
				t.Errorf("Expected distance %v, got %v", tt.expected, dist)
			}
		})
	}
}

func TestBSPPlaneClassify(t *testing.T) {
	plane := NewBSPPlane(vec3d.T{0, 0, 1}, -10.0)

	tests := []struct {
		name     string
		point    vec3d.T
		expected int
	}{
		{"Point on plane", vec3d.T{0, 0, 10}, 0},
		{"Point above plane", vec3d.T{0, 0, 15}, 1},
		{"Point below plane", vec3d.T{0, 0, 5}, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := plane.Classify(tt.point)
			if result != tt.expected {
				t.Errorf("Expected classification %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestBSPPlaneFlip(t *testing.T) {
	plane := NewBSPPlane(vec3d.T{1, 2, 3}, 10.0)

	originalNormal := vec3d.T{plane.Normal[0], plane.Normal[1], plane.Normal[2]}
	originalDistance := plane.Distance

	plane.Flip()

	if plane.Normal[0] != -originalNormal[0] ||
		plane.Normal[1] != -originalNormal[1] ||
		plane.Normal[2] != -originalNormal[2] {
		t.Errorf("Normal not flipped correctly: expected %v, got %v",
			vec3d.T{-originalNormal[0], -originalNormal[1], -originalNormal[2]},
			plane.Normal)
	}
	if plane.Distance != -originalDistance {
		t.Errorf("Distance not flipped correctly: expected %v, got %v", -originalDistance, plane.Distance)
	}
}

func TestNewBSPPolygon(t *testing.T) {
	vertices := []*BSPVertex{
		{Position: vec3d.T{0, 0, 0}},
		{Position: vec3d.T{1, 0, 0}},
		{Position: vec3d.T{0, 1, 0}},
	}

	poly := NewBSPPolygon(vertices)

	if poly == nil {
		t.Fatal("Expected polygon, got nil")
	}
	if len(poly.Vertices) != 3 {
		t.Errorf("Expected 3 vertices, got %d", len(poly.Vertices))
	}
	if poly.Plane == nil {
		t.Error("Expected plane to be calculated")
	}
}

func TestNewBSPPolygonInsufficientVertices(t *testing.T) {
	vertices := []*BSPVertex{
		{Position: vec3d.T{0, 0, 0}},
		{Position: vec3d.T{1, 0, 0}},
	}

	poly := NewBSPPolygon(vertices)

	if poly != nil {
		t.Error("Expected nil for polygon with insufficient vertices")
	}
}

func TestBSPPolygonFlip(t *testing.T) {
	vertices := []*BSPVertex{
		{Position: vec3d.T{0, 0, 0}},
		{Position: vec3d.T{1, 0, 0}},
		{Position: vec3d.T{0, 1, 0}},
	}

	poly := NewBSPPolygon(vertices)
	originalOrder := make([]*BSPVertex, len(poly.Vertices))
	copy(originalOrder, poly.Vertices)

	poly.Flip()

	for i := 0; i < len(originalOrder); i++ {
		if poly.Vertices[i] != originalOrder[len(originalOrder)-1-i] {
			t.Errorf("Vertex order not flipped correctly at index %d", i)
		}
	}
}

func TestBSPPolygonTriangles(t *testing.T) {
	tests := []struct {
		name          string
		vertices      []*BSPVertex
		expectedCount int
	}{
		{
			name: "Triangle",
			vertices: []*BSPVertex{
				{Position: vec3d.T{0, 0, 0}},
				{Position: vec3d.T{1, 0, 0}},
				{Position: vec3d.T{0, 1, 0}},
			},
			expectedCount: 1,
		},
		{
			name: "Quad",
			vertices: []*BSPVertex{
				{Position: vec3d.T{0, 0, 0}},
				{Position: vec3d.T{1, 0, 0}},
				{Position: vec3d.T{1, 1, 0}},
				{Position: vec3d.T{0, 1, 0}},
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poly := NewBSPPolygon(tt.vertices)
			triangles := poly.Triangles()

			if len(triangles) != tt.expectedCount {
				t.Errorf("Expected %d triangles, got %d", tt.expectedCount, len(triangles))
			}

			for _, tri := range triangles {
				if len(tri.Vertices) != 3 {
					t.Errorf("Triangle should have 3 vertices, got %d", len(tri.Vertices))
				}
			}
		})
	}
}

func TestBSPPolygonCenter(t *testing.T) {
	vertices := []*BSPVertex{
		{Position: vec3d.T{0, 0, 0}},
		{Position: vec3d.T{2, 0, 0}},
		{Position: vec3d.T{0, 2, 0}},
	}

	poly := NewBSPPolygon(vertices)
	center := poly.Center()

	if center[0] != 2.0/3.0 || center[1] != 2.0/3.0 || center[2] != 0 {
		t.Errorf("Expected center [0.666..., 0.666..., 0], got %v", center)
	}
}

func TestBSPPolygonCenterEmpty(t *testing.T) {
	vertices := []*BSPVertex{}

	poly := NewBSPPolygon(vertices)
	if poly == nil {
		poly = &BSPPolygon{Vertices: vertices}
	}
	center := poly.Center()

	if center[0] != 0 || center[1] != 0 || center[2] != 0 {
		t.Errorf("Expected center [0, 0, 0] for empty polygon, got %v", center)
	}
}

func TestBSPPolygonsClone(t *testing.T) {
	vertices := []*BSPVertex{
		{Position: vec3d.T{0, 0, 0}, Normal: vec3d.T{0, 0, 1}},
		{Position: vec3d.T{1, 0, 0}, Normal: vec3d.T{0, 0, 1}},
		{Position: vec3d.T{0, 1, 0}, Normal: vec3d.T{0, 0, 1}},
	}

	poly := NewBSPPolygon(vertices)
	polygons := BSPPolygons{poly}
	clone := polygons.Clone()

	if len(clone) != len(polygons) {
		t.Errorf("Expected clone to have same length, got %d vs %d", len(clone), len(polygons))
	}

	if clone[0] == polygons[0] {
		t.Error("Expected clone to have different polygon instances")
	}

	if clone[0].Vertices[0] == polygons[0].Vertices[0] {
		t.Error("Expected clone to have different vertex instances")
	}
}

func TestBSPPolygonsBounds(t *testing.T) {
	vertices := []*BSPVertex{
		{Position: vec3d.T{0, 0, 0}},
		{Position: vec3d.T{2, 0, 0}},
		{Position: vec3d.T{0, 2, 0}},
	}

	poly := NewBSPPolygon(vertices)
	polygons := BSPPolygons{poly}

	min, max := polygons.Bounds()

	if min[0] != 0 || min[1] != 0 || min[2] != 0 {
		t.Errorf("Expected min [0, 0, 0], got %v", min)
	}
	if max[0] != 2 || max[1] != 2 || max[2] != 0 {
		t.Errorf("Expected max [2, 2, 0], got %v", max)
	}
}

func TestBSPPolygonsBoundsEmpty(t *testing.T) {
	polygons := BSPPolygons{}
	min, max := polygons.Bounds()

	if min[0] != 0 || min[1] != 0 || min[2] != 0 {
		t.Errorf("Expected min [0, 0, 0] for empty polygons, got %v", min)
	}
	if max[0] != 0 || max[1] != 0 || max[2] != 0 {
		t.Errorf("Expected max [0, 0, 0] for empty polygons, got %v", max)
	}
}

func TestBSPPolygonsAllVertices(t *testing.T) {
	vertices1 := []*BSPVertex{
		{Position: vec3d.T{0, 0, 0}},
		{Position: vec3d.T{1, 0, 0}},
		{Position: vec3d.T{0, 1, 0}},
	}

	vertices2 := []*BSPVertex{
		{Position: vec3d.T{1, 0, 0}},
		{Position: vec3d.T{2, 0, 0}},
		{Position: vec3d.T{1, 1, 0}},
	}

	poly1 := NewBSPPolygon(vertices1)
	poly2 := NewBSPPolygon(vertices2)
	polygons := BSPPolygons{poly1, poly2}

	allVertices := polygons.AllVertices()

	if len(allVertices) != 6 {
		t.Errorf("Expected 6 vertices, got %d", len(allVertices))
	}
}

func TestBSPVertex(t *testing.T) {
	pos := vec3d.T{1, 2, 3}
	normal := vec3d.T{0, 0, 1}
	color := [3]uint8{255, 0, 0}
	alpha := uint8(128)
	uv := vec2d.T{0.5, 0.5}

	vertex := &BSPVertex{
		Position: pos,
		Normal:   normal,
		Color:    color,
		Alpha:    alpha,
		UV:       uv,
	}

	if vertex.Position != pos {
		t.Errorf("Expected position %v, got %v", pos, vertex.Position)
	}
	if vertex.Normal != normal {
		t.Errorf("Expected normal %v, got %v", normal, vertex.Normal)
	}
	if vertex.Color != color {
		t.Errorf("Expected color %v, got %v", color, vertex.Color)
	}
	if vertex.Alpha != alpha {
		t.Errorf("Expected alpha %d, got %d", alpha, vertex.Alpha)
	}
	if vertex.UV != uv {
		t.Errorf("Expected UV %v, got %v", uv, vertex.UV)
	}
}

func TestPlaneFromBSPVertices(t *testing.T) {
	vertices := []*BSPVertex{
		{Position: vec3d.T{0, 0, 0}},
		{Position: vec3d.T{1, 0, 0}},
		{Position: vec3d.T{0, 1, 0}},
	}

	plane := PlaneFromBSPVertices(vertices)

	if plane == nil {
		t.Fatal("Expected plane, got nil")
	}

	normal := plane.Normal
	if normal[0] != 0 || normal[1] != 0 || normal[2] != 1 {
		t.Errorf("Expected normal [0, 0, 1], got %v", normal)
	}
}

func TestPlaneFromBSPVerticesInsufficient(t *testing.T) {
	vertices := []*BSPVertex{
		{Position: vec3d.T{0, 0, 0}},
		{Position: vec3d.T{1, 0, 0}},
	}

	plane := PlaneFromBSPVertices(vertices)

	if plane != nil {
		t.Error("Expected nil for insufficient vertices")
	}
}
