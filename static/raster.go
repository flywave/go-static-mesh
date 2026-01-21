package static

import (
	"fmt"
	"io"
	"os"

	"github.com/flywave/go-cog"
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
	if len(g.Data) == 0 {
		return g.NoData
	}
	minHeight := g.Data[0]
	for _, v := range g.Data {
		if v != g.NoData && v < minHeight {
			minHeight = v
		}
	}
	return minHeight
}

func (g *ElevationGrid) GetMaxHeight() float64 {
	if len(g.Data) == 0 {
		return g.NoData
	}
	maxHeight := g.Data[0]
	for _, v := range g.Data {
		if v != g.NoData && v > maxHeight {
			maxHeight = v
		}
	}
	return maxHeight
}

func (g *ElevationGrid) ToTileData() *TileData {
	if g.Data == nil {
		return nil
	}

	tileData := NewTileData([2]uint32{uint32(g.Width), uint32(g.Height)}, BORDER_NONE)
	copy(tileData.Datas, g.Data)
	tileData.Box = g.Bounds
	tileData.Boxsrs = g.Srs
	tileData.NoData = g.NoData

	return tileData
}

func NewElevationGridFromTileData(tileData *TileData, srs geo.Proj) *ElevationGrid {
	if tileData == nil {
		return nil
	}

	data, size, transform := tileData.GetExtend()
	cellSize := [2]float64{transform[1], transform[5]}

	return &ElevationGrid{
		Width:    int(size[0]),
		Height:   int(size[1]),
		Data:     data,
		MinX:     tileData.Box.Min[0],
		MinY:     tileData.Box.Min[1],
		CellSize: cellSize[0],
		NoData:   tileData.NoData,
		Bounds:   tileData.Box,
		Srs:      srs,
	}
}

type RasterProvider interface {
	TileProvider

	GetElevation(lng, lat float64) float64
	GetElevationGrid() *ElevationGrid
	GetElevationTile(coord [3]int) (*ElevationGrid, error)
	GetTileData(coord [3]int) (*TileData, error)
	LoadDEM(r io.Reader, mode RasterDemMode) (*TileData, error)
}

type MapboxRasterProvider struct {
	mode     RasterDemMode
	tileSize uint32
	bounds   vec2d.Rect
	srs      geo.Proj
	noData   float64
}

func NewMapboxRasterProvider(mode RasterDemMode) *MapboxRasterProvider {
	return &MapboxRasterProvider{
		mode:     mode,
		tileSize: 256,
		bounds:   vec2d.Rect{},
		srs:      nil,
		noData:   -9999,
	}
}

func (p *MapboxRasterProvider) SetTileSize(size uint32) {
	p.tileSize = size
}

func (p *MapboxRasterProvider) SetNoData(noData float64) {
	p.noData = noData
}

func (p *MapboxRasterProvider) Attribution() string {
	return ""
}

func (p *MapboxRasterProvider) Grid() *geo.TileGrid {
	return &geo.TileGrid{}
}

func (p *MapboxRasterProvider) Bounds() vec2d.Rect {
	return p.bounds
}

func (p *MapboxRasterProvider) Srs() geo.Proj {
	return p.srs
}

func (p *MapboxRasterProvider) GetElevation(lng, lat float64) float64 {
	return p.noData
}

func (p *MapboxRasterProvider) GetElevationGrid() *ElevationGrid {
	return nil
}

func (p *MapboxRasterProvider) GetElevationTile(coord [3]int) (*ElevationGrid, error) {
	return nil, nil
}

func (p *MapboxRasterProvider) GetTileData(coord [3]int) (*TileData, error) {
	return nil, nil
}

func (p *MapboxRasterProvider) LoadDEM(r io.Reader, mode RasterDemMode) (*TileData, error) {
	demIO := &DemIO{Mode: mode}
	return demIO.Decode(r)
}

type GeoTIFFRasterProvider struct {
	filename string
	grid     *geo.TileGrid
	bounds   vec2d.Rect
	srs      geo.Proj
	reader   *cog.Reader
	noData   float64
}

func NewGeoTIFFRasterProvider(filename string) (*GeoTIFFRasterProvider, error) {
	reader := cog.Read(filename)
	if reader == nil || len(reader.Data) == 0 {
		return nil, fmt.Errorf("failed to read GeoTIFF file: %s", filename)
	}

	p := &GeoTIFFRasterProvider{
		filename: filename,
		reader:   reader,
		noData:   -9999,
	}

	if len(reader.Data) > 0 {
		bounds := reader.GetBounds(0)
		p.bounds = bounds

		if epsgCode, err := reader.GetEPSGCode(0); err == nil && epsgCode != 0 {
			p.srs = geo.NewProj(uint32(epsgCode))
		} else {
			p.srs = geo.NewProj(4326)
		}

		if noData := reader.GetNoData(0); noData != nil {
			p.noData = *noData
		}

		size := reader.GetSize(0)
		tileSize0 := uint32(size[0])
		tileSize1 := uint32(size[1])
		tileSize := [2]uint32{tileSize0, tileSize1}
		p.grid = &geo.TileGrid{
			Srs:      p.srs,
			TileSize: tileSize[:],
		}
	}

	return p, nil
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
	if p.reader == nil {
		return p.noData
	}

	grid := p.GetElevationGrid()
	if grid == nil {
		return p.noData
	}

	return grid.GetElevation(lng, lat)
}

func (p *GeoTIFFRasterProvider) GetElevationGrid() *ElevationGrid {
	if p.reader == nil || len(p.reader.Data) == 0 {
		return nil
	}

	size := p.reader.GetSize(0)
	data := p.reader.Data[0]

	var elevations []float64
	switch v := data.(type) {
	case []uint16:
		elevations = make([]float64, len(v))
		for i, val := range v {
			elevations[i] = float64(val)
		}
	case []int16:
		elevations = make([]float64, len(v))
		for i, val := range v {
			elevations[i] = float64(val)
		}
	case []float32:
		elevations = make([]float64, len(v))
		for i, val := range v {
			elevations[i] = float64(val)
		}
	case []float64:
		elevations = make([]float64, len(v))
		copy(elevations, v)
	default:
		return nil
	}

	pixelSize := p.reader.GetPixelSize(0)
	minX := p.bounds.Min[0]
	minY := p.bounds.Min[1]
	if p.srs != nil {
		transform := p.reader.GetGeoTransform(0)
		minX = transform[3]
		minY = transform[5] + transform[4]*float64(size[1])
	}

	return &ElevationGrid{
		Width:    int(size[0]),
		Height:   int(size[1]),
		Data:     elevations,
		MinX:     minX,
		MinY:     minY,
		CellSize: pixelSize[0],
		NoData:   p.noData,
		Bounds:   p.bounds,
		Srs:      p.srs,
	}
}

func (p *GeoTIFFRasterProvider) GetElevationTile(coord [3]int) (*ElevationGrid, error) {
	return p.GetElevationGrid(), nil
}

func (p *GeoTIFFRasterProvider) GetTileData(coord [3]int) (*TileData, error) {
	grid := p.GetElevationGrid()
	if grid == nil {
		return nil, nil
	}

	return grid.ToTileData(), nil
}

func (p *GeoTIFFRasterProvider) LoadDEM(r io.Reader, mode RasterDemMode) (*TileData, error) {
	file, err := os.CreateTemp("", "geotiff-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(file.Name())
	defer file.Close()

	_, err = io.Copy(file, r)
	if err != nil {
		return nil, err
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	reader := cog.Read(file.Name())
	if reader == nil || len(reader.Data) == 0 {
		return nil, fmt.Errorf("failed to read GeoTIFF data")
	}

	size := reader.GetSize(0)
	data := reader.Data[0]

	var elevations []float64
	switch v := data.(type) {
	case []uint16:
		elevations = make([]float64, len(v))
		for i, val := range v {
			elevations[i] = float64(val)
		}
	case []int16:
		elevations = make([]float64, len(v))
		for i, val := range v {
			elevations[i] = float64(val)
		}
	case []float32:
		elevations = make([]float64, len(v))
		for i, val := range v {
			elevations[i] = float64(val)
		}
	case []float64:
		elevations = make([]float64, len(v))
		copy(elevations, v)
	default:
		return nil, fmt.Errorf("unsupported data type")
	}

	tileData := NewTileData(size, BORDER_NONE)
	tileData.Datas = elevations
	tileData.NoData = -9999

	if bounds := reader.GetBounds(0); bounds.Min[0] != bounds.Max[0] || bounds.Min[1] != bounds.Max[1] {
		tileData.Box = bounds
	}

	if epsgCode, err := reader.GetEPSGCode(0); err == nil && epsgCode != 0 {
		tileData.Boxsrs = geo.NewProj(uint32(epsgCode))
	} else {
		tileData.Boxsrs = geo.NewProj(4326)
	}

	if noData := reader.GetNoData(0); noData != nil {
		tileData.NoData = *noData
	}

	return tileData, nil
}
