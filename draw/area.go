package draw

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/gg"
	"github.com/flywave/go-geo"
	"github.com/flywave/go-static-mesh/utils"
	"github.com/flywave/go-tesselator"
)

type Area struct {
	MapObject
	Positions []vec2d.T
	Srs       geo.Proj
	Color     color.Color
	Fill      color.Color
	Weight    float64
	Height    float64
}

func NewArea(positions []vec2d.T, srs geo.Proj, col color.Color, fill color.Color, weight float64) *Area {
	return NewAreaWithHeight(positions, srs, col, fill, weight, 10.0)
}

func NewAreaWithHeight(positions []vec2d.T, srs geo.Proj, col color.Color, fill color.Color, weight float64, height float64) *Area {
	a := new(Area)
	a.Positions = positions
	a.Srs = srs
	a.Color = col
	a.Fill = fill
	a.Weight = weight
	a.Height = height
	return a
}

func ParseAreaString(s string) (*Area, error) {
	area := new(Area)
	area.Color = color.RGBA{0xff, 0, 0, 0xff}
	area.Fill = color.Transparent
	area.Weight = 5.0
	area.Height = 10.0

	for _, ss := range strings.Split(s, "|") {
		if ok, suffix := utils.HasPrefix(ss, "color:"); ok {
			var err error
			area.Color, err = ParseColorString(suffix)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "fill:"); ok {
			var err error
			area.Fill, err = ParseColorString(suffix)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "weight:"); ok {
			var err error
			area.Weight, err = strconv.ParseFloat(suffix, 64)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "height:"); ok {
			var err error
			area.Height, err = strconv.ParseFloat(suffix, 64)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "epsg:"); ok {
			epsg, err := strconv.ParseInt(suffix, 10, 64)
			if err != nil {
				return nil, err
			}
			area.Srs = geo.NewProj(int(epsg))
		} else {
			lat, lng, err := ParseLatLon(ss)
			if err != nil {
				return nil, err
			}
			area.Positions = append(area.Positions, vec2d.T{lat, lng})
		}
	}

	if area.Srs == nil {
		area.Srs = geo.NewProj("EPSG:4326")
	}
	return area, nil
}

func (p *Area) ExtraMarginPixels() (float64, float64, float64, float64) {
	return p.Weight, p.Weight, p.Weight, p.Weight
}

func (p *Area) Bounds() vec2d.Rect {
	r := vec2d.Rect{Min: vec2d.MaxVal, Max: vec2d.MinVal}
	for _, ll := range p.Positions {
		r.Extend(&ll)
	}
	return r
}

func (p *Area) SrsProj() geo.Proj {
	return p.Srs
}

func (p *Area) Draw(gc *gg.Context, trans *Transformer) {
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
	gc.ClosePath()
	gc.SetColor(p.Fill)
	gc.FillPreserve()
	gc.SetColor(p.Color)
	gc.Stroke()
}

func (a *Area) ExtrudeToMesh(meshBuilder interface{}, height float64) error {
	if len(a.Positions) < 3 {
		return nil
	}

	if height <= 0 {
		height = a.Height
	}

	type meshWithVertices interface {
		AppendVertex(x, y, z float64) uint32
		AppendTriangle(a, b, c uint32)
	}

	m, ok := meshBuilder.(meshWithVertices)
	if !ok {
		return fmt.Errorf("invalid mesh builder type")
	}

	topIndices, _, err := tesselator.Tesselate([]tesselator.Contour{
		makeContourFromPositions(a.Positions, height),
	}, tesselator.WindingRuleOdd)
	if err != nil {
		return fmt.Errorf("failed to triangulate top area: %w", err)
	}

	topVertexStart := uint32(0)
	for _, pos := range a.Positions {
		m.AppendVertex(pos[0], pos[1], height)
	}

	for i := 0; i < len(topIndices); i += 3 {
		m.AppendTriangle(topVertexStart+uint32(topIndices[i]), topVertexStart+uint32(topIndices[i+1]), topVertexStart+uint32(topIndices[i+2]))
	}

	bottomIndices, _, err := tesselator.Tesselate([]tesselator.Contour{
		makeContourFromPositions(a.Positions, 0),
	}, tesselator.WindingRuleOdd)
	if err != nil {
		return fmt.Errorf("failed to triangulate bottom area: %w", err)
	}

	bottomVertexStart := uint32(len(a.Positions))
	for _, pos := range a.Positions {
		m.AppendVertex(pos[0], pos[1], 0)
	}

	for i := 0; i < len(bottomIndices); i += 3 {
		idx0 := bottomVertexStart + uint32(bottomIndices[i])
		idx1 := bottomVertexStart + uint32(bottomIndices[i+1])
		idx2 := bottomVertexStart + uint32(bottomIndices[i+2])
		m.AppendTriangle(idx2, idx1, idx0)
	}

	for i := 0; i < len(a.Positions); i++ {
		next := (i + 1) % len(a.Positions)

		topIdx := topVertexStart + uint32(i)
		topNextIdx := topVertexStart + uint32(next)
		bottomIdx := bottomVertexStart + uint32(i)
		bottomNextIdx := bottomVertexStart + uint32(next)

		m.AppendTriangle(bottomIdx, topIdx, bottomNextIdx)
		m.AppendTriangle(bottomNextIdx, topIdx, topNextIdx)
	}

	return nil
}

func makeContourFromPositions(positions []vec2d.T, height float64) tesselator.Contour {
	contour := make(tesselator.Contour, len(positions))
	for i, pos := range positions {
		contour[i] = tesselator.Vertex{
			X: float32(pos[0]),
			Y: float32(pos[1]),
			Z: float32(height),
		}
	}
	return contour
}

func (a *Area) DrawToTexture(dc *gg.Context, trans *Transformer) image.Image {
	dc.Clear()
	a.Draw(dc, trans)
	return dc.Image()
}
