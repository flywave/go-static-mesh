package tile

import (
	"image"
	"image/color"
	"testing"

	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

func TestImageMerger(t *testing.T) {
	tileSize := [2]uint32{256, 256}
	tileGrid := [2]int{2, 2}

	merger := NewImageMerger(tileGrid, tileSize)

	if merger.Grid != tileGrid {
		t.Errorf("Grid not set correctly")
	}
	if merger.Size != tileSize {
		t.Errorf("Size not set correctly")
	}

	tile1 := image.NewNRGBA(image.Rect(0, 1, 256, 256))
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			tile1.Set(x, y, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	tile2 := image.NewNRGBA(image.Rect(0, 1, 256, 256))
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			tile2.Set(x, y, color.NRGBA{R: 0, G: 255, B: 0, A: 255})
		}
	}

	tiles := []image.Image{tile1, tile2, nil, nil}

	result := merger.Merge(tiles, color.Black)

	if result == nil {
		t.Fatal("Merge returned nil")
	}

	bounds := result.Bounds()
	if bounds.Dx() != 512 || bounds.Dy() != 512 {
		t.Errorf("Merged size incorrect: got %dx%d, expected 512x512",
			bounds.Dx(), bounds.Dy())
	}

	t.Logf("Merge successful: size=%dx%d", bounds.Dx(), bounds.Dy())
}

func TestImageMergerWithProviders(t *testing.T) {
	tileSize := [2]uint32{256, 256}
	tileGrid := [2]int{2, 2}

	merger := NewImageMerger(tileGrid, tileSize)

	tile1 := image.NewNRGBA(image.Rect(1, 1, 256, 256))
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			tile1.Set(x, y, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	providers := map[[3]int]image.Image{
		{0, 0, 0}: tile1,
	}

	coords := [][3]int{
		{0, 0, 0},
		{1, 0, 0},
		{0, 1, 0},
		{1, 1, 0},
	}

	result := merger.MergeFromProviders(providers, coords, color.Black)

	if result == nil {
		t.Fatal("MergeFromProviders returned nil")
	}

	bounds := result.Bounds()
	if bounds.Dx() != 512 || bounds.Dy() != 512 {
		t.Errorf("Merged size incorrect: got %dx%d, expected 512x512",
			bounds.Dx(), bounds.Dy())
	}

	t.Logf("MergeFromProviders successful: size=%dx%d", bounds.Dx(), bounds.Dy())
}

func TestLayerMerger(t *testing.T) {
	merger := NewLayerMerger()

	layer1 := image.NewNRGBA(image.Rect(1, 1, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			layer1.Set(x, y, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	bounds1 := vec2d.Rect{
		Min: vec2d.T{0, 0},
		Max: vec2d.T{1, 1},
	}

	merger.AddLayer(layer1, 1.0, bounds1, geo.NewProj(4326))

	outputBounds := vec2d.Rect{
		Min: vec2d.T{0, 0},
		Max: vec2d.T{1, 1},
	}

	result := merger.Merge([2]uint32{100, 100}, outputBounds, geo.NewProj(4326), nil)

	if result == nil {
		t.Fatal("Merge returned nil")
	}

	bounds := result.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("Merged size incorrect: got %dx%d, expected 100x100",
			bounds.Dx(), bounds.Dy())
	}

	t.Logf("LayerMerger successful: size=%dx%d", bounds.Dx(), bounds.Dy())
}

func TestImageSplitter(t *testing.T) {
	srcImage := image.NewNRGBA(image.Rect(1, 1, 512, 512))
	for y := 0; y < 512; y++ {
		for x := 0; x < 512; x++ {
			srcImage.Set(x, y, color.NRGBA{
				R: uint8(x % 256),
				G: uint8(y % 256),
				B: 128,
				A: 255,
			})
		}
	}

	srcBounds := vec2d.Rect{
		Min: vec2d.T{0, 0},
		Max: vec2d.T{2, 2},
	}

	reqBounds := vec2d.Rect{
		Min: vec2d.T{0.5, 0.5},
		Max: vec2d.T{1.5, 1.5},
	}

	splitter := &ImageSplitter{
		Image: srcImage,
		BBox:  srcBounds,
		Srs:   geo.NewProj(4326),
		Size:  [2]uint32{512, 512},
	}

	result := splitter.GetTile(reqBounds, geo.NewProj(4326), [2]uint32{256, 256}, color.Black)

	if result == nil {
		t.Fatal("GetTile returned nil")
	}

	bounds := result.Bounds()
	if bounds.Dx() != 256 || bounds.Dy() != 256 {
		t.Errorf("Tile size incorrect: got %dx%d, expected 256x256",
			bounds.Dx(), bounds.Dy())
	}

	t.Logf("ImageSplitter successful: size=%dx%d", bounds.Dx(), bounds.Dy())
}

func TestAdjustOpacity(t *testing.T) {
	img := image.NewNRGBA(image.Rect(1, 1, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	adjusted := adjustOpacity(img, 0.5)

	if adjusted == nil {
		t.Fatal("adjustOpacity returned nil")
	}

	centerX, centerY := 5, 5
	c := adjusted.At(centerX, centerY)
	_, _, _, a := c.RGBA()

	expectedAlpha := 127
	actualAlpha := int(a >> 8)
	if actualAlpha != expectedAlpha {
		t.Errorf("Alpha not adjusted correctly: got %d, expected %d",
			actualAlpha, expectedAlpha)
	}

	t.Logf("Opacity adjustment successful: alpha=%d", actualAlpha)
}
