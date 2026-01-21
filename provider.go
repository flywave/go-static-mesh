package static

import (
	"image"

	"github.com/flywave/go-geo"
)

type TileProvider interface {
	Attribution() string
	Grid() *geo.TileGrid
}

type TileFetcher interface {
	Fetch(coord [3]int) (image.Image, error)
}

func NewTileFetcher(p TileProvider) TileFetcher {
	return nil
}
