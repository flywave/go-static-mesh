package mesh

import (
	"image"
	"image/color"
	"testing"

	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

type mockImageTileProvider struct{}

func (m *mockImageTileProvider) GetImageTile(coord [3]int) (image.Image, error) {
	return nil, nil
}

func (m *mockImageTileProvider) Attribution() string {
	return ""
}

func (m *mockImageTileProvider) Grid() *geo.TileGrid {
	return nil
}

func (m *mockImageTileProvider) Bounds() vec2d.Rect {
	return vec2d.Rect{}
}

func (m *mockImageTileProvider) Srs() geo.Proj {
	return geo.NewProj(4326)
}

func TestNewTextureSource(t *testing.T) {
	bounds := vec2d.Rect{
		Min: vec2d.T{0, 0},
		Max: vec2d.T{1, 1},
	}
	srs := geo.NewProj(4326)
	provider := &mockImageTileProvider{}
	source := NewTextureSource(bounds, srs, provider)

	if source == nil {
		t.Error("NewTextureSource returned nil")
	}

	if source.GetTexture() != nil {
		t.Error("GetTexture should return nil initially")
	}
}

func TestDefaultTextureOptions(t *testing.T) {
	opts := DefaultTextureOptions()

	if opts == nil {
		t.Fatal("DefaultTextureOptions returned nil")
	}

	if opts.Resolution != 2048 {
		t.Errorf("Expected resolution 2048, got %d", opts.Resolution)
	}

	if opts.Format != "png" {
		t.Errorf("Expected format png, got %s", opts.Format)
	}

	if opts.Transparent != false {
		t.Error("Expected Transparent to be false")
	}

	if opts.Opacity != 1.0 {
		t.Errorf("Expected opacity 1.0, got %f", opts.Opacity)
	}
}

func TestNewTileMerger(t *testing.T) {
	merger := NewTileMerger(256)

	if merger == nil {
		t.Error("NewTileMerger returned nil")
	}

	if merger.tileSize != 256 {
		t.Errorf("Expected tileSize 256, got %d", merger.tileSize)
	}
}

func TestTileMergerMergeTiles(t *testing.T) {
	merger := NewTileMerger(256)

	tiles := []*TextureTile{}

	img, err := merger.MergeTiles(tiles, vec2d.Rect{})
	if err != nil {
		t.Errorf("MergeTiles returned error: %v", err)
	}

	if img != nil {
		t.Error("MergeTiles should return nil for empty tiles")
	}

	singleTile := &TextureTile{
		Coord: [3]int{0, 0, 1},
		Image: image.NewRGBA(image.Rect(0, 0, 256, 256)),
		X:     0,
		Y:     0,
	}

	tiles = append(tiles, singleTile)

	img, err = merger.MergeTiles(tiles, vec2d.Rect{})
	if err != nil {
		t.Errorf("MergeTiles returned error for single tile: %v", err)
	}

	if img == nil {
		t.Error("MergeTiles should return image for single tile")
	}
}

func TestTextureGeneratorCalculateTextureBounds(t *testing.T) {
	generator := NewTextureGenerator(256)

	vertices := []vec3d.T{
		{0, 0, 0},
		{1, 0, 0},
		{1, 1, 0},
		{0, 1, 0},
	}

	bounds := generator.CalculateTextureBounds(vertices)

	if bounds.Min[0] != 0 || bounds.Min[1] != 0 {
		t.Errorf("Expected Min {0, 0}, got %v", bounds.Min)
	}

	if bounds.Max[0] != 1 || bounds.Max[1] != 1 {
		t.Errorf("Expected Max {1, 1}, got %v", bounds.Max)
	}

	emptyVertices := []vec3d.T{}
	emptyBounds := generator.CalculateTextureBounds(emptyVertices)

	if emptyBounds.Min[0] != 0 || emptyBounds.Max[0] != 0 {
		t.Error("Expected empty bounds for empty vertices")
	}
}

func TestTextureGeneratorCalculateUVs(t *testing.T) {
	generator := NewTextureGenerator(256)

	bounds := vec2d.Rect{
		Min: vec2d.T{0, 0},
		Max: vec2d.T{1, 1},
	}

	vertices := []vec3d.T{
		{0, 0, 0},
		{0.5, 0.5, 0},
		{1, 1, 0},
	}

	uvs := generator.CalculateUVs(vertices, bounds)

	if len(uvs) != len(vertices) {
		t.Errorf("Expected %d UVs, got %d", len(vertices), len(uvs))
	}

	expectedUVs := []vec2d.T{
		{0, 0},
		{0.5, 0.5},
		{1, 1},
	}

	for i, uv := range uvs {
		if uv[0] != expectedUVs[i][0] || uv[1] != expectedUVs[i][1] {
			t.Errorf("UV %d: expected %v, got %v", i, expectedUVs[i], uv)
		}
	}

	emptyVertices := []vec3d.T{}
	emptyUVs := generator.CalculateUVs(emptyVertices, bounds)

	if emptyUVs != nil {
		t.Error("Expected nil UVs for empty vertices")
	}
}

func TestTextureGeneratorCropToBounds(t *testing.T) {
	generator := NewTextureGenerator(256)

	img := image.NewRGBA(image.Rect(0, 0, 512, 512))

	srcBounds := vec2d.Rect{
		Min: vec2d.T{0, 0},
		Max: vec2d.T{2, 2},
	}

	dstBounds := vec2d.Rect{
		Min: vec2d.T{0, 0},
		Max: vec2d.T{1, 1},
	}

	cropped := generator.CropToBounds(img, srcBounds, dstBounds, nil, nil)

	if cropped == nil {
		t.Error("CropToBounds returned nil")
	}

	if cropped.Bounds().Dx() > img.Bounds().Dx() || cropped.Bounds().Dy() > img.Bounds().Dy() {
		t.Error("Cropped image should be smaller or equal to original")
	}
}

func TestTextureGeneratorResizeImage(t *testing.T) {
	generator := NewTextureGenerator(256)

	src := image.NewRGBA(image.Rect(0, 0, 512, 512))
	dst := image.NewRGBA(image.Rect(0, 0, 256, 256))

	src.Set(100, 100, color.RGBA{255, 0, 0, 255})

	generator.resizeImage(src, dst)

	if dst.Bounds().Dx() != 256 || dst.Bounds().Dy() != 256 {
		t.Errorf("Expected destination size 256x256, got %dx%d", dst.Bounds().Dx(), dst.Bounds().Dy())
	}
}
