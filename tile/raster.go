package tile

import (
	"fmt"
	"io"
	"os"

	gdal "github.com/flywave/flywave-gdal"
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

func (g *ElevationGrid) GetWidth() int {
	return g.Width
}

func (g *ElevationGrid) GetHeight() int {
	return g.Height
}

func (g *ElevationGrid) GetData() []float64 {
	return g.Data
}

func (g *ElevationGrid) GetMinX() float64 {
	return g.MinX
}

func (g *ElevationGrid) GetMinY() float64 {
	return g.MinY
}

func (g *ElevationGrid) GetCellSize() float64 {
	return g.CellSize
}

func (g *ElevationGrid) GetNoData() float64 {
	return g.NoData
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
	tempFile *os.File
	grid     *geo.TileGrid
	bounds   vec2d.Rect
	srs      geo.Proj
	reader   *cog.Reader
	noData   float64
	gdalDS   gdal.Dataset
}

func NewGeoTIFFRasterProvider(filename string) (*GeoTIFFRasterProvider, error) {
	ds, err := gdal.Open(filename, gdal.ReadOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to open GeoTIFF with GDAL: %w", err)
	}

	p := &GeoTIFFRasterProvider{
		filename: filename,
		gdalDS:   ds,
		noData:   -9999,
	}

	boundsArr := ds.Bounds()
	p.bounds = vec2d.Rect{
		Min: vec2d.T{boundsArr[0], boundsArr[1]},
		Max: vec2d.T{boundsArr[2], boundsArr[3]},
	}

	p.srs = geo.NewProj(4326)

	shape := ds.Shape()
	tileSize := [2]uint32{uint32(shape[0]), uint32(shape[1])}
	p.grid = &geo.TileGrid{
		Srs:      p.srs,
		TileSize: tileSize[:],
	}

	nodatavals := ds.Nodatavals()
	if len(nodatavals) > 0 {
		p.noData = nodatavals[0]
	}

	return p, nil
}

func NewGeoTIFFRasterProviderFromReader(r io.Reader) (*GeoTIFFRasterProvider, error) {
	file, err := os.CreateTemp("", "geotiff-raster-*.tif")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	_, err = io.Copy(file, r)
	if err != nil {
		file.Close()
		os.Remove(file.Name())
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}

	err = file.Close()
	if err != nil {
		os.Remove(file.Name())
		return nil, fmt.Errorf("failed to close temp file: %w", err)
	}

	p, err := NewGeoTIFFRasterProvider(file.Name())
	if err != nil {
		os.Remove(file.Name())
		return nil, err
	}

	p.tempFile = file
	return p, nil
}

func (p *GeoTIFFRasterProvider) Close() error {
	if !p.gdalDS.IsClosed() {
		p.gdalDS.Close()
	}

	if p.tempFile != nil {
		err := os.Remove(p.tempFile.Name())
		p.tempFile = nil
		return err
	}

	return nil
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
	if p.gdalDS.IsClosed() {
		return nil
	}

	shape := p.gdalDS.Shape()
	width := shape[0]
	height := shape[1]

	buffer := make([]float64, width*height)
	err := p.gdalDS.ReadRaster(0, 0, width, height, buffer, width, height, []int{1}, gdal.Float64, 0, 0, 0, gdal.GRA_NearestNeighbour)
	if err != nil {
		return nil
	}

	geoTransform := p.gdalDS.GeoTransform()
	cellSize := geoTransform[1]

	return &ElevationGrid{
		Width:    width,
		Height:   height,
		Data:     buffer,
		MinX:     p.bounds.Min[0],
		MinY:     p.bounds.Min[1],
		CellSize: cellSize,
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

	ds, err := gdal.Open(file.Name(), gdal.ReadOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to open GeoTIFF: %w", err)
	}
	defer ds.Close()

	shape := ds.Shape()
	width := shape[0]
	height := shape[1]

	buffer := make([]float64, width*height)
	err = ds.ReadRaster(0, 0, width, height, buffer, width, height, []int{1}, gdal.Float64, 0, 0, 0, gdal.GRA_NearestNeighbour)
	if err != nil {
		return nil, fmt.Errorf("failed to read raster data: %w", err)
	}

	tileData := NewTileData([2]uint32{uint32(width), uint32(height)}, BORDER_NONE)
	tileData.Datas = buffer
	tileData.NoData = p.noData

	boundsArr := ds.Bounds()
	tileData.Box = vec2d.Rect{
		Min: vec2d.T{boundsArr[0], boundsArr[1]},
		Max: vec2d.T{boundsArr[2], boundsArr[3]},
	}

	tileData.Boxsrs = p.srs

	nodatavals := ds.Nodatavals()
	if len(nodatavals) > 0 {
		tileData.NoData = nodatavals[0]
	}

	return tileData, nil
}
