package static

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/flywave/go-geo"
	qm "github.com/flywave/go-quantized-mesh"
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
	url           string
	extensionFlag qm.TerrainExtensionFlag
	grid          *geo.TileGrid
	bounds        vec2d.Rect
	srs           geo.Proj
	client        *http.Client
	tileCache     map[[3]int]*TinMesh
	cacheMutex    sync.RWMutex
	maxCacheSize  int
	maxRetries    int
	timeout       time.Duration
}

func NewCesiumQuantizedMeshProvider(url string) *CesiumQuantizedMeshProvider {
	grid := &geo.TileGrid{}
	return &CesiumQuantizedMeshProvider{
		url:           url,
		extensionFlag: qm.Ext_None,
		grid:          grid,
		bounds:        vec2d.Rect{Min: vec2d.T{-85.0, -180.0}, Max: vec2d.T{85.0, 180.0}},
		srs:           geo.NewProj(4326),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		tileCache:    make(map[[3]int]*TinMesh),
		maxCacheSize: 100,
		maxRetries:   3,
		timeout:      30 * time.Second,
	}
}

func NewCesiumQuantizedMeshProviderWithConfig(url string, config *DecoderConfig) *CesiumQuantizedMeshProvider {
	grid := &geo.TileGrid{}
	extensionFlag := qm.Ext_None

	if config != nil {
		if config.ExtensionHeader {
			if config.VertexNormals {
				extensionFlag |= qm.Ext_Light
			}
			if config.WaterMask {
				extensionFlag |= qm.Ext_WaterMask
			}
			if config.Metadata {
				extensionFlag |= qm.Ext_Metadata
			}
		}
	}

	timeout := 30 * time.Second
	if config != nil && config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	maxRetries := 3
	if config != nil && config.MaxRetries > 0 {
		maxRetries = config.MaxRetries
	}

	return &CesiumQuantizedMeshProvider{
		url:           url,
		extensionFlag: extensionFlag,
		grid:          grid,
		bounds:        vec2d.Rect{Min: vec2d.T{-85.0, -180.0}, Max: vec2d.T{85.0, 180.0}},
		srs:           geo.NewProj(4326),
		client: &http.Client{
			Timeout: timeout,
		},
		tileCache:    make(map[[3]int]*TinMesh),
		maxCacheSize: 100,
		maxRetries:   maxRetries,
		timeout:      timeout,
	}
}

type DecoderConfig struct {
	ExtensionHeader bool
	WaterMask       bool
	VertexNormals   bool
	Metadata        bool
	MaxRetries      int
	Timeout         int
}

func NewDecoderConfig() *DecoderConfig {
	return &DecoderConfig{
		ExtensionHeader: true,
		VertexNormals:   true,
		WaterMask:       false,
		Metadata:        false,
		MaxRetries:      3,
		Timeout:         30,
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
	p.cacheMutex.RLock()
	if cached, ok := p.tileCache[coord]; ok {
		p.cacheMutex.RUnlock()
		return cached, nil
	}
	p.cacheMutex.RUnlock()

	url := fmt.Sprintf(p.url, coord[2], coord[0], coord[1])
	data, err := p.fetchWithRetry(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tile %v: %w", coord, err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("empty tile data for %v", coord)
	}

	qmTile := &qm.QuantizedMeshTile{}
	reader := bytes.NewReader(data)
	err = qmTile.Read(reader, p.extensionFlag)
	if err != nil {
		return nil, fmt.Errorf("failed to decode quantized mesh: %w", err)
	}

	tinMesh := p.convertToTinMesh(qmTile, coord)

	p.cacheMutex.Lock()
	if len(p.tileCache) >= p.maxCacheSize {
		p.evictCache()
	}
	p.tileCache[coord] = tinMesh
	p.cacheMutex.Unlock()

	return tinMesh, nil
}

func (p *CesiumQuantizedMeshProvider) fetchWithRetry(url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt < p.maxRetries; attempt++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Accept", qm.GetTerrainMime(p.extensionFlag))

		resp, err := p.client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("tile not found")
		}

		if resp.StatusCode != http.StatusOK {
			data, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("http error %d: %s", resp.StatusCode, string(data))
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		return data, nil
	}

	return nil, lastErr
}

func (p *CesiumQuantizedMeshProvider) convertToTinMesh(qmTile *qm.QuantizedMeshTile, coord [3]int) *TinMesh {
	meshData, err := qmTile.GetMesh()
	if err != nil || meshData == nil {
		return &TinMesh{
			Bounds:    p.calculateTileBounds(coord),
			Srs:       p.srs,
			MinHeight: 0,
			MaxHeight: 0,
		}
	}

	vertices := make([]vec3d.T, len(meshData.Vertices))
	for i, v := range meshData.Vertices {
		vertices[i] = vec3d.T{v[0], v[1], v[2]}
	}

	indices := make([]uint32, len(meshData.Faces)*3)
	idx := 0
	for _, f := range meshData.Faces {
		indices[idx] = uint32(f[0])
		indices[idx+1] = uint32(f[1])
		indices[idx+2] = uint32(f[2])
		idx += 3
	}

	var northEdge, southEdge, westEdge, eastEdge []uint16
	if qmTile.Edge != nil {
		if edge16, ok := qmTile.Edge.(*qm.Indices16); ok {
			count := edge16.GetIndexCount()
			if count > 0 {
				splitSize := count / 4
				if splitSize > 0 {
					northEdge = make([]uint16, splitSize)
					southEdge = make([]uint16, splitSize)
					westEdge = make([]uint16, splitSize)
					eastEdge = make([]uint16, splitSize)
					for i := 0; i < splitSize; i++ {
						northEdge[i] = uint16(edge16.GetIndex(i))
						southEdge[i] = uint16(edge16.GetIndex(i + splitSize))
						westEdge[i] = uint16(edge16.GetIndex(i + splitSize*2))
						eastEdge[i] = uint16(edge16.GetIndex(i + splitSize*3))
					}
				}
			}
		}
	}

	minHeight := 0.0
	maxHeight := 0.0
	if len(vertices) > 0 {
		minHeight = vertices[0][2]
		maxHeight = vertices[0][2]
		for _, v := range vertices {
			if v[2] < minHeight {
				minHeight = v[2]
			}
			if v[2] > maxHeight {
				maxHeight = v[2]
			}
		}
	}

	return &TinMesh{
		Vertices:    vertices,
		Indices:     indices,
		EdgeIndices: nil,
		NorthEdge:   northEdge,
		SouthEdge:   southEdge,
		WestEdge:    westEdge,
		EastEdge:    eastEdge,
		MinHeight:   minHeight,
		MaxHeight:   maxHeight,
		Bounds:      p.calculateTileBounds(coord),
		Srs:         p.srs,
	}
}

func (p *CesiumQuantizedMeshProvider) calculateTileBounds(coord [3]int) vec2d.Rect {
	x, y, z := coord[0], coord[1], coord[2]

	n := 1 << z
	lon1 := float64(x)/float64(n)*360.0 - 180.0
	lon2 := float64(x+1)/float64(n)*360.0 - 180.0

	lat1 := math.Atan(math.Sinh(math.Pi * (1 - 2*float64(y)/float64(n))))
	lat1 = lat1 * 180.0 / math.Pi

	lat2 := math.Atan(math.Sinh(math.Pi * (1 - 2*float64(y+1)/float64(n))))
	lat2 = lat2 * 180.0 / math.Pi

	bounds := vec2d.Rect{
		Min: vec2d.T{math.Min(lat1, lat2), lon1},
		Max: vec2d.T{math.Max(lat1, lat2), lon2},
	}

	return bounds
}

func (p *CesiumQuantizedMeshProvider) evictCache() {
	keys := make([][3]int, 0, len(p.tileCache))
	for k := range p.tileCache {
		keys = append(keys, k)
	}

	if len(keys) > p.maxCacheSize/2 {
		for _, k := range keys[:len(keys)-p.maxCacheSize/2] {
			delete(p.tileCache, k)
		}
	}
}

func (p *CesiumQuantizedMeshProvider) GetHeight(lng, lat float64) float64 {
	if p.srs == nil {
		return 0
	}

	tileCoord := p.lonLatToTileCoord(lng, lat, 15)
	tinMesh, err := p.GetMeshTile(tileCoord)
	if err != nil || tinMesh == nil {
		return 0
	}

	localX, localY := p.lonLatToLocal(lng, lat, tileCoord)

	for i := 0; i < len(tinMesh.Vertices); i += 3 {
		if i+2 >= len(tinMesh.Vertices) {
			break
		}
		v0 := tinMesh.Vertices[tinMesh.Indices[i]]
		v1 := tinMesh.Vertices[tinMesh.Indices[i+1]]
		v2 := tinMesh.Vertices[tinMesh.Indices[i+2]]

		if p.pointInTriangle(localX, localY, v0[0], v0[1], v1[0], v1[1], v2[0], v2[1]) {
			return p.interpolateHeight(localX, localY, v0, v1, v2)
		}
	}

	return 0
}

func (p *CesiumQuantizedMeshProvider) GetNormal(lng, lat float64) vec3d.T {
	height := p.GetHeight(lng, lat)
	if height == 0 {
		return vec3d.T{0, 0, 1}
	}

	delta := 1.0
	h1 := p.GetHeight(lng, lat+delta)
	h2 := p.GetHeight(lng+delta, lat)

	dx := h2 - height
	dy := h1 - height

	normal := vec3d.T{-dx, -dy, delta}
	length := math.Sqrt(normal[0]*normal[0] + normal[1]*normal[1] + normal[2]*normal[2])
	if length > 0 {
		normal[0] /= length
		normal[1] /= length
		normal[2] /= length
	}

	return normal
}

func (p *CesiumQuantizedMeshProvider) GetMesh() (*TinMesh, error) {
	zoom := 12
	coords := p.calculateTileCoordsForBounds(zoom, p.bounds)

	var merged *TinMesh
	for _, coord := range coords {
		tileMesh, err := p.GetMeshTile(coord)
		if err != nil {
			continue
		}
		if merged == nil {
			merged = tileMesh
		} else {
			merged = p.mergeMeshes(merged, tileMesh)
		}
	}

	return merged, nil
}

func (p *CesiumQuantizedMeshProvider) ClearCache() {
	p.cacheMutex.Lock()
	defer p.cacheMutex.Unlock()
	p.tileCache = make(map[[3]int]*TinMesh)
}

func (p *CesiumQuantizedMeshProvider) GetCacheSize() int {
	p.cacheMutex.RLock()
	defer p.cacheMutex.RUnlock()
	return len(p.tileCache)
}

func (p *CesiumQuantizedMeshProvider) SetMaxCacheSize(size int) {
	p.cacheMutex.Lock()
	defer p.cacheMutex.Unlock()
	p.maxCacheSize = size
	for len(p.tileCache) > size {
		p.evictCache()
	}
}

func (p *CesiumQuantizedMeshProvider) lonLatToTileCoord(lng, lat float64, zoom int) [3]int {
	n := 1 << zoom
	x := int((lng + 180.0) / 360.0 * float64(n))
	y := int((1.0 - math.Log(math.Tan(lat*math.Pi/180.0)+1.0/math.Cos(lat*math.Pi/180.0))/math.Pi) / 2.0 * float64(n))

	if x < 0 {
		x = 0
	}
	if x >= n {
		x = n - 1
	}
	if y < 0 {
		y = 0
	}
	if y >= n {
		y = n - 1
	}

	return [3]int{x, y, zoom}
}

func (p *CesiumQuantizedMeshProvider) lonLatToLocal(lng, lat float64, coord [3]int) (float64, float64) {
	bounds := p.calculateTileBounds(coord)
	localX := (lng - bounds.Min[1]) / (bounds.Max[1] - bounds.Min[1])
	localY := (lat - bounds.Min[0]) / (bounds.Max[0] - bounds.Min[0])
	return localX, localY
}

func (p *CesiumQuantizedMeshProvider) pointInTriangle(px, py, x0, y0, x1, y1, x2, y2 float64) bool {
	denom := (y1-y2)*(x0-x2) + (x2-x1)*(y0-y2)
	if math.Abs(denom) < 1e-10 {
		return false
	}

	a := ((y1-y2)*(px-x2) + (x2-x1)*(py-y2)) / denom
	b := ((y2-y0)*(px-x2) + (x0-x2)*(py-y2)) / denom
	c := 1.0 - a - b

	return a >= 0 && b >= 0 && c >= 0
}

func (p *CesiumQuantizedMeshProvider) interpolateHeight(px, py float64, v0, v1, v2 vec3d.T) float64 {
	denom := (v1[1]-v2[1])*(v0[0]-v2[0]) + (v2[0]-v1[0])*(v0[1]-v2[1])
	if math.Abs(denom) < 1e-10 {
		return v0[2]
	}

	w0 := ((v1[1]-v2[1])*(px-v2[0]) + (v2[0]-v1[0])*(py-v2[1])) / denom
	w1 := ((v2[1]-v0[1])*(px-v2[0]) + (v0[0]-v2[0])*(py-v2[1])) / denom
	w2 := 1.0 - w0 - w1

	return w0*v0[2] + w1*v1[2] + w2*v2[2]
}

func (p *CesiumQuantizedMeshProvider) calculateTileCoordsForBounds(zoom int, bounds vec2d.Rect) [][3]int {
	minLat := bounds.Min[0]
	maxLat := bounds.Max[0]
	minLon := bounds.Min[1]
	maxLon := bounds.Max[1]

	if minLat < -85.0511 {
		minLat = -85.0511
	}
	if maxLat > 85.0511 {
		maxLat = 85.0511
	}

	n := 1 << zoom

	minX := int((minLon + 180.0) / 360.0 * float64(n))
	maxX := int((maxLon + 180.0) / 360.0 * float64(n))
	minY := int((1.0 - math.Log(math.Tan(maxLat*math.Pi/180.0)+1.0/math.Cos(maxLat*math.Pi/180.0))) / math.Pi / 2.0 * float64(n))
	maxY := int((1.0 - math.Log(math.Tan(minLat*math.Pi/180.0)+1.0/math.Cos(minLat*math.Pi/180.0))) / math.Pi / 2.0 * float64(n))

	if minX < 0 {
		minX = 0
	}
	if maxX >= n {
		maxX = n - 1
	}
	if minY < 0 {
		minY = 0
	}
	if maxY >= n {
		maxY = n - 1
	}

	if minX > maxX || minY > maxY {
		return [][3]int{}
	}

	coords := make([][3]int, 0, (maxX-minX+1)*(maxY-minY+1))
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			coords = append(coords, [3]int{x, y, zoom})
		}
	}

	return coords
}

func (p *CesiumQuantizedMeshProvider) mergeMeshes(mesh1, mesh2 *TinMesh) *TinMesh {
	if mesh1 == nil {
		return mesh2
	}
	if mesh2 == nil {
		return mesh1
	}

	merged := &TinMesh{
		Vertices:    make([]vec3d.T, 0, len(mesh1.Vertices)+len(mesh2.Vertices)),
		Indices:     make([]uint32, 0, len(mesh1.Indices)+len(mesh2.Indices)),
		EdgeIndices: nil,
		NorthEdge:   nil,
		SouthEdge:   nil,
		WestEdge:    nil,
		EastEdge:    nil,
		MinHeight:   math.Min(mesh1.MinHeight, mesh2.MinHeight),
		MaxHeight:   math.Max(mesh1.MaxHeight, mesh2.MaxHeight),
		Srs:         mesh1.Srs,
	}

	merged.Vertices = append(merged.Vertices, mesh1.Vertices...)
	merged.Vertices = append(merged.Vertices, mesh2.Vertices...)

	baseIndex := uint32(len(mesh1.Vertices))
	for _, idx := range mesh2.Indices {
		merged.Indices = append(merged.Indices, idx+baseIndex)
	}
	merged.Indices = append(merged.Indices, mesh1.Indices...)

	merged.Bounds.Min[0] = math.Min(mesh1.Bounds.Min[0], mesh2.Bounds.Min[0])
	merged.Bounds.Min[1] = math.Min(mesh1.Bounds.Min[1], mesh2.Bounds.Min[1])
	merged.Bounds.Max[0] = math.Max(mesh1.Bounds.Max[0], mesh2.Bounds.Max[0])
	merged.Bounds.Max[1] = math.Max(mesh1.Bounds.Max[1], mesh2.Bounds.Max[1])

	return merged
}
