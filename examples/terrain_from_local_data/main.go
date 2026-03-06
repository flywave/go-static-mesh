package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-static-mesh/mesh"
	"github.com/flywave/go-static-mesh/mesh/builder"
	"github.com/flywave/go-static-mesh/mesh/writer"
	"github.com/flywave/go-static-mesh/tile"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

type LocalTileProvider struct {
	providers map[[3]int]*tile.GeoTIFFRasterProvider
	grid      *geo.TileGrid
	bounds    vec2d.Rect
	srs       geo.Proj
}

func NewLocalTileProvider(dataDir string) (*LocalTileProvider, error) {
	p := &LocalTileProvider{
		providers: make(map[[3]int]*tile.GeoTIFFRasterProvider),
		srs:       geo.NewProj(4326),
	}

	minX, maxX := int(^uint(0)>>1), 0
	minY, maxY := int(^uint(0)>>1), 0
	var zoom int

	files, err := filepath.Glob(filepath.Join(dataDir, "*.tif"))
	if err != nil {
		return nil, fmt.Errorf("failed to list tif files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no tif files found in %s", dataDir)
	}

	for _, file := range files {
		basename := filepath.Base(file)
		parts := strings.Split(strings.TrimSuffix(basename, ".tif"), "_")
		if len(parts) != 3 {
			log.Printf("skipping file with invalid name format: %s", basename)
			continue
		}

		var z, x, y int
		_, err := fmt.Sscanf(basename, "%d_%d_%d.tif", &z, &x, &y)
		if err != nil {
			log.Printf("failed to parse tile coordinates from %s: %v", basename, err)
			continue
		}

		if zoom == 0 {
			zoom = z
		}

		provider, err := tile.NewGeoTIFFRasterProvider(file)
		if err != nil {
			log.Printf("failed to create provider for %s: %v", file, err)
			continue
		}

		coord := [3]int{x, y, z}
		p.providers[coord] = provider

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

		log.Printf("加载瓦片: %d/%d/%d from %s", z, x, y, basename)
	}

	tileCount := 1 << zoom
	minLon := float64(minX)/float64(tileCount)*360.0 - 180.0
	maxLon := float64(maxX+1)/float64(tileCount)*360.0 - 180.0
	maxLat := 85.0511 * (1 - 2*float64(minY)/float64(tileCount))
	minLat := 85.0511 * (1 - 2*float64(maxY+1)/float64(tileCount))

	p.bounds = vec2d.Rect{
		Min: vec2d.T{minLon, minLat},
		Max: vec2d.T{maxLon, maxLat},
	}

	p.grid = &geo.TileGrid{
		Srs:      p.srs,
		TileSize: []uint32{256, 256},
	}

	log.Printf("总共加载 %d 个瓦片", len(p.providers))
	log.Printf("边界范围: [%.6f, %.6f] - [%.6f, %.6f]",
		p.bounds.Min[0], p.bounds.Min[1],
		p.bounds.Max[0], p.bounds.Max[1])

	return p, nil
}

func (p *LocalTileProvider) Attribution() string {
	return "Local DEM Data"
}

func (p *LocalTileProvider) Grid() *geo.TileGrid {
	return p.grid
}

func (p *LocalTileProvider) Bounds() vec2d.Rect {
	return p.bounds
}

func (p *LocalTileProvider) Srs() geo.Proj {
	return p.srs
}

func (p *LocalTileProvider) GetElevationTile(coord [3]int) (*tile.ElevationGrid, error) {
	provider, exists := p.providers[coord]
	if !exists {
		return nil, fmt.Errorf("tile %d/%d/%d not found", coord[2], coord[0], coord[1])
	}
	return provider.GetElevationGrid(), nil
}

func (p *LocalTileProvider) GetTileData(coord [3]int) (*tile.TileData, error) {
	provider, exists := p.providers[coord]
	if !exists {
		return nil, fmt.Errorf("tile %d/%d/%d not found", coord[2], coord[0], coord[1])
	}
	return provider.GetTileData(coord)
}

func main() {
	fmt.Println("=== 地形网格生成示例 ===")
	fmt.Println()

	dataDir := "data/dem"

	provider, err := NewLocalTileProvider(dataDir)
	if err != nil {
		log.Fatalf("创建 provider 失败: %v", err)
	}

	fmt.Println()
	fmt.Println("步骤 1: 创建 Builder")
	b := builder.NewBuilder()

	fmt.Println("步骤 2: 设置边界")
	b.SetBounds(provider.Bounds(), provider.Srs())

	fmt.Println("步骤 3: 设置垂直夸张")
	b.SetVerticalExaggeration(1.0)

	fmt.Println("步骤 4: 设置基础高程")
	b.SetBaseElevation(0.0)

	fmt.Println("步骤 5: 配置网格闭合")
	b.SetCloseMesh(true, 100.0)

	fmt.Println()
	fmt.Println("步骤 6: 构建地形网格")
	fmt.Println("  注意: 需要集成以下组件:")
	fmt.Println("    - TIN 生成算法")
	fmt.Println("    - DEM 数据采样")
	fmt.Println("    - 纹理映射")
	fmt.Println()

	m := &mesh.Mesh{
		Bounds: provider.Bounds(),
		Srs:    provider.Srs(),
	}

	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("创建输出目录失败: %v", err)
	}

	outputFile := filepath.Join(outputDir, "terrain.gltf")
	fmt.Printf("步骤 7: 保存为 GLTF 格式 -> %s\n", outputFile)

	w := writer.NewGltfWriter()

	if err := w.Write(m, outputFile); err != nil {
		log.Printf("写入 GLTF 失败: %v (这是预期的，因为 mesh 为空)", err)
	}

	fmt.Println()
	fmt.Println("=== 示例完成 ===")
	fmt.Println()
	fmt.Println("接下来的步骤:")
	fmt.Println("  1. 实现 TIN 生成算法")
	fmt.Println("  2. 从 GeoTIFF 采样高程数据")
	fmt.Println("  3. 生成三角网格")
	fmt.Println("  4. 可选: 添加卫星影像纹理")
	fmt.Println("  5. 导出为 GLTF 格式")

	for _, provider := range provider.providers {
		provider.Close()
	}
}
