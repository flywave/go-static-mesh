package tile

import (
	draw "github.com/flywave/go-static-mesh/draw"

	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

type TileProvider interface {
	draw.TileProvider
}

type TileFetcher interface {
	Fetch(coord [3]int) ([]byte, error)
}

type ImageTileProvider interface {
	TileProvider
	draw.ImageTileProvider
}

type TileProviderMode int

const (
	TileProviderModeXYZ TileProviderMode = iota
	TileProviderModeTMS
)

type TileProviderConfig struct {
	Mode        TileProviderMode
	URL         string
	MaxRetries  int
	Timeout     int
	UserAgent   string
	Headers     map[string]string
	Attribution string
	Grid        *geo.TileGrid
	Bounds      vec2d.Rect
}
