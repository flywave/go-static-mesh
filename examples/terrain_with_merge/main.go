package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-static-mesh/mesh"
	"github.com/flywave/go-static-mesh/mesh/builder"
	"github.com/flywave/go-static-mesh/mesh/writer"
	"github.com/flywave/go-static-mesh/tile"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

type MergedTileProvider struct {
	mergedGrid *tile.ElevationGrid
	bounds     vec2d.Rect
	srs        geo.Proj
	grid       *geo.TileGrid
}

func NewMergedTileProvider(dataDir string) (*MergedTileProvider, error) {
	files, err := filepath.Glob(filepath.Join(dataDir, "*.tif"))
	if err != nil {
		return nil, fmt.Errorf("failed to list tif files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no tif files found in %s", dataDir)
	}

	minX, maxX := int(^uint(0)>>1), 0
	minY, maxY := int(^uint(0)>>1), 0
	var zoom int

	providers := make(map[[3]int]*tile.GeoTIFFRasterProvider)

	for _, file := range files {
		basename := filepath.Base(file)
		var z, x, y int
		_, err := fmt.Sscanf(basename, "%d_%d_%d.tif", &z, &x, &y)
		if err != nil {
			log.Printf("skipping file with invalid name: %s", basename)
			continue
		}

		if zoom == 0 {
			zoom = z
		}

		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}

		provider, err := tile.NewGeoTIFFRasterProvider(file)
		if err != nil {
			log.Printf("failed to load %s: %v", file, err)
			continue
		}

		coord := [3]int{x, y, z}
		providers[coord] = provider
		log.Printf("loaded tile %d/%d/%d", z, x, y)
	}

	defer func() {
		for _, p := range providers {
			p.Close()
		}
	}()

	gridCols := maxX - minX + 1
	gridRows := maxY - minY + 1

	coords := make([][3]int, gridCols*gridRows)
	idx := 0
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			coords[idx] = [3]int{x, y, zoom}
			idx++
		}
	}

	log.Printf("Merging %d tiles (%dx%d grid)...", len(coords), gridCols, gridRows)

	merger := tile.NewRasterMerger([2]int{gridCols, gridRows}, [2]uint32{256, 256})

	mergedData := merger.MergeFromProviders(providers, coords, tile.BORDER_NONE)
	if mergedData == nil {
		return nil, fmt.Errorf("failed to merge tiles")
	}

	mergedGrid := tile.NewElevationGridFromTileData(mergedData, geo.NewProj(4326))
	if mergedGrid == nil {
		return nil, fmt.Errorf("failed to create elevation grid from merged data")
	}

	log.Printf("Merged grid size: %dx%d", mergedGrid.Width, mergedGrid.Height)
	log.Printf("Merged bounds: [%.6f, %.6f] - [%.6f, %.6f]",
		mergedGrid.Bounds.Min[0], mergedGrid.Bounds.Min[1],
		mergedGrid.Bounds.Max[0], mergedGrid.Bounds.Max[1])

	return &MergedTileProvider{
		mergedGrid: mergedGrid,
		bounds:     mergedGrid.Bounds,
		srs:        geo.NewProj(4326),
		grid: &geo.TileGrid{
			Srs:      geo.NewProj(4326),
			TileSize: []uint32{uint32(mergedGrid.Width), uint32(mergedGrid.Height)},
		},
	}, nil
}

func (p *MergedTileProvider) Attribution() string {
	return "Merged DEM Data"
}

func (p *MergedTileProvider) Grid() *geo.TileGrid {
	return p.grid
}

func (p *MergedTileProvider) Bounds() vec2d.Rect {
	return p.bounds
}

func (p *MergedTileProvider) Srs() geo.Proj {
	return p.srs
}

func (p *MergedTileProvider) GetElevationGrid() *tile.ElevationGrid {
	return p.mergedGrid
}

func (p *MergedTileProvider) GetElevation(lng, lat float64) float64 {
	return p.mergedGrid.GetElevation(lng, lat)
}

func main() {
	fmt.Println("=== 地形网格生成 - 使用瓦片合并算法 ===")
	fmt.Println()

	dataDir := "data/dem"

	provider, err := NewMergedTileProvider(dataDir)
	if err != nil {
		log.Fatalf("创建合并 provider 失败: %v", err)
	}

	fmt.Println()
	fmt.Println("步骤 1: 创建 Builder")
	b := builder.NewBuilder()

	fmt.Println("步骤 2: 设置边界")
	b.SetBounds(provider.Bounds(), provider.Srs())

	fmt.Println("步骤 3: 创建并配置 TIN Generator")
	tinGen := mesh.NewTINGenerator()
	tinGen.SetMaxError(2.0)
	tinGen.SetSrcProj(geo.NewProj(4326))

	fmt.Println("步骤 4: 设置 TIN Generator")
	b.SetTINGenerator(tinGen)

	fmt.Println("步骤 5: 设置 Raster Provider")
	b.SetRasterProvider(provider)

	fmt.Println("步骤 6: 设置垂直夸张")
	b.SetVerticalExaggeration(1.0)

	fmt.Println("步骤 7: 设置基础高程")
	b.SetBaseElevation(0.0)

	fmt.Println("步骤 8: 配置网格闭合")
	b.SetCloseMesh(true, 100.0)

	fmt.Println()
	fmt.Println("步骤 9: 构建地形网格")
	terrainMesh, err := b.BuildForDisplay()
	if err != nil {
		log.Fatalf("构建地形网格失败: %v", err)
	}

	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("创建输出目录失败: %v", err)
	}

	outputFile := filepath.Join(outputDir, "terrain_merged.glb")
	fmt.Printf("步骤 10: 保存为 GLB 格式 -> %s\n", outputFile)

	w := writer.NewGltfWriter()
	if err := w.Write(terrainMesh, outputFile); err != nil {
		log.Fatalf("写入 GLTF 失败: %v", err)
	}

	fmt.Println()
	fmt.Println("✓ 地形网格生成成功！")
	fmt.Printf("  - 顶点数: %d\n", len(terrainMesh.Vertices))
	fmt.Printf("  - 三角形数: %d\n", len(terrainMesh.Indices)/3)
	fmt.Printf("  - 输出文件: %s\n", outputFile)

	stat, _ := os.Stat(outputFile)
	fmt.Printf("  - 文件大小: %.2f MB\n", float64(stat.Size())/1024/1024)

	fmt.Println()
	fmt.Println("=== 完成 ===")
	fmt.Println()
	fmt.Println("实现的功能:")
	fmt.Println("  ✅ 1. 瓦片合并算法 (RasterMerger)")
	fmt.Println("  ✅ 2. 多瓦片自动拼接")
	fmt.Println("  ✅  ✅ 3. TIN 网格生成")
	fmt.Println("  ✅ 4. GLTF 格式导出")
}
