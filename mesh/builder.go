package mesh

import (
	"fmt"
	"github.com/flywave/go-geo"
	draw "github.com/flywave/go-static-mesh/draw"
	static "github.com/flywave/go-static-mesh/static"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
	"image"
	"math"
	"sync"
	"time"
)

type Tile struct {
	Coord     [3]int
	Zoom      int
	Data      []float64
	Elevation *float64
	Bounds    vec2d.Rect
	Width     int
	Height    int
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
	tileCache            TileCache
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

func (b *Builder) SetTileCache(cache TileCache) {
	b.tileCache = cache
}

func (b *Builder) GetCacheStats() CacheStats {
	if b.tileCache == nil {
		return CacheStats{}
	}
	return b.tileCache.Stats()
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

	var tinMesh interface{}
	var err error

	if b.rasterProvider != nil {
		tinMesh, err = b.generateTINFromRaster()
		if err != nil {
			return nil, fmt.Errorf("failed to generate TIN from raster: %w", err)
		}
	} else if b.tinMeshProvider != nil {
		type providerWithMesh interface {
			GetMesh() (interface{}, error)
		}
		provider, ok := b.tinMeshProvider.(providerWithMesh)
		if ok {
			tinMesh, err = provider.GetMesh()
			if err != nil {
				return nil, fmt.Errorf("failed to get TIN mesh: %w", err)
			}
		} else {
			return nil, fmt.Errorf("tin mesh provider does not support GetMesh")
		}
	}

	if tinMesh == nil {
		return nil, fmt.Errorf("failed to generate TIN mesh")
	}

	b.applyVerticalExaggeration(tinMesh)
	b.applyBaseElevation(tinMesh)

	type meshWithConversion interface {
		GetVertices() []vec3d.T
		GetIndices() []uint32
		GetMinHeight() float64
		GetBounds() vec2d.Rect
		GetSrs() geo.Proj
	}

	m, ok := tinMesh.(meshWithConversion)
	if !ok {
		return nil, fmt.Errorf("invalid TIN mesh type")
	}

	mesh := &Mesh{
		Vertices: m.GetVertices(),
		Indices:  m.GetIndices(),
		Bounds:   m.GetBounds(),
		Srs:      m.GetSrs(),
	}

	mesh.CalculateNormals()

	err = b.addGeoDataToMesh(mesh, tinMesh, isPrint)
	if err != nil {
		return nil, fmt.Errorf("failed to add geo data to mesh: %w", err)
	}

	if isPrint && b.closeMesh {
		closer := &SimpleCloser{}
		closedMesh, err := closer.CloseSurfaceMesh(tinMesh, b.baseThickness)
		if err != nil {
			return nil, fmt.Errorf("failed to close mesh: %w", err)
		}
		return closedMesh, nil
	}

	if b.imageryProvider != nil {
		textureMesh, err := b.GenerateTexture()
		if err == nil {
			mesh.Texture = textureMesh.Texture
			if len(textureMesh.UVs) > 0 {
				mesh.UVs = textureMesh.UVs
			}
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
	if b.rasterProvider == nil {
		return nil, fmt.Errorf("raster provider not set")
	}

	type providerWithTileGetter interface {
		GetTileData(coord [3]int) (*static.TileData, error)
		Grid() *geo.TileGrid
	}

	provider, ok := b.rasterProvider.(providerWithTileGetter)
	if !ok {
		return nil, fmt.Errorf("raster provider does not support tile fetching")
	}

	grid := provider.Grid()
	if grid == nil {
		return nil, fmt.Errorf("provider grid not set")
	}

	coords := b.calculateTileCoords(zoom)
	if len(coords) == 0 {
		return nil, fmt.Errorf("no tiles to fetch for given bounds")
	}

	var wg sync.WaitGroup
	tiles := make([]*Tile, len(coords))
	tilesCh := make(chan struct {
		index int
		tile  *Tile
	}, len(coords))

	for i, coord := range coords {
		wg.Add(1)
		go func(idx int, c [3]int) {
			defer wg.Done()

			var tile *Tile

			if b.tileCache != nil {
				if cachedTile, found := b.tileCache.Get(c); found {
					tile = cachedTile
					tilesCh <- struct {
						index int
						tile  *Tile
					}{idx, tile}
					return
				}
			}

			var tileData *static.TileData
			var err error

			for retry := 0; retry < b.tileErrorHandler.MaxRetries; retry++ {
				tileData, err = provider.GetTileData(c)
				if err == nil {
					break
				}
				time.Sleep(time.Second * time.Duration(retry+1))
			}

			if err != nil {
				if b.tileErrorHandler.SkipMissing {
					tilesCh <- struct {
						index int
						tile  *Tile
					}{idx, nil}
					return
				} else if b.tileErrorHandler.UseNoData {
					tileData = b.createNoDataTileData(c, zoom)
				} else {
					tilesCh <- struct {
						index int
						tile  *Tile
					}{idx, nil}
					return
				}
			}

			tileBounds := b.calculateTileBounds(c, zoom)
			tile = &Tile{
				Coord:  c,
				Zoom:   zoom,
				Bounds: tileBounds,
			}

			if tileData != nil {
				tile.Width = int(tileData.Size[0])
				tile.Height = int(tileData.Size[1])

				grid := static.NewElevationGridFromTileData(tileData, b.srs)
				if grid != nil {
					tile.Data = grid.Data
				}
			}

			if b.tileCache != nil && tile != nil && tile.Data != nil {
				b.tileCache.Put(c, tile)
			}

			tilesCh <- struct {
				index int
				tile  *Tile
			}{idx, tile}
		}(i, coord)
	}

	go func() {
		wg.Wait()
		close(tilesCh)
	}()

	for result := range tilesCh {
		if result.tile != nil {
			tiles[result.index] = result.tile
		}
	}

	validTiles := make([]*Tile, 0, len(tiles))
	for _, tile := range tiles {
		if tile != nil && tile.Data != nil {
			validTiles = append(validTiles, tile)
		}
	}

	return validTiles, nil
}

func (b *Builder) handleTileError(tile [3]int, err error) error {
	return err
}

func (b *Builder) mergeTilesToGrid(tiles []*Tile) interface{} {
	if len(tiles) == 0 {
		return nil
	}

	minX := math.Inf(1)
	minY := math.Inf(1)
	maxX := math.Inf(-1)
	maxY := math.Inf(-1)
	tileSize := 0

	for _, tile := range tiles {
		if tile.Bounds.Min[0] < minX {
			minX = tile.Bounds.Min[0]
		}
		if tile.Bounds.Min[1] < minY {
			minY = tile.Bounds.Min[1]
		}
		if tile.Bounds.Max[0] > maxX {
			maxX = tile.Bounds.Max[0]
		}
		if tile.Bounds.Max[1] > maxY {
			maxY = tile.Bounds.Max[1]
		}
		if tile.Width > tileSize {
			tileSize = tile.Width
		}
	}

	if tileSize == 0 {
		tileSize = 256
	}

	width := int(math.Ceil((maxX - minX) / (float64(tileSize) / 256.0)))
	height := int(math.Ceil((maxY - minY) / (float64(tileSize) / 256.0)))

	gridWidth := width
	gridHeight := height

	grid := &static.ElevationGrid{
		Width:    gridWidth,
		Height:   gridHeight,
		MinX:     minX,
		MinY:     minY,
		CellSize: (maxX - minX) / float64(gridWidth),
		NoData:   -9999.0,
		Data:     make([]float64, gridWidth*gridHeight),
		Bounds:   b.bounds,
		Srs:      b.srs,
	}

	for _, tile := range tiles {
		if tile.Data == nil {
			continue
		}

		startX := int((tile.Bounds.Min[0] - minX) / grid.CellSize)
		startY := int((tile.Bounds.Min[1] - minY) / grid.CellSize)

		for y := 0; y < tile.Height && startY+y < gridHeight; y++ {
			for x := 0; x < tile.Width && startX+x < gridWidth; x++ {
				srcIdx := y*tile.Width + x
				dstIdx := (startY+y)*gridWidth + (startX + x)

				if srcIdx < len(tile.Data) && dstIdx < len(grid.Data) {
					grid.Data[dstIdx] = tile.Data[srcIdx]
				}
			}
		}
	}

	return grid
}

func (b *Builder) generateTINFromRaster() (interface{}, error) {
	if b.rasterProvider == nil {
		return nil, fmt.Errorf("raster provider not set")
	}

	if b.tinGenerator == nil {
		return nil, fmt.Errorf("TIN generator not set")
	}

	type providerWithGrid interface {
		GetElevationGrid() *static.ElevationGrid
	}

	provider, ok := b.rasterProvider.(providerWithGrid)
	if ok {
		grid := provider.GetElevationGrid()
		if grid != nil {
			tinMesh, err := b.tinGenerator.GenerateFromRaster(grid)
			if err != nil {
				return nil, fmt.Errorf("failed to generate TIN from raster: %w", err)
			}
			return tinMesh, nil
		}
	}

	zoom, err := b.determineZoom()
	if err != nil {
		return nil, fmt.Errorf("failed to determine zoom: %w", err)
	}

	tiles, err := b.fetchTiles(zoom)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tiles: %w", err)
	}

	if len(tiles) == 0 {
		return nil, fmt.Errorf("no tiles fetched")
	}

	grid := b.mergeTilesToGrid(tiles)
	if grid == nil {
		return nil, fmt.Errorf("failed to merge tiles to grid")
	}

	elevationGrid, ok := grid.(*static.ElevationGrid)
	if !ok {
		return nil, fmt.Errorf("invalid grid type")
	}

	tinMesh, err := b.tinGenerator.GenerateFromRaster(elevationGrid)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TIN from raster: %w", err)
	}

	return tinMesh, nil
}

func (b *Builder) applyVerticalExaggeration(mesh interface{}) {
	type meshWithVertices interface {
		GetVertices() []vec3d.T
	}

	m, ok := mesh.(meshWithVertices)
	if !ok {
		return
	}

	vertices := m.GetVertices()
	for i := range vertices {
		vertices[i][2] *= b.verticalExaggeration
	}
}

func (b *Builder) applyBaseElevation(mesh interface{}) {
	type meshWithVertices interface {
		GetVertices() []vec3d.T
	}

	m, ok := mesh.(meshWithVertices)
	if !ok {
		return
	}

	vertices := m.GetVertices()
	for i := range vertices {
		vertices[i][2] += b.baseElevation
	}
}

func (b *Builder) calculateTileCoords(zoom int) [][3]int {
	if b.srs == nil {
		b.srs = geo.NewProj(4326)
	}

	boundsWGS84 := b.bounds
	if !b.srs.Eq(geo.NewProj(4326)) {
		boundsWGS84 = b.srs.TransformRectTo(geo.NewProj(4326), b.bounds, 16)
	}

	minLat := boundsWGS84.Min[0]
	maxLat := boundsWGS84.Max[0]
	minLon := boundsWGS84.Min[1]
	maxLon := boundsWGS84.Max[1]

	if minLat < -85.0511 {
		minLat = -85.0511
	}
	if maxLat > 85.0511 {
		maxLat = 85.0511
	}

	n := math.Pow(2, float64(zoom))

	minX := int(math.Floor(((minLon + 180.0) / 360.0) * n))
	maxX := int(math.Floor(((maxLon + 180.0) / 360.0) * n))

	minY := int(math.Floor((1.0 - math.Log(math.Tan(maxLat*math.Pi/180.0)+1.0/math.Cos(maxLat*math.Pi/180.0))/math.Pi) / 2.0 * n))
	maxY := int(math.Floor((1.0 - math.Log(math.Tan(minLat*math.Pi/180.0)+1.0/math.Cos(minLat*math.Pi/180.0))/math.Pi) / 2.0 * n))

	if minX < 0 {
		minX = 0
	}
	if maxX >= int(n) {
		maxX = int(n) - 1
	}
	if minY < 0 {
		minY = 0
	}
	if maxY >= int(n) {
		maxY = int(n) - 1
	}

	if minX > maxX || minY > maxY {
		return [][3]int{}
	}

	capacity := (maxX - minX + 1) * (maxY - minY + 1)
	if capacity < 0 || capacity > 100000 {
		capacity = 0
	}

	coords := make([][3]int, 0, capacity)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			coords = append(coords, [3]int{x, y, zoom})
		}
	}

	return coords
}

func (b *Builder) calculateTileBounds(coord [3]int, zoom int) vec2d.Rect {
	x, y, z := coord[0], coord[1], coord[2]

	n := math.Pow(2, float64(z))
	lon1 := float64(x)/n*360.0 - 180.0
	lon2 := float64(x+1)/n*360.0 - 180.0

	lat1 := math.Atan(math.Sinh(math.Pi * (1 - 2*float64(y)/n)))
	lat1 = lat1 * 180.0 / math.Pi

	lat2 := math.Atan(math.Sinh(math.Pi * (1 - 2*float64(y+1)/n)))
	lat2 = lat2 * 180.0 / math.Pi

	bounds := vec2d.Rect{
		Min: vec2d.T{math.Min(lat1, lat2), lon1},
		Max: vec2d.T{math.Max(lat1, lat2), lon2},
	}

	if b.srs != nil && !b.srs.Eq(geo.NewProj(4326)) {
		bounds = geo.NewProj(4326).TransformRectTo(b.srs, bounds, 16)
	}

	return bounds
}

func (b *Builder) createNoDataTileData(coord [3]int, zoom int) *static.TileData {
	bounds := b.calculateTileBounds(coord, zoom)

	tileData := static.NewTileData([2]uint32{256, 256}, static.BORDER_NONE)
	for i := range tileData.Datas {
		tileData.Datas[i] = b.tileErrorHandler.NoDataValue
	}
	tileData.Box = bounds
	tileData.Boxsrs = b.srs
	tileData.NoData = b.tileErrorHandler.NoDataValue

	return tileData
}
