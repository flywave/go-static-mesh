package static

import (
	"image"

	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

type ImageryProvider interface {
	TileProvider
	TileFetcher

	GetImageTile(coord [3]int) (image.Image, error)
	GetImageBounds() vec2d.Rect
}

type GenericTileProvider struct {
	config *TileProviderConfig
	grid   *geo.TileGrid
	bounds vec2d.Rect
	srs    geo.Proj
}

func NewTileProvider(url string, mode TileProviderMode) *GenericTileProvider {
	grid := &geo.TileGrid{}
	return &GenericTileProvider{
		config: &TileProviderConfig{
			Mode:    mode,
			URL:     url,
			Headers: make(map[string]string),
		},
		grid:   grid,
		bounds: vec2d.Rect{Min: vec2d.T{-85.0, -180.0}, Max: vec2d.T{85.0, 180.0}},
		srs:    geo.NewProj(4326),
	}
}

func NewTileProviderWithConfig(config *TileProviderConfig) *GenericTileProvider {
	grid := &geo.TileGrid{}
	return &GenericTileProvider{
		config: config,
		grid:   grid,
		bounds: vec2d.Rect{Min: vec2d.T{-85.0, -180.0}, Max: vec2d.T{85.0, 180.0}},
		srs:    grid.Srs,
	}
}

func (p *GenericTileProvider) Attribution() string {
	if p.config.Attribution != "" {
		return p.config.Attribution
	}
	return ""
}

func (p *GenericTileProvider) Grid() *geo.TileGrid {
	if p.grid != nil {
		return p.grid
	}
	return p.config.Grid
}

func (p *GenericTileProvider) Bounds() vec2d.Rect {
	return p.bounds
}

func (p *GenericTileProvider) Srs() geo.Proj {
	return p.srs
}

func (p *GenericTileProvider) Fetch(coord [3]int) ([]byte, error) {
	return nil, nil
}

func (p *GenericTileProvider) GetImageTile(coord [3]int) (image.Image, error) {
	return nil, nil
}

func (p *GenericTileProvider) GetImageBounds() vec2d.Rect {
	return p.bounds
}
