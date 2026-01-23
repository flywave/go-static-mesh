package extruder

import (
	"fmt"

	"github.com/flywave/go-static-mesh/draw"
	"github.com/flywave/go-static-mesh/mesh"
	bsp "github.com/flywave/go-static-mesh/mesh/bsp"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

type BooleanExtruder struct {
	baseMesh  *mesh.Mesh
	operation bsp.BSPOperation
}

func NewBooleanExtruder() *BooleanExtruder {
	return &BooleanExtruder{
		baseMesh:  &mesh.Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}},
		operation: bsp.BSPOperationUnion,
	}
}

func (e *BooleanExtruder) SetOperation(operation bsp.BSPOperation) {
	e.operation = operation
}

func (e *BooleanExtruder) SetBaseMesh(mesh *mesh.Mesh) {
	e.baseMesh = mesh
}

func (e *BooleanExtruder) ExtrudeAndApply(path *draw.Path, height float64) error {
	tempMesh := &mesh.Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}

	extruder := NewPathExtruder()
	options := &ExtrudeOptions{
		Resolution: 1.0,
		Radius:     path.Weight / 2.0,
		Segments:   16,
	}

	err := extruder.ExtrudeToMeshWithResolution(path, tempMesh, height, options)
	if err != nil {
		return fmt.Errorf("failed to extrude path: %w", err)
	}

	return e.applyBoolean(tempMesh)
}

func (e *BooleanExtruder) ExtrudeAreaAndApply(area *draw.Area, height float64) error {
	tempMesh := &mesh.Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}

	extruder := NewAreaExtruder()
	err := extruder.ExtrudeToMesh(area, tempMesh, height)
	if err != nil {
		return fmt.Errorf("failed to extrude area: %w", err)
	}

	return e.applyBoolean(tempMesh)
}

func (e *BooleanExtruder) ExtrudeAreaAndApplyWithResolution(area *draw.Area, height, resolution float64) error {
	tempMesh := &mesh.Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}

	extruder := NewAreaExtruder()
	err := extruder.ExtrudeToMeshWithResolution(area, tempMesh, height, resolution)
	if err != nil {
		return fmt.Errorf("failed to extrude area: %w", err)
	}

	return e.applyBoolean(tempMesh)
}

func (e *BooleanExtruder) applyBoolean(newMesh *mesh.Mesh) error {
	if len(e.baseMesh.Indices) == 0 {
		e.baseMesh.Vertices = append(e.baseMesh.Vertices, newMesh.Vertices...)
		e.baseMesh.Indices = append(e.baseMesh.Indices, newMesh.Indices...)
		return nil
	}

	resultMesh, err := bsp.PerformBoolean(e.baseMesh, newMesh, e.operation)
	if err != nil {
		return fmt.Errorf("failed to perform boolean operation: %w", err)
	}

	e.baseMesh = resultMesh
	return nil
}

func (e *BooleanExtruder) GetResult() *mesh.Mesh {
	return e.baseMesh
}

func (e *BooleanExtruder) Reset() {
	e.baseMesh = &mesh.Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}
	e.operation = bsp.BSPOperationUnion
}

type MultiExtrusionOptions struct {
	Operation  bsp.BSPOperation
	BaseMesh   *mesh.Mesh
	Height     float64
	Resolution float64
}

func ExtrudeMultipleWithBoolean(geoData []draw.MapObject, options *MultiExtrusionOptions) (*mesh.Mesh, error) {
	if options == nil {
		options = &MultiExtrusionOptions{
			Operation:  bsp.BSPOperationUnion,
			BaseMesh:   &mesh.Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}},
			Height:     10.0,
			Resolution: 1.0,
		}
	}

	if options.BaseMesh == nil {
		options.BaseMesh = &mesh.Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}
	}

	for _, geoObj := range geoData {
		var tempMesh *mesh.Mesh
		var err error

		tempMesh = &mesh.Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}
		err = geoObj.ExtrudeToMesh(tempMesh, options.Height)

		if err != nil {
			return nil, fmt.Errorf("failed to extrude geo object: %w", err)
		}

		if len(options.BaseMesh.Indices) == 0 {
			options.BaseMesh = tempMesh
		} else {
			resultMesh, err := bsp.PerformBoolean(options.BaseMesh, tempMesh, options.Operation)
			if err != nil {
				return nil, fmt.Errorf("failed to perform boolean operation: %w", err)
			}
			options.BaseMesh = resultMesh
		}
	}

	if len(options.BaseMesh.Vertices) > 0 {
		options.BaseMesh.CalculateNormals()
	}

	return options.BaseMesh, nil
}

func ExtrudeAndSubtract(base *mesh.Mesh, cutouts []draw.MapObject, height float64) (*mesh.Mesh, error) {
	result := &mesh.Mesh{
		Vertices: make([]vec3d.T, len(base.Vertices)),
		Indices:  make([]uint32, len(base.Indices)),
	}
	copy(result.Vertices, base.Vertices)
	copy(result.Indices, base.Indices)

	areaExtruder := NewAreaExtruder()

	for _, cutout := range cutouts {
		if area, ok := cutout.(*draw.Area); ok {
			cutoutMesh := &mesh.Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}
			err := areaExtruder.ExtrudeToMesh(area, cutoutMesh, height)
			if err != nil {
				return nil, fmt.Errorf("failed to extrude cutout: %w", err)
			}

			resultMesh, err := bsp.PerformBoolean(result, cutoutMesh, bsp.BSPOperationSubtraction)
			if err != nil {
				return nil, fmt.Errorf("failed to perform subtraction: %w", err)
			}
			result = resultMesh
		}
	}

	return result, nil
}

func ExtrudeAndIntersect(meshA *mesh.Mesh, meshB *mesh.Mesh, height float64) (*mesh.Mesh, error) {
	if len(meshA.Vertices) == 0 || len(meshB.Vertices) == 0 {
		return &mesh.Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}, nil
	}

	result, err := bsp.PerformBoolean(meshA, meshB, bsp.BSPOperationIntersection)
	if err != nil {
		return nil, fmt.Errorf("failed to perform intersection: %w", err)
	}

	return result, nil
}
