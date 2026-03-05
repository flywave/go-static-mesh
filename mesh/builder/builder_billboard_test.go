package builder

import (
	"image/color"
	"testing"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-static-mesh/draw"
	"github.com/flywave/go-static-mesh/mesh"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

func TestBuilderAddBillboard(t *testing.T) {
	builder := NewBuilder()

	pos := vec2d.T{30.0, 120.0}
	srs := geo.NewProj(4326)
	text := "Test Billboard"

	billboard := draw.NewBillboard(pos, srs, text)
	billboard.SetMode(draw.BillboardModeDisplay)

	builder.AddBillboard(billboard)

	if len(builder.geoData) != 1 {
		t.Errorf("Expected 1 geo data object, got %d", len(builder.geoData))
	}

	addedBillboard, ok := builder.geoData[0].(*draw.Billboard)
	if !ok {
		t.Fatal("Expected geo data to be a Billboard")
	}

	if addedBillboard.Text != text {
		t.Errorf("Expected text '%s', got '%s'", text, addedBillboard.Text)
	}

	t.Logf("✅ Billboard added successfully: text='%s'", addedBillboard.Text)
}

func TestBuilderAddBillboards(t *testing.T) {
	builder := NewBuilder()

	billboards := []*draw.Billboard{
		draw.NewBillboard(vec2d.T{30.0, 120.0}, geo.NewProj(4326), "A"),
		draw.NewBillboard(vec2d.T{30.1, 120.1}, geo.NewProj(4326), "B"),
		draw.NewBillboard(vec2d.T{30.2, 120.2}, geo.NewProj(4326), "C"),
	}

	builder.AddBillboards(billboards)

	if len(builder.geoData) != 3 {
		t.Errorf("Expected 3 geo data objects, got %d", len(builder.geoData))
	}

	for i, obj := range builder.geoData {
		billboard, ok := obj.(*draw.Billboard)
		if !ok {
			t.Errorf("Object %d is not a Billboard", i)
			continue
		}

		expectedText := string(rune('A' + i))
		if billboard.Text != expectedText {
			t.Errorf("Object %d: expected text '%s', got '%s'", i, expectedText, billboard.Text)
		}
	}

	t.Logf("✅ %d billboards added successfully", len(billboards))
}

func TestBuilderExtractBillboards(t *testing.T) {
	builder := NewBuilder()

	billboards := []*draw.Billboard{
		draw.NewBillboard(vec2d.T{30.0, 120.0}, geo.NewProj(4326), "Billboard1"),
		draw.NewBillboard(vec2d.T{30.1, 120.1}, geo.NewProj(4326), "Billboard2"),
	}

	path := draw.NewPath([]vec2d.T{{0, 0}, {1, 0}, {1, 1}}, geo.NewProj(4326), color.White, 1.0)

	for _, billboard := range billboards {
		builder.AddGeoData(billboard)
	}
	builder.AddGeoData(path)

	if len(builder.geoData) != 3 {
		t.Fatalf("Expected 3 geo data objects, got %d", len(builder.geoData))
	}

	extracted := builder.extractBillboards()

	if len(extracted) != 2 {
		t.Errorf("Expected 2 extracted billboards, got %d", len(extracted))
	}

	if len(builder.geoData) != 1 {
		t.Errorf("Expected 1 remaining geo data object, got %d", len(builder.geoData))
	}

	t.Logf("✅ Extracted %d billboards, %d remaining objects", len(extracted), len(builder.geoData))
}

func TestBuilderProcessBillboards(t *testing.T) {
	builder := NewBuilder()

	billboards := []*draw.Billboard{
		draw.NewBillboard(vec2d.T{0, 0}, geo.NewProj(4326), "Test1"),
		draw.NewBillboard(vec2d.T{10, 10}, geo.NewProj(4326), "Test2"),
	}

	for _, billboard := range billboards {
		builder.AddBillboard(billboard)
	}

	result, err := builder.processBillboards(nil)
	if err != nil {
		t.Fatalf("Failed to process billboards: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result mesh")
	}

	if len(result.Vertices) == 0 {
		t.Error("Expected vertices in result mesh")
	}

	t.Logf("✅ Processed %d billboards: %d vertices, %d triangles",
		len(billboards), len(result.Vertices), result.TriangleCount())
}

func TestBuilderCombineMeshes(t *testing.T) {
	mesh1 := &mesh.Mesh{
		Vertices: []vec3d.T{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}},
		Indices:  []uint32{0, 1, 2},
	}

	mesh2 := &mesh.Mesh{
		Vertices: []vec3d.T{{2, 0, 0}, {3, 0, 0}, {3, 1, 0}},
		Indices:  []uint32{0, 1, 2},
	}

	builder := NewBuilder()
	combined, err := builder.combineMeshes(mesh1, mesh2)
	if err != nil {
		t.Fatalf("Failed to combine meshes: %v", err)
	}

	if len(combined.Vertices) != 6 {
		t.Errorf("Expected 6 vertices, got %d", len(combined.Vertices))
	}

	if len(combined.Indices) != 6 {
		t.Errorf("Expected 6 indices, got %d", len(combined.Indices))
	}

	t.Logf("✅ Combined meshes: %d vertices, %d indices", len(combined.Vertices), len(combined.Indices))
}

func TestBuilderMixedGeoData(t *testing.T) {
	builder := NewBuilder()

	bounds := vec2d.Rect{
		Min: vec2d.T{29.9, 119.9},
		Max: vec2d.T{30.1, 120.1},
	}
	builder.SetBounds(bounds, geo.NewProj(4326))

	builder.AddBillboard(draw.NewBillboard(vec2d.T{30.0, 120.0}, geo.NewProj(4326), "A"))
	builder.AddGeoData(draw.NewPath([]vec2d.T{{0, 0}, {1, 0}, {1, 1}}, geo.NewProj(4326), color.White, 1.0))
	builder.AddBillboard(draw.NewBillboard(vec2d.T{30.05, 120.05}, geo.NewProj(4326), "B"))

	if len(builder.geoData) != 3 {
		t.Errorf("Expected 3 geo data objects, got %d", len(builder.geoData))
	}

	billboards := builder.extractBillboards()

	if len(billboards) != 2 {
		t.Errorf("Expected 2 billboards, got %d", len(billboards))
	}

	if len(builder.geoData) != 1 {
		t.Errorf("Expected 1 non-billboard object, got %d", len(builder.geoData))
	}

	t.Logf("✅ Mixed geo data handled correctly: %d billboards, %d other objects",
		len(billboards), len(builder.geoData))
}
