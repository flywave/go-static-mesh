package static

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"encoding/xml"
	"image/color"

	"github.com/flywave/go-geo"
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

func (p *GPXProvider) Bounds() vec2d.Rect {
	r := vec2d.Rect{
		Min: vec2d.T{90.0, 180.0},
		Max: vec2d.T{-90.0, -180.0},
	}
	for _, path := range p.paths {
		bounds := path.Bounds()
		r.Min[0] = math.Min(r.Min[0], bounds.Min[0])
		r.Min[1] = math.Min(r.Min[1], bounds.Min[1])
		r.Max[0] = math.Max(r.Max[0], bounds.Max[0])
		r.Max[1] = math.Max(r.Max[1], bounds.Max[1])
	}
	return r
}

func (p *GPXProvider) Srs() geo.Proj {
	return geo.NewProj(4326)
}

func (p *GPXProvider) Attribution() string {
	return ""
}

func (p *GPXProvider) Grid() *geo.TileGrid {
	return &geo.TileGrid{}
}

type GeoJSONProvider struct {
	paths  []*draw.Path
	areas  []*draw.Area
	points []*draw.Marker
}

func NewGeoJSONProvider(filename string) (*GeoJSONProvider, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return NewGeoJSONProviderFromReader(file)
}

func NewGeoJSONProviderFromReader(reader io.Reader) (*GeoJSONProvider, error) {
	geoJSON := make(map[string]interface{})
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&geoJSON); err != nil {
		return nil, err
	}

	provider := &GeoJSONProvider{
		paths:  make([]*draw.Path, 0),
		areas:  make([]*draw.Area, 0),
		points: make([]*draw.Marker, 0),
	}

	if features, ok := geoJSON["features"].([]interface{}); ok {
		for _, feature := range features {
			if featureMap, ok := feature.(map[string]interface{}); ok {
				provider.parseFeature(featureMap)
			}
		}
	}

	return provider, nil
}

func (p *GeoJSONProvider) parseFeature(feature map[string]interface{}) {
	geometry, ok := feature["geometry"].(map[string]interface{})
	if !ok {
		return
	}

	geomType, _ := geometry["type"].(string)
	coords, _ := geometry["coordinates"].([]interface{})

	switch geomType {
	case "Point":
		if len(coords) >= 2 {
			lng := 0.0
			lat := 0.0
			if x, ok := coords[0].(float64); ok {
				lng = x
			}
			if y, ok := coords[1].(float64); ok {
				lat = y
			}
			marker := draw.NewMarker(vec2d.T{lat, lng}, geo.NewProj(4326), color.RGBA{0, 0, 0xff, 0xff}, 10.0)
			p.points = append(p.points, marker)
		}

	case "LineString":
		if coords != nil {
			path := draw.NewPath(nil, nil, color.RGBA{0, 0, 0xff, 0xff}, 2.0)
			for _, coord := range coords {
				if coordArray, ok := coord.([]interface{}); ok && len(coordArray) >= 2 {
					lng := 0.0
					lat := 0.0
					if x, ok := coordArray[0].(float64); ok {
						lng = x
					}
					if y, ok := coordArray[1].(float64); ok {
						lat = y
					}
					path.Positions = append(path.Positions, vec2d.T{lat, lng})
				}
			}
			if len(path.Positions) > 0 {
				p.paths = append(p.paths, path)
			}
		}

	case "Polygon":
		if len(coords) > 0 {
			ring, ok := coords[0].([]interface{})
			if ok {
				area := draw.NewArea(nil, geo.NewProj(4326), color.RGBA{0, 0xff, 0, 0x80}, color.RGBA{0, 0, 0, 0}, 2.0)
				for _, coord := range ring {
					if coordArray, ok := coord.([]interface{}); ok && len(coordArray) >= 2 {
						lng := 0.0
						lat := 0.0
						if x, ok := coordArray[0].(float64); ok {
							lng = x
						}
						if y, ok := coordArray[1].(float64); ok {
							lat = y
						}
						area.Positions = append(area.Positions, vec2d.T{lat, lng})
					}
				}
				if len(area.Positions) > 0 {
					p.areas = append(p.areas, area)
				}
			}
		}

	case "MultiLineString":
		for _, line := range coords {
			if lineArray, ok := line.([]interface{}); ok {
				path := draw.NewPath(nil, nil, color.RGBA{0, 0, 0xff, 0xff}, 2.0)
				for _, coord := range lineArray {
					if coordArray, ok := coord.([]interface{}); ok && len(coordArray) >= 2 {
						lng := 0.0
						lat := 0.0
						if x, ok := coordArray[0].(float64); ok {
							lng = x
						}
						if y, ok := coordArray[1].(float64); ok {
							lat = y
						}
						path.Positions = append(path.Positions, vec2d.T{lat, lng})
					}
				}
				if len(path.Positions) > 0 {
					p.paths = append(p.paths, path)
				}
			}
		}

	case "MultiPolygon":
		for _, polygon := range coords {
			if polygonArray, ok := polygon.([]interface{}); ok && len(polygonArray) > 0 {
				ring, ok := polygonArray[0].([]interface{})
				if ok {
					area := draw.NewArea(nil, geo.NewProj(4326), color.RGBA{0, 0xff, 0, 0x80}, color.RGBA{0, 0, 0, 0}, 2.0)
					for _, coord := range ring {
						if coordArray, ok := coord.([]interface{}); ok && len(coordArray) >= 2 {
							lng := 0.0
							lat := 0.0
							if x, ok := coordArray[0].(float64); ok {
								lng = x
							}
							if y, ok := coordArray[1].(float64); ok {
								lat = y
							}
							area.Positions = append(area.Positions, vec2d.T{lat, lng})
						}
					}
					if len(area.Positions) > 0 {
						p.areas = append(p.areas, area)
					}
				}
			}
		}
	}
}

func (p *GeoJSONProvider) GetPaths() []*draw.Path {
	return p.paths
}

func (p *GeoJSONProvider) GetAreas() []*draw.Area {
	return p.areas
}

func (p *GeoJSONProvider) GetPoints() []*draw.Marker {
	return p.points
}

func (p *GeoJSONProvider) Bounds() vec2d.Rect {
	r := vec2d.Rect{
		Min: vec2d.T{90.0, 180.0},
		Max: vec2d.T{-90.0, -180.0},
	}
	for _, path := range p.paths {
		bounds := path.Bounds()
		r.Min[0] = math.Min(r.Min[0], bounds.Min[0])
		r.Min[1] = math.Min(r.Min[1], bounds.Min[1])
		r.Max[0] = math.Max(r.Max[0], bounds.Max[0])
		r.Max[1] = math.Max(r.Max[1], bounds.Max[1])
	}
	for _, area := range p.areas {
		bounds := area.Bounds()
		r.Min[0] = math.Min(r.Min[0], bounds.Min[0])
		r.Min[1] = math.Min(r.Min[1], bounds.Min[1])
		r.Max[0] = math.Max(r.Max[0], bounds.Max[0])
		r.Max[1] = math.Max(r.Max[1], bounds.Max[1])
	}
	for _, marker := range p.points {
		bounds := marker.Bounds()
		r.Min[0] = math.Min(r.Min[0], bounds.Min[0])
		r.Min[1] = math.Min(r.Min[1], bounds.Min[1])
		r.Max[0] = math.Max(r.Max[0], bounds.Max[0])
		r.Max[1] = math.Max(r.Max[1], bounds.Max[1])
	}
	return r
}

func (p *GeoJSONProvider) Srs() geo.Proj {
	return geo.NewProj(4326)
}

func (p *GeoJSONProvider) Attribution() string {
	return ""
}

func (p *GeoJSONProvider) Grid() *geo.TileGrid {
	return &geo.TileGrid{}
}

type KMLProvider struct {
	paths  []*draw.Path
	areas  []*draw.Area
	points []*draw.Marker
}

func NewKMLProvider(filename string) (*KMLProvider, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return NewKMLProviderFromReader(file)
}

func NewKMLProviderFromReader(reader io.Reader) (*KMLProvider, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	provider := &KMLProvider{
		paths:  make([]*draw.Path, 0),
		areas:  make([]*draw.Area, 0),
		points: make([]*draw.Marker, 0),
	}

	xmlDoc := bytes.NewBuffer(data)
	decoder := xml.NewDecoder(xmlDoc)

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		switch se := token.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "Placemark":
				provider.parsePlacemark(decoder)
			}
		}
	}

	return provider, nil
}

func (p *KMLProvider) parsePlacemark(decoder *xml.Decoder) {
	var coords string

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch se := token.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "Point":
				coords = p.parseCoordinates(decoder)
				if len(coords) > 0 {
					parts := strings.Split(strings.TrimSpace(coords), ",")
					if len(parts) >= 2 {
						lng, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
						lat, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
						marker := draw.NewMarker(vec2d.T{lat, lng}, geo.NewProj(4326), color.RGBA{0, 0, 0xff, 0xff}, 10.0)
						p.points = append(p.points, marker)
					}
				}
			case "LineString":
				coords = p.parseCoordinates(decoder)
				path := draw.NewPath(nil, geo.NewProj(4326), color.RGBA{0, 0, 0xff, 0xff}, 2.0)

				if len(coords) > 0 {
					coordPairs := strings.Split(strings.TrimSpace(coords), " ")
					for _, pair := range coordPairs {
						parts := strings.Split(strings.TrimSpace(pair), ",")
						if len(parts) >= 2 {
							lng, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
							lat, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
							path.Positions = append(path.Positions, vec2d.T{lat, lng})
						}
					}
				}

				if len(path.Positions) > 0 {
					p.paths = append(p.paths, path)
				}
			case "Polygon":
				coords = p.parseCoordinates(decoder)
				area := draw.NewArea(nil, geo.NewProj(4326), color.RGBA{0, 0xff, 0, 0x80}, color.RGBA{0, 0, 0, 0}, 2.0)

				if len(coords) > 0 {
					coordPairs := strings.Split(strings.TrimSpace(coords), " ")
					for _, pair := range coordPairs {
						parts := strings.Split(strings.TrimSpace(pair), ",")
						if len(parts) >= 2 {
							lng, _ := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
							lat, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
							area.Positions = append(area.Positions, vec2d.T{lat, lng})
						}
					}
				}

				if len(area.Positions) > 0 {
					p.areas = append(p.areas, area)
				}
			}
		case xml.EndElement:
			if se.Name.Local == "Placemark" {
				return
			}
		}
	}
}

func (p *KMLProvider) parseCoordinates(decoder *xml.Decoder) string {
	for {
		token, err := decoder.Token()
		if err != nil {
			return ""
		}

		switch se := token.(type) {
		case xml.StartElement:
			if se.Name.Local == "coordinates" {
				var val string
				if err := decoder.DecodeElement(&val, &se); err == nil {
					return val
				}
			}
		case xml.EndElement:
			if se.Name.Local == "coordinates" {
				return ""
			}
		}
	}
}

func (p *KMLProvider) GetPaths() []*draw.Path {
	return p.paths
}

func (p *KMLProvider) GetAreas() []*draw.Area {
	return p.areas
}

func (p *KMLProvider) GetPoints() []*draw.Marker {
	return p.points
}

func (p *KMLProvider) Bounds() vec2d.Rect {
	r := vec2d.Rect{
		Min: vec2d.T{90.0, 180.0},
		Max: vec2d.T{-90.0, -180.0},
	}
	for _, path := range p.paths {
		bounds := path.Bounds()
		r.Min[0] = math.Min(r.Min[0], bounds.Min[0])
		r.Min[1] = math.Min(r.Min[1], bounds.Min[1])
		r.Max[0] = math.Max(r.Max[0], bounds.Max[0])
		r.Max[1] = math.Max(r.Max[1], bounds.Max[1])
	}
	for _, area := range p.areas {
		bounds := area.Bounds()
		r.Min[0] = math.Min(r.Min[0], bounds.Min[0])
		r.Min[1] = math.Min(r.Min[1], bounds.Min[1])
		r.Max[0] = math.Max(r.Max[0], bounds.Max[0])
		r.Max[1] = math.Max(r.Max[1], bounds.Max[1])
	}
	for _, marker := range p.points {
		bounds := marker.Bounds()
		r.Min[0] = math.Min(r.Min[0], bounds.Min[0])
		r.Min[1] = math.Min(r.Min[1], bounds.Min[1])
		r.Max[0] = math.Max(r.Max[0], bounds.Max[0])
		r.Max[1] = math.Max(r.Max[1], bounds.Max[1])
	}
	return r
}

func (p *KMLProvider) Srs() geo.Proj {
	return geo.NewProj(4326)
}

func (p *KMLProvider) Attribution() string {
	return ""
}

func (p *KMLProvider) Grid() *geo.TileGrid {
	return &geo.TileGrid{}
}
