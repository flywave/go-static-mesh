package draw

import (
	"fmt"
	"image"
	"log"
	"os"
	"strconv"
	"strings"

	_ "image/jpeg"
	_ "image/png"

	vec2d "github.com/flywave/go3d/float64/vec2"

	"github.com/flywave/gg"
	"github.com/flywave/go-geo"
	"github.com/flywave/go-static-mesh/utils"
)

type ImageMarker struct {
	MapObject
	Position vec2d.T
	Srs      geo.Proj
	Img      image.Image
	OffsetX  float64
	OffsetY  float64
	Height   float64
}

func NewImageMarker(pos vec2d.T, srs geo.Proj, img image.Image, offsetX, offsetY float64) *ImageMarker {
	return NewImageMarkerWithHeight(pos, srs, img, offsetX, offsetY, 10.0)
}

func NewImageMarkerWithHeight(pos vec2d.T, srs geo.Proj, img image.Image, offsetX, offsetY float64, height float64) *ImageMarker {
	m := new(ImageMarker)
	m.Position = pos
	m.Srs = srs
	m.Img = img
	m.OffsetX = offsetX
	m.OffsetY = offsetY
	m.Height = height
	return m
}

func ParseImageMarkerString(s string) ([]*ImageMarker, error) {
	markers := make([]*ImageMarker, 0)

	var img image.Image = nil
	offsetX := 0.0
	offsetY := 0.0
	var height float64 = 10.0
	epsg := int64(4326)

	for _, ss := range strings.Split(s, "|") {
		if ok, suffix := utils.HasPrefix(ss, "image:"); ok {
			file, err := os.Open(suffix)
			if err != nil {
				return nil, err
			}
			defer file.Close()

			img, _, err = image.Decode(file)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "offsetx:"); ok {
			var err error
			offsetX, err = strconv.ParseFloat(suffix, 64)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "offsety:"); ok {
			var err error
			offsetY, err = strconv.ParseFloat(suffix, 64)
			if err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "height:"); ok {
			var err error
			height, err = strconv.ParseFloat(suffix, 64)
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
			if img == nil {
				return nil, fmt.Errorf("cannot create an ImageMarker without an image: %s", s)
			}
			m := NewImageMarkerWithHeight(vec2d.T{lat, lng}, geo.NewProj(int(epsg)), img, offsetX, offsetY, height)
			markers = append(markers, m)
		}
	}
	return markers, nil
}

func (m *ImageMarker) SetImage(img image.Image) {
	m.Img = img
}

func (m *ImageMarker) SetOffsetX(offset float64) {
	m.OffsetX = offset
}

func (m *ImageMarker) SetOffsetY(offset float64) {
	m.OffsetY = offset
}

func (m *ImageMarker) ExtraMarginPixels() (float64, float64, float64, float64) {
	size := m.Img.Bounds().Size()
	return m.OffsetX, m.OffsetY, float64(size.X) - m.OffsetX, float64(size.Y) - m.OffsetY
}

func (m *ImageMarker) Bounds() vec2d.Rect {
	r := vec2d.Rect{Min: vec2d.MaxVal, Max: vec2d.MinVal}
	r.Extend(&m.Position)
	return r
}

func (m *ImageMarker) SrsProj() geo.Proj {
	return m.Srs
}

func (m *ImageMarker) Draw(gc *gg.Context, trans *Transformer) {
	if !CanDisplay(m.Position) {
		log.Printf("ImageMarker coordinates not displayable: %f/%f", m.Position[0], m.Position[1])
		return
	}

	x, y := trans.LatLngToXY(m.Position, m.Srs)
	gc.DrawImage(m.Img, int(x-m.OffsetX), int(y-m.OffsetY))
}

func (m *ImageMarker) ExtrudeToMesh(meshBuilder interface{}, height float64) error {
	return m.ExtrudeToMeshWithTerrain(meshBuilder, height, nil)
}

func (m *ImageMarker) ExtrudeToMeshWithTerrain(meshBuilder interface{}, height float64, terrain TerrainMesh) error {
	if height <= 0 {
		height = m.Height
	}

	if height <= 0 || m.Img == nil {
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

	getBottomZ := func(pos vec2d.T) float64 {
		if terrain != nil {
			terrainHeight := sampleTerrainHeight(pos, terrain)
			if terrainHeight > 0 {
				return terrainHeight
			}
		}
		return 0
	}

	type meshWithVertices interface {
		AppendVertex(x, y, z float64) uint32
		AppendTriangle(a, b, c uint32)
	}

	mesh, ok := meshBuilder.(meshWithVertices)
	if !ok {
		return fmt.Errorf("invalid mesh builder type")
	}

	size := m.Img.Bounds().Size()
	width := float64(size.X)
	imgHeight := float64(size.Y)

	corners := []vec2d.T{
		{m.Position[0], m.Position[1]},
		{m.Position[0] + width, m.Position[1]},
		{m.Position[0] + width, m.Position[1] + imgHeight},
		{m.Position[0], m.Position[1] + imgHeight},
	}

	v0 := mesh.AppendVertex(corners[0][0], corners[0][1], getZ(corners[0]))
	v1 := mesh.AppendVertex(corners[1][0], corners[1][1], getZ(corners[1]))
	v2 := mesh.AppendVertex(corners[2][0], corners[2][1], getZ(corners[2]))
	v3 := mesh.AppendVertex(corners[3][0], corners[3][1], getZ(corners[3]))

	v4 := mesh.AppendVertex(corners[0][0], corners[0][1], getBottomZ(corners[0]))
	v5 := mesh.AppendVertex(corners[1][0], corners[1][1], getBottomZ(corners[1]))
	v6 := mesh.AppendVertex(corners[2][0], corners[2][1], getBottomZ(corners[2]))
	v7 := mesh.AppendVertex(corners[3][0], corners[3][1], getBottomZ(corners[3]))

	mesh.AppendTriangle(v0, v1, v3)
	mesh.AppendTriangle(v1, v2, v3)
	mesh.AppendTriangle(v4, v7, v5)
	mesh.AppendTriangle(v5, v6, v7)

	mesh.AppendTriangle(v4, v0, v1)
	mesh.AppendTriangle(v4, v1, v5)
	mesh.AppendTriangle(v5, v1, v2)
	mesh.AppendTriangle(v5, v2, v6)
	mesh.AppendTriangle(v6, v2, v3)
	mesh.AppendTriangle(v6, v3, v7)
	mesh.AppendTriangle(v7, v3, v0)
	mesh.AppendTriangle(v7, v0, v4)

	return nil
}
