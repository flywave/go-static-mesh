package main

import (
	"fmt"
	"image"
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

type TextureAwareTerrainProvider struct {
	mergedDEM      *tile.ElevationGrid
	mergedTexture  image.Image
	bounds         vec2d.Rect
	srs            geo.Proj
	grid           *geo.TileGrid
	textureEnabled bool
}

func NewTextureAwareTerrainProvider(demDir, textureDir string, enableTexture bool) (*TextureAwareTerrainProvider, error) {
	mergedDEM, err := loadAndMergeDEM(demDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load DEM: %w", err)
	}

	var mergedTexture image.Image
	if enableTexture && textureDir != "" {
		mergedTexture, err = loadAndMergeTexture(textureDir, mergedDEM.Bounds)
		if err != nil {
			log.Printf("Warning: failed to load texture: %v", err)
		}
	}

	return &TextureAwareTerrainProvider{
		mergedDEM:     mergedDEM,
		mergedTexture: mergedTexture,
		bounds:        mergedDEM.Bounds,
		srs:           geo.NewProj(4326),
		grid: &geo.TileGrid{
			Srs:      geo.NewProj(4326),
			TileSize: []uint32{uint32(mergedDEM.Width), uint32(mergedDEM.Height)},
		},
		textureEnabled: mergedTexture != nil,
	}, nil
}

func loadAndMergeDEM(dataDir string) (*tile.ElevationGrid, error) {
	files, err := filepath.Glob(filepath.Join(dataDir, "*.tif"))
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no DEM files found")
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
			continue
		}

		coord := [3]int{x, y, z}
		providers[coord] = provider
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

	log.Printf("Merging %d DEM tiles (%dx%d grid)...", len(coords), gridCols, gridRows)

	merger := tile.NewRasterMerger([2]int{gridCols, gridRows}, [2]uint32{256, 256})
	mergedData := merger.MergeFromProviders(providers, coords, tile.BORDER_NONE)

	mergedGrid := tile.NewElevationGridFromTileData(mergedData, geo.NewProj(4326))

	log.Printf("DEM merged: size=%dx%d", mergedGrid.Width, mergedGrid.Height)

	return mergedGrid, nil
}

func loadAndMergeTexture(textureDir string, demBounds vec2d.Rect) (image.Image, error) {
	files, err := filepath.Glob(filepath.Join(textureDir, "*.webp"))
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		files, err = filepath.Glob(filepath.Join(textureDir, "*.png"))
		if err != nil {
			return nil, err
		}
	}

	if len(files) == 0 {
		files, err = filepath.Glob(filepath.Join(textureDir, "*.jpg"))
		if err != nil {
			return nil, err
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no texture files found")
	}

	log.Printf("Loading %d texture files...", len(files))

	layers := make([]image.Image, 0)
	opacities := make([]float64, 0)
	bounds := make([]vec2d.Rect, 0)
	srs := make([]geo.Proj, 0)

	for _, file := range files {
		img, err := loadImage(file)
		if err != nil {
			log.Printf("Failed to load %s: %v", file, err)
			continue
		}

		layers = append(layers, img)
		opacities = append(opacities, 1.0)
		bounds = append(bounds, demBounds)
		srs = append(srs, geo.NewProj(4326))
	}

	if len(layers) == 0 {
		return nil, fmt.Errorf("no valid texture layers")
	}

	merger := tile.NewLayerMerger()
	for i := range layers {
		merger.AddLayer(layers[i], opacities[i], bounds[i], srs[i])
	}

	mergedTexture := merger.Merge(
		[2]uint32{uint32(1024), uint32(1024)},
		demBounds,
		geo.NewProj(4326),
		nil,
	)

	log.Printf("Texture merged: size=%dx%d",
		mergedTexture.Bounds().Dx(), mergedTexture.Bounds().Dy())

	return mergedTexture, nil
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (p *TextureAwareTerrainProvider) Attribution() string {
	return "Merged DEM + Texture Data"
}

func (p *TextureAwareTerrainProvider) Grid() *geo.TileGrid {
	return p.grid
}

func (p *TextureAwareTerrainProvider) Bounds() vec2d.Rect {
	return p.bounds
}

func (p *TextureAwareTerrainProvider) Srs() geo.Proj {
	return p.srs
}

func (p *TextureAwareTerrainProvider) GetElevationGrid() *tile.ElevationGrid {
	return p.mergedDEM
}

func (p *TextureAwareTerrainProvider) GetElevation(lng, lat float64) float64 {
	return p.mergedDEM.GetElevation(lng, lat)
}

func (p *TextureAwareTerrainProvider) GetTexture() image.Image {
	return p.mergedTexture
}

func main() {
	fmt.Println("=== 地形 + 纹理生成示例 ===")
	fmt.Println()

	demDir := "../../data/dem"
	textureDir := "../../data/satellite"
	enableTexture := true

	provider, err := NewTextureAwareTerrainProvider(demDir, textureDir, enableTexture)
	if err != nil {
		log.Fatalf("创建 provider 失败: %v", err)
	}

	fmt.Println()
	fmt.Println("步骤 1: 创建 Builder")
	b := builder.NewBuilder()

	fmt.Println("步骤 2: 设置边界")
	b.SetBounds(provider.Bounds(), provider.Srs())

	fmt.Println("步骤 3: 配置 TIN Generator")
	tinGen := mesh.NewTINGenerator()
	tinGen.SetMaxError(2.0)
	tinGen.SetSrcProj(geo.NewProj(4326))

	fmt.Println("步骤 4: 设置 providers")
	b.SetTINGenerator(tinGen)
	b.SetRasterProvider(provider)

	if provider.textureEnabled {
		fmt.Println("步骤 5: 添加纹理")
		textureMesh := &mesh.Mesh{
			Texture: provider.GetTexture(),
		}
		b.SetTexture(textureMesh)
	}

	fmt.Println("步骤 6: 配置网格参数")
	b.SetVerticalExaggeration(1.0)
	b.SetBaseElevation(0.0)
	b.SetCloseMesh(true, 100.0)

	fmt.Println()
	fmt.Println("步骤 7: 构建地形网格")
	terrainMesh, err := b.BuildForDisplay()
	if err != nil {
		log.Fatalf("构建失败: %v", err)
	}

	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("创建输出目录失败: %v", err)
	}

	outputFile := filepath.Join(outputDir, "terrain_textured.gltf")
	fmt.Printf("步骤 8: 保存为 GLTF -> %s\n", outputFile)

	w := writer.NewGltfWriter()
	if err := w.Write(terrainMesh, outputFile); err != nil {
		log.Fatalf("写入失败: %v", err)
	}

	fmt.Println()
	fmt.Println("✓ 地形 + 纹理生成成功！")
	fmt.Printf("  - 顶点数: %d\n", len(terrainMesh.Vertices))
	fmt.Printf("  - 三角形数: %d\n", len(terrainMesh.Indices)/3)
	if provider.textureEnabled {
		fmt.Printf("  - 纹理: 已应用\n")
	} else {
		fmt.Printf("  - 纹理: 无\n")
	}
	fmt.Printf("  - 输出文件: %s\n", outputFile)

	stat, _ := os.Stat(outputFile)
	fmt.Printf("  - 文件大小: %.2f MB\n", float64(stat.Size())/1024/1024)

	fmt.Println()
	fmt.Println("=== 完成 ===")
}
