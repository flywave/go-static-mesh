package draw

import (
	"image/color"
	"testing"

	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

func TestNewArea(t *testing.T) {
	positions := []vec2d.T{
		{35.6895, 139.6917},
		{35.6896, 139.6918},
		{35.6897, 139.6919},
	}

	area := NewArea(positions, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 128}, 5.0)

	if area == nil {
		t.Fatal("NewArea returned nil")
	}

	if len(area.Positions) != 3 {
		t.Errorf("expected 3 positions, got %d", len(area.Positions))
	}

	if area.Weight != 5.0 {
		t.Errorf("expected weight 5.0, got %f", area.Weight)
	}
}

func TestAreaBounds(t *testing.T) {
	positions := []vec2d.T{
		{35.6895, 139.6917},
		{35.6896, 139.6918},
		{35.6897, 139.6919},
	}

	area := NewArea(positions, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 128}, 5.0)

	bounds := area.Bounds()

	if bounds.Min[0] != 35.6895 || bounds.Min[1] != 139.6917 {
		t.Errorf("unexpected bounds min: %v", bounds.Min)
	}

	if bounds.Max[0] != 35.6897 || bounds.Max[1] != 139.6919 {
		t.Errorf("unexpected bounds max: %v", bounds.Max)
	}
}

func TestAreaSrsProj(t *testing.T) {
	srs := geo.NewProj(4326)
	positions := []vec2d.T{{35.6895, 139.6917}}

	area := NewArea(positions, srs, color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 128}, 5.0)

	retrievedSrs := area.SrsProj()

	if !retrievedSrs.Eq(srs) {
		t.Error("SRS not retrieved correctly")
	}
}

func TestAreaExtraMarginPixels(t *testing.T) {
	positions := []vec2d.T{{35.6895, 139.6917}}

	area := NewArea(positions, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 128}, 5.0)

	l, top, r, b := area.ExtraMarginPixels()

	_ = l
	_ = top
	_ = r
	_ = b
}

func TestNewPath(t *testing.T) {
	positions := []vec2d.T{
		{35.6895, 139.6917},
		{35.6896, 139.6918},
	}

	path := NewPath(positions, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, 5.0)

	if path == nil {
		t.Fatal("NewPath returned nil")
	}

	if len(path.Positions) != 2 {
		t.Errorf("expected 2 positions, got %d", len(path.Positions))
	}

	if path.Weight != 5.0 {
		t.Errorf("expected weight 5.0, got %f", path.Weight)
	}

	if path.Height != 10.0 {
		t.Errorf("expected default height 10.0, got %f", path.Height)
	}
}

func TestPathBounds(t *testing.T) {
	positions := []vec2d.T{
		{35.6895, 139.6917},
		{35.6896, 139.6918},
		{35.6897, 139.6919},
	}

	path := NewPath(positions, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, 5.0)

	bounds := path.Bounds()

	if bounds.Min[0] != 35.6895 || bounds.Min[1] != 139.6917 {
		t.Errorf("unexpected bounds min: %v", bounds.Min)
	}

	if bounds.Max[0] != 35.6897 || bounds.Max[1] != 139.6919 {
		t.Errorf("unexpected bounds max: %v", bounds.Max)
	}
}

func TestPathSrsProj(t *testing.T) {
	srs := geo.NewProj(4326)
	positions := []vec2d.T{{35.6895, 139.6917}}

	path := NewPath(positions, srs, color.RGBA{255, 0, 0, 255}, 5.0)

	retrievedSrs := path.SrsProj()

	if !retrievedSrs.Eq(srs) {
		t.Error("SRS not retrieved correctly")
	}
}

func TestPathExtraMarginPixels(t *testing.T) {
	positions := []vec2d.T{{35.6895, 139.6917}}

	path := NewPath(positions, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, 5.0)

	l, top, r, b := path.ExtraMarginPixels()

	_ = l
	_ = top
	_ = r
	_ = b
}

func TestNewMarker(t *testing.T) {
	position := vec2d.T{35.6895, 139.6917}

	marker := NewMarker(position, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, 10.0)

	if marker == nil {
		t.Fatal("NewMarker returned nil")
	}

	if marker.Position != position {
		t.Error("position not set correctly")
	}

	if marker.Size != 10.0 {
		t.Errorf("expected size 10.0, got %f", marker.Size)
	}
}

func TestNewMarkerWithHeight(t *testing.T) {
	position := vec2d.T{35.6895, 139.6917}

	marker := NewMarkerWithHeight(position, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, 10.0, 15.0)

	if marker == nil {
		t.Fatal("NewMarkerWithHeight returned nil")
	}

	if marker.Height != 15.0 {
		t.Errorf("expected height 15.0, got %f", marker.Height)
	}
}

func TestMarkerBounds(t *testing.T) {
	position := vec2d.T{35.6895, 139.6917}

	marker := NewMarker(position, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, 10.0)

	bounds := marker.Bounds()

	if bounds.Min != position || bounds.Max != position {
		t.Errorf("marker bounds should be a point at the position")
	}
}

func TestMarkerSrsProj(t *testing.T) {
	srs := geo.NewProj(4326)
	position := vec2d.T{35.6895, 139.6917}

	marker := NewMarker(position, srs, color.RGBA{255, 0, 0, 255}, 10.0)

	retrievedSrs := marker.SrsProj()

	if !retrievedSrs.Eq(srs) {
		t.Error("SRS not retrieved correctly")
	}
}

func TestMarkerExtraMarginPixels(t *testing.T) {
	position := vec2d.T{35.6895, 139.6917}

	marker := NewMarker(position, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, 10.0)

	l, top, r, b := marker.ExtraMarginPixels()

	_ = l
	_ = top
	_ = r
	_ = b
}

func TestMarkerSetLabelColor(t *testing.T) {
	position := vec2d.T{35.6895, 139.6917}
	marker := NewMarker(position, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, 10.0)

	labelColor := color.RGBA{0, 255, 0, 255}
	marker.SetLabelColor(labelColor)

	if marker.LabelColor != labelColor {
		t.Error("label color not set correctly")
	}
}

func TestNewCircle(t *testing.T) {
	position := vec2d.T{35.6895, 139.6917}

	circle := NewCircle(position, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 128}, 100.0, 5.0)

	if circle == nil {
		t.Fatal("NewCircle returned nil")
	}

	if circle.Position != position {
		t.Error("position not set correctly")
	}

	if circle.Radius != 100.0 {
		t.Errorf("expected radius 100.0, got %f", circle.Radius)
	}

	if circle.Weight != 5.0 {
		t.Errorf("expected weight 5.0, got %f", circle.Weight)
	}
}

func TestNewCircleWithHeight(t *testing.T) {
	position := vec2d.T{35.6895, 139.6917}

	circle := NewCircleWithHeight(position, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 128}, 100.0, 5.0, 15.0)

	if circle == nil {
		t.Fatal("NewCircleWithHeight returned nil")
	}

	if circle.Height != 15.0 {
		t.Errorf("expected height 15.0, got %f", circle.Height)
	}
}

func TestCircleBounds(t *testing.T) {
	position := vec2d.T{35.6895, 139.6917}

	circle := NewCircle(position, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 128}, 100.0, 5.0)

	bounds := circle.Bounds()

	// Circle bounds should include the radius
	if bounds.Min[0] > position[0] || bounds.Max[0] < position[0] {
		t.Error("circle bounds should contain the center position")
	}
}

func TestCircleSrsProj(t *testing.T) {
	srs := geo.NewProj(4326)
	position := vec2d.T{35.6895, 139.6917}

	circle := NewCircle(position, srs, color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 128}, 100.0, 5.0)

	retrievedSrs := circle.SrsProj()

	if !retrievedSrs.Eq(srs) {
		t.Error("SRS not retrieved correctly")
	}
}

func TestCircleExtraMarginPixels(t *testing.T) {
	position := vec2d.T{35.6895, 139.6917}

	circle := NewCircle(position, geo.NewProj(4326), color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 128}, 100.0, 5.0)

	l, top, r, b := circle.ExtraMarginPixels()

	_ = l
	_ = top
	_ = r
	_ = b
}
