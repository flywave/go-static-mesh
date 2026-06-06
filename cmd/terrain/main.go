package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-static-mesh/mesh"
	"github.com/flywave/go-static-mesh/mesh/builder"
	"github.com/flywave/go-static-mesh/mesh/writer"
	"github.com/flywave/go-static-mesh/tile"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

type config struct {
	demDir              string
	satelliteDir        string
	outputDir           string
	outputFormats       []string
	maxError            float64
	verticalExaggeration float64
	baseElevation       float64
	closeMesh           bool
	baseThickness       float64
	zoom                int
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.demDir, "dem", "data/dem", "DEM tiles directory")
	flag.StringVar(&cfg.satelliteDir, "satellite", "", "Satellite imagery directory (optional)")
	flag.StringVar(&cfg.outputDir, "output", "output", "Output directory")
	formats := flag.String("format", "glb", "Output formats (comma-separated: glb,gltf,stl,obj,mst)")
	flag.Float64Var(&cfg.maxError, "max-error", 2.0, "TIN max error")
	flag.Float64Var(&cfg.verticalExaggeration, "exaggeration", 1.0, "Vertical exaggeration")
	flag.Float64Var(&cfg.baseElevation, "base-elevation", 0.0, "Base elevation offset")
	flag.BoolVar(&cfg.closeMesh, "close", false, "Close mesh for 3D printing")
	flag.Float64Var(&cfg.baseThickness, "thickness", 5.0, "Base thickness for closed mesh")
	flag.IntVar(&cfg.zoom, "zoom", 0, "Tile zoom level (0=auto)")
	flag.Parse()

	cfg.outputFormats = strings.Split(*formats, ",")
	for i := range cfg.outputFormats {
		cfg.outputFormats[i] = strings.TrimSpace(cfg.outputFormats[i])
	}

	if err := run(cfg); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run(cfg config) error {
	demFiles, err := filepath.Glob(filepath.Join(cfg.demDir, "*.webp"))
	if err != nil {
		return fmt.Errorf("list DEM files: %w", err)
	}
	if len(demFiles) == 0 {
		demFiles, err = filepath.Glob(filepath.Join(cfg.demDir, "*.tif"))
		if err != nil {
			return fmt.Errorf("list DEM files: %w", err)
		}
	}
	if len(demFiles) == 0 {
		return fmt.Errorf("no DEM files found in %s", cfg.demDir)
	}

	ext := filepath.Ext(demFiles[0])
	log.Printf("Found %d DEM files (format: %s)", len(demFiles), ext)

	var mergedDEM *tile.ElevationGrid
	if ext == ".webp" {
		mergedDEM, err = mergeDEMWebP(demFiles)
	} else {
		mergedDEM, err = mergeDEMTiff(demFiles)
	}
	if err != nil {
		return fmt.Errorf("merge DEM: %w", err)
	}

	var mergedTexture image.Image
	if cfg.satelliteDir != "" {
		mergedTexture, err = mergeSatellite(cfg.satelliteDir, mergedDEM.Bounds)
		if err != nil {
			log.Printf("Warning: failed to load satellite: %v", err)
		}
	}

	b := builder.NewBuilder()
	b.SetBounds(mergedDEM.Bounds, geo.NewProj(4326))

	tinGen := mesh.NewTINGenerator()
	tinGen.SetMaxError(cfg.maxError)
	tinGen.SetSrcProj(geo.NewProj(4326))
	b.SetTINGenerator(tinGen)

	provider := &mergedProvider{mergedDEM: mergedDEM, bounds: mergedDEM.Bounds}
	b.SetRasterProvider(provider)
	b.SetVerticalExaggeration(cfg.verticalExaggeration)
	b.SetBaseElevation(cfg.baseElevation)

	if cfg.closeMesh {
		b.SetCloseMesh(true, cfg.baseThickness)
	}

	if mergedTexture != nil {
		textureMesh := &mesh.Mesh{Texture: mergedTexture}
		b.SetTexture(textureMesh)
	}

	if cfg.zoom > 0 {
		b.SetZoom(cfg.zoom)
	}

	log.Println("Building terrain mesh...")
	terrainMesh, err := b.BuildForDisplay()
	if err != nil {
		return fmt.Errorf("build mesh: %w", err)
	}

	log.Printf("Mesh built: %d vertices, %d triangles", len(terrainMesh.Vertices), terrainMesh.TriangleCount())

	if err := os.MkdirAll(cfg.outputDir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	for _, format := range cfg.outputFormats {
		outputFile := filepath.Join(cfg.outputDir, "terrain."+format)
		if err := writeMesh(terrainMesh, outputFile, format); err != nil {
			log.Printf("Warning: failed to write %s: %v", format, err)
		} else {
			stat, _ := os.Stat(outputFile)
			log.Printf("Written %s: %.2f MB", outputFile, float64(stat.Size())/1024/1024)
		}
	}

	return nil
}

func writeMesh(m *mesh.Mesh, path, format string) error {
	switch format {
	case "glb":
		return writer.NewGltfWriter().Write(m, path)
	case "gltf":
		w := writer.NewGLTFWriter(false, true)
		return w.Write(m, path)
	case "stl":
		return writer.NewSTLWriter(false).WriteFile(m, path)
	case "obj":
		return writer.NewOBJWriter(true, true).Write(m, path)
	case "mst":
		return writer.NewMstWriter().Write(m, path)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

type mergedProvider struct {
	mergedDEM *tile.ElevationGrid
	bounds    vec2d.Rect
}

func (p *mergedProvider) Attribution() string { return "DEM Data" }
func (p *mergedProvider) Grid() *geo.TileGrid {
	return &geo.TileGrid{
		Srs:      geo.NewProj(4326),
		TileSize: []uint32{uint32(p.mergedDEM.Width), uint32(p.mergedDEM.Height)},
	}
}
func (p *mergedProvider) Bounds() vec2d.Rect     { return p.bounds }
func (p *mergedProvider) Srs() geo.Proj           { return geo.NewProj(4326) }
func (p *mergedProvider) GetElevationGrid() *tile.ElevationGrid { return p.mergedDEM }

func mergeDEMWebP(files []string) (*tile.ElevationGrid, error) {
	minX, maxX := int(^uint(0)>>1), 0
	minY, maxY := int(^uint(0)>>1), 0
	var zoom int
	tileData := make(map[[3]int]*tile.TileData)

	for _, file := range files {
		basename := filepath.Base(file)
		var z, x, y int
		if _, err := fmt.Sscanf(basename, "%d_%d_%d.webp", &z, &x, &y); err != nil {
			continue
		}
		if zoom == 0 {
			zoom = z
		}
		if x < minX { minX = x }
		if x > maxX { maxX = x }
		if y < minY { minY = y }
		if y > maxY { maxY = y }

		f, err := os.Open(file)
		if err != nil { continue }
		provider := tile.NewMapboxRasterProvider(tile.ModeMapbox)
		data, err := provider.LoadDEM(f, tile.ModeMapbox)
		f.Close()
		if err != nil { continue }
		tileData[[3]int{x, y, z}] = data
	}

	gridCols := maxX - minX + 1
	gridRows := maxY - minY + 1
	coords := make([][3]int, 0, gridCols*gridRows)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			coords = append(coords, [3]int{x, y, zoom})
		}
	}

	tiles := make([]*tile.ElevationGrid, len(coords))
	for i, coord := range coords {
		if data, ok := tileData[coord]; ok {
			if grid := tile.NewElevationGridFromTileData(data, geo.NewProj(4326)); grid != nil {
				tiles[i] = grid
			}
		}
	}

	var tileSize uint32
	for _, g := range tiles {
		if g != nil { tileSize = uint32(g.Width); break }
	}
	merger := tile.NewRasterMerger([2]int{gridCols, gridRows}, [2]uint32{tileSize, tileSize})
	mergedData := merger.Merge(tiles, tile.BORDER_NONE)
	if mergedData == nil {
		return nil, fmt.Errorf("merge failed")
	}

	n := float64(int(1) << zoom)
	minLon := float64(minX)/n*360.0 - 180.0
	maxLon := float64(maxX+1)/n*360.0 - 180.0
	latNorth := math.Atan(math.Sinh(math.Pi*(1-2*float64(minY)/n))) * 180.0 / math.Pi
	latSouth := math.Atan(math.Sinh(math.Pi*(1-2*float64(maxY+1)/n))) * 180.0 / math.Pi
	mergedData.Box = vec2d.Rect{Min: vec2d.T{latSouth, minLon}, Max: vec2d.T{latNorth, maxLon}}
	mergedData.Boxsrs = geo.NewProj(4326)

	return tile.NewElevationGridFromTileData(mergedData, geo.NewProj(4326)), nil
}

func mergeDEMTiff(files []string) (*tile.ElevationGrid, error) {
	minX, maxX := int(^uint(0)>>1), 0
	minY, maxY := int(^uint(0)>>1), 0
	var zoom int
	providers := make(map[[3]int]*tile.GeoTIFFRasterProvider)

	for _, file := range files {
		basename := filepath.Base(file)
		var z, x, y int
		if _, err := fmt.Sscanf(basename, "%d_%d_%d.tif", &z, &x, &y); err != nil {
			continue
		}
		if zoom == 0 { zoom = z }
		if x < minX { minX = x }
		if x > maxX { maxX = x }
		if y < minY { minY = y }
		if y > maxY { maxY = y }
		provider, err := tile.NewGeoTIFFRasterProvider(file)
		if err != nil { continue }
		providers[[3]int{x, y, z}] = provider
	}
	defer func() {
		for _, p := range providers { p.Close() }
	}()

	gridCols := maxX - minX + 1
	gridRows := maxY - minY + 1
	coords := make([][3]int, 0, gridCols*gridRows)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			coords = append(coords, [3]int{x, y, zoom})
		}
	}

	var tileSize uint32
	tiles := make([]*tile.ElevationGrid, len(coords))
	for i, coord := range coords {
		if p, ok := providers[coord]; ok {
			if g := p.GetElevationGrid(); g != nil {
				tiles[i] = g
				if tileSize == 0 { tileSize = uint32(g.Width) }
			}
		}
	}
	merger := tile.NewRasterMerger([2]int{gridCols, gridRows}, [2]uint32{tileSize, tileSize})
	mergedData := merger.Merge(tiles, tile.BORDER_NONE)
	if mergedData == nil {
		return nil, fmt.Errorf("merge failed")
	}
	return tile.NewElevationGridFromTileData(mergedData, geo.NewProj(4326)), nil
}

func mergeSatellite(dir string, demBounds vec2d.Rect) (image.Image, error) {
	exts := []string{"*.webp", "*.png", "*.jpg"}
	var files []string
	for _, ext := range exts {
		var err error
		files, err = filepath.Glob(filepath.Join(dir, ext))
		if err != nil { return nil, err }
		if len(files) > 0 { break }
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no satellite files found")
	}

	var zoom int
	minX, maxX := int(^uint(0)>>1), 0
	minY, maxY := int(^uint(0)>>1), 0
	tileImages := make(map[[3]int]image.Image)

	for _, file := range files {
		base := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
		parts := strings.Split(base, "_")
		if len(parts) < 3 { continue }
		z, _ := strconv.Atoi(parts[len(parts)-3])
		x, _ := strconv.Atoi(parts[len(parts)-2])
		y, _ := strconv.Atoi(parts[len(parts)-1])
		if zoom == 0 { zoom = z }
		if x < minX { minX = x }
		if x > maxX { maxX = x }
		if y < minY { minY = y }
		if y > maxY { maxY = y }

		f, err := os.Open(file)
		if err != nil { continue }
		img, _, err := image.Decode(f)
		f.Close()
		if err != nil { continue }
		tileImages[[3]int{x, y, z}] = img
	}

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

	tiles := make([]image.Image, len(coords))
	var imgSize uint32
	for i, coord := range coords {
		if img, ok := tileImages[coord]; ok {
			tiles[i] = img
			if imgSize == 0 { imgSize = uint32(img.Bounds().Dx()) }
		}
	}

	merger := tile.NewImageMerger([2]int{gridCols, gridRows}, [2]uint32{imgSize, imgSize})
	return merger.Merge(tiles, nil), nil
}
