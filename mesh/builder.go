package mesh

import (
	"fmt"
	"github.com/flywave/go-geo"
	draw "github.com/flywave/go-static-mesh/draw"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
	"image"
	"math"
	"time"
)

type Tile struct {
	Coord  [3]int
	Zoom   int
	Data   []byte
	Bounds vec2d.Rect
	Width  int
	Height int
}

type TileErrorHandler struct {
	SkipMissing bool
	UseNoData   bool
	NoDataValue float64
	MaxRetries  int
	Timeout     time.Duration
}

type Builder struct {
	tinGenerator         TINGenerator
	rasterProvider       interface{}
	tinMeshProvider      interface{}
	imageryProvider      interface{}
	model3DProvider      interface{}
	geoData              []draw.MapObject
	bounds               vec2d.Rect
	srs                  geo.Proj
	zoom                 *int
	autoZoomMin          int
	autoZoomMax          int
	resolution           float64
	verticalExaggeration float64
	baseElevation        float64
	extrudeGeoData       bool
	geoDataHeight        float64
	closeMesh            bool
	baseThickness        float64
	tileErrorHandler     TileErrorHandler
}

func NewBuilder() *Builder {
	return &Builder{
		geoData:              []draw.MapObject{},
		autoZoomMin:          13,
		autoZoomMax:          17,
		resolution:           1.0,
		verticalExaggeration: 1.0,
		tileErrorHandler: TileErrorHandler{
			SkipMissing: true,
			MaxRetries:  3,
			Timeout:     30 * time.Second,
		},
	}
}

func (b *Builder) SetTINGenerator(generator TINGenerator) {
	b.tinGenerator = generator
}

func (b *Builder) SetRasterProvider(provider interface{}) {
	b.rasterProvider = provider
	b.tinMeshProvider = nil
}

func (b *Builder) SetTinMeshProvider(provider interface{}) {
	b.tinMeshProvider = provider
	b.rasterProvider = nil
}

func (b *Builder) AddImageryProvider(provider interface{}) {
	b.imageryProvider = provider
}

func (b *Builder) AddGeoData(obj draw.MapObject) {
	b.geoData = append(b.geoData, obj)
}

func (b *Builder) SetBounds(bounds vec2d.Rect, srs geo.Proj) {
	b.bounds = bounds
	b.srs = srs
}

func (b *Builder) SetZoom(zoom int) {
	b.zoom = &zoom
}

func (b *Builder) SetAutoZoomRange(minZoom, maxZoom int) {
	b.autoZoomMin = minZoom
	b.autoZoomMax = maxZoom
}

func (b *Builder) SetVerticalExaggeration(scale float64) {
	b.verticalExaggeration = scale
}

func (b *Builder) SetBaseElevation(elevation float64) {
	b.baseElevation = elevation
}

func (b *Builder) SetExtrudeGeoData(extrude bool, height float64) {
	b.extrudeGeoData = extrude
	b.geoDataHeight = height
}

func (b *Builder) SetCloseMesh(close bool, thickness float64) {
	b.closeMesh = close
	b.baseThickness = thickness
}

func (b *Builder) SetTileErrorHandler(handler TileErrorHandler) {
	b.tileErrorHandler = handler
}

func (b *Builder) GenerateTexture() (*Mesh, error) {
	if b.imageryProvider == nil {
		return nil, fmt.Errorf("imagery provider not set")
	}

	if b.bounds.Min[0] >= b.bounds.Max[0] || b.bounds.Min[1] >= b.bounds.Max[1] {
		return nil, fmt.Errorf("bounds not set properly")
	}

	zoom, err := b.determineZoom()
	if err != nil {
		return nil, fmt.Errorf("failed to determine zoom: %w", err)
	}

	provider, ok := b.imageryProvider.(interface {
		GetImageTile(coord [3]int) (image.Image, error)
	})
	if !ok {
		return nil, fmt.Errorf("imagery provider does not support image tiles")
	}

	fetcher := NewTileFetcher(provider)
	tiles, err := fetcher.FetchTiles(b.bounds, zoom, b.srs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tiles: %w", err)
	}

	if len(tiles) == 0 {
		return nil, fmt.Errorf("no tiles fetched")
	}

	tileSize := 256
	for _, tile := range tiles {
		if tile.Image != nil {
			tileSize = tile.Image.Bounds().Dx()
			break
		}
	}

	generator := NewTextureGenerator(tileSize)
	texture, err := generator.Generate(tiles, b.bounds, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate texture: %w", err)
	}

	texture = generator.DrawGeoObjects(texture, b.geoData, b.bounds, b.srs)

	textureBounds := generator.CalculateTextureBounds([]vec3d.T{
		{b.bounds.Min[0], b.bounds.Min[1], 0},
		{b.bounds.Max[0], b.bounds.Min[1], 0},
		{b.bounds.Max[0], b.bounds.Max[1], 0},
		{b.bounds.Min[0], b.bounds.Max[1], 0},
	})

	mesh := &Mesh{
		Texture: texture,
		Bounds:  b.bounds,
		Srs:     b.srs,
	}

	if texture != nil {
		mesh.UVs = generator.CalculateUVs([]vec3d.T{
			{b.bounds.Min[0], b.bounds.Min[1], 0},
			{b.bounds.Max[0], b.bounds.Min[1], 0},
			{b.bounds.Max[0], b.bounds.Max[1], 0},
			{b.bounds.Min[0], b.bounds.Max[1], 0},
		}, textureBounds)
	}

	return mesh, nil
}

func (b *Builder) BuildForPrint() (*Mesh, error) {
	return b.build(true)
}

func (b *Builder) BuildForDisplay() (*Mesh, error) {
	return b.build(false)
}

func (b *Builder) BuildForDisplayWithTexture() (*Mesh, error) {
	return b.build(false)
}

func (b *Builder) build(isPrint bool) (*Mesh, error) {
	if b.rasterProvider == nil && b.tinMeshProvider == nil {
		return nil, fmt.Errorf("no raster or tin mesh provider set")
	}

	if b.bounds.Min[0] >= b.bounds.Max[0] || b.bounds.Min[1] >= b.bounds.Max[1] {
		return nil, fmt.Errorf("bounds not set properly")
	}

	mesh := &Mesh{
		Bounds: b.bounds,
		Srs:    b.srs,
	}

	if isPrint && b.closeMesh {
		closer := &SimpleCloser{}
		if tinMesh, ok := b.tinMeshProvider.(interface{ GetVertices() interface{} }); ok {
			return closer.CloseSurfaceMesh(tinMesh, b.baseThickness)
		}
	}

	return mesh, nil
}

func (b *Builder) determineZoom() (int, error) {
	if b.zoom != nil {
		return *b.zoom, nil
	}

	if b.autoZoomMin >= b.autoZoomMax {
		return b.autoZoomMin, nil
	}

	bounds := b.bounds
	if bounds.Min[0] >= bounds.Max[0] || bounds.Min[1] >= bounds.Max[1] {
		return b.autoZoomMin, nil
	}

	width := bounds.Max[0] - bounds.Min[0]
	height := bounds.Max[1] - bounds.Min[1]

	estimatedZoom := int(math.Log2(180.0 / math.Max(width, height)))

	if estimatedZoom < b.autoZoomMin {
		return b.autoZoomMin, nil
	}
	if estimatedZoom > b.autoZoomMax {
		return b.autoZoomMax, nil
	}

	return estimatedZoom, nil
}

func (b *Builder) fetchTiles(zoom int) ([]*Tile, error) {
	return []*Tile{}, nil
}

func (b *Builder) handleTileError(tile [3]int, err error) error {
	return err
}

func (b *Builder) mergeTilesToGrid(tiles []*Tile) interface{} {
	return nil
}

func (b *Builder) generateTINFromRaster() (interface{}, error) {
	return nil, nil
}

func (b *Builder) applyVerticalExaggeration(mesh interface{}) {
}

func (b *Builder) applyBaseElevation(mesh interface{}) {
}
