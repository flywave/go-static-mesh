package draw

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"strconv"
	"strings"

	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"

	"github.com/flywave/gg"
	"github.com/flywave/go-geo"
	"github.com/flywave/go-static-mesh/utils"
)

type BillboardMode int

const (
	BillboardModePrint BillboardMode = iota
	BillboardModeDisplay
)

type Billboard struct {
	MapObject
	Position     vec2d.T
	Srs          geo.Proj
	Text         string
	Width        float64
	Height       float64
	Thickness    float64
	Rotation     float64
	TextDepth    float64
	FontSize     float64
	Mode         BillboardMode
	FontParser   FontParser
	Color        color.Color
	Background   color.Color
	FontPath     string
	TextureDPI   int
	TextureScale float64
}

func NewBillboard(pos vec2d.T, srs geo.Proj, text string) *Billboard {
	return NewBillboardWithSize(pos, srs, text, 10.0, 5.0, 0.5)
}

func NewBillboardWithSize(pos vec2d.T, srs geo.Proj, text string, width, height, thickness float64) *Billboard {
	b := &Billboard{
		Position:     pos,
		Srs:          srs,
		Text:         text,
		Width:        width,
		Height:       height,
		Thickness:    thickness,
		Rotation:     0,
		TextDepth:    0.3,
		FontSize:     height * 0.6,
		Mode:         BillboardModeDisplay,
		FontParser:   nil,
		Color:        color.RGBA{0x00, 0x00, 0x00, 0xff},
		Background:   color.RGBA{0xff, 0xff, 0xff, 0xff},
		FontPath:     "",
		TextureDPI:   300,
		TextureScale: 1.0,
	}
	return b
}

func ParseBillboardString(s string) (*Billboard, error) {
	billboard := &Billboard{
		Text:       "",
		Width:      10.0,
		Height:     5.0,
		Thickness:  0.5,
		Rotation:   0,
		TextDepth:  0.3,
		Mode:       BillboardModeDisplay,
		Color:      color.RGBA{0x00, 0x00, 0x00, 0xff},
		Background: color.RGBA{0xff, 0xff, 0xff, 0xff},
	}

	for _, ss := range strings.Split(s, "|") {
		if ok, suffix := utils.HasPrefix(ss, "text:"); ok {
			billboard.Text = suffix
		} else if ok, suffix := utils.HasPrefix(ss, "width:"); ok {
			var err error
			if billboard.Width, err = strconv.ParseFloat(suffix, 64); err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "height:"); ok {
			var err error
			if billboard.Height, err = strconv.ParseFloat(suffix, 64); err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "thickness:"); ok {
			var err error
			if billboard.Thickness, err = strconv.ParseFloat(suffix, 64); err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "rotation:"); ok {
			var err error
			if billboard.Rotation, err = strconv.ParseFloat(suffix, 64); err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "textdepth:"); ok {
			var err error
			if billboard.TextDepth, err = strconv.ParseFloat(suffix, 64); err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "fontsize:"); ok {
			var err error
			if billboard.FontSize, err = strconv.ParseFloat(suffix, 64); err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "mode:"); ok {
			if suffix == "print" {
				billboard.Mode = BillboardModePrint
			} else {
				billboard.Mode = BillboardModeDisplay
			}
		} else if ok, suffix := utils.HasPrefix(ss, "color:"); ok {
			var err error
			if billboard.Color, err = ParseColorString(suffix); err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "background:"); ok {
			var err error
			if billboard.Background, err = ParseColorString(suffix); err != nil {
				return nil, err
			}
		} else if ok, suffix := utils.HasPrefix(ss, "epsg:"); ok {
			epsg, err := strconv.ParseInt(suffix, 10, 64)
			if err != nil {
				return nil, err
			}
			billboard.Srs = geo.NewProj(int(epsg))
		} else {
			lat, lng, err := ParseLatLon(ss)
			if err != nil {
				return nil, err
			}
			billboard.Position = vec2d.T{lat, lng}
		}
	}

	if billboard.Srs == nil {
		billboard.Srs = geo.NewProj("EPSG:4326")
	}

	if billboard.FontSize == 0 {
		billboard.FontSize = billboard.Height * 0.6
	}

	return billboard, nil
}

func (b *Billboard) SetRotation(angle float64) {
	b.Rotation = angle
}

func (b *Billboard) SetMode(mode BillboardMode) {
	b.Mode = mode
}

func (b *Billboard) SetFontParser(parser FontParser) {
	b.FontParser = parser
}

func (b *Billboard) SetTextDepth(depth float64) {
	b.TextDepth = depth
}

func (b *Billboard) SetFontSize(size float64) {
	b.FontSize = size
}

func (b *Billboard) SetColor(col color.Color) {
	b.Color = col
}

func (b *Billboard) SetBackground(col color.Color) {
	b.Background = col
}

func (b *Billboard) SetFontPath(path string) {
	b.FontPath = path
}

func (b *Billboard) SetTextureDPI(dpi int) {
	b.TextureDPI = dpi
}

func (b *Billboard) SetTextureScale(scale float64) {
	b.TextureScale = scale
}

func (b *Billboard) ExtraMarginPixels() (float64, float64, float64, float64) {
	margin := math.Max(b.Width, b.Height) / 2.0
	return margin, margin, margin, margin
}

func (b *Billboard) Bounds() vec2d.Rect {
	corners := b.GetRotatedCorners()
	r := vec2d.Rect{Min: vec2d.MaxVal, Max: vec2d.MinVal}
	for _, corner := range corners {
		r.Extend(&corner)
	}
	return r
}

func (b *Billboard) SrsProj() geo.Proj {
	return b.Srs
}

func (b *Billboard) GetRotatedCorners() [4]vec2d.T {
	halfW := b.Width / 2.0
	halfH := b.Height / 2.0

	corners := [4]vec2d.T{
		{-halfW, -halfH},
		{halfW, -halfH},
		{halfW, halfH},
		{-halfW, halfH},
	}

	rad := b.Rotation * math.Pi / 180.0
	cosR := math.Cos(rad)
	sinR := math.Sin(rad)

	for i := range corners {
		x := corners[i][0]
		y := corners[i][1]
		corners[i][0] = x*cosR - y*sinR
		corners[i][1] = x*sinR + y*cosR

		corners[i][0] += b.Position[0]
		corners[i][1] += b.Position[1]
	}

	return corners
}

func (b *Billboard) GetRotated3DCorners(baseZ, topZ float64) [8]vec3d.T {
	halfW := b.Width / 2.0
	halfH := b.Height / 2.0

	corners := [8]vec3d.T{
		{-halfW, -halfH, baseZ},
		{halfW, -halfH, baseZ},
		{halfW, halfH, baseZ},
		{-halfW, halfH, baseZ},
		{-halfW, -halfH, topZ},
		{halfW, -halfH, topZ},
		{halfW, halfH, topZ},
		{-halfW, halfH, topZ},
	}

	rad := b.Rotation * math.Pi / 180.0
	cosR := math.Cos(rad)
	sinR := math.Sin(rad)

	for i := range corners {
		y := corners[i][1]
		z := corners[i][2]
		corners[i][1] = y*cosR - z*sinR
		corners[i][2] = y*sinR + z*cosR

		corners[i][0] += b.Position[0]
		corners[i][1] += b.Position[1]
	}

	return corners
}

func (b *Billboard) Draw(gc *gg.Context, trans *Transformer) {
	if b.Text == "" {
		return
	}

	corners := b.GetRotatedCorners()

	gc.ClearPath()
	gc.SetLineWidth(1.0)

	screenCorners := make([]vec2d.T, 4)
	for i, corner := range corners {
		x, y := trans.LatLngToXY(corner, b.Srs)
		screenCorners[i] = vec2d.T{x, y}
	}

	gc.MoveTo(screenCorners[0][0], screenCorners[0][1])
	for i := 1; i < 4; i++ {
		gc.LineTo(screenCorners[i][0], screenCorners[i][1])
	}
	gc.ClosePath()

	gc.SetColor(b.Background)
	gc.FillPreserve()
	gc.SetColor(b.Color)
	gc.Stroke()

	centerX := (screenCorners[0][0] + screenCorners[2][0]) / 2.0
	centerY := (screenCorners[0][1] + screenCorners[2][1]) / 2.0

	gc.SetColor(b.Color)
	gc.DrawStringAnchored(b.Text, centerX, centerY, 0.5, 0.5)
}

func (b *Billboard) ExtrudeToMesh(meshBuilder interface{}, height float64) error {
	return b.ExtrudeToMeshWithTerrain(meshBuilder, height, nil)
}

func (b *Billboard) ExtrudeToMeshWithTerrain(meshBuilder interface{}, height float64, terrain TerrainMesh) error {
	type meshWithVertices interface {
		AppendVertex(x, y, z float64) uint32
		AppendTriangle(a, b, c uint32)
		GetVertices() interface{}
	}

	var vertices []vec3d.T
	var indices []uint32
	var appendVertex func(x, y, z float64) uint32
	var appendTriangle func(a, b, c uint32)

	if m, ok := meshBuilder.(meshWithVertices); ok {
		appendVertex = m.AppendVertex
		appendTriangle = m.AppendTriangle
	} else if m, ok := meshBuilder.(interface {
		GetVertices() interface{}
	}); ok {
		if v := m.GetVertices(); v != nil {
			if verts, ok := v.([]vec3d.T); ok {
				vertices = verts
				indices = []uint32{}
				appendVertex = func(x, y, z float64) uint32 {
					vertices = append(vertices, vec3d.T{x, y, z})
					return uint32(len(vertices) - 1)
				}
				appendTriangle = func(a, b, c uint32) {
					indices = append(indices, a, b, c)
				}
			}
		}
	}

	if appendVertex == nil || appendTriangle == nil {
		return fmt.Errorf("invalid mesh builder type")
	}

	corners := b.GetRotatedCorners()

	baseHeight := 0.0
	if terrain != nil {
		heights := make([]float64, 4)
		for i, corner := range corners {
			heights[i] = sampleTerrainHeight(corner, terrain)
		}
		minHeight := heights[0]
		for _, h := range heights {
			if h < minHeight {
				minHeight = h
			}
		}
		baseHeight = minHeight
	}

	rotated3D := b.GetRotated3DCorners(0, b.Thickness)

	minRotatedZ := rotated3D[0][2]
	for i := 1; i < 8; i++ {
		if rotated3D[i][2] < minRotatedZ {
			minRotatedZ = rotated3D[i][2]
		}
	}

	vertexIndices := make([]uint32, 8)
	for i := 0; i < 8; i++ {
		z := rotated3D[i][2] - minRotatedZ + baseHeight
		vertexIndices[i] = appendVertex(rotated3D[i][0], rotated3D[i][1], z)
	}

	appendTriangle(vertexIndices[0], vertexIndices[1], vertexIndices[2])
	appendTriangle(vertexIndices[0], vertexIndices[2], vertexIndices[3])

	appendTriangle(vertexIndices[6], vertexIndices[5], vertexIndices[4])
	appendTriangle(vertexIndices[7], vertexIndices[6], vertexIndices[4])

	appendTriangle(vertexIndices[0], vertexIndices[4], vertexIndices[1])
	appendTriangle(vertexIndices[1], vertexIndices[4], vertexIndices[5])

	appendTriangle(vertexIndices[1], vertexIndices[5], vertexIndices[2])
	appendTriangle(vertexIndices[2], vertexIndices[5], vertexIndices[6])

	appendTriangle(vertexIndices[2], vertexIndices[6], vertexIndices[3])
	appendTriangle(vertexIndices[3], vertexIndices[6], vertexIndices[7])

	appendTriangle(vertexIndices[3], vertexIndices[7], vertexIndices[0])
	appendTriangle(vertexIndices[0], vertexIndices[7], vertexIndices[4])

	return nil
}

func (b *Billboard) DrawToTexture(dc *gg.Context, trans *Transformer) image.Image {
	dc.Clear()
	b.Draw(dc, trans)
	return dc.Image()
}

func (b *Billboard) CreateDisplayTexture() (image.Image, error) {
	if b.Text == "" {
		return nil, fmt.Errorf("text is empty")
	}

	dpi := b.TextureDPI
	if dpi <= 0 {
		dpi = 300
	}

	scale := b.TextureScale
	if scale <= 0 {
		scale = 1.0
	}

	textureWidth := int(float64(512) * scale)
	textureHeight := int(float64(256) * scale)

	if b.Width > 0 && b.Height > 0 {
		aspectRatio := b.Width / b.Height
		if aspectRatio > 2.0 {
			textureWidth = int(float64(1024) * scale)
			textureHeight = int(float64(512) * scale)
		} else if aspectRatio < 0.5 {
			textureWidth = int(float64(256) * scale)
			textureHeight = int(float64(512) * scale)
		}
	}

	textureWidth = clampInt(textureWidth, 64, 4096)
	textureHeight = clampInt(textureHeight, 64, 4096)

	if textureWidth <= 0 || textureHeight <= 0 {
		return nil, fmt.Errorf("invalid texture dimensions: %dx%d", textureWidth, textureHeight)
	}

	dc := gg.NewContext(textureWidth, textureHeight)

	dc.SetColor(b.Background)
	dc.Clear()

	fontSize := float64(textureHeight) * 0.5

	var fontPath string
	if b.FontPath != "" {
		fontPath = b.FontPath
	} else {
		fontPath = findSystemFont()
		if fontPath == "" {
			return nil, fmt.Errorf("no default font available, please specify FontPath")
		}
	}

	if err := dc.LoadFontFace(fontPath, fontSize); err != nil {
		return nil, fmt.Errorf("failed to load font: %w", err)
	}

	dc.SetColor(b.Color)

	textWidth, textHeight := dc.MeasureString(b.Text)

	centerX := float64(textureWidth) / 2.0
	centerY := float64(textureHeight) / 2.0

	scaleX := float64(textureWidth) * 0.9 / textWidth
	scaleY := float64(textureHeight) * 0.9 / textHeight
	scaleFactor := math.Min(scaleX, scaleY)

	if scaleFactor < 1.0 {
		fontSize = fontSize * scaleFactor
		if err := dc.LoadFontFace(fontPath, fontSize); err != nil {
			return nil, fmt.Errorf("failed to load font with scaled size: %w", err)
		}
	}

	dc.DrawStringAnchored(b.Text, centerX, centerY, 0.5, 0.5)

	return dc.Image(), nil
}

func (b *Billboard) calculateTextureFontSizeInPoints() float64 {
	fontSizeInMeters := b.FontSize
	if fontSizeInMeters <= 0 {
		fontSizeInMeters = b.Height * 0.6
	}

	inch := 0.0254
	fontSizeInPoints := (fontSizeInMeters / inch) * 72.0

	return fontSizeInPoints
}

func findSystemFont() string {
	fonts := []string{
		"fonts/DejaVuSans.ttf",
		"../fonts/DejaVuSans.ttf",
		"../../fonts/DejaVuSans.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
		"/usr/share/fonts/TTF/DejaVuSans.ttf",
		"C:\\Windows\\Fonts\\arial.ttf",
		"/System/Library/Fonts/Helvetica.ttc",
	}

	for _, font := range fonts {
		if _, err := os.Stat(font); err == nil {
			return font
		}
	}

	return ""
}

func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (b *Billboard) calculateTextureFontSize() float64 {
	dpi := b.TextureDPI
	if dpi <= 0 {
		dpi = 300
	}

	inch := 0.0254
	fontSizeInMeters := b.FontSize
	if fontSizeInMeters <= 0 {
		fontSizeInMeters = b.Height * 0.6
	}

	fontSizeInPixels := (fontSizeInMeters / inch) * float64(dpi)

	return fontSizeInPixels
}
