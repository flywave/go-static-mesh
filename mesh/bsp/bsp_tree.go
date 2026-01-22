package bsp

import (
	"math"

	"github.com/flywave/go-static-mesh/mesh"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

type BSPOperation int

const (
	BSPOperationUnion BSPOperation = iota
	BSPOperationIntersection
	BSPOperationSubtraction
)

type BSPNode struct {
	Plane    *BSPPlane
	Polygons BSPPolygons
	Front    *BSPNode
	Back     *BSPNode
}

type BSPTree struct {
	Root *BSPNode
}

func NewBSPTree(polygons BSPPolygons) *BSPTree {
	return &BSPTree{
		Root: buildBSPNode(polygons),
	}
}

func buildBSPNode(polygons BSPPolygons) *BSPNode {
	if len(polygons) == 0 {
		return nil
	}

	node := &BSPNode{}

	if len(polygons) == 1 {
		node.Polygons = polygons
		node.Plane = polygons[0].Plane
		return node
	}

	node.Plane = selectSplitPlane(polygons)

	var frontPolygons, backPolygons, coplanarFrontPolygons, coplanarBackPolygons BSPPolygons

	for _, poly := range polygons {
		var front, back, coplanarFront, coplanarBack BSPPolygons
		node.Plane.SplitPolygon(poly, &coplanarFront, &coplanarBack, &front, &back)

		frontPolygons = append(frontPolygons, front...)
		backPolygons = append(backPolygons, back...)
		coplanarFrontPolygons = append(coplanarFrontPolygons, coplanarFront...)
		coplanarBackPolygons = append(coplanarBackPolygons, coplanarBack...)
	}

	node.Polygons = coplanarFrontPolygons

	if len(frontPolygons) > 0 {
		node.Front = buildBSPNode(frontPolygons)
	}

	if len(backPolygons) > 0 {
		node.Back = buildBSPNode(backPolygons)
	}

	return node
}

func selectSplitPlane(polygons BSPPolygons) *BSPPlane {
	if len(polygons) == 0 {
		return nil
	}

	minSplits := math.MaxInt32
	bestPlane := polygons[0].Plane

	for _, poly := range polygons {
		if poly.Plane == nil {
			continue
		}

		splits := 0
		for _, other := range polygons {
			if other == poly || other.Plane == nil {
				continue
			}

			var front, back BSPPolygons
			poly.Plane.SplitPolygon(other, &BSPPolygons{}, &BSPPolygons{}, &front, &back)

			if len(front) > 0 && len(back) > 0 {
				splits++
			}
		}

		if splits < minSplits {
			minSplits = splits
			bestPlane = poly.Plane
		}
	}

	return bestPlane
}

func (tree *BSPTree) ClipTo(clipper *BSPTree) {
	if tree.Root == nil {
		return
	}
	tree.Root = clipPolygons(tree.Root, clipper.Root, false)
}

func clipPolygons(node, clipper *BSPNode, isBack bool) *BSPNode {
	if node == nil {
		return nil
	}

	result := &BSPNode{
		Plane:    node.Plane,
		Polygons: BSPPolygons{},
	}

	if len(node.Polygons) > 0 {
		for _, poly := range node.Polygons {
			if classifyPolygons(clipper, poly, isBack) {
				result.Polygons = append(result.Polygons, poly)
			}
		}
	}

	result.Front = clipPolygons(node.Front, clipper, isBack)
	result.Back = clipPolygons(node.Back, clipper, isBack)

	return result
}

func classifyPolygons(clipper *BSPNode, poly *BSPPolygon, isBack bool) bool {
	if clipper == nil {
		return !isBack
	}

	plane := clipper.Plane
	if plane == nil {
		return classifyPolygons(clipper.Front, poly, isBack) && classifyPolygons(clipper.Back, poly, isBack)
	}

	if poly.Plane == nil {
		poly.Plane = PlaneFromBSPVertices(poly.Vertices)
	}

	var frontPolygons, backPolygons BSPPolygons
	plane.SplitPolygon(poly, &BSPPolygons{}, &BSPPolygons{}, &frontPolygons, &backPolygons)

	if len(frontPolygons) > 0 && len(backPolygons) > 0 {
		return false
	}

	if len(frontPolygons) > 0 {
		return classifyPolygons(clipper.Front, poly, isBack)
	}

	if len(backPolygons) > 0 {
		return classifyPolygons(clipper.Back, poly, isBack)
	}

	planeDist := plane.DistanceToPoint(poly.Vertices[0].Position)
	frontSide := planeDist >= 0

	if isBack {
		return !frontSide
	}
	return frontSide
}

func (tree *BSPTree) Inverse() {
	if tree.Root == nil {
		return
	}
	inverseNode(tree.Root)
}

func inverseNode(node *BSPNode) {
	if node == nil {
		return
	}

	for _, poly := range node.Polygons {
		poly.Flip()
	}

	node.Front, node.Back = node.Back, node.Front

	inverseNode(node.Front)
	inverseNode(node.Back)
}

func (tree *BSPTree) CollectPolygons() BSPPolygons {
	if tree.Root == nil {
		return BSPPolygons{}
	}
	return collectPolygons(tree.Root)
}

func collectPolygons(node *BSPNode) BSPPolygons {
	if node == nil {
		return BSPPolygons{}
	}

	result := BSPPolygons{}
	result = append(result, node.Polygons...)
	result = append(result, collectPolygons(node.Front)...)
	result = append(result, collectPolygons(node.Back)...)

	return result
}

func PerformBoolean(meshA, meshB *mesh.Mesh, operation BSPOperation) (*mesh.Mesh, error) {
	polygonsA := meshToBSPPolygons(meshA)
	polygonsB := meshToBSPPolygons(meshB)

	if len(polygonsA) == 0 {
		return bspPolygonsToMesh(polygonsB), nil
	}

	if len(polygonsB) == 0 {
		return bspPolygonsToMesh(polygonsA), nil
	}

	treeA := NewBSPTree(polygonsA)
	treeB := NewBSPTree(polygonsB)

	switch operation {
	case BSPOperationUnion:
		treeA.ClipTo(treeB)
		treeB.ClipTo(treeA)
		treeB.Inverse()
		treeB.ClipTo(treeA)
		treeB.Inverse()
		resultPolygons := treeA.CollectPolygons()
		resultPolygons = append(resultPolygons, treeB.CollectPolygons()...)
		return bspPolygonsToMesh(resultPolygons), nil

	case BSPOperationIntersection:
		treeA.Inverse()
		treeA.ClipTo(treeB)
		treeB.ClipTo(treeA)
		treeA.Inverse()
		treeB.ClipTo(treeA)
		resultPolygons := treeA.CollectPolygons()
		resultPolygons = append(resultPolygons, treeB.CollectPolygons()...)
		return bspPolygonsToMesh(resultPolygons), nil

	case BSPOperationSubtraction:
		treeA.Inverse()
		treeA.ClipTo(treeB)
		treeB.ClipTo(treeA)
		treeA.Inverse()
		resultPolygons := treeA.CollectPolygons()
		return bspPolygonsToMesh(resultPolygons), nil
	}

	return &mesh.Mesh{}, nil
}

func meshToBSPPolygons(mesh *mesh.Mesh) BSPPolygons {
	polygons := make(BSPPolygons, 0, len(mesh.Indices)/3)

	for i := 0; i < len(mesh.Indices); i += 3 {
		if i+2 >= len(mesh.Indices) {
			continue
		}

		idx0 := mesh.Indices[i]
		idx1 := mesh.Indices[i+1]
		idx2 := mesh.Indices[i+2]

		if idx0 >= uint32(len(mesh.Vertices)) ||
			idx1 >= uint32(len(mesh.Vertices)) ||
			idx2 >= uint32(len(mesh.Vertices)) {
			continue
		}

		v0 := mesh.Vertices[idx0]
		v1 := mesh.Vertices[idx1]
		v2 := mesh.Vertices[idx2]

		normal := calculateTriangleNormal(&v0, &v1, &v2)

		poly := &BSPPolygon{
			Vertices: []*BSPVertex{
				{
					Position: v0,
					Normal:   normal,
					Color:    [3]uint8{255, 255, 255},
					Alpha:    255,
					UV:       vec2d.T{0, 0},
				},
				{
					Position: v1,
					Normal:   normal,
					Color:    [3]uint8{255, 255, 255},
					Alpha:    255,
					UV:       vec2d.T{0, 0},
				},
				{
					Position: v2,
					Normal:   normal,
					Color:    [3]uint8{255, 255, 255},
					Alpha:    255,
					UV:       vec2d.T{0, 0},
				},
			},
		}

		poly.Plane = PlaneFromBSPVertices(poly.Vertices)

		polygons = append(polygons, poly)
	}

	return polygons
}

func bspPolygonsToMesh(polygons BSPPolygons) *mesh.Mesh {
	mesh := &mesh.Mesh{
		Vertices: []vec3d.T{},
		Indices:  []uint32{},
	}

	triangles := BSPPolygons{}
	for _, poly := range polygons {
		tri := poly.Triangles()
		if tri != nil {
			triangles = append(triangles, tri...)
		}
	}

	for _, poly := range triangles {
		if len(poly.Vertices) != 3 {
			continue
		}

		baseIdx := uint32(len(mesh.Vertices))

		for _, v := range poly.Vertices {
			mesh.Vertices = append(mesh.Vertices, v.Position)
		}

		mesh.Indices = append(mesh.Indices,
			baseIdx, baseIdx+1, baseIdx+2,
		)
	}

	if len(mesh.Vertices) > 0 {
		mesh.CalculateNormals()
	}

	return mesh
}

func calculateTriangleNormal(v0, v1, v2 *vec3d.T) vec3d.T {
	edge1 := vec3d.T{
		(*v1)[0] - (*v0)[0],
		(*v1)[1] - (*v0)[1],
		(*v1)[2] - (*v0)[2],
	}
	edge2 := vec3d.T{
		(*v2)[0] - (*v0)[0],
		(*v2)[1] - (*v0)[1],
		(*v2)[2] - (*v0)[2],
	}

	normal := vec3d.T{
		edge1[1]*edge2[2] - edge1[2]*edge2[1],
		edge1[2]*edge2[0] - edge1[0]*edge2[2],
		edge1[0]*edge2[1] - edge1[1]*edge2[0],
	}

	length := math.Sqrt(normal[0]*normal[0] + normal[1]*normal[1] + normal[2]*normal[2])
	if length > 0 {
		normal[0] /= length
		normal[1] /= length
		normal[2] /= length
	}

	return normal
}
