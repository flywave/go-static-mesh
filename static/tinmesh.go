package static

import (
	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

type TinMesh struct {
	Vertices    []vec3d.T
	Indices     []uint32
	EdgeIndices []uint32
	NorthEdge   []uint16
	SouthEdge   []uint16
	WestEdge    []uint16
	EastEdge    []uint16
	MinHeight   float64
	MaxHeight   float64
	Bounds      vec2d.Rect
	Srs         geo.Proj
}

func (m *TinMesh) GetVertices() []vec3d.T {
	return m.Vertices
}

func (m *TinMesh) GetIndices() []uint32 {
	return m.Indices
}

func (m *TinMesh) GetMinHeight() float64 {
	return m.MinHeight
}

func (m *TinMesh) GetBounds() vec2d.Rect {
	return m.Bounds
}

func (m *TinMesh) GetSrs() geo.Proj {
	return m.Srs
}

func (m *TinMesh) GetElevation(x, y float64) float64 {
	return 0
}

func (m *TinMesh) GetNormal(x, y float64) vec3d.T {
	return vec3d.T{0, 0, 1}
}

type TinMeshProvider interface {
	TileProvider

	GetMeshTile(coord [3]int) (*TinMesh, error)
	GetMesh() (*TinMesh, error)
	GetHeight(lng, lat float64) float64
	GetNormal(lng, lat float64) vec3d.T
}

type CesiumQuantizedMeshProvider struct {
	url    string
	config interface{}
	grid   *geo.TileGrid
	bounds vec2d.Rect
	srs    geo.Proj
}

func NewCesiumQuantizedMeshProvider(url string) *CesiumQuantizedMeshProvider {
	grid := &geo.TileGrid{}
	return &CesiumQuantizedMeshProvider{
		url:    url,
		grid:   grid,
		bounds: vec2d.Rect{Min: vec2d.T{-85.0, -180.0}, Max: vec2d.T{85.0, 180.0}},
		srs:    geo.NewProj(4326),
	}
}

func NewCesiumQuantizedMeshProviderWithConfig(url string, config interface{}) *CesiumQuantizedMeshProvider {
	grid := &geo.TileGrid{}
	return &CesiumQuantizedMeshProvider{
		url:    url,
		config: config,
		grid:   grid,
		bounds: vec2d.Rect{Min: vec2d.T{-85.0, -180.0}, Max: vec2d.T{85.0, 180.0}},
		srs:    geo.NewProj(4326),
	}
}

func (p *CesiumQuantizedMeshProvider) Attribution() string {
	return ""
}

func (p *CesiumQuantizedMeshProvider) Grid() *geo.TileGrid {
	return p.grid
}

func (p *CesiumQuantizedMeshProvider) Bounds() vec2d.Rect {
	return p.bounds
}

func (p *CesiumQuantizedMeshProvider) Srs() geo.Proj {
	return p.srs
}

func (p *CesiumQuantizedMeshProvider) GetMeshTile(coord [3]int) (*TinMesh, error) {
	return nil, nil
}

func (p *CesiumQuantizedMeshProvider) GetHeight(lng, lat float64) float64 {
	return 0
}

func (p *CesiumQuantizedMeshProvider) GetNormal(lng, lat float64) vec3d.T {
	return vec3d.T{0, 0, 1}
}

func (p *CesiumQuantizedMeshProvider) GetMesh() (*TinMesh, error) {
	return nil, nil
}
