package tile

import (
	"testing"

	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

func TestRasterMerger(t *testing.T) {
	tileSize := [2]uint32{256, 256}
	tileGrid := [2]int{2, 2}

	merger := NewRasterMerger(tileGrid, tileSize)

	if merger.Grid != tileGrid {
		t.Errorf("Grid not set correctly")
	}
	if merger.Size != tileSize {
		t.Errorf("Size not set correctly")
	}

	tile1 := &ElevationGrid{
		Width:  256,
		Height: 256,
		Data:   make([]float64, 256*256),
		Bounds: vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{1, 1}},
		Srs:    geo.NewProj(4326),
	}

	for i := range tile1.Data {
		tile1.Data[i] = float64(i)
	}

	tile2 := &ElevationGrid{
		Width:  256,
		Height: 256,
		Data:   make([]float64, 256*256),
		Bounds: vec2d.Rect{Min: vec2d.T{1, 0}, Max: vec2d.T{2, 1}},
		Srs:    geo.NewProj(4326),
	}

	for i := range tile2.Data {
		tile2.Data[i] = float64(i) * 2
	}

	tiles := []*ElevationGrid{tile1, tile2, nil, nil}

	result := merger.Merge(tiles, BORDER_NONE)

	if result == nil {
		t.Fatal("Merge returned nil")
	}

	expectedWidth := uint32(256 * 2)
	expectedHeight := uint32(256 * 2)
	if result.Size[0] != expectedWidth || result.Size[1] != expectedHeight {
		t.Errorf("Merged size incorrect: got [%d, %d], expected [%d, %d]",
			result.Size[0], result.Size[1], expectedWidth, expectedHeight)
	}

	if len(result.Datas) != int(expectedWidth*expectedHeight) {
		t.Errorf("Data length incorrect: got %d, expected %d",
			len(result.Datas), expectedWidth*expectedHeight)
	}

	if result.Box.Min[0] != 0 || result.Box.Max[0] != 2 {
		t.Errorf("Merged bounds incorrect: got %v", result.Box)
	}

	t.Logf("Merge successful: size=[%d,%d], data_len=%d",
		result.Size[0], result.Size[1], len(result.Datas))
}

func TestRasterMergerWithBorders(t *testing.T) {
	tileSize := [2]uint32{256, 256}
	tileGrid := [2]int{2, 2}

	merger := NewRasterMerger(tileGrid, tileSize)

	tile1 := &ElevationGrid{
		Width:  256,
		Height: 256,
		Data:   make([]float64, 256*256),
		Bounds: vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{1, 1}},
		Srs:    geo.NewProj(4326),
	}

	for i := range tile1.Data {
		tile1.Data[i] = 100.0
	}

	tiles := []*ElevationGrid{tile1, nil, nil, nil}

	result := merger.Merge(tiles, BORDER_BILATERAL)

	if result == nil {
		t.Fatal("Merge returned nil")
	}

	if !result.IsBilateral() {
		t.Error("Result should be bilateral")
	}

	if len(result.LeftBorder) != int(result.Size[1]) {
		t.Errorf("Left border length incorrect: got %d, expected %d",
			len(result.LeftBorder), result.Size[1])
	}

	t.Logf("Merge with borders successful: bilateral=%v", result.IsBilateral())
}

func TestTileOffset(t *testing.T) {
	tileSize := [2]uint32{256, 256}
	tileGrid := [2]int{4, 4}

	merger := NewRasterMerger(tileGrid, tileSize)

	tests := []struct {
		index    int
		expected [2]int
	}{
		{0, [2]int{0, 0}},
		{1, [2]int{256, 0}},
		{2, [2]int{512, 0}},
		{3, [2]int{768, 0}},
		{4, [2]int{0, 256}},
		{5, [2]int{256, 256}},
		{15, [2]int{768, 768}},
	}

	for _, tt := range tests {
		result := merger.tileOffset(tt.index)
		if result != tt.expected {
			t.Errorf("tileOffset(%d) = %v, expected %v", tt.index, result, tt.expected)
		}
	}
}

func TestMergeFromProviders(t *testing.T) {
	t.Skip("Requires actual GeoTIFF files")
}

func TestTiledRaster(t *testing.T) {
	tileSize := [2]uint32{256, 256}
	tileGrid := [2]int{2, 2}

	tiles := make([]*ElevationGrid, 4)
	for i := range tiles {
		tiles[i] = &ElevationGrid{
			Width:  256,
			Height: 256,
			Data:   make([]float64, 256*256),
			Bounds: vec2d.Rect{
				Min: vec2d.T{float64(i % 2), float64(i / 2)},
				Max: vec2d.T{float64(i%2 + 1), float64(i/2 + 1)},
			},
			Srs: geo.NewProj(4326),
		}
		for j := range tiles[i].Data {
			tiles[i].Data[j] = float64(i*1000 + j)
		}
	}

	tr := NewTiledRaster(tiles, tileGrid, tileSize)

	result := tr.GetMergedRaster(BORDER_NONE)

	if result == nil {
		t.Fatal("GetMergedRaster returned nil")
	}

	if result.Size[0] != 512 || result.Size[1] != 512 {
		t.Errorf("Merged size incorrect: got [%d, %d], expected [512, 512]",
			result.Size[0], result.Size[1])
	}

	t.Logf("TiledRaster merge successful: size=[%d,%d]", result.Size[0], result.Size[1])
}

func TestJoinRects(t *testing.T) {
	empty := vec2d.Rect{}

	r1 := vec2d.Rect{
		Min: vec2d.T{0, 0},
		Max: vec2d.T{1, 1},
	}

	r2 := vec2d.Rect{
		Min: vec2d.T{1, 0},
		Max: vec2d.T{2, 1},
	}

	result1 := joinRects(empty, r1)
	if result1 != r1 {
		t.Error("joinRects with empty should return r1")
	}

	result2 := joinRects(r1, empty)
	if result2 != r1 {
		t.Error("joinRects with empty should return r1")
	}

	result3 := joinRects(r1, r2)
	expected := vec2d.Rect{
		Min: vec2d.T{0, 0},
		Max: vec2d.T{2, 1},
	}
	if result3 != expected {
		t.Errorf("joinRects incorrect: got %v, expected %v", result3, expected)
	}
}
