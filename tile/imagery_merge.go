package tile

import (
	"image"
	"image/color"
	"math"

	"github.com/flywave/gg"
	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
	"github.com/flywave/imaging"
)

type ImageMerger struct {
	Grid [2]int
	Size [2]uint32
}

func NewImageMerger(tileGrid [2]int, tileSize [2]uint32) *ImageMerger {
	return &ImageMerger{
		Grid: tileGrid,
		Size: tileSize,
	}
}

func (m *ImageMerger) Merge(tiles []image.Image, backgroundColor color.Color) image.Image {
	if m.Grid[0] == 1 && m.Grid[1] == 1 {
		if len(tiles) >= 1 && tiles[0] != nil {
			return tiles[0]
		}
	}

	srcSize := m.srcSize()
	result := image.NewNRGBA(image.Rect(0, 0, int(srcSize[0]), int(srcSize[1])))

	if backgroundColor != nil {
		for y := 0; y < int(srcSize[1]); y++ {
			for x := 0; x < int(srcSize[0]); x++ {
				result.Set(x, y, backgroundColor)
			}
		}
	}

	dc := gg.NewContextForImage(result)

	for i, tile := range tiles {
		if tile == nil {
			continue
		}

		pos := m.tileOffset(i)
		resized := imaging.Resize(tile, int(m.Size[0]), int(m.Size[1]), imaging.Lanczos)
		dc.DrawImage(resized, pos[0], pos[1])
	}

	return dc.Image()
}

func (m *ImageMerger) MergeFromProviders(providers map[[3]int]image.Image, coords [][3]int, backgroundColor color.Color) image.Image {
	tiles := make([]image.Image, len(coords))

	for i, coord := range coords {
		img, exists := providers[coord]
		if !exists {
			continue
		}
		tiles[i] = img
	}

	return m.Merge(tiles, backgroundColor)
}

func (m *ImageMerger) srcSize() [2]uint32 {
	width := uint32(m.Grid[0]) * m.Size[0]
	height := uint32(m.Grid[1]) * m.Size[1]
	return [2]uint32{width, height}
}

func (m *ImageMerger) tileOffset(i int) [2]int {
	x := int(math.Mod(float64(i), float64(m.Grid[0]))) * int(m.Size[0])
	y := int(math.Floor(float64(i)/float64(m.Grid[0]))) * int(m.Size[1])
	return [2]int{x, y}
}

type LayerMerger struct {
	Layers    []image.Image
	Opacities []float64
	Bounds    []vec2d.Rect
	Srs       []geo.Proj
}

func NewLayerMerger() *LayerMerger {
	return &LayerMerger{
		Layers:    make([]image.Image, 0),
		Opacities: make([]float64, 0),
		Bounds:    make([]vec2d.Rect, 0),
		Srs:       make([]geo.Proj, 0),
	}
}

func (l *LayerMerger) AddLayer(img image.Image, opacity float64, bounds vec2d.Rect, srs geo.Proj) {
	l.Layers = append(l.Layers, img)
	l.Opacities = append(l.Opacities, opacity)
	l.Bounds = append(l.Bounds, bounds)
	l.Srs = append(l.Srs, srs)
}

func (l *LayerMerger) Merge(outputSize [2]uint32, outputBounds vec2d.Rect, outputSrs geo.Proj, backgroundColor color.Color) image.Image {
	result := image.NewNRGBA(image.Rect(0, 0, int(outputSize[0]), int(outputSize[1])))

	if backgroundColor != nil {
		for y := 0; y < int(outputSize[1]); y++ {
			for x := 0; x < int(outputSize[0]); x++ {
				result.Set(x, y, backgroundColor)
			}
		}
	}

	dc := gg.NewContextForImage(result)

	for i, layer := range l.Layers {
		if layer == nil {
			continue
		}

		layerBounds := l.Bounds[i]
		opacity := l.Opacities[i]
		layerSrs := l.Srs[i]

		if !isRectEmpty(layerBounds) && !isRectEmpty(outputBounds) {
			if layerSrs != nil && outputSrs != nil && !layerSrs.Eq(outputSrs) {
				layerBounds = layerSrs.TransformRectTo(outputSrs, layerBounds, 16)
			}

			offset := calculateImageOffset(layerBounds, outputBounds, outputSize)

			if opacity < 1.0 {
				layer = adjustOpacity(layer, opacity)
			}

			resized := imaging.Resize(layer, int(outputSize[0]), int(outputSize[1]), imaging.Lanczos)
			dc.DrawImage(resized, offset[0], offset[1])
		} else {
			if opacity < 1.0 {
				layer = adjustOpacity(layer, opacity)
			}
			resized := imaging.Resize(layer, int(outputSize[0]), int(outputSize[1]), imaging.Lanczos)
			dc.DrawImage(resized, 0, 0)
		}
	}

	return dc.Image()
}

func calculateImageOffset(srcBounds, dstBounds vec2d.Rect, dstSize [2]uint32) [2]int {
	if isRectEmpty(srcBounds) || isRectEmpty(dstBounds) {
		return [2]int{0, 0}
	}

	facX := (dstBounds.Min[0] - srcBounds.Min[0]) / (srcBounds.Max[0] - srcBounds.Min[0])
	facY := (srcBounds.Max[1] - dstBounds.Max[1]) / (srcBounds.Max[1] - srcBounds.Min[1])

	return [2]int{
		int(facX * float64(dstSize[0])),
		int(facY * float64(dstSize[1])),
	}
}

func adjustOpacity(img image.Image, opacity float64) *image.NRGBA {
	bounds := img.Bounds()
	result := image.NewNRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			newAlpha := uint8(float64(a>>8) * opacity)
			result.SetNRGBA(x, y, color.NRGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: newAlpha,
			})
		}
	}

	return result
}

type TiledImage struct {
	Tiles    []image.Image
	TileGrid [2]int
	TileSize [2]uint32
	BBox     vec2d.Rect
	Srs      geo.Proj
}

func NewTiledImage(tiles []image.Image, tileGrid [2]int, tileSize [2]uint32, bbox vec2d.Rect, srs geo.Proj) *TiledImage {
	return &TiledImage{
		Tiles:    tiles,
		TileGrid: tileGrid,
		TileSize: tileSize,
		BBox:     bbox,
		Srs:      srs,
	}
}

func (t *TiledImage) GetMergedImage(backgroundColor color.Color) image.Image {
	merger := NewImageMerger(t.TileGrid, t.TileSize)
	return merger.Merge(t.Tiles, backgroundColor)
}

func (t *TiledImage) Resample(reqBounds vec2d.Rect, reqSrs geo.Proj, outSize [2]uint32, backgroundColor color.Color) image.Image {
	merged := t.GetMergedImage(backgroundColor)

	splitter := &ImageSplitter{
		Image: merged,
		BBox:  t.BBox,
		Srs:   t.Srs,
		Size:  [2]uint32{uint32(merged.Bounds().Dx()), uint2uint(merged.Bounds().Dy())},
	}

	return splitter.GetTile(reqBounds, reqSrs, outSize, backgroundColor)
}

type ImageSplitter struct {
	Image image.Image
	BBox  vec2d.Rect
	Srs   geo.Proj
	Size  [2]uint32
}

func (s *ImageSplitter) GetTile(reqBounds vec2d.Rect, reqSrs geo.Proj, outSize [2]uint32, backgroundColor color.Color) image.Image {
	srcBounds := s.BBox
	if s.Srs != nil && reqSrs != nil && !s.Srs.Eq(reqSrs) {
		srcBounds = s.Srs.TransformRectTo(reqSrs, srcBounds, 16)
	}

	offset := calculateImageOffset(srcBounds, reqBounds, s.Size)

	minX := offset[0]
	minY := offset[1]
	maxX := minX + int(outSize[0])
	maxY := minY + int(outSize[1])

	bounds := s.Image.Bounds()

	minX = max(0, minX)
	minY = max(0, minY)
	maxX = min(bounds.Dx(), maxX)
	maxY = min(bounds.Dy(), maxY)

	if maxX <= minX || maxY <= minY {
		result := image.NewNRGBA(image.Rect(0, 0, int(outSize[0]), int(outSize[1])))
		if backgroundColor != nil {
			for y := 0; y < int(outSize[1]); y++ {
				for x := 0; x < int(outSize[0]); x++ {
					result.Set(x, y, backgroundColor)
				}
			}
		}
		return result
	}

	cropped := imaging.Crop(s.Image, image.Rect(minX, minY, maxX, maxY))

	result := image.NewNRGBA(image.Rect(0, 0, int(outSize[0]), int(outSize[1])))
	if backgroundColor != nil {
		for y := 0; y < int(outSize[1]); y++ {
			for x := 0; x < int(outSize[0]); x++ {
				result.Set(x, y, backgroundColor)
			}
		}
	}

	dc := gg.NewContextForImage(result)
	drawX := max(0, -offset[0])
	drawY := max(0, -offset[1])
	dc.DrawImage(cropped, drawX, drawY)

	return dc.Image()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func uint2uint(v int) uint32 {
	if v < 0 {
		return 0
	}
	return uint32(v)
}
