package static

import (
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

type Model3D struct {
	ID          string
	Name        string
	Position    vec2d.T
	Elevation   float64
	Rotation    vec3d.T
	Scale       vec3d.T
	Mesh        *TinMesh
	Format      string
	TextureData []byte
	TextureType string
	Metadata    map[string]interface{}
}

type Model3DProvider interface {
	GetModels(bounds vec2d.Rect) ([]Model3D, error)
	GetModel(id string) (*Model3D, error)
	LoadModelFromFile(id, filepath string) error
	LoadModelFromData(id string, data []byte, format string) error
}
