package static

import (
	"io"
	"os"

	"image/color"

	"github.com/flywave/go-gpx"
	draw "github.com/flywave/go-static-mesh/draw"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

type GeoDataProvider interface {
	GetPaths() []*draw.Path
	GetAreas() []*draw.Area
	GetPoints() []*draw.Marker
}

type GPXProvider struct {
	paths []*draw.Path
}

func NewGPXProvider(filename string) (*GPXProvider, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return NewGPXProviderFromReader(file)
}

func NewGPXProviderFromReader(reader io.Reader) (*GPXProvider, error) {
	gpxData, err := gpx.Read(reader)
	if err != nil {
		return nil, err
	}

	provider := &GPXProvider{
		paths: make([]*draw.Path, 0),
	}

	for _, trk := range gpxData.Trk {
		for _, seg := range trk.TrkSeg {
			path := draw.NewPath(nil, nil, color.RGBA{0xff, 0, 0, 0xff}, 2.0)
			for _, pt := range seg.TrkPt {
				path.Positions = append(path.Positions, vec2d.T{pt.Lat, pt.Lon})
			}
			if len(path.Positions) > 0 {
				provider.paths = append(provider.paths, path)
			}
		}
	}

	return provider, nil
}

func (p *GPXProvider) GetPaths() []*draw.Path {
	return p.paths
}

func (p *GPXProvider) GetAreas() []*draw.Area {
	return nil
}

func (p *GPXProvider) GetPoints() []*draw.Marker {
	return nil
}
