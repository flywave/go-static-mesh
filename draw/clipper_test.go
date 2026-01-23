package draw

import (
	"testing"

	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
	"image/color"
)

func TestNewClipper(t *testing.T) {
	clipper := NewClipper()
	if clipper == nil {
		t.Fatal("NewClipper returned nil")
	}
}

func TestClipPathToBounds_PathInside(t *testing.T) {
	clipper := NewClipper()

	positions := []vec2d.T{
		{0.5, 0.5},
		{1.5, 0.5},
		{1.5, 1.5},
		{0.5, 1.5},
		{0.5, 0.5},
	}

	path := NewPathWithHeight(positions, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, 5.0, 10.0)

	bounds := vec2d.Rect{
		Min: vec2d.T{0.0, 0.0},
		Max: vec2d.T{2.0, 2.0},
	}

	clipped := clipper.ClipPathToBounds(path, bounds)

	if clipped == nil {
		t.Fatal("Clipped path is nil when path is inside bounds")
	}

	if len(clipped.Positions) < 2 {
		t.Errorf("Clipped path has too few positions: got %d, want >= 2", len(clipped.Positions))
	}
}

func TestClipPathToBounds_PathOutside(t *testing.T) {
	clipper := NewClipper()

	positions := []vec2d.T{
		{5.0, 5.0},
		{6.0, 5.0},
		{6.0, 6.0},
		{5.0, 6.0},
		{5.0, 5.0},
	}

	path := NewPathWithHeight(positions, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, 5.0, 10.0)

	bounds := vec2d.Rect{
		Min: vec2d.T{0.0, 0.0},
		Max: vec2d.T{2.0, 2.0},
	}

	clipped := clipper.ClipPathToBounds(path, bounds)

	if clipped != nil {
		t.Fatal("Clipped path should be nil when path is outside bounds")
	}
}

func TestClipPathToBounds_PathPartiallyOverlapping(t *testing.T) {
	clipper := NewClipper()

	positions := []vec2d.T{
		{1.5, 1.5},
		{3.5, 1.5},
		{3.5, 3.5},
		{1.5, 3.5},
		{1.5, 1.5},
	}

	path := NewPathWithHeight(positions, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, 5.0, 10.0)

	bounds := vec2d.Rect{
		Min: vec2d.T{0.0, 0.0},
		Max: vec2d.T{2.0, 2.0},
	}

	clipped := clipper.ClipPathToBounds(path, bounds)

	if clipped == nil {
		t.Fatal("Clipped path should not be nil when path partially overlaps bounds")
	}

	if len(clipped.Positions) < 2 {
		t.Errorf("Clipped path has too few positions: got %d, want >= 2", len(clipped.Positions))
	}

	for _, pos := range clipped.Positions {
		if pos[0] < bounds.Min[0] || pos[0] > bounds.Max[0] ||
			pos[1] < bounds.Min[1] || pos[1] > bounds.Max[1] {
			t.Errorf("Clipped position %v is outside bounds", pos)
		}
	}
}

func TestClipAreaToBounds_AreaInside(t *testing.T) {
	clipper := NewClipper()

	positions := []vec2d.T{
		{0.5, 0.5},
		{1.5, 0.5},
		{1.5, 1.5},
		{0.5, 1.5},
		{0.5, 0.5},
	}

	area := NewAreaWithHeight(positions, geo.NewProj(4326), color.RGBA{255, 0, 0, 255},
		color.RGBA{0, 255, 0, 128}, 5.0, 10.0)

	bounds := vec2d.Rect{
		Min: vec2d.T{0.0, 0.0},
		Max: vec2d.T{2.0, 2.0},
	}

	clipped := clipper.ClipAreaToBounds(area, bounds)

	if clipped == nil {
		t.Fatal("Clipped area is nil when area is inside bounds")
	}

	if len(clipped.Positions) < 3 {
		t.Errorf("Clipped area has too few positions: got %d, want >= 3", len(clipped.Positions))
	}
}

func TestClipAreaToBounds_AreaOutside(t *testing.T) {
	clipper := NewClipper()

	positions := []vec2d.T{
		{5.0, 5.0},
		{6.0, 5.0},
		{6.0, 6.0},
		{5.0, 6.0},
		{5.0, 5.0},
	}

	area := NewAreaWithHeight(positions, geo.NewProj(4326), color.RGBA{255, 0, 0, 255},
		color.RGBA{0, 255, 0, 128}, 5.0, 10.0)

	bounds := vec2d.Rect{
		Min: vec2d.T{0.0, 0.0},
		Max: vec2d.T{2.0, 2.0},
	}

	clipped := clipper.ClipAreaToBounds(area, bounds)

	if clipped != nil {
		t.Fatal("Clipped area should be nil when area is outside bounds")
	}
}

func TestClipAreaToBounds_AreaPartiallyOverlapping(t *testing.T) {
	clipper := NewClipper()

	positions := []vec2d.T{
		{1.5, 1.5},
		{3.5, 1.5},
		{3.5, 3.5},
		{1.5, 3.5},
		{1.5, 1.5},
	}

	area := NewAreaWithHeight(positions, geo.NewProj(4326), color.RGBA{255, 0, 0, 255},
		color.RGBA{0, 255, 0, 128}, 5.0, 10.0)

	bounds := vec2d.Rect{
		Min: vec2d.T{0.0, 0.0},
		Max: vec2d.T{2.0, 2.0},
	}

	clipped := clipper.ClipAreaToBounds(area, bounds)

	if clipped == nil {
		t.Fatal("Clipped area should not be nil when area partially overlaps bounds")
	}

	if len(clipped.Positions) < 3 {
		t.Errorf("Clipped area has too few positions: got %d, want >= 3", len(clipped.Positions))
	}

	for _, pos := range clipped.Positions {
		if pos[0] < bounds.Min[0] || pos[0] > bounds.Max[0] ||
			pos[1] < bounds.Min[1] || pos[1] > bounds.Max[1] {
			t.Errorf("Clipped position %v is outside bounds", pos)
		}
	}
}

func TestClipPathToBounds_EmptyPath(t *testing.T) {
	clipper := NewClipper()

	positions := []vec2d.T{}
	path := NewPathWithHeight(positions, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, 5.0, 10.0)

	bounds := vec2d.Rect{
		Min: vec2d.T{0.0, 0.0},
		Max: vec2d.T{2.0, 2.0},
	}

	clipped := clipper.ClipPathToBounds(path, bounds)

	if clipped != nil {
		t.Fatal("Clipped path should be nil for empty path")
	}
}

func TestClipAreaToBounds_TriangleArea(t *testing.T) {
	clipper := NewClipper()

	positions := []vec2d.T{
		{0.5, 0.5},
		{1.5, 0.5},
		{1.0, 1.5},
		{0.5, 0.5},
	}

	area := NewAreaWithHeight(positions, geo.NewProj(4326), color.RGBA{255, 0, 0, 255},
		color.RGBA{0, 255, 0, 128}, 5.0, 10.0)

	bounds := vec2d.Rect{
		Min: vec2d.T{0.0, 0.0},
		Max: vec2d.T{2.0, 2.0},
	}

	clipped := clipper.ClipAreaToBounds(area, bounds)

	if clipped == nil {
		t.Fatal("Clipped area is nil when triangle is inside bounds")
	}

	if len(clipped.Positions) < 3 {
		t.Errorf("Clipped area has too few positions: got %d, want >= 3", len(clipped.Positions))
	}
}
