package mesh

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/flywave/gg"
	"github.com/flywave/go-geo"
	tile "github.com/flywave/go-static-mesh/tile"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

type TextureSource struct {
	bounds   vec2d.Rect
	srs      geo.Proj
	provider tile.ImageTileProvider
	zoom     int
}

func NewTextureSource(bounds vec2d.Rect, srs geo.Proj, provider tile.ImageTileProvider) *TextureSource {
	return &TextureSource{
		bounds:   bounds,
		srs:      srs,
		provider: provider,
		zoom:     15,
	}
}

func (s *TextureSource) SetZoom(zoom int) {
	s.zoom = zoom
}

func (s *TextureSource) GetTexture() image.Image {
	if s.provider == nil {
		return nil
	}

	tileFetcher := NewTileFetcher(s.provider)
	tiles, err := tileFetcher.FetchTiles(s.bounds, s.zoom, s.srs)
	if err != nil || len(tiles) == 0 {
		return nil
	}

	tileSize := 256
	if len(tiles) > 0 && tiles[0].Image != nil {
		bounds := tiles[0].Image.Bounds()
		tileSize = bounds.Dx()
	}

	merger := NewTileMerger(tileSize)
	mergedImg, err := merger.MergeTiles(tiles, s.bounds)
	if err != nil {
		return nil
	}

	return mergedImg
}

type TileMerger struct {
	tileSize     int
	blendEdges   bool
	blendWidth   int
	blendOverlap int
}

func NewTileMerger(tileSize int) *TileMerger {
	return &TileMerger{
		tileSize:     tileSize,
		blendEdges:   true,
		blendWidth:   8,
		blendOverlap: 2,
	}
}

func (m *TileMerger) SetBlendEdges(enabled bool, width int) {
	m.blendEdges = enabled
	if width > 0 {
		m.blendWidth = width
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
		if tile.Image == nil {
			continue
		}

		x := (tile.X - minX) * m.tileSize
		y := (tile.Y - minY) * m.tileSize

		if m.blendEdges {
			blendedImage := m.applyEdgeBlending(tile.Image, tile.X-minX, tile.Y-minY, gridWidth, gridHeight)
			dc.DrawImage(blendedImage, x, y)
		} else {
			dc.DrawImage(tile.Image, x, y)
		}
	}

	return dc.Image(), nil
}

func (m *TileMerger) applyEdgeBlending(img image.Image, gridX, gridY, gridWidth, gridHeight int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	rgba := image.NewRGBA(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			alphaFactor := 1.0

			if m.blendWidth > 0 {
				leftEdge := x < m.blendWidth
				rightEdge := x >= width-m.blendWidth
				topEdge := y < m.blendWidth
				bottomEdge := y >= height-m.blendWidth

				blendX := 0.0
				if leftEdge {
					blendX = float64(x) / float64(m.blendWidth)
				} else if rightEdge {
					blendX = float64(width-x) / float64(m.blendWidth)
				}

				blendY := 0.0
				if topEdge {
					blendY = float64(y) / float64(m.blendWidth)
				} else if bottomEdge {
					blendY = float64(height-y) / float64(m.blendWidth)
				}

				blendFactor := math.Max(blendX, blendY)
				if blendFactor > 0 {
					alphaFactor = blendFactor
				}
			}

			newAlpha := float64(a) * alphaFactor
			rgba.SetRGBA(x, y, color.RGBA{
				R: uint8(float64(r>>8) * alphaFactor),
				G: uint8(float64(g>>8) * alphaFactor),
				B: uint8(float64(b>>8) * alphaFactor),
				A: uint8(newAlpha / 256),
			})
		}
	}

	return rgba
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
	merger        *TileMerger
	interpolation string
	sharpen       bool
}

func NewTextureGenerator(tileSize int) *TextureGenerator {
	return &TextureGenerator{
		merger:        NewTileMerger(tileSize),
		interpolation: "bilinear",
		sharpen:       false,
	}
}

func (g *TextureGenerator) SetInterpolation(method string) {
	g.interpolation = method
}

func (g *TextureGenerator) SetSharpen(enabled bool) {
	g.sharpen = enabled
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

	switch g.interpolation {
	case "bicubic":
		g.resizeBicubic(src, dst)
	case "lanczos":
		g.resizeLanczos(src, dst)
	default:
		g.resizeBilinear(src, dst)
	}

	if g.sharpen {
		g.sharpenImage(dst)
	}
}

func (g *TextureGenerator) resizeBilinear(src image.Image, dst *image.RGBA) {
	srcW := src.Bounds().Dx()
	srcH := src.Bounds().Dy()
	dstW := dst.Bounds().Dx()
	dstH := dst.Bounds().Dy()

	xRatio := float64(srcW-1) / float64(dstW)
	yRatio := float64(srcH-1) / float64(dstH)

	for y := 0; y < dstH; y++ {
		for x := 0; x < dstW; x++ {
			srcX := float64(x) * xRatio
			srcY := float64(y) * yRatio

			x0 := int(srcX)
			y0 := int(srcY)
			x1 := minInt(x0+1, srcW-1)
			y1 := minInt(y0+1, srcH-1)

			xWeight := srcX - float64(x0)
			yWeight := srcY - float64(y0)

			c00 := getRGBA(src, x0, y0)
			c10 := getRGBA(src, x1, y0)
			c01 := getRGBA(src, x0, y1)
			c11 := getRGBA(src, x1, y1)

			c0 := blendRGBA(c00, c10, xWeight)
			c1 := blendRGBA(c01, c11, xWeight)
			final := blendRGBA(c0, c1, yWeight)

			dst.SetRGBA(x, y, final)
		}
	}
}

func (g *TextureGenerator) resizeBicubic(src image.Image, dst *image.RGBA) {
	srcW := src.Bounds().Dx()
	srcH := src.Bounds().Dy()
	dstW := dst.Bounds().Dx()
	dstH := dst.Bounds().Dy()

	for y := 0; y < dstH; y++ {
		for x := 0; x < dstW; x++ {
			srcX := float64(x) * float64(srcW) / float64(dstW)
			srcY := float64(y) * float64(srcH) / float64(dstH)

			srcX = clampFloat(srcX, 0, float64(srcW-1))
			srcY = clampFloat(srcY, 0, float64(srcH-1))

			r := g.bicubicInterpolate(src, srcX, srcY, 0)
			gComp := g.bicubicInterpolate(src, srcX, srcY, 1)
			b := g.bicubicInterpolate(src, srcX, srcY, 2)

			dst.SetRGBA(x, y, color.RGBA{
				R: uint8(clampFloat(r, 0, 255)),
				G: uint8(clampFloat(gComp, 0, 255)),
				B: uint8(clampFloat(b, 0, 255)),
				A: 255,
			})
		}
	}
}

func (g *TextureGenerator) resizeLanczos(src image.Image, dst *image.RGBA) {
	srcW := src.Bounds().Dx()
	srcH := src.Bounds().Dy()
	dstW := dst.Bounds().Dx()
	dstH := dst.Bounds().Dy()

	a := 3.0

	for y := 0; y < dstH; y++ {
		for x := 0; x < dstW; x++ {
			srcX := float64(x) * float64(srcW-1) / float64(dstW-1)
			srcY := float64(y) * float64(srcH-1) / float64(dstH-1)

			srcX = clampFloat(srcX, 0, float64(srcW-1))
			srcY = clampFloat(srcY, 0, float64(srcH-1))

			r := g.lanczosInterpolate(src, srcX, srcY, 0, a)
			gComp := g.lanczosInterpolate(src, srcX, srcY, 1, a)
			b := g.lanczosInterpolate(src, srcX, srcY, 2, a)

			dst.SetRGBA(x, y, color.RGBA{
				R: uint8(clampFloat(r, 0, 255)),
				G: uint8(clampFloat(gComp, 0, 255)),
				B: uint8(clampFloat(b, 0, 255)),
				A: 255,
			})
		}
	}
}

func (g *TextureGenerator) bicubicInterpolate(img image.Image, x, y float64, channel int) float64 {
	x0 := int(math.Floor(x)) - 1
	y0 := int(math.Floor(y)) - 1

	sum := 0.0
	weightSum := 0.0

	for j := 0; j <= 3; j++ {
		for i := 0; i <= 3; i++ {
			srcX := clampInt(x0+i, 0, img.Bounds().Dx()-1)
			srcY := clampInt(y0+j, 0, img.Bounds().Dy()-1)

			c := getRGBA(img, srcX, srcY)
			var val float64
			switch channel {
			case 0:
				val = float64(c.R)
			case 1:
				val = float64(c.G)
			case 2:
				val = float64(c.B)
			default:
				val = float64(c.A)
			}

			dx := x - float64(srcX)
			dy := y - float64(srcY)
			weight := g.cubicWeight(dx) * g.cubicWeight(dy)

			sum += val * weight
			weightSum += weight
		}
	}

	if weightSum > 0 {
		return sum / weightSum
	}
	return 0
}

func (g *TextureGenerator) lanczosInterpolate(img image.Image, x, y float64, channel int, a float64) float64 {
	radius := int(math.Ceil(a))

	sum := 0.0
	weightSum := 0.0

	for j := -radius; j <= radius; j++ {
		for i := -radius; i <= radius; i++ {
			srcX := clampInt(int(math.Round(x))+i, 0, img.Bounds().Dx()-1)
			srcY := clampInt(int(math.Round(y))+j, 0, img.Bounds().Dy()-1)

			c := getRGBA(img, srcX, srcY)
			var val float64
			switch channel {
			case 0:
				val = float64(c.R)
			case 1:
				val = float64(c.G)
			case 2:
				val = float64(c.B)
			default:
				val = float64(c.A)
			}

			dx := float64(i)
			dy := float64(j)
			weight := g.lanczosWeight(dx, a) * g.lanczosWeight(dy, a)

			sum += val * weight
			weightSum += weight
		}
	}

	if weightSum > 0 {
		return sum / weightSum
	}
	return 0
}

func (g *TextureGenerator) cubicWeight(x float64) float64 {
	absX := math.Abs(x)
	if absX <= 1 {
		return 1.5*absX*absX*absX - 2.5*absX*absX + 1
	}
	if absX <= 2 {
		return -0.5*absX*absX*absX + 2.5*absX*absX - 4*absX + 2
	}
	return 0
}

func (g *TextureGenerator) lanczosWeight(x, a float64) float64 {
	if x == 0 {
		return 1
	}
	if math.Abs(x) >= a {
		return 0
	}
	px := math.Pi * x
	sinc := math.Sin(px) / px
	lanczos := math.Sin(px/a) / (px / a)
	return sinc * lanczos
}

func (g *TextureGenerator) sharpenImage(img *image.RGBA) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	kernel := [9]float32{
		0, -1, 0,
		-1, 5, -1,
		0, -1, 0,
	}

	temp := image.NewRGBA(bounds)

	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			var sumR, sumG, sumB float32

			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					c := img.RGBAAt(x+kx, y+ky)
					weight := kernel[(ky+1)*3+(kx+1)]
					sumR += float32(c.R) * weight
					sumG += float32(c.G) * weight
					sumB += float32(c.B) * weight
				}
			}

			temp.SetRGBA(x, y, color.RGBA{
				R: uint8(clampFloat(float64(sumR), 0, 255)),
				G: uint8(clampFloat(float64(sumG), 0, 255)),
				B: uint8(clampFloat(float64(sumB), 0, 255)),
				A: img.RGBAAt(x, y).A,
			})
		}
	}

	draw.Draw(img, bounds, temp, bounds.Min, draw.Src)
}

func getRGBA(img image.Image, x, y int) color.RGBA {
	if x < 0 || y < 0 || x >= img.Bounds().Dx() || y >= img.Bounds().Dy() {
		return color.RGBA{}
	}

	if rgba, ok := img.(*image.RGBA); ok {
		return rgba.RGBAAt(x, y)
	}

	r, g, b, a := img.At(x, y).RGBA()
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
}

func blendRGBA(c1, c2 color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c1.R)*(1-t) + float64(c2.R)*t),
		G: uint8(float64(c1.G)*(1-t) + float64(c2.G)*t),
		B: uint8(float64(c1.B)*(1-t) + float64(c2.B)*t),
		A: uint8(float64(c1.A)*(1-t) + float64(c2.A)*t),
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func clampInt(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func clampFloat(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func (g *TextureGenerator) DrawGeoObjects(img image.Image, objects interface{}, bounds vec2d.Rect, srs geo.Proj) image.Image {
	return img
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
