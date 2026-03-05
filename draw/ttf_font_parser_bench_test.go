package draw

import (
	"os"
	"testing"
)

func BenchmarkTTFFontParserParseFont(b *testing.B) {
	fontPath := GetTestFontPath()
	if fontPath == "" {
		b.Skip("Font file not found, skipping benchmark")
	}

	data, err := os.ReadFile(fontPath)
	if err != nil {
		b.Fatalf("Failed to read font file: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		parser := NewTTFFontParser()
		err := parser.ParseFont(data)
		if err != nil {
			b.Fatalf("Failed to parse font: %v", err)
		}
	}
}

func BenchmarkTTFFontParserGetGlyphPaths(b *testing.B) {
	fontPath := GetTestFontPath()
	if fontPath == "" {
		b.Skip("Font file not found, skipping benchmark")
	}

	parser, err := NewTTFFontParserFromFile(fontPath)
	if err != nil {
		b.Fatalf("Failed to create parser: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := parser.GetGlyphPaths('A')
		if err != nil {
			b.Fatalf("Failed to get glyph paths: %v", err)
		}
	}
}

func BenchmarkTTFFontParserGetTextPaths(b *testing.B) {
	fontPath := GetTestFontPath()
	if fontPath == "" {
		b.Skip("Font file not found, skipping benchmark")
	}

	parser, err := NewTTFFontParserFromFile(fontPath)
	if err != nil {
		b.Fatalf("Failed to create parser: %v", err)
	}

	text := "Hello World"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, err := parser.GetTextPaths(text, 72.0)
		if err != nil {
			b.Fatalf("Failed to get text paths: %v", err)
		}
	}
}

func BenchmarkTTFFontParserGetTextPathsLong(b *testing.B) {
	fontPath := GetTestFontPath()
	if fontPath == "" {
		b.Skip("Font file not found, skipping benchmark")
	}

	parser, err := NewTTFFontParserFromFile(fontPath)
	if err != nil {
		b.Fatalf("Failed to create parser: %v", err)
	}

	text := "The quick brown fox jumps over the lazy dog. " +
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit."

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, err := parser.GetTextPaths(text, 72.0)
		if err != nil {
			b.Fatalf("Failed to get text paths: %v", err)
		}
	}
}

func BenchmarkFolderFontCacheLoad(b *testing.B) {
	fontsDir := GetTestFontsDir()
	if fontsDir == "" {
		b.Skip("Fonts directory not found, skipping benchmark")
	}

	dejaVuNamer := func(fontData FontData) string {
		return "DejaVuSans.ttf"
	}

	cache := NewFolderFontCache(fontsDir)
	cache.namer = dejaVuNamer

	fontData := FontData{
		Name:   "DejaVu",
		Family: FontFamilySans,
		Style:  FontStyleNormal,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := cache.Load(fontData)
		if err != nil {
			b.Fatalf("Failed to load font: %v", err)
		}
	}
}

func BenchmarkSyncFolderFontCacheLoad(b *testing.B) {
	fontsDir := GetTestFontsDir()
	if fontsDir == "" {
		b.Skip("Fonts directory not found, skipping benchmark")
	}

	dejaVuNamer := func(fontData FontData) string {
		return "DejaVuSans.ttf"
	}

	cache := NewSyncFolderFontCache(fontsDir)
	cache.setNamer(dejaVuNamer)

	fontData := FontData{
		Name:   "DejaVu",
		Family: FontFamilySans,
		Style:  FontStyleNormal,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := cache.Load(fontData)
		if err != nil {
			b.Fatalf("Failed to load font: %v", err)
		}
	}
}

func BenchmarkSyncFolderFontCacheConcurrent(b *testing.B) {
	fontsDir := GetTestFontsDir()
	if fontsDir == "" {
		b.Skip("Fonts directory not found, skipping benchmark")
	}

	dejaVuNamer := func(fontData FontData) string {
		return "DejaVuSans.ttf"
	}

	cache := NewSyncFolderFontCache(fontsDir)
	cache.setNamer(dejaVuNamer)

	fontData := FontData{
		Name:   "DejaVu",
		Family: FontFamilySans,
		Style:  FontStyleNormal,
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := cache.Load(fontData)
			if err != nil {
				b.Fatalf("Failed to load font: %v", err)
			}
		}
	})
}

func BenchmarkTTFFontParserGetTextWidth(b *testing.B) {
	fontPath := GetTestFontPath()
	if fontPath == "" {
		b.Skip("Font file not found, skipping benchmark")
	}

	parser, err := NewTTFFontParserFromFile(fontPath)
	if err != nil {
		b.Fatalf("Failed to create parser: %v", err)
	}

	text := "Hello World"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = parser.GetTextWidth(text, 72.0)
	}
}

func BenchmarkTTFFontParserGetTextHeight(b *testing.B) {
	fontPath := GetTestFontPath()
	if fontPath == "" {
		b.Skip("Font file not found, skipping benchmark")
	}

	parser, err := NewTTFFontParserFromFile(fontPath)
	if err != nil {
		b.Fatalf("Failed to create parser: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = parser.GetTextHeight(72.0)
	}
}
