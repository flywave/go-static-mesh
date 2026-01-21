package static

import (
	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

type ElevationGrid struct {
	Width    int
	Height   int
	Data     []float64
	MinX     float64
	MinY     float64
	CellSize float64
	NoData   float64
	Bounds   vec2d.Rect
	Srs      geo.Proj
}

func (g *ElevationGrid) GetElevation(x, y float64) float64 {
	gridX := int((x - g.MinX) / g.CellSize)
	gridY := int((y - g.MinY) / g.CellSize)

	if gridX < 0 || gridX >= g.Width || gridY < 0 || gridY >= g.Height {
		return g.NoData
	}

	return g.Data[gridY*g.Width+gridX]
}

func (g *ElevationGrid) GetElevationByGrid(gridX, gridY int) float64 {
	if gridX < 0 || gridX >= g.Width || gridY < 0 || gridY >= g.Height {
		return g.NoData
	}

	return g.Data[gridY*g.Width+gridX]
}

func (g *ElevationGrid) GetBounds() vec2d.Rect {
	return g.Bounds
}

func (g *ElevationGrid) GetMinHeight() float64 {
	minHeight := g.Data[0]
	for _, v := range g.Data {
		if v != g.NoData && v < minHeight {
			minHeight = v
		}
	}
	return minHeight
}

func (g *ElevationGrid) GetMaxHeight() float64 {
	maxHeight := g.Data[0]
	for _, v := range g.Data {
		if v != g.NoData && v > maxHeight {
			maxHeight = v
		}
	}
	return maxHeight
}

type RasterProvider interface {
	TileProvider

	GetElevation(lng, lat float64) float64
	GetElevationGrid() *ElevationGrid
	GetElevationTile(coord [3]int) (*ElevationGrid, error)
}

type GeoTIFFRasterProvider struct {
	filename string
	grid     *geo.TileGrid
	bounds   vec2d.Rect
	srs      geo.Proj
}

func NewGeoTIFFRasterProvider(filename string) (*GeoTIFFRasterProvider, error) {
	return &GeoTIFFRasterProvider{
		filename: filename,
	}, nil
}

func (p *GeoTIFFRasterProvider) Attribution() string {
	return ""
}

func (p *GeoTIFFRasterProvider) Grid() *geo.TileGrid {
	return p.grid
}

func (p *GeoTIFFRasterProvider) Bounds() vec2d.Rect {
	return p.bounds
}

func (p *GeoTIFFRasterProvider) Srs() geo.Proj {
	return p.srs
}

func (p *GeoTIFFRasterProvider) GetElevation(lng, lat float64) float64 {
	return 0
}

func (p *GeoTIFFRasterProvider) GetElevationGrid() *ElevationGrid {
	return nil
}

func (p *GeoTIFFRasterProvider) GetElevationTile(coord [3]int) (*ElevationGrid, error) {
	return nil, nil
}
