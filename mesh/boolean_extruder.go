package mesh

import (
	"fmt"

	"github.com/flywave/go-static-mesh/draw"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

type BooleanExtruder struct {
	baseMesh  *Mesh
	operation BSPOperation
}

func NewBooleanExtruder() *BooleanExtruder {
	return &BooleanExtruder{
		baseMesh:  &Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}},
		operation: BSPOperationUnion,
	}
}

func (e *BooleanExtruder) SetOperation(operation BSPOperation) {
	e.operation = operation
}

func (e *BooleanExtruder) SetBaseMesh(mesh *Mesh) {
	e.baseMesh = mesh
}

func (e *BooleanExtruder) ExtrudeAndApply(path *draw.Path, height float64) error {
	tempMesh := &Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}

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
	tempMesh := &Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}

	extruder := NewAreaExtruder()
	err := extruder.ExtrudeToMesh(area, tempMesh, height)
	if err != nil {
		return fmt.Errorf("failed to extrude area: %w", err)
	}

	return e.applyBoolean(tempMesh)
}

func (e *BooleanExtruder) ExtrudeAreaAndApplyWithResolution(area *draw.Area, height, resolution float64) error {
	tempMesh := &Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}

	extruder := NewAreaExtruder()
	err := extruder.ExtrudeToMeshWithResolution(area, tempMesh, height, resolution)
	if err != nil {
		return fmt.Errorf("failed to extrude area: %w", err)
	}

	return e.applyBoolean(tempMesh)
}

func (e *BooleanExtruder) applyBoolean(newMesh *Mesh) error {
	if len(e.baseMesh.Indices) == 0 {
		e.baseMesh.Vertices = append(e.baseMesh.Vertices, newMesh.Vertices...)
		e.baseMesh.Indices = append(e.baseMesh.Indices, newMesh.Indices...)
		return nil
	}

	resultMesh, err := PerformBoolean(e.baseMesh, newMesh, e.operation)
	if err != nil {
		return fmt.Errorf("failed to perform boolean operation: %w", err)
	}

	e.baseMesh = resultMesh
	return nil
}

func (e *BooleanExtruder) GetResult() *Mesh {
	return e.baseMesh
}

func (e *BooleanExtruder) Reset() {
	e.baseMesh = &Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}
	e.operation = BSPOperationUnion
}

type MultiExtrusionOptions struct {
	Operation  BSPOperation
	BaseMesh   *Mesh
	Height     float64
	Resolution float64
}

func ExtrudeMultipleWithBoolean(geoData []draw.MapObject, options *MultiExtrusionOptions) (*Mesh, error) {
	if options == nil {
		options = &MultiExtrusionOptions{
			Operation:  BSPOperationUnion,
			BaseMesh:   &Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}},
			Height:     10.0,
			Resolution: 1.0,
		}
	}

	if options.BaseMesh == nil {
		options.BaseMesh = &Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}
	}

	pathExtruder := NewPathExtruder()
	areaExtruder := NewAreaExtruder()

	for _, geoObj := range geoData {
		var tempMesh *Mesh
		var err error

		if meshObj, ok := geoObj.(MeshObject); ok {
			tempMesh = &Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}
			err = meshObj.ExtrudeToMesh(tempMesh, options.Height)
		} else if path, ok := geoObj.(*draw.Path); ok {
			tempMesh = &Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}
			extruderOptions := &ExtrudeOptions{
				Radius:   path.Weight / 2.0,
				Segments: 16,
			}
			err = pathExtruder.ExtrudeToMeshWithResolution(path, tempMesh, options.Height, extruderOptions)
		} else if area, ok := geoObj.(*draw.Area); ok {
			tempMesh = &Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}
			err = areaExtruder.ExtrudeToMeshWithResolution(area, tempMesh, options.Height, options.Resolution)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to extrude geo object: %w", err)
		}

		if len(options.BaseMesh.Indices) == 0 {
			options.BaseMesh = tempMesh
		} else {
			resultMesh, err := PerformBoolean(options.BaseMesh, tempMesh, options.Operation)
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

func ExtrudeAndSubtract(base *Mesh, cutouts []draw.MapObject, height float64) (*Mesh, error) {
	result := &Mesh{
		Vertices: make([]vec3d.T, len(base.Vertices)),
		Indices:  make([]uint32, len(base.Indices)),
	}
	copy(result.Vertices, base.Vertices)
	copy(result.Indices, base.Indices)

	areaExtruder := NewAreaExtruder()

	for _, cutout := range cutouts {
		if area, ok := cutout.(*draw.Area); ok {
			cutoutMesh := &Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}
			err := areaExtruder.ExtrudeToMesh(area, cutoutMesh, height)
			if err != nil {
				return nil, fmt.Errorf("failed to extrude cutout: %w", err)
			}

			resultMesh, err := PerformBoolean(result, cutoutMesh, BSPOperationSubtraction)
			if err != nil {
				return nil, fmt.Errorf("failed to perform subtraction: %w", err)
			}
			result = resultMesh
		}
	}

	return result, nil
}

func ExtrudeAndIntersect(meshA *Mesh, meshB *Mesh, height float64) (*Mesh, error) {
	if len(meshA.Vertices) == 0 || len(meshB.Vertices) == 0 {
		return &Mesh{Vertices: []vec3d.T{}, Indices: []uint32{}}, nil
	}

	result, err := PerformBoolean(meshA, meshB, BSPOperationIntersection)
	if err != nil {
		return nil, fmt.Errorf("failed to perform intersection: %w", err)
	}

	return result, nil
}
