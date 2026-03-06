package tile

import (
	"math"

	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

type RasterMerger struct {
	Grid    [2]int
	Size    [2]uint32
	BBox    vec2d.Rect
	BBoxSrs geo.Proj
}

func NewRasterMerger(tileGrid [2]int, tileSize [2]uint32) *RasterMerger {
	return &RasterMerger{
		Grid: tileGrid,
		Size: tileSize,
	}
}

func isRectEmpty(r vec2d.Rect) bool {
	return r.Min[0] == 0 && r.Min[1] == 0 && r.Max[0] == 0 && r.Max[1] == 0
}

func (m *RasterMerger) Merge(tiles []*ElevationGrid, borderMode BorderMode) *TileData {
	srcSize := m.srcSize()
	tileData := NewTileData(srcSize, borderMode)

	bbox := m.BBox
	var bboxSrs geo.Proj = m.BBoxSrs

	for i, grid := range tiles {
		if grid == nil {
			continue
		}

		pos := m.tileOffset(i)

		srcTileData := grid.ToTileData()
		if srcTileData == nil {
			continue
		}

		m.copyTileData(tileData, srcTileData, pos)

		if !isRectEmpty(grid.Bounds) {
			if isRectEmpty(bbox) {
				bbox = grid.Bounds
			} else {
				bbox = joinRects(bbox, grid.Bounds)
			}
			if grid.Srs != nil {
				bboxSrs = grid.Srs
			}
		}
	}

	tileData.Box = bbox
	tileData.Boxsrs = bboxSrs

	return tileData
}

func (m *RasterMerger) MergeFromProviders(providers map[[3]int]*GeoTIFFRasterProvider, coords [][3]int, borderMode BorderMode) *TileData {
	tiles := make([]*ElevationGrid, len(coords))

	for i, coord := range coords {
		provider, exists := providers[coord]
		if !exists {
			continue
		}

		grid := provider.GetElevationGrid()
		if grid == nil {
			continue
		}

		tiles[i] = grid
	}

	return m.Merge(tiles, borderMode)
}

func (m *RasterMerger) srcSize() [2]uint32 {
	width := m.Size[0] * uint32(m.Grid[0])
	height := m.Size[1] * uint32(m.Grid[1])
	return [2]uint32{width, height}
}

func (m *RasterMerger) tileOffset(i int) [2]int {
	x := int(math.Mod(float64(i), float64(m.Grid[0]))) * int(m.Size[0])
	y := int(math.Floor(float64(i)/float64(m.Grid[0]))) * int(m.Size[1])
	return [2]int{x, y}
}

func (m *RasterMerger) copyTileData(dst, src *TileData, pos [2]int) {
	if pos[0]+int(src.Size[0]) > int(dst.Size[0]) || pos[1]+int(src.Size[1]) > int(dst.Size[1]) {
		return
	}

	xOffset, yOffset := pos[0], pos[1]
	for y := yOffset; y < yOffset+int(src.Size[1]); y++ {
		srcStart := (y - yOffset) * int(src.Size[0])
		srcEnd := srcStart + int(src.Size[0])
		dstStart := y*int(dst.Size[0]) + xOffset
		dstEnd := dstStart + int(src.Size[0])

		copy(dst.Datas[dstStart:dstEnd], src.Datas[srcStart:srcEnd])
	}

	m.copyBorders(dst, src, pos)
}

func (m *RasterMerger) copyBorders(dst, src *TileData, pos [2]int) {
	if !dst.HasBorder() || !src.HasBorder() {
		return
	}

	xOffset, yOffset := pos[0], pos[1]

	if xOffset == 0 && dst.HasBorder() && len(src.LeftBorder) > 0 {
		for y := 0; y < int(src.Size[1]) && yOffset+y < len(dst.LeftBorder); y++ {
			if y < len(src.LeftBorder) {
				dst.LeftBorder[yOffset+y] = src.LeftBorder[y]
			}
		}
	}

	if xOffset+int(src.Size[0]) == int(dst.Size[0]) && dst.IsBilateral() && len(src.RightBorder) > 0 {
		for y := 0; y < int(src.Size[1]) && yOffset+y < len(dst.RightBorder); y++ {
			if y < len(src.RightBorder) {
				dst.RightBorder[yOffset+y] = src.RightBorder[y]
			}
		}
	}

	if yOffset == 0 && dst.HasBorder() && len(src.TopBorder) > 0 {
		offx := 1
		if xOffset == 0 {
			offx = 0
		}
		for x := 0; x < int(src.Size[0]) && offx+xOffset+x < len(dst.TopBorder); x++ {
			if offx+x < len(src.TopBorder) {
				dst.TopBorder[offx+xOffset+x] = src.TopBorder[offx+x]
			}
		}
	}

	if yOffset+int(src.Size[1]) == int(dst.Size[1]) && dst.IsBilateral() && len(src.BottomBorder) > 0 {
		offx := 1
		if xOffset == 0 {
			offx = 0
		}
		for x := 0; x < int(src.Size[0]) && offx+xOffset+x < len(dst.BottomBorder); x++ {
			if offx+x < len(src.BottomBorder) {
				dst.BottomBorder[offx+xOffset+x] = src.BottomBorder[offx+x]
			}
		}
	}
}

func joinRects(r1, r2 vec2d.Rect) vec2d.Rect {
	if isRectEmpty(r1) {
		return r2
	}
	if isRectEmpty(r2) {
		return r1
	}

	minX := r1.Min[0]
	minY := r1.Min[1]
	maxX := r1.Max[0]
	maxY := r1.Max[1]

	if r2.Min[0] < minX {
		minX = r2.Min[0]
	}
	if r2.Min[1] < minY {
		minY = r2.Min[1]
	}
	if r2.Max[0] > maxX {
		maxX = r2.Max[0]
	}
	if r2.Max[1] > maxY {
		maxY = r2.Max[1]
	}

	return vec2d.Rect{
		Min: vec2d.T{minX, minY},
		Max: vec2d.T{maxX, maxY},
	}
}

type TiledRaster struct {
	Tiles    []*ElevationGrid
	TileGrid [2]int
	TileSize [2]uint32
}

func NewTiledRaster(tiles []*ElevationGrid, tileGrid [2]int, tileSize [2]uint32) *TiledRaster {
	return &TiledRaster{
		Tiles:    tiles,
		TileGrid: tileGrid,
		TileSize: tileSize,
	}
}

func (t *TiledRaster) GetMergedRaster(borderMode BorderMode) *TileData {
	m := NewRasterMerger(t.TileGrid, t.TileSize)
	return m.Merge(t.Tiles, borderMode)
}
