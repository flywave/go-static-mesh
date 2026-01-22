package draw

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/gg"
	"github.com/flywave/go-geo"
	"github.com/flywave/go-gpx"
	"github.com/flywave/go-static-mesh/utils"
)

type Path struct {
	MapObject
	Positions []vec2d.T
	Srs       geo.Proj
	Color     color.Color
	Weight    float64
	Height    float64
}

func NewPath(positions []vec2d.T, srs geo.Proj, col color.Color, weight float64) *Path {
	return NewPathWithHeight(positions, srs, col, weight, 10.0)
}

func NewPathWithHeight(positions []vec2d.T, srs geo.Proj, col color.Color, weight float64, height float64) *Path {
	p := new(Path)
	p.Positions = positions
	p.Color = col
	p.Weight = weight
	p.Height = height
	p.Srs = srs

	return p
}

func ParsePathString(s string) ([]*Path, error) {
	paths := make([]*Path, 0)
	currentPath := new(Path)
	currentPath.Color = color.RGBA{0xff, 0, 0, 0xff}
	currentPath.Weight = 5.0
	currentPath.Height = 10.0

	for _, ss := range strings.Split(s, "|") {
		if ok, suffix := utils.HasPrefix(ss, "color:"); ok {
			var err error
			if currentPath.Color, err = ParseColorString(suffix); err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "weight:"); ok {
			var err error
			if currentPath.Weight, err = strconv.ParseFloat(suffix, 64); err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "height:"); ok {
			var err error
			if currentPath.Height, err = strconv.ParseFloat(suffix, 64); err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "gpx:"); ok {
			gpxData, err := gpx.Read(bytes.NewBufferString(suffix))
			if err != nil {
				return nil, err
			}
			for _, trk := range gpxData.Trk {
				for _, seg := range trk.TrkSeg {
					p := new(Path)
					p.Color = currentPath.Color
					p.Weight = currentPath.Weight
					for _, pt := range seg.TrkPt {
						p.Positions = append(p.Positions, vec2d.T{pt.Lat, pt.Lon})
					}
					if len(p.Positions) > 0 {
						paths = append(paths, p)
					}
				}
			}
		} else if ok, suffix := utils.HasPrefix(ss, "epsg:"); ok {
			epsg, err := strconv.ParseInt(suffix, 10, 64)
			if err != nil {
				return nil, err
			}
			currentPath.Srs = geo.NewProj(int(epsg))
		} else {
			lat, lng, err := ParseLatLon(ss)
			if err != nil {
				return nil, err
			}
			currentPath.Positions = append(currentPath.Positions, vec2d.T{lat, lng})
		}
	}

	if currentPath.Srs == nil {
		currentPath.Srs = geo.NewProj("EPSG:4326")
	}

	if len(currentPath.Positions) > 0 {
		paths = append(paths, currentPath)
	}
	return paths, nil
}

func (p *Path) ExtraMarginPixels() (float64, float64, float64, float64) {
	return p.Weight, p.Weight, p.Weight, p.Weight
}

func (p *Path) Bounds() vec2d.Rect {
	r := vec2d.Rect{Min: vec2d.MaxVal, Max: vec2d.MinVal}
	for _, ll := range p.Positions {
		r.Extend(&ll)
	}
	return r
}

func (m *Path) SrsProj() geo.Proj {
	return m.Srs
}

func (p *Path) Draw(gc *gg.Context, trans *Transformer) {
	if len(p.Positions) <= 1 {
		return
	}

	gc.ClearPath()
	gc.SetLineWidth(p.Weight)
	gc.SetLineCap(gg.LineCapRound)
	gc.SetLineJoin(gg.LineJoinRound)
	for _, ll := range p.Positions {
		gc.LineTo(trans.LatLngToXY(ll, p.Srs))
	}
	gc.SetColor(p.Color)
	gc.Stroke()
}

func (p *Path) ExtrudeToMesh(meshBuilder interface{}, height float64) error {
	return p.ExtrudeToMeshWithTerrain(meshBuilder, height, nil)
}

func (p *Path) ExtrudeToMeshWithTerrain(meshBuilder interface{}, height float64, terrain TerrainMesh) error {
	if len(p.Positions) < 2 {
		return nil
	}

	if height <= 0 {
		height = p.Height
	}

	getZ := func(pos vec2d.T) float64 {
		if terrain != nil {
			terrainHeight := sampleTerrainHeight(pos, terrain)
			if terrainHeight > 0 {
				return terrainHeight + height
			}
		}
		return height
	}

	type meshWithVertices interface {
		Vertices() interface{}
		AppendVertex(x, y, z float64) uint32
		AppendTriangle(a, b, c uint32)
	}

	m, ok := meshBuilder.(meshWithVertices)
	if !ok {
		return fmt.Errorf("invalid mesh builder type")
	}

	for i := 0; i < len(p.Positions)-1; i++ {
		start := p.Positions[i]
		end := p.Positions[i+1]

		dx := end[0] - start[0]
		dy := end[1] - start[1]
		length := math.Sqrt(dx*dx + dy*dy)

		if length < 0.0001 {
			continue
		}

		perpX := -dy / length
		perpY := dx / length

		offsetX := (p.Weight / 2.0) * perpX
		offsetY := (p.Weight / 2.0) * perpY

		v0 := m.AppendVertex(start[0]+offsetX, start[1]+offsetY, getZ(start))
		v1 := m.AppendVertex(end[0]+offsetX, end[1]+offsetY, getZ(end))
		v2 := m.AppendVertex(end[0]-offsetX, end[1]-offsetY, getZ(end))
		v3 := m.AppendVertex(start[0]-offsetX, start[1]-offsetY, getZ(start))

		v4 := m.AppendVertex(start[0]+offsetX, start[1]+offsetY, 0)
		v5 := m.AppendVertex(end[0]+offsetX, end[1]+offsetY, 0)
		v6 := m.AppendVertex(end[0]-offsetX, end[1]-offsetY, 0)
		v7 := m.AppendVertex(start[0]-offsetX, start[1]-offsetY, 0)

		m.AppendTriangle(v0, v1, v3)
		m.AppendTriangle(v1, v2, v3)
		m.AppendTriangle(v4, v7, v5)
		m.AppendTriangle(v7, v6, v5)

		m.AppendTriangle(v0, v4, v1)
		m.AppendTriangle(v1, v4, v5)
		m.AppendTriangle(v1, v5, v2)
		m.AppendTriangle(v2, v5, v6)
		m.AppendTriangle(v2, v6, v3)
		m.AppendTriangle(v3, v6, v7)
		m.AppendTriangle(v3, v7, v0)
	}

	return nil
}

func (p *Path) DrawToTexture(dc *gg.Context, trans *Transformer) image.Image {
	dc.Clear()
	p.Draw(dc, trans)
	return dc.Image()
}
