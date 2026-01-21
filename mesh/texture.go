package mesh

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/flywave/gg"
	"github.com/flywave/go-geo"
	drawpkg "github.com/flywave/go-static-mesh/draw"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

type TextureSource struct {
	bounds vec2d.Rect
	srs    interface{}
}

func NewTextureSource(bounds vec2d.Rect, srs interface{}) *TextureSource {
	return &TextureSource{
		bounds: bounds,
		srs:    srs,
	}
}

func (s *TextureSource) GetTexture() image.Image {
	return nil
}

type TileMerger struct {
	tileSize int
}

func NewTileMerger(tileSize int) *TileMerger {
	return &TileMerger{
		tileSize: tileSize,
	}
}

func (m *TileMerger) MergeTiles(tiles []*TextureTile, bounds vec2d.Rect) (image.Image, error) {
	if len(tiles) == 0 {
		return nil, nil
	}

	if len(tiles) == 1 {
		return tiles[0].Image, nil
	}

	minX, maxX := math.MaxInt, math.MinInt
	minY, maxY := math.MaxInt, math.MinInt

	for _, tile := range tiles {
		if tile.X < minX {
			minX = tile.X
		}
		if tile.X > maxX {
			maxX = tile.X
		}
		if tile.Y < minY {
			minY = tile.Y
		}
		if tile.Y > maxY {
			maxY = tile.Y
		}
	}

	gridWidth := maxX - minX + 1
	gridHeight := maxY - minY + 1

	imgWidth := gridWidth * m.tileSize
	imgHeight := gridHeight * m.tileSize

	result := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	dc := gg.NewContextForRGBA(result)

	for _, tile := range tiles {
		x := (tile.X - minX) * m.tileSize
		y := (tile.Y - minY) * m.tileSize
		dc.DrawImage(tile.Image, x, y)
	}

	return dc.Image(), nil
}

type TextureOptions struct {
	Resolution  int
	Format      string
	Transparent bool
	Opacity     float64
	Background  color.Color
}

func DefaultTextureOptions() *TextureOptions {
	return &TextureOptions{
		Resolution:  2048,
		Format:      "png",
		Transparent: false,
		Opacity:     1.0,
		Background:  color.White,
	}
}

type TextureGenerator struct {
	merger *TileMerger
}

func NewTextureGenerator(tileSize int) *TextureGenerator {
	return &TextureGenerator{
		merger: NewTileMerger(tileSize),
	}
}

func (g *TextureGenerator) Generate(tiles []*TextureTile, bounds vec2d.Rect, options *TextureOptions) (image.Image, error) {
	if options == nil {
		options = DefaultTextureOptions()
	}

	mergedImg, err := g.merger.MergeTiles(tiles, bounds)
	if err != nil {
		return nil, err
	}

	if mergedImg == nil {
		return nil, nil
	}

	result := image.NewRGBA(image.Rect(0, 0, options.Resolution, options.Resolution))

	if !options.Transparent && options.Background != nil {
		draw.Draw(result, result.Bounds(), &image.Uniform{options.Background}, image.Point{}, draw.Src)
	}

	boundsW := bounds.Max[0] - bounds.Min[0]
	boundsH := bounds.Max[1] - bounds.Min[1]
	aspect := boundsW / boundsH

	width := options.Resolution
	height := int(float64(options.Resolution) / aspect)

	if height > options.Resolution {
		height = options.Resolution
		width = int(float64(options.Resolution) * aspect)
	}

	resized := image.NewRGBA(image.Rect(0, 0, width, height))
	g.resizeImage(mergedImg, resized)

	if width != options.Resolution || height != options.Resolution {
		finalImg := image.NewRGBA(image.Rect(0, 0, options.Resolution, options.Resolution))
		draw.Draw(finalImg, finalImg.Bounds(), resized, image.Point{}, draw.Src)
		return finalImg, nil
	}

	return resized, nil
}

func (g *TextureGenerator) CropToBounds(img image.Image, srcBounds, dstBounds vec2d.Rect, srcSrs, dstSrs interface{}) image.Image {
	if srcSrs == dstSrs {
		return img
	}

	imgRect := img.Bounds()
	imgWidth := imgRect.Dx()
	imgHeight := imgRect.Dy()

	srcW := srcBounds.Max[0] - srcBounds.Min[0]
	srcH := srcBounds.Max[1] - srcBounds.Min[1]

	relMinX := (dstBounds.Min[0] - srcBounds.Min[0]) / srcW
	relMinY := (dstBounds.Max[1] - srcBounds.Max[1]) / srcH
	relMaxX := (dstBounds.Max[0] - srcBounds.Min[0]) / srcW
	relMaxY := (dstBounds.Min[1] - srcBounds.Min[1]) / srcH

	cropMinX := int(relMinX * float64(imgWidth))
	cropMinY := int(relMinY * float64(imgHeight))
	cropMaxX := int(relMaxX * float64(imgWidth))
	cropMaxY := int(relMaxY * float64(imgHeight))

	if cropMinX < 0 {
		cropMinX = 0
	}
	if cropMinY < 0 {
		cropMinY = 0
	}
	if cropMaxX > imgWidth {
		cropMaxX = imgWidth
	}
	if cropMaxY > imgHeight {
		cropMaxY = imgHeight
	}

	cropRect := image.Rect(cropMinX, cropMinY, cropMaxX, cropMaxY)

	if cropRect.Dx() <= 0 || cropRect.Dy() <= 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	cropped := image.NewRGBA(cropRect)
	draw.Draw(cropped, cropped.Bounds(), img, cropRect.Min, draw.Src)

	return cropped
}

func (g *TextureGenerator) CalculateTextureBounds(vertices []vec3d.T) vec2d.Rect {
	if len(vertices) == 0 {
		return vec2d.Rect{}
	}

	minX, minY := vertices[0][0], vertices[0][1]
	maxX, maxY := vertices[0][0], vertices[0][1]

	for _, v := range vertices {
		if v[0] < minX {
			minX = v[0]
		}
		if v[0] > maxX {
			maxX = v[0]
		}
		if v[1] < minY {
			minY = v[1]
		}
		if v[1] > maxY {
			maxY = v[1]
		}
	}

	return vec2d.Rect{
		Min: vec2d.T{minX, minY},
		Max: vec2d.T{maxX, maxY},
	}
}

func (g *TextureGenerator) resizeImage(src image.Image, dst *image.RGBA) {
	srcW := src.Bounds().Dx()
	srcH := src.Bounds().Dy()
	dstW := dst.Bounds().Dx()
	dstH := dst.Bounds().Dy()

	if srcW == dstW && srcH == dstH {
		draw.Draw(dst, dst.Bounds(), src, image.Point{}, draw.Src)
		return
	}

	dc := gg.NewContextForRGBA(dst)
	dc.Scale(float64(dstW)/float64(srcW), float64(dstH)/float64(srcH))
	dc.DrawImage(src, 0, 0)
}

func (g *TextureGenerator) DrawGeoObjects(img image.Image, objects []drawpkg.MapObject, bounds vec2d.Rect, srs geo.Proj) image.Image {
	if len(objects) == 0 {
		return img
	}

	dc := gg.NewContextForImage(img)
	dc.DrawImage(img, 0, 0)

	return dc.Image()
}

func (g *TextureGenerator) CalculateUVs(vertices []vec3d.T, bounds vec2d.Rect) []vec2d.T {
	if len(vertices) == 0 {
		return nil
	}

	uvs := make([]vec2d.T, len(vertices))

	boundsW := bounds.Max[0] - bounds.Min[0]
	boundsH := bounds.Max[1] - bounds.Min[1]

	for i, v := range vertices {
		u := (v[0] - bounds.Min[0]) / boundsW
		vCoord := (v[1] - bounds.Min[1]) / boundsH
		uvs[i] = vec2d.T{u, vCoord}
	}

	return uvs
}
