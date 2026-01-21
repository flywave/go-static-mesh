package static

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strings"
	"time"

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
