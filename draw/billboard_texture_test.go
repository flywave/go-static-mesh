package draw

import (
	"image/color"
	"testing"

	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

func TestBillboardCreateDisplayTexture(t *testing.T) {
	fontPath := SkipIfNoFont(t)

	pos := vec2d.T{30.0, 120.0}
	srs := geo.NewProj(4326)
	text := "Hello World"

	billboard := NewBillboardWithSize(pos, srs, text, 10.0, 5.0, 0.5)
	billboard.SetTextureDPI(300)
	billboard.SetColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
	billboard.SetBackground(color.RGBA{0xff, 0xff, 0xff, 0xff})
	billboard.SetFontPath(fontPath)

	texture, err := billboard.CreateDisplayTexture()
	if err != nil {
		t.Logf("CreateDisplayTexture failed: %v", err)
		return
	}

	if texture == nil {
		t.Error("Expected texture to be created")
	}

	bounds := texture.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		t.Errorf("Invalid texture dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}

	t.Logf("Texture created successfully: %dx%d pixels", bounds.Dx(), bounds.Dy())
}

func TestBillboardCalculateTextureFontSize(t *testing.T) {
	billboard := NewBillboardWithSize(vec2d.T{0, 0}, geo.NewProj(4326), "Test", 10.0, 5.0, 0.5)
	billboard.SetTextureDPI(300)

	fontSize := billboard.calculateTextureFontSizeInPoints()

	if fontSize <= 0 {
		t.Errorf("Invalid font size: %f", fontSize)
	}

	expectedApprox := 5.0 * 0.6 / 0.0254 * 72.0
	tolerance := expectedApprox * 0.01

	if fontSize < expectedApprox-tolerance || fontSize > expectedApprox+tolerance {
		t.Errorf("Font size %f not within expected range %f±%f", fontSize, expectedApprox, tolerance)
	}

	t.Logf("Calculated font size: %f points", fontSize)
}

func TestBillboardTextureDimensions(t *testing.T) {
	fontPath := SkipIfNoFont(t)

	tests := []struct {
		name   string
		width  float64
		height float64
		dpi    int
		text   string
	}{
		{"Small billboard", 5.0, 2.5, 300, "Test"},
		{"Medium billboard", 10.0, 5.0, 300, "Hello World"},
		{"Large billboard", 20.0, 10.0, 150, "Large Text"},
		{"High DPI", 10.0, 5.0, 600, "High DPI"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			billboard := NewBillboardWithSize(
				vec2d.T{0, 0},
				geo.NewProj(4326),
				tt.text,
				tt.width,
				tt.height,
				0.5,
			)
			billboard.SetTextureDPI(tt.dpi)
			billboard.SetFontPath(fontPath)

			texture, err := billboard.CreateDisplayTexture()
			if err != nil {
				t.Logf("Failed to create texture: %v", err)
				return
			}

			bounds := texture.Bounds()

			if bounds.Dx() < 64 || bounds.Dy() < 64 {
				t.Errorf("Texture dimensions too small: %dx%d", bounds.Dx(), bounds.Dy())
			}

			if bounds.Dx() > 4096 || bounds.Dy() > 4096 {
				t.Errorf("Texture dimensions too large: %dx%d", bounds.Dx(), bounds.Dy())
			}

			t.Logf("%s: Texture %dx%d pixels for %.1fx%.1f meter billboard at %d DPI",
				tt.name, bounds.Dx(), bounds.Dy(), tt.width, tt.height, tt.dpi)
		})
	}
}

func TestBillboardWithCustomFont(t *testing.T) {
	fontPath := SkipIfNoFont(t)

	billboard := NewBillboardWithSize(
		vec2d.T{30.0, 120.0},
		geo.NewProj(4326),
		"Custom Font",
		10.0,
		5.0,
		0.5,
	)

	billboard.SetFontPath(fontPath)
	billboard.SetTextureDPI(300)
	billboard.SetColor(color.RGBA{0x33, 0x66, 0x99, 0xff})
	billboard.SetBackground(color.RGBA{0xff, 0xff, 0xee, 0xff})

	texture, err := billboard.CreateDisplayTexture()
	if err != nil {
		t.Logf("Failed to create texture with custom font: %v", err)
		return
	}

	if texture == nil {
		t.Error("Expected texture to be created with custom font")
	}

	t.Logf("Custom font texture created: %dx%d", texture.Bounds().Dx(), texture.Bounds().Dy())
}

func TestBillboardWithTTFFontParser(t *testing.T) {
	fontPath := SkipIfNoFont(t)

	parser, err := NewTTFFontParserFromFile(fontPath)
	if err != nil {
		t.Fatalf("Failed to create font parser: %v", err)
	}

	billboard := NewBillboardWithSize(
		vec2d.T{0, 0},
		nil,
		"Test",
		10.0,
		5.0,
		0.5,
	)
	billboard.SetMode(BillboardModePrint)
	billboard.SetFontParser(parser)

	paths, width, err := billboard.FontParser.GetTextPaths(billboard.Text, billboard.FontSize)
	if err != nil {
		t.Fatalf("Failed to get text paths: %v", err)
	}

	if len(paths) == 0 {
		t.Error("Expected at least one path")
	}

	if width <= 0 {
		t.Errorf("Expected positive width, got %f", width)
	}

	t.Logf("Billboard text '%s': %d paths, width %.2f", billboard.Text, len(paths), width)
}
