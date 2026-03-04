package bsp

import (
	"math"

	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

type BSPVertex struct {
	Position vec3d.T
	Normal   vec3d.T
	Color    [3]uint8
	Alpha    uint8
	UV       vec2d.T
}

type BSPPolygon struct {
	Plane    *BSPPlane
	Vertices []*BSPVertex
	Shared   bool
}

type BSPPlane struct {
	Normal   vec3d.T
	Distance float64
}

type BSPPolygons []*BSPPolygon

func NewBSPPlane(normal vec3d.T, distance float64) *BSPPlane {
	return &BSPPlane{
		Normal:   normal,
		Distance: distance,
	}
}

func (p *BSPPlane) DistanceToPoint(point vec3d.T) float64 {
	return vec3d.Dot(&p.Normal, &point) + p.Distance
}

func (p *BSPPlane) Flip() {
	p.Normal[0] = -p.Normal[0]
	p.Normal[1] = -p.Normal[1]
	p.Normal[2] = -p.Normal[2]
	p.Distance = -p.Distance
}

func (p *BSPPlane) Classify(point vec3d.T) int {
	dist := p.DistanceToPoint(point)
	if dist < -1e-6 {
		return -1
	} else if dist > 1e-6 {
		return 1
	}
	return 0
}

func (p *BSPPlane) SplitPolygon(poly *BSPPolygon, coplanarFront, coplanarBack, front, back *BSPPolygons) {
	if poly.Plane == nil {
		poly.Plane = PlaneFromBSPVertices(poly.Vertices)
	}

	classifications := make([]int, len(poly.Vertices))
	for i, v := range poly.Vertices {
		classifications[i] = p.Classify(v.Position)
	}

	var frontVerts []*BSPVertex
	var backVerts []*BSPVertex
	var coplanarVerts []*BSPVertex

	for i, classification := range classifications {
		v := poly.Vertices[i]
		nextV := poly.Vertices[(i+1)%len(poly.Vertices)]
		nextClassification := classifications[(i+1)%len(classifications)]

		switch {
		case classification == -1 && nextClassification == -1:
			backVerts = append(backVerts, v)
		case classification == 1 && nextClassification == 1:
			frontVerts = append(frontVerts, v)
		case classification == 1 && nextClassification == -1:
			frontVerts = append(frontVerts, v)
			intersectionPoint := p.LineIntersection(v.Position, nextV.Position)
			intersectionVertex := &BSPVertex{
				Position: intersectionPoint,
				Normal:   v.Normal,
				Color:    v.Color,
				Alpha:    v.Alpha,
				UV:       v.UV,
			}
			frontVerts = append(frontVerts, intersectionVertex)
			backVerts = append(backVerts, intersectionVertex)
		case classification == -1 && nextClassification == 1:
			backVerts = append(backVerts, v)
			intersectionPoint := p.LineIntersection(v.Position, nextV.Position)
			intersectionVertex := &BSPVertex{
				Position: intersectionPoint,
				Normal:   v.Normal,
				Color:    v.Color,
				Alpha:    v.Alpha,
				UV:       v.UV,
			}
			backVerts = append(backVerts, intersectionVertex)
			frontVerts = append(frontVerts, intersectionVertex)
		case classification == 0 || nextClassification == 0:
			coplanarVerts = append(coplanarVerts, v)
		}
	}

	if len(frontVerts) >= 3 {
		*front = append(*front, NewBSPPolygon(frontVerts))
	}
	if len(backVerts) >= 3 {
		*back = append(*back, NewBSPPolygon(backVerts))
	}
	if len(coplanarVerts) >= 3 {
		coplanarPoly := NewBSPPolygon(coplanarVerts)
		coplanarPoly.Shared = true
		if vec3d.Dot(&p.Normal, &coplanarPoly.Plane.Normal) >= 0 {
			*coplanarFront = append(*coplanarFront, coplanarPoly)
		} else {
			*coplanarBack = append(*coplanarBack, coplanarPoly)
		}
	}
}

func (p *BSPPlane) LineIntersection(a, b vec3d.T) vec3d.T {
	ab := vec3d.Sub(&b, &a)
	dotNormalAb := vec3d.Dot(&p.Normal, &ab)

	const EPSILON = 1e-10
	if math.Abs(dotNormalAb) < EPSILON {
		return a
	}

	dotNormalA := vec3d.Dot(&p.Normal, &a)
	t := (-p.Distance - dotNormalA) / dotNormalAb

	t = math.Max(0, math.Min(1, t))

	abScaled := vec3d.T{
		ab[0] * t,
		ab[1] * t,
		ab[2] * t,
	}
	result := vec3d.Add(&a, &abScaled)
	return result
}

func NewBSPPolygon(vertices []*BSPVertex) *BSPPolygon {
	if len(vertices) < 3 {
		return nil
	}

	poly := &BSPPolygon{
		Vertices: vertices,
	}

	if len(vertices) >= 3 {
		poly.Plane = PlaneFromBSPVertices(vertices)
	}

	return poly
}

func PlaneFromBSPVertices(vertices []*BSPVertex) *BSPPlane {
	if len(vertices) < 3 {
		return nil
	}

	a := vertices[0].Position
	b := vertices[1].Position
	c := vertices[2].Position

	ab := [3]float64{
		b[0] - a[0],
		b[1] - a[1],
		b[2] - a[2],
	}
	ac := [3]float64{
		c[0] - a[0],
		c[1] - a[1],
		c[2] - a[2],
	}

	normal := [3]float64{
		ab[1]*ac[2] - ab[2]*ac[1],
		ab[2]*ac[0] - ab[0]*ac[2],
		ab[0]*ac[1] - ab[1]*ac[0],
	}
	length := math.Sqrt(normal[0]*normal[0] + normal[1]*normal[1] + normal[2]*normal[2])
	if length > 0 {
		normal[0] /= length
		normal[1] /= length
		normal[2] /= length
	}

	distance := -(normal[0]*a[0] + normal[1]*a[1] + normal[2]*a[2])

	return &BSPPlane{
		Normal:   normal,
		Distance: distance,
	}
}

func (p *BSPPolygon) Flip() {
	for i, j := 0, len(p.Vertices)-1; i < j; i, j = i+1, j-1 {
		p.Vertices[i], p.Vertices[j] = p.Vertices[j], p.Vertices[i]
	}
	if p.Plane != nil {
		p.Plane.Flip()
	}
}

func (p *BSPPolygon) Triangles() []*BSPPolygon {
	if len(p.Vertices) < 3 {
		return nil
	}

	if len(p.Vertices) == 3 {
		return []*BSPPolygon{p}
	}

	triangles := make([]*BSPPolygon, 0, len(p.Vertices)-2)
	for i := 1; i < len(p.Vertices)-1; i++ {
		triVerts := []*BSPVertex{
			p.Vertices[0],
			p.Vertices[i],
			p.Vertices[i+1],
		}
		tri := NewBSPPolygon(triVerts)
		if tri != nil {
			triangles = append(triangles, tri)
		}
	}

	return triangles
}

func (p *BSPPolygon) Center() vec3d.T {
	if len(p.Vertices) == 0 {
		return vec3d.T{0, 0, 0}
	}

	center := vec3d.T{0, 0, 0}
	for _, v := range p.Vertices {
		center[0] += v.Position[0]
		center[1] += v.Position[1]
		center[2] += v.Position[2]
	}

	center[0] /= float64(len(p.Vertices))
	center[1] /= float64(len(p.Vertices))
	center[2] /= float64(len(p.Vertices))

	return center
}

func (ps BSPPolygons) Clone() BSPPolygons {
	clone := make(BSPPolygons, len(ps))
	for i, p := range ps {
		newVertices := make([]*BSPVertex, len(p.Vertices))
		for j, v := range p.Vertices {
			newVertices[j] = &BSPVertex{
				Position: v.Position,
				Normal:   v.Normal,
				Color:    v.Color,
				Alpha:    v.Alpha,
				UV:       v.UV,
			}
		}
		clone[i] = &BSPPolygon{
			Plane:    p.Plane,
			Vertices: newVertices,
			Shared:   p.Shared,
		}
	}
	return clone
}

func (ps BSPPolygons) Bounds() (vec3d.T, vec3d.T) {
	if len(ps) == 0 {
		return vec3d.T{0, 0, 0}, vec3d.T{0, 0, 0}
	}

	min := vec3d.T{math.Inf(1), math.Inf(1), math.Inf(1)}
	max := vec3d.T{math.Inf(-1), math.Inf(-1), math.Inf(-1)}

	for _, p := range ps {
		for _, v := range p.Vertices {
			for i := 0; i < 3; i++ {
				if v.Position[i] < min[i] {
					min[i] = v.Position[i]
				}
				if v.Position[i] > max[i] {
					max[i] = v.Position[i]
				}
			}
		}
	}

	return min, max
}

func (ps BSPPolygons) AllVertices() []*BSPVertex {
	var vertices []*BSPVertex
	for _, p := range ps {
		vertices = append(vertices, p.Vertices...)
	}
	return vertices
}
