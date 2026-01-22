package static

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/flywave/go-cog"
	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

type ImageryProvider interface {
	ImageTileProvider
	TileFetcher

	GetImageBounds() vec2d.Rect
}

type GenericTileProvider struct {
	config *TileProviderConfig
	grid   *geo.TileGrid
	bounds vec2d.Rect
	srs    geo.Proj
	client *http.Client
}

func NewTileProvider(url string, mode TileProviderMode) *GenericTileProvider {
	grid := &geo.TileGrid{}
	config := &TileProviderConfig{
		Mode:    mode,
		URL:     url,
		Headers: make(map[string]string),
	}

	if mode == TileProviderModeTMS {
		config.URL = strings.Replace(url, "{z}", "%d", -1)
		config.URL = strings.Replace(config.URL, "{x}", "%d", -1)
		config.URL = strings.Replace(config.URL, "{y}", "%d", -1)
	} else {
		config.URL = strings.Replace(url, "{z}", "%d", -1)
		config.URL = strings.Replace(config.URL, "{x}", "%d", -1)
		config.URL = strings.Replace(config.URL, "{y}", "%d", -1)
	}

	return &GenericTileProvider{
		config: config,
		grid:   grid,
		bounds: vec2d.Rect{Min: vec2d.T{-85.0, -180.0}, Max: vec2d.T{85.0, 180.0}},
		srs:    geo.NewProj(4326),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func NewTileProviderWithConfig(config *TileProviderConfig) *GenericTileProvider {
	grid := &geo.TileGrid{}
	if config.Grid != nil {
		grid = config.Grid
	}

	timeout := 30 * time.Second
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	return &GenericTileProvider{
		config: config,
		grid:   grid,
		bounds: vec2d.Rect{Min: vec2d.T{-85.0, -180.0}, Max: vec2d.T{85.0, 180.0}},
		srs:    grid.Srs,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (p *GenericTileProvider) Attribution() string {
	if p.config != nil && p.config.Attribution != "" {
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

func (p *GenericTileProvider) buildTileURL(x, y, z int) string {
	url := p.config.URL
	return fmt.Sprintf(url, z, x, y)
}

func (p *GenericTileProvider) Fetch(coord [3]int) ([]byte, error) {
	x, y, z := coord[0], coord[1], coord[2]

	maxRetries := 3
	if p.config.MaxRetries > 0 {
		maxRetries = p.config.MaxRetries
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		url := p.buildTileURL(x, y, z)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		if p.config.UserAgent != "" {
			req.Header.Set("User-Agent", p.config.UserAgent)
		}

		for key, value := range p.config.Headers {
			req.Header.Set(key, value)
		}

		resp, err := p.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to fetch tile: %w", err)
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			return nil, fmt.Errorf("tile not found: %d/%d/%d", z, x, y)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("tile request failed with status %d", resp.StatusCode)
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = fmt.Errorf("failed to read tile data: %w", err)
			continue
		}

		return data, nil
	}

	return nil, lastErr
}

func (p *GenericTileProvider) GetImageTile(coord [3]int) (image.Image, error) {
	data, err := p.Fetch(coord)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("empty tile data")
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		_, err1 := png.Decode(bytes.NewReader(data))
		_, err2 := jpeg.Decode(bytes.NewReader(data))
		if err1 == nil {
			img, _ = png.Decode(bytes.NewReader(data))
		} else if err2 == nil {
			img, _ = jpeg.Decode(bytes.NewReader(data))
		} else {
			return nil, fmt.Errorf("failed to decode tile image: %w", err)
		}
	}

	return img, nil
}

func (p *GenericTileProvider) GetImageBounds() vec2d.Rect {
	return p.bounds
}

type GeoTIFFImageryProvider struct {
	filename string
	tempFile *os.File
	grid     *geo.TileGrid
	bounds   vec2d.Rect
	srs      geo.Proj
	reader   *cog.Reader
	image    image.Image
}

func NewGeoTIFFImageryProvider(filename string) (*GeoTIFFImageryProvider, error) {
	reader := cog.Read(filename)
	if reader == nil || len(reader.Data) == 0 {
		return nil, fmt.Errorf("failed to read GeoTIFF file: %s", filename)
	}

	p := &GeoTIFFImageryProvider{
		filename: filename,
		reader:   reader,
	}

	if len(reader.Data) > 0 {
		bounds := reader.GetBounds(0)
		p.bounds = bounds

		if epsgCode, err := reader.GetEPSGCode(0); err == nil && epsgCode != 0 {
			p.srs = geo.NewProj(uint32(epsgCode))
		} else {
			p.srs = geo.NewProj(4326)
		}

		size := reader.GetSize(0)
		tileSize0 := uint32(size[0])
		tileSize1 := uint32(size[1])
		tileSize := [2]uint32{tileSize0, tileSize1}
		p.grid = &geo.TileGrid{
			Srs:      p.srs,
			TileSize: tileSize[:],
		}

		if img, ok := reader.Data[0].(image.Image); ok {
			p.image = img
		}
	}

	return p, nil
}

func NewGeoTIFFImageryProviderFromReader(r io.Reader) (*GeoTIFFImageryProvider, error) {
	file, err := os.CreateTemp("", "geotiff-imagery-*.tif")
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

	p, err := NewGeoTIFFImageryProvider(file.Name())
	if err != nil {
		os.Remove(file.Name())
		return nil, err
	}

	p.tempFile = file
	return p, nil
}

func (p *GeoTIFFImageryProvider) Close() error {
	if p.tempFile != nil {
		err := os.Remove(p.tempFile.Name())
		p.tempFile = nil
		return err
	}

	return nil
}

func (p *GeoTIFFImageryProvider) Attribution() string {
	return ""
}

func (p *GeoTIFFImageryProvider) Grid() *geo.TileGrid {
	return p.grid
}

func (p *GeoTIFFImageryProvider) Bounds() vec2d.Rect {
	return p.bounds
}

func (p *GeoTIFFImageryProvider) Srs() geo.Proj {
	return p.srs
}

func (p *GeoTIFFImageryProvider) GetImageTile(coord [3]int) (image.Image, error) {
	if p.image != nil {
		return p.image, nil
	}

	return nil, fmt.Errorf("no image data available")
}

func (p *GeoTIFFImageryProvider) GetImageBounds() vec2d.Rect {
	return p.bounds
}

func (p *GeoTIFFImageryProvider) Fetch(coord [3]int) ([]byte, error) {
	img, err := p.GetImageTile(coord)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	return buf.Bytes(), nil
}

func NewTileFetcher(provider interface {
	GetImageTile(coord [3]int) (image.Image, error)
}) *TileFetcherImpl {
	return &TileFetcherImpl{provider: provider}
}

type TileFetcherImpl struct {
	provider interface {
		GetImageTile(coord [3]int) (image.Image, error)
	}
}

func (f *TileFetcherImpl) Fetch(coord [3]int) (image.Image, error) {
	return f.provider.GetImageTile(coord)
}
