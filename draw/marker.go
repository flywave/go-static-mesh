package draw

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"strconv"
	"strings"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/gg"
	"github.com/flywave/go-geo"
	"github.com/flywave/go-static-mesh/utils"
)

type Marker struct {
	MapObject
	Position     vec2d.T
	Srs          geo.Proj
	Color        color.Color
	Size         float64
	Label        string
	LabelColor   color.Color
	LabelXOffset float64
	LabelYOffset float64
	Height       float64
	TipOffset    float64
}

func NewMarker(pos vec2d.T, srs geo.Proj, col color.Color, size float64) *Marker {
	return NewMarkerWithHeight(pos, srs, col, size, 10.0)
}

func NewMarkerWithHeight(pos vec2d.T, srs geo.Proj, col color.Color, size float64, height float64) *Marker {
	return NewMarkerWithTipOffset(pos, srs, col, size, height, 0.1)
}

func NewMarkerWithTipOffset(pos vec2d.T, srs geo.Proj, col color.Color, size float64, height float64, tipOffset float64) *Marker {
	m := new(Marker)
	m.Position = pos
	m.Srs = srs
	m.Color = col
	m.Size = size
	m.Label = ""
	if Luminance(m.Color) >= 0.5 {
		m.LabelColor = color.RGBA{0x00, 0x00, 0x00, 0xff}
	} else {
		m.LabelColor = color.RGBA{0xff, 0xff, 0xff, 0xff}
	}
	m.LabelXOffset = 0.5
	m.LabelYOffset = 0.5
	m.Height = height
	m.TipOffset = tipOffset
	return m
}

func parseSizeString(s string) (float64, error) {
	switch {
	case s == "mid":
		return 16.0, nil
	case s == "small":
		return 12.0, nil
	case s == "tiny":
		return 8.0, nil
	}

	if floatValue, err := strconv.ParseFloat(s, 64); err == nil && floatValue > 0 {
		return floatValue, nil
	}

	return 0.0, fmt.Errorf("cannot parse size string: '%s'", s)
}

func ParseLabelOffset(s string) (float64, error) {
	if floatValue, err := strconv.ParseFloat(s, 64); err == nil && floatValue > 0 {
		return floatValue, nil
	}

	return 0.5, fmt.Errorf("cannot parse label offset: '%s'", s)
}

func ParseMarkerString(s string) ([]*Marker, error) {
	markers := make([]*Marker, 0)

	var markerColor color.Color = color.RGBA{0xff, 0, 0, 0xff}
	size := 16.0
	label := ""
	labelXOffset := 0.5
	labelYOffset := 0.5
	height := 10.0
	tipOffset := 0.1
	epsg := int64(4326)

	var labelColor color.Color

	for _, ss := range strings.Split(s, "|") {
		if ok, suffix := utils.HasPrefix(ss, "color:"); ok {
			var err error
			markerColor, err = ParseColorString(suffix)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "label:"); ok {
			label = suffix
		} else if ok, suffix := utils.HasPrefix(ss, "size:"); ok {
			var err error
			size, err = parseSizeString(suffix)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "labelcolor:"); ok {
			var err error
			labelColor, err = ParseColorString(suffix)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "labelxoffset:"); ok {
			var err error
			labelXOffset, err = ParseLabelOffset(suffix)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "labelyoffset:"); ok {
			var err error
			labelYOffset, err = ParseLabelOffset(suffix)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "height:"); ok {
			var err error
			height, err = strconv.ParseFloat(suffix, 64)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "tipoffset:"); ok {
			var err error
			tipOffset, err = strconv.ParseFloat(suffix, 64)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "epsg:"); ok {
			var err error
			epsg, err = strconv.ParseInt(suffix, 10, 64)
			if err != nil {
				return nil, err
			}
		} else {
			lat, lng, err := ParseLatLon(ss)
			if err != nil {
				return nil, err
			}
			m := NewMarkerWithTipOffset(vec2d.T{lat, lng}, geo.NewProj(int(epsg)), markerColor, size, height, tipOffset)
			m.Label = label
			if labelColor != nil {
				m.SetLabelColor(labelColor)
			}
			m.LabelXOffset = labelXOffset
			m.LabelYOffset = labelYOffset
			markers = append(markers, m)
		}
	}
	return markers, nil
}

func (m *Marker) SetLabelColor(col color.Color) {
	m.LabelColor = col
}

func (m *Marker) ExtraMarginPixels() (float64, float64, float64, float64) {
	return 0.5*m.Size + 1.0, 1.5*m.Size + 1.0, 0.5*m.Size + 1.0, 1.0
}

func (m *Marker) Bounds() vec2d.Rect {
	r := vec2d.Rect{Min: vec2d.MaxVal, Max: vec2d.MinVal}
	r.Extend(&m.Position)
	return r
}

func (m *Marker) SrsProj() geo.Proj {
	return m.Srs
}

func (m *Marker) Draw(gc *gg.Context, trans *Transformer) {
	if !CanDisplay(m.Position) {
		log.Printf("Marker coordinates not displayable: %f/%f", m.Position[0], m.Position[0])
		return
	}

	gc.ClearPath()
	gc.SetLineJoin(gg.LineJoinRound)
	gc.SetLineWidth(1.0)

	radius := 0.5 * m.Size
	x, y := trans.LatLngToXY(m.Position, m.Srs)
	gc.DrawArc(x, y-m.Size, radius, (90.0+60.0)*math.Pi/180.0, (360.0+90.0-60.0)*math.Pi/180.0)
	gc.LineTo(x, y)
	gc.ClosePath()
	gc.SetColor(m.Color)
	gc.FillPreserve()
	gc.SetRGB(0, 0, 0)
	gc.Stroke()

	if m.Label != "" {
		gc.SetColor(m.LabelColor)
		gc.DrawStringAnchored(m.Label, x, y-m.Size, m.LabelXOffset, m.LabelYOffset)
	}
}

func (m *Marker) ExtrudeToMesh(meshBuilder interface{}, height float64) error {
	return m.ExtrudeToMeshWithTerrain(meshBuilder, height, nil)
}

func (m *Marker) ExtrudeToMeshWithTerrain(meshBuilder interface{}, height float64, terrain TerrainMesh) error {
	if height <= 0 {
		height = m.Height
	}

	if height <= 0 || m.Size <= 0 {
		return nil
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
		AppendVertex(x, y, z float64) uint32
		AppendTriangle(a, b, c uint32)
	}

	mesh, ok := meshBuilder.(meshWithVertices)
	if !ok {
		return fmt.Errorf("invalid mesh builder type")
	}

	arcSegments := 32
	arcRadius := m.Size / 2.0
	arcCenterY := m.Size
	startAngle := (90.0 + 60.0) * math.Pi / 180.0
	endAngle := (360.0 + 90.0 - 60.0) * math.Pi / 180.0

	topVertices := make([]uint32, arcSegments+2)
	bottomVertices := make([]uint32, arcSegments+2)

	for i := 0; i <= arcSegments; i++ {
		angle := startAngle + float64(i)/float64(arcSegments)*(endAngle-startAngle)
		x := arcRadius * math.Cos(angle)
		y := arcCenterY + arcRadius*math.Sin(angle)

		pos := vec2d.T{m.Position[0] + x, m.Position[1] + y}
		topVertices[i] = mesh.AppendVertex(pos[0], pos[1], getZ(pos))
		bottomVertices[i] = mesh.AppendVertex(pos[0], pos[1], 0)
	}

	tipIndex := arcSegments + 1
	topTipZ := getZ(m.Position)
	bottomTipZ := -m.TipOffset
	topVertices[tipIndex] = mesh.AppendVertex(m.Position[0], m.Position[1], topTipZ)
	bottomVertices[tipIndex] = mesh.AppendVertex(m.Position[0], m.Position[1], bottomTipZ)

	for i := 0; i < arcSegments; i++ {
		topTipIdx := topVertices[tipIndex]
		bottomTipIdx := bottomVertices[tipIndex]
		if i == 0 {
			angle := startAngle + float64(0)/float64(arcSegments)*(endAngle-startAngle)
			x := arcRadius * math.Cos(angle)
			y := arcCenterY + arcRadius*math.Sin(angle)
			pos := vec2d.T{m.Position[0] + x, m.Position[1] + y}
			topTipIdx = mesh.AppendVertex(pos[0], pos[1], getZ(pos))
			bottomTipIdx = mesh.AppendVertex(pos[0], pos[1], bottomTipZ)
		}
		mesh.AppendTriangle(topTipIdx, topVertices[i], topVertices[i+1])
		mesh.AppendTriangle(bottomTipIdx, bottomVertices[i+1], bottomVertices[i])
	}

	for i := 0; i <= arcSegments; i++ {
		next := (i + 1) % (arcSegments + 1)

		mesh.AppendTriangle(topVertices[i], topVertices[next], bottomVertices[i])
		mesh.AppendTriangle(topVertices[next], bottomVertices[next], bottomVertices[i])
	}

	return nil
}
