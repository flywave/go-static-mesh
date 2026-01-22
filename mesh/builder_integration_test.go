package mesh

import (
	"testing"

	"github.com/flywave/go-geo"
	static "github.com/flywave/go-static-mesh/static"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

type MockRasterProvider struct {
	grid *static.ElevationGrid
}

func NewMockRasterProvider() *MockRasterProvider {
	width, height := 256, 256
	data := make([]float64, width*height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := y*width + x
			data[i] = float64(x+y) * 0.1
		}
	}

	grid := &static.ElevationGrid{
		Width:    width,
		Height:   height,
		Data:     data,
		MinX:     0.0,
		MinY:     0.0,
		CellSize: 0.001,
		NoData:   -9999.0,
		Bounds:   vec2d.Rect{Min: vec2d.T{0, 0}, Max: vec2d.T{0.256, 0.256}},
		Srs:      geo.NewProj(4326),
	}

	return &MockRasterProvider{grid: grid}
}

func (p *MockRasterProvider) Attribution() string {
	return "Mock Raster Provider"
}

func (p *MockRasterProvider) Grid() *geo.TileGrid {
	return &geo.TileGrid{
		Srs: geo.NewProj(4326),
	}
}

func (p *MockRasterProvider) Bounds() vec2d.Rect {
	return p.grid.Bounds
}

func (p *MockRasterProvider) Srs() geo.Proj {
	return p.grid.Srs
}

func (p *MockRasterProvider) GetElevation(lng, lat float64) float64 {
	return p.grid.GetElevation(lng, lat)
}

func (p *MockRasterProvider) GetElevationGrid() *static.ElevationGrid {
	return p.grid
}

func (p *MockRasterProvider) GetElevationTile(coord [3]int) (*static.ElevationGrid, error) {
	return p.grid, nil
}

func (p *MockRasterProvider) GetTileData(coord [3]int) (*static.TileData, error) {
	return p.grid.ToTileData(), nil
}

func TestBuilderBuildFromRaster(t *testing.T) {
	provider := NewMockRasterProvider()
	builder := NewBuilder()

	tinGenerator := NewTINGenerator()
	builder.SetTINGenerator(tinGenerator)

	builder.SetRasterProvider(provider)
	builder.SetBounds(provider.Bounds(), provider.Srs())
	builder.SetZoom(15)

	mesh, err := builder.BuildForDisplay()
	if err != nil {
		t.Fatalf("BuildForDisplay failed: %v", err)
	}

	if mesh == nil {
		t.Fatal("Mesh is nil")
	}

	if len(mesh.Vertices) == 0 {
		t.Error("No vertices in mesh")
	}

	if len(mesh.Indices) == 0 {
		t.Error("No indices in mesh")
	}

	t.Logf("Mesh built successfully: %d vertices, %d triangles",
		len(mesh.Vertices), len(mesh.Indices)/3)
}

func TestBuilderBuildForPrint(t *testing.T) {
	provider := NewMockRasterProvider()
	builder := NewBuilder()

	tinGenerator := NewTINGenerator()
	builder.SetTINGenerator(tinGenerator)

	builder.SetRasterProvider(provider)
	builder.SetBounds(provider.Bounds(), provider.Srs())
	builder.SetZoom(15)
	builder.SetCloseMesh(true, 5.0)

	mesh, err := builder.BuildForPrint()
	if err != nil {
		t.Fatalf("BuildForPrint failed: %v", err)
	}

	if mesh == nil {
		t.Fatal("Mesh is nil")
	}

	if len(mesh.Vertices) == 0 {
		t.Error("No vertices in mesh")
	}

	if len(mesh.Indices) == 0 {
		t.Error("No indices in mesh")
	}

	t.Logf("Print mesh built successfully: %d vertices, %d triangles",
		len(mesh.Vertices), len(mesh.Indices)/3)
}

func TestBuilderAutoZoom(t *testing.T) {
	provider := NewMockRasterProvider()
	builder := NewBuilder()

	tinGenerator := NewTINGenerator()
	builder.SetTINGenerator(tinGenerator)

	builder.SetRasterProvider(provider)
	builder.SetBounds(provider.Bounds(), provider.Srs())
	builder.SetAutoZoomRange(10, 18)

	zoom, err := builder.determineZoom()
	if err != nil {
		t.Fatalf("determineZoom failed: %v", err)
	}

	if zoom < 10 || zoom > 18 {
		t.Errorf("Zoom %d out of range [10, 18]", zoom)
	}

	t.Logf("Auto-determined zoom: %d", zoom)
}
