package draw

import (
	"image/color"
	"testing"

	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

func TestBillboardCompleteWorkflow(t *testing.T) {
	t.Run("DisplayMode_Complete", func(t *testing.T) {
		// 1. 创建 billboard
		pos := vec2d.T{30.0, 120.0}
		srs := geo.NewProj(4326)
		text := "Complete Test"
		width := 15.0
		height := 8.0
		thickness := 0.8

		billboard := NewBillboardWithSize(pos, srs, text, width, height, thickness)

		// 2. 验证基本属性
		if billboard.Position != pos {
			t.Errorf("Position not set correctly")
		}
		if billboard.Text != text {
			t.Errorf("Text not set correctly")
		}
		if billboard.Width != width {
			t.Errorf("Width not set correctly")
		}
		if billboard.Height != height {
			t.Errorf("Height not set correctly")
		}
		if billboard.Thickness != thickness {
			t.Errorf("Thickness not set correctly")
		}

		// 3. 配置 Display 模式
		billboard.SetMode(BillboardModeDisplay)
		if billboard.Mode != BillboardModeDisplay {
			t.Error("Mode not set to Display")
		}

		fontPath := SkipIfNoFont(t)
		billboard.SetFontPath(fontPath)
		if billboard.FontPath != fontPath {
			t.Error("FontPath not set correctly")
		}

		billboard.SetColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
		billboard.SetBackground(color.RGBA{0xff, 0xff, 0xff, 0xff})
		billboard.SetTextureDPI(300)
		billboard.SetTextureScale(1.5)

		// 4. 生成纹理
		texture, err := billboard.CreateDisplayTexture()
		if err != nil {
			t.Fatalf("Failed to create texture: %v", err)
		}

		if texture == nil {
			t.Fatal("Texture is nil")
		}

		bounds := texture.Bounds()
		if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
			t.Errorf("Invalid texture dimensions: %dx%d", bounds.Dx(), bounds.Dy())
		}

		t.Logf("✅ Display Mode Complete: texture %dx%d", bounds.Dx(), bounds.Dy())
	})

	t.Run("PrintMode_Complete", func(t *testing.T) {
		fontPath := SkipIfNoFont(t)

		// 1. 创建字体解析器
		parser, err := NewTTFFontParserFromFile(fontPath)
		if err != nil {
			t.Fatalf("Failed to create font parser: %v", err)
		}

		// 2. 创建 billboard
		pos := vec2d.T{30.0, 120.0}
		srs := geo.NewProj(4326)
		text := "3D"

		billboard := NewBillboardWithSize(pos, srs, text, 10.0, 5.0, 0.5)

		// 3. 配置 Print 模式
		billboard.SetMode(BillboardModePrint)
		if billboard.Mode != BillboardModePrint {
			t.Error("Mode not set to Print")
		}

		billboard.SetFontParser(parser)
		if billboard.FontParser == nil {
			t.Error("FontParser not set")
		}

		billboard.SetTextDepth(0.3)
		if billboard.TextDepth != 0.3 {
			t.Error("TextDepth not set correctly")
		}

		billboard.SetFontSize(3.0)
		if billboard.FontSize != 3.0 {
			t.Error("FontSize not set correctly")
		}

		// 4. 获取文字路径
		paths, width, err := billboard.FontParser.GetTextPaths(billboard.Text, billboard.FontSize)
		if err != nil {
			t.Fatalf("Failed to get text paths: %v", err)
		}

		if len(paths) == 0 {
			t.Error("No paths returned")
		}

		if width <= 0 {
			t.Errorf("Invalid width: %f", width)
		}

		t.Logf("✅ Print Mode Complete: %d paths, width %.2f", len(paths), width)
	})

	t.Run("CoordinateSystem", func(t *testing.T) {
		pos := vec2d.T{30.0, 120.0}
		srs := geo.NewProj(4326)

		billboard := NewBillboardWithSize(pos, srs, "Test", 10.0, 5.0, 0.5)

		// 测试2D角点
		corners2D := billboard.GetRotatedCorners()
		if len(corners2D) != 4 {
			t.Errorf("Expected 4 corners, got %d", len(corners2D))
		}

		// 验证角点位置
		for i, corner := range corners2D {
			if corner[0] < pos[0]-6.0 || corner[0] > pos[0]+6.0 {
				t.Errorf("Corner %d X out of bounds: %f", i, corner[0])
			}
			if corner[1] < pos[1]-3.0 || corner[1] > pos[1]+3.0 {
				t.Errorf("Corner %d Y out of bounds: %f", i, corner[1])
			}
		}

		// 测试3D角点
		corners3D := billboard.GetRotated3DCorners(0, 0.5)
		if len(corners3D) != 8 {
			t.Errorf("Expected 8 3D corners, got %d", len(corners3D))
		}

		// 测试边界
		bounds := billboard.Bounds()
		if bounds.Min[0] >= bounds.Max[0] || bounds.Min[1] >= bounds.Max[1] {
			t.Error("Invalid bounds")
		}

		// 测试 SRS
		if billboard.SrsProj() == nil {
			t.Error("SRS is nil")
		}

		t.Logf("✅ Coordinate System: 2D corners=%d, 3D corners=%d, bounds=%v",
			len(corners2D), len(corners3D), bounds)
	})

	t.Run("Rotation", func(t *testing.T) {
		billboard := NewBillboard(vec2d.T{0, 0}, geo.NewProj(4326), "Test")

		// 测试无旋转
		billboard.SetRotation(0)
		corners0 := billboard.GetRotatedCorners()

		// 测试90度旋转
		billboard.SetRotation(90)
		corners90 := billboard.GetRotatedCorners()

		// 验证旋转后的角点不同
		different := false
		for i := 0; i < 4; i++ {
			if corners0[i] != corners90[i] {
				different = true
				break
			}
		}

		if !different {
			t.Error("Rotation did not change corners")
		}

		t.Logf("✅ Rotation: corners changed after rotation")
	})

	t.Run("StringParsing", func(t *testing.T) {
		tests := []string{
			"30.0,120.0|text:Hello|width:10|height:5|thickness:0.5",
			"30.0,120.0|text:Test|mode:display",
			"30.0,120.0|text:3D|mode:print|fontsize:3",
		}

		for i, testStr := range tests {
			billboard, err := ParseBillboardString(testStr)
			if err != nil {
				t.Errorf("Test %d: Failed to parse: %v", i+1, err)
				continue
			}

			if billboard == nil {
				t.Errorf("Test %d: Billboard is nil", i+1)
				continue
			}

			t.Logf("✅ Parse Test %d: text='%s', mode=%d", i+1, billboard.Text, billboard.Mode)
		}
	})

	t.Run("ExtraMarginPixels", func(t *testing.T) {
		billboard := NewBillboardWithSize(vec2d.T{0, 0}, geo.NewProj(4326), "Test", 10.0, 5.0, 0.5)

		top, right, bottom, left := billboard.ExtraMarginPixels()

		expectedMargin := maxFloat64(billboard.Width, billboard.Height) / 2.0

		if top != expectedMargin || right != expectedMargin ||
			bottom != expectedMargin || left != expectedMargin {
			t.Errorf("Margin values incorrect: got (%.2f, %.2f, %.2f, %.2f), expected %.2f",
				top, right, bottom, left, expectedMargin)
		}

		t.Logf("✅ Extra Margin: %.2f pixels on all sides", expectedMargin)
	})
}

func TestBillboardEdgeCases(t *testing.T) {
	t.Run("EmptyText", func(t *testing.T) {
		billboard := NewBillboard(vec2d.T{0, 0}, geo.NewProj(4326), "")

		texture, err := billboard.CreateDisplayTexture()
		if err == nil {
			t.Error("Expected error for empty text")
		}

		if texture != nil {
			t.Error("Expected nil texture for empty text")
		}

		t.Logf("✅ Empty text handled correctly: %v", err)
	})

	t.Run("VeryLongText", func(t *testing.T) {
		fontPath := SkipIfNoFont(t)

		longText := "This is a very long text that should be scaled down to fit the texture properly"

		billboard := NewBillboardWithSize(vec2d.T{0, 0}, geo.NewProj(4326), longText, 10.0, 5.0, 0.5)
		billboard.SetFontPath(fontPath)

		texture, err := billboard.CreateDisplayTexture()
		if err != nil {
			t.Fatalf("Failed to create texture for long text: %v", err)
		}

		if texture == nil {
			t.Fatal("Texture is nil for long text")
		}

		t.Logf("✅ Long text handled: %d chars", len(longText))
	})

	t.Run("ZeroDimensions", func(t *testing.T) {
		billboard := NewBillboardWithSize(vec2d.T{0, 0}, geo.NewProj(4326), "Test", 0, 0, 0)

		if billboard.Width != 0 || billboard.Height != 0 {
			t.Error("Zero dimensions not preserved")
		}

		// 纹理生成应该使用默认值
		fontPath := SkipIfNoFont(t)
		billboard.SetFontPath(fontPath)

		texture, err := billboard.CreateDisplayTexture()
		if err != nil {
			t.Logf("Zero dimensions error (expected): %v", err)
		} else {
			t.Logf("✅ Zero dimensions handled: texture %dx%d",
				texture.Bounds().Dx(), texture.Bounds().Dy())
		}
	})

	t.Run("InvalidFontPath", func(t *testing.T) {
		billboard := NewBillboard(vec2d.T{0, 0}, geo.NewProj(4326), "Test")
		billboard.SetFontPath("/nonexistent/font.ttf")

		texture, err := billboard.CreateDisplayTexture()
		if err == nil {
			t.Error("Expected error for invalid font path")
		}

		if texture != nil {
			t.Error("Expected nil texture for invalid font path")
		}

		t.Logf("✅ Invalid font path handled: %v", err)
	})

	t.Run("SpecialCharacters", func(t *testing.T) {
		fontPath := SkipIfNoFont(t)

		specialText := "Test!@#$%^&*()"

		billboard := NewBillboard(vec2d.T{0, 0}, geo.NewProj(4326), specialText)
		billboard.SetFontPath(fontPath)

		texture, err := billboard.CreateDisplayTexture()
		if err != nil {
			t.Fatalf("Failed to handle special characters: %v", err)
		}
		_ = texture

		t.Logf("✅ Special characters handled: '%s'", specialText)
	})
}

func TestBillboardPerformance(t *testing.T) {
	fontPath := SkipIfNoFont(t)

	t.Run("TextureGenerationPerformance", func(t *testing.T) {
		billboard := NewBillboardWithSize(vec2d.T{0, 0}, geo.NewProj(4326), "Performance Test", 10.0, 5.0, 0.5)
		billboard.SetFontPath(fontPath)
		billboard.SetTextureDPI(300)

		// 生成10次，测量性能
		for i := 0; i < 10; i++ {
			texture, err := billboard.CreateDisplayTexture()
			if err != nil {
				t.Fatalf("Failed on iteration %d: %v", i, err)
			}
			_ = texture // 使用结果避免编译器警告
		}

		t.Logf("✅ Generated 10 textures successfully")
	})

	t.Run("MultipleBillboards", func(t *testing.T) {
		count := 10
		billboards := make([]*Billboard, count)

		for i := 0; i < count; i++ {
			pos := vec2d.T{30.0 + float64(i)*0.1, 120.0 + float64(i)*0.1}
			billboards[i] = NewBillboard(pos, geo.NewProj(4326), "Test")
			billboards[i].SetFontPath(fontPath)
		}

		// 为每个生成纹理
		for i, billboard := range billboards {
			texture, err := billboard.CreateDisplayTexture()
			if err != nil {
				t.Errorf("Failed for billboard %d: %v", i, err)
			}
			_ = texture
		}

		t.Logf("✅ Generated textures for %d billboards", count)
	})
}

func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
