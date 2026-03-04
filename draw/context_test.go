package draw

import (
	"image/color"
	"testing"

	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

func TestNewContext(t *testing.T) {
	ctx := NewContext()

	if ctx == nil {
		t.Fatal("NewContext returned nil")
	}

	if ctx.width != 512 {
		t.Errorf("expected default width 512, got %d", ctx.width)
	}

	if ctx.height != 512 {
		t.Errorf("expected default height 512, got %d", ctx.height)
	}

	// objects slice is lazily initialized when first used
}

func TestContextSetSize(t *testing.T) {
	ctx := NewContext()

	ctx.SetSize(800, 600)

	if ctx.width != 800 {
		t.Errorf("expected width 800, got %d", ctx.width)
	}

	if ctx.height != 600 {
		t.Errorf("expected height 600, got %d", ctx.height)
	}
}

func TestContextSetZoom(t *testing.T) {
	ctx := NewContext()

	zoom := 15
	ctx.SetZoom(zoom)

	if ctx.zoom == nil || *ctx.zoom != zoom {
		t.Errorf("expected zoom %d, got %v", zoom, ctx.zoom)
	}
}

func TestContextSetBoundingBox(t *testing.T) {
	ctx := NewContext()

	bbox := vec2d.Rect{Min: vec2d.T{30.0, 120.0}, Max: vec2d.T{31.0, 121.0}}
	srs := geo.NewProj(4326)

	ctx.SetBoundingBox(bbox, srs)

	if ctx.boundingBox == nil {
		t.Fatal("bounding box not set")
	}

	if *ctx.boundingBox != bbox {
		t.Error("bounding box not set correctly")
	}
}

func TestContextSetBackground(t *testing.T) {
	ctx := NewContext()

	bgColor := color.RGBA{255, 0, 0, 255}
	ctx.SetBackground(bgColor)

	if ctx.background != bgColor {
		t.Error("background color not set correctly")
	}
}

func TestContextAddMarker(t *testing.T) {
	ctx := NewContext()

	position := vec2d.T{35.6895, 139.6917}
	marker := NewMarker(position, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, 5.0)

	ctx.AddMarker(marker)

	if len(ctx.objects) != 1 {
		t.Errorf("expected 1 object, got %d", len(ctx.objects))
	}
}

func TestContextClearMarkers(t *testing.T) {
	ctx := NewContext()

	position := vec2d.T{35.6895, 139.6917}
	marker := NewMarker(position, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, 5.0)
	ctx.AddMarker(marker)

	// Add a path to test filtering
	pathPositions := []vec2d.T{{35.6895, 139.6917}, {35.6896, 139.6918}}
	path := NewPath(pathPositions, geo.NewProj(4326), color.RGBA{0, 255, 0, 255}, 5.0)
	ctx.AddPath(path)

	ctx.ClearMarkers()

	if len(ctx.objects) != 1 {
		t.Errorf("expected 1 object after clearing markers, got %d", len(ctx.objects))
	}
}

func TestContextAddPath(t *testing.T) {
	ctx := NewContext()

	positions := []vec2d.T{{35.6895, 139.6917}, {35.6896, 139.6918}}
	path := NewPath(positions, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, 5.0)

	ctx.AddPath(path)

	if len(ctx.objects) != 1 {
		t.Errorf("expected 1 object, got %d", len(ctx.objects))
	}
}

func TestContextClearPaths(t *testing.T) {
	ctx := NewContext()

	positions := []vec2d.T{{35.6895, 139.6917}, {35.6896, 139.6918}}
	path := NewPath(positions, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, 5.0)
	ctx.AddPath(path)

	// Add a marker to test filtering
	position := vec2d.T{35.6895, 139.6917}
	marker := NewMarker(position, geo.NewProj(4326), color.RGBA{0, 255, 0, 255}, 5.0)
	ctx.AddMarker(marker)

	ctx.ClearPaths()

	if len(ctx.objects) != 1 {
		t.Errorf("expected 1 object after clearing paths, got %d", len(ctx.objects))
	}
}

func TestContextAddArea(t *testing.T) {
	ctx := NewContext()

	positions := []vec2d.T{
		{35.6895, 139.6917},
		{35.6896, 139.6918},
		{35.6897, 139.6919},
	}
	area := NewArea(positions, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 128}, 5.0)

	ctx.AddArea(area)

	if len(ctx.objects) != 1 {
		t.Errorf("expected 1 object, got %d", len(ctx.objects))
	}
}

func TestContextClearAreas(t *testing.T) {
	ctx := NewContext()

	positions := []vec2d.T{
		{35.6895, 139.6917},
		{35.6896, 139.6918},
		{35.6897, 139.6919},
	}
	area := NewArea(positions, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 128}, 5.0)
	ctx.AddArea(area)

	// Add a path to test filtering
	pathPositions := []vec2d.T{{35.6895, 139.6917}, {35.6896, 139.6918}}
	path := NewPath(pathPositions, geo.NewProj(4326), color.RGBA{0, 255, 0, 255}, 5.0)
	ctx.AddPath(path)

	ctx.ClearAreas()

	if len(ctx.objects) != 1 {
		t.Errorf("expected 1 object after clearing areas, got %d", len(ctx.objects))
	}
}

func TestContextAddCircle(t *testing.T) {
	ctx := NewContext()

	circle := NewCircle(vec2d.T{35.6895, 139.6917}, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 128}, 100.0, 5.0)

	ctx.AddCircle(circle)

	if len(ctx.objects) != 1 {
		t.Errorf("expected 1 object, got %d", len(ctx.objects))
	}
}

func TestContextClearCircles(t *testing.T) {
	ctx := NewContext()

	circle := NewCircle(vec2d.T{35.6895, 139.6917}, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 128}, 100.0, 5.0)
	ctx.AddCircle(circle)

	// Add a path to test filtering
	pathPositions := []vec2d.T{{35.6895, 139.6917}, {35.6896, 139.6918}}
	path := NewPath(pathPositions, geo.NewProj(4326), color.RGBA{0, 255, 0, 255}, 5.0)
	ctx.AddPath(path)

	ctx.ClearCircles()

	if len(ctx.objects) != 1 {
		t.Errorf("expected 1 object after clearing circles, got %d", len(ctx.objects))
	}
}

func TestContextAddObject(t *testing.T) {
	ctx := NewContext()

	positions := []vec2d.T{{35.6895, 139.6917}, {35.6896, 139.6918}}
	path := NewPath(positions, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, 5.0)

	ctx.AddObject(path)

	if len(ctx.objects) != 1 {
		t.Errorf("expected 1 object, got %d", len(ctx.objects))
	}
}

func TestContextClearObjects(t *testing.T) {
	ctx := NewContext()

	positions := []vec2d.T{{35.6895, 139.6917}, {35.6896, 139.6918}}
	path := NewPath(positions, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, 5.0)
	ctx.AddObject(path)

	ctx.ClearObjects()

	if len(ctx.objects) != 0 {
		t.Errorf("expected 0 objects after clear, got %d", len(ctx.objects))
	}
}

func TestContextClearOverlays(t *testing.T) {
	ctx := NewContext()

	// Add an overlay (we can use nil for testing)
	ctx.overlays = append(ctx.overlays, nil)

	ctx.ClearOverlays()

	if ctx.overlays != nil {
		t.Error("overlays should be nil after clear")
	}
}

func TestContextOverrideAttribution(t *testing.T) {
	ctx := NewContext()

	attribution := "Custom Attribution"
	ctx.OverrideAttribution(attribution)

	if ctx.overrideAttribution == nil || *ctx.overrideAttribution != attribution {
		t.Error("attribution not overridden correctly")
	}
}
