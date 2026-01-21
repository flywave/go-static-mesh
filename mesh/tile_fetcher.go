package mesh

import (
	"fmt"
	"image"
	"log"
	"math"
	"sync"

	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

type TextureTile struct {
	Coord  [3]int
	Image  image.Image
	X      int
	Y      int
	Bounds vec2d.Rect
}

type TileFetcher struct {
	provider interface {
		GetImageTile(coord [3]int) (image.Image, error)
	}
}

func NewTileFetcher(provider interface {
	GetImageTile(coord [3]int) (image.Image, error)
}) *TileFetcher {
	return &TileFetcher{provider: provider}
}

func (f *TileFetcher) FetchTiles(bounds vec2d.Rect, zoom int, srs geo.Proj) ([]*TextureTile, error) {
	tileCoords := f.calculateTileCoords(bounds, zoom, srs)
	if len(tileCoords) == 0 {
		return nil, fmt.Errorf("no tiles to fetch for given bounds")
	}

	var wg sync.WaitGroup
	tiles := make([]*TextureTile, len(tileCoords))
	tilesCh := make(chan struct {
		index int
		tile  *TextureTile
	}, len(tileCoords))

	for i, coord := range tileCoords {
		wg.Add(1)
		go func(idx int, c [3]int) {
			defer wg.Done()

			img, err := f.provider.GetImageTile(c)
			if err != nil {
				log.Printf("Failed to fetch tile %d/%d/%d: %v", c[0], c[1], c[2], err)
				tilesCh <- struct {
					index int
					tile  *TextureTile
				}{idx, nil}
				return
			}

			tileBounds := f.calculateTileBounds(c, zoom, srs)
			tile := &TextureTile{
				Coord:  c,
				Image:  img,
				X:      idx,
				Y:      0,
				Bounds: tileBounds,
			}
			tilesCh <- struct {
				index int
				tile  *TextureTile
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

	validTiles := make([]*TextureTile, 0, len(tiles))
	for _, tile := range tiles {
		if tile != nil {
			validTiles = append(validTiles, tile)
		}
	}

	return validTiles, nil
}

func (f *TileFetcher) calculateTileCoords(bounds vec2d.Rect, zoom int, srs geo.Proj) [][3]int {
	if srs == nil {
		srs = geo.NewProj(4326)
	}

	boundsWGS84 := bounds
	if !srs.Eq(geo.NewProj(4326)) {
		boundsWGS84 = srs.TransformRectTo(geo.NewProj(4326), bounds, 16)
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

func (f *TileFetcher) calculateTileBounds(coord [3]int, zoom int, srs geo.Proj) vec2d.Rect {
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

	if srs != nil && !srs.Eq(geo.NewProj(4326)) {
		bounds = geo.NewProj(4326).TransformRectTo(srs, bounds, 16)
	}

	return bounds
}
