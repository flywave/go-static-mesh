package draw

import (
	"os"
	"sync"
	"testing"
)

func TestFontParserInterface(t *testing.T) {
	fontPath := SkipIfNoFont(t)

	var _ FontParser = (*TTFFontParser)(nil)
	var _ FontParser = (*DefaultFontParser)(nil)

	t.Log("TTFFontParser implements FontParser interface")
	t.Log("DefaultFontParser implements FontParser interface")
	t.Logf("Using font: %s", fontPath)
}

func TestTTFFontParserParseFont(t *testing.T) {
	fontPath := SkipIfNoFont(t)

	data, err := os.ReadFile(fontPath)
	if err != nil {
		t.Fatalf("Failed to read font file: %v", err)
	}

	parser := NewTTFFontParser()

	if parser.font != nil {
		t.Error("Expected font to be nil before parsing")
	}

	err = parser.ParseFont(data)
	if err != nil {
		t.Fatalf("Failed to parse font: %v", err)
	}

	if parser.font == nil {
		t.Error("Expected font to be set after parsing")
	}
}

func TestTTFFontParserParseInvalidData(t *testing.T) {
	parser := NewTTFFontParser()

	invalidData := []byte("this is not a valid font file")

	err := parser.ParseFont(invalidData)
	if err == nil {
		t.Error("Expected error when parsing invalid font data")
	}

	t.Logf("Got expected error: %v", err)
}

func TestTTFFontParserParseEmptyData(t *testing.T) {
	parser := NewTTFFontParser()

	emptyData := []byte{}

	err := parser.ParseFont(emptyData)
	if err == nil {
		t.Error("Expected error when parsing empty data")
	}

	t.Logf("Got expected error: %v", err)
}

func TestTTFFontParserGetGlyphPathsWithoutFont(t *testing.T) {
	parser := NewTTFFontParser()

	paths, err := parser.GetGlyphPaths('A')

	if err != ErrFontNotLoaded {
		t.Errorf("Expected ErrFontNotLoaded, got: %v", err)
	}

	if paths != nil {
		t.Error("Expected nil paths when font not loaded")
	}
}

func TestTTFFontParserGetTextPathsWithoutFont(t *testing.T) {
	parser := NewTTFFontParser()

	paths, width, err := parser.GetTextPaths("Test", 72.0)

	if err != ErrFontNotLoaded {
		t.Errorf("Expected ErrFontNotLoaded, got: %v", err)
	}

	if paths != nil {
		t.Error("Expected nil paths when font not loaded")
	}

	if width != 0 {
		t.Errorf("Expected width 0, got: %f", width)
	}
}

func TestTTFFontParserGetGlyphPathsInvalidChar(t *testing.T) {
	fontPath := SkipIfNoFont(t)

	parser, err := NewTTFFontParserFromFile(fontPath)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	invalidChar := rune(0xFFFF)
	paths, err := parser.GetGlyphPaths(invalidChar)

	if err == nil {
		t.Error("Expected error for invalid character")
	}

	t.Logf("Got expected error for invalid char: %v", err)

	_ = paths
}

func TestTTFFontParserDifferentCharacters(t *testing.T) {
	fontPath := SkipIfNoFont(t)

	parser, err := NewTTFFontParserFromFile(fontPath)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	tests := []struct {
		name string
		char rune
	}{
		{"Uppercase A", 'A'},
		{"Lowercase a", 'a'},
		{"Digit 0", '0'},
		{"Digit 9", '9'},
		{"Space", ' '},
		{"Exclamation", '!'},
		{"At sign", '@'},
		{"Chinese character 中", '中'},
		{"Japanese character あ", 'あ'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths, err := parser.GetGlyphPaths(tt.char)

			if tt.char == ' ' {
				if err != nil {
					t.Logf("Space character returned error: %v", err)
				}
				return
			}

			if err != nil {
				t.Logf("Character %c returned error: %v", tt.char, err)
				return
			}

			if len(paths) == 0 {
				t.Logf("Character %c has no paths", tt.char)
				return
			}

			t.Logf("Character %c: %d paths", tt.char, len(paths))
		})
	}
}

func TestTTFFontParserEmptyText(t *testing.T) {
	fontPath := SkipIfNoFont(t)

	parser, err := NewTTFFontParserFromFile(fontPath)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	paths, width, err := parser.GetTextPaths("", 72.0)
	if err != nil {
		t.Fatalf("Failed to get text paths: %v", err)
	}

	if len(paths) != 0 {
		t.Errorf("Expected 0 paths for empty text, got %d", len(paths))
	}

	if width != 0 {
		t.Errorf("Expected width 0 for empty text, got %f", width)
	}
}

func TestTTFFontParserDifferentFontSizes(t *testing.T) {
	fontPath := SkipIfNoFont(t)

	parser, err := NewTTFFontParserFromFile(fontPath)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	text := "Test"
	sizes := []float64{10, 20, 50, 72, 100, 200}

	for _, size := range sizes {
		t.Run(string(rune(int(size))), func(t *testing.T) {
			paths, width, err := parser.GetTextPaths(text, size)
			if err != nil {
				t.Fatalf("Failed to get text paths: %v", err)
			}

			if len(paths) == 0 {
				t.Error("Expected at least one path")
			}

			if width <= 0 {
				t.Errorf("Expected positive width, got %f", width)
			}

			t.Logf("Size %.0f: %d paths, width %.2f", size, len(paths), width)
		})
	}
}

func TestTTFFontParserLongText(t *testing.T) {
	fontPath := SkipIfNoFont(t)

	parser, err := NewTTFFontParserFromFile(fontPath)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	longText := "The quick brown fox jumps over the lazy dog. " +
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ abcdefghijklmnopqrstuvwxyz 0123456789"

	paths, width, err := parser.GetTextPaths(longText, 72.0)
	if err != nil {
		t.Fatalf("Failed to get text paths: %v", err)
	}

	if len(paths) == 0 {
		t.Error("Expected at least one path")
	}

	if width <= 0 {
		t.Errorf("Expected positive width, got %f", width)
	}

	t.Logf("Long text: %d paths, width %.2f", len(paths), width)
}

func TestTTFFontParserRepeatedText(t *testing.T) {
	fontPath := SkipIfNoFont(t)

	parser, err := NewTTFFontParserFromFile(fontPath)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	text1 := "AAA"
	text2 := "ABC"

	paths1, width1, err := parser.GetTextPaths(text1, 72.0)
	if err != nil {
		t.Fatalf("Failed to get text paths: %v", err)
	}

	paths2, width2, err := parser.GetTextPaths(text2, 72.0)
	if err != nil {
		t.Fatalf("Failed to get text paths: %v", err)
	}

	t.Logf("Text '%s': %d paths, width %.2f", text1, len(paths1), width1)
	t.Logf("Text '%s': %d paths, width %.2f", text2, len(paths2), width2)

	if len(paths1) == 0 || len(paths2) == 0 {
		t.Error("Expected at least one path for each text")
	}
}

func TestTTFFontParserGetFont(t *testing.T) {
	fontPath := SkipIfNoFont(t)

	parser, err := NewTTFFontParserFromFile(fontPath)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	font := parser.GetFont()
	if font == nil {
		t.Error("Expected font to be returned")
	}
}

func TestSyncFolderFontCacheConcurrent(t *testing.T) {
	fontsDir := SkipIfNoFontsDir(t)

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

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			font, err := cache.Load(fontData)
			if err != nil {
				t.Errorf("Goroutine %d: Failed to load font: %v", id, err)
				return
			}

			if font == nil {
				t.Errorf("Goroutine %d: Expected font to be loaded", id)
				return
			}

			t.Logf("Goroutine %d: Font loaded successfully", id)
		}(i)
	}

	wg.Wait()
	t.Log("Concurrent font loading test passed")
}

func TestSetFontCache(t *testing.T) {
	originalCache := GetGlobalFontCache()

	customCache := NewFolderFontCache("/tmp/fonts")
	SetFontCache(customCache)

	if GetGlobalFontCache() != customCache {
		t.Error("Expected custom cache to be set")
	}

	SetFontCache(nil)

	if GetGlobalFontCache() != originalCache {
		t.Error("Expected original cache to be restored")
	}

	t.Log("SetFontCache test passed")
}

func TestFontDataStyles(t *testing.T) {
	tests := []struct {
		name  string
		style FontStyle
	}{
		{"Normal", FontStyleNormal},
		{"Bold", FontStyleBold},
		{"Italic", FontStyleItalic},
		{"BoldItalic", FontStyleBold | FontStyleItalic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fontData := FontData{
				Name:   "Test",
				Family: FontFamilySans,
				Style:  tt.style,
			}

			if fontData.Style != tt.style {
				t.Errorf("Expected style %d, got %d", tt.style, fontData.Style)
			}
		})
	}
}

func TestFontDataFamilies(t *testing.T) {
	tests := []struct {
		name   string
		family FontFamily
	}{
		{"Sans", FontFamilySans},
		{"Serif", FontFamilySerif},
		{"Mono", FontFamilyMono},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fontData := FontData{
				Name:   "Test",
				Family: tt.family,
				Style:  FontStyleNormal,
			}

			if fontData.Family != tt.family {
				t.Errorf("Expected family %d, got %d", tt.family, fontData.Family)
			}
		})
	}
}

func TestTTFFontParserNewFromFileNotFound(t *testing.T) {
	nonExistentPath := "/tmp/nonexistent_font_12345.ttf"

	parser, err := NewTTFFontParserFromFile(nonExistentPath)
	if err == nil {
		t.Error("Expected error when font file not found")
	}

	if parser != nil {
		t.Error("Expected nil parser when font file not found")
	}

	t.Logf("Got expected error: %v", err)
}

func TestDefaultFontParser(t *testing.T) {
	parser := NewDefaultFontParser()

	err := parser.ParseFont([]byte("test"))
	if err != nil {
		t.Error("DefaultFontParser.ParseFont should return nil")
	}

	paths, err := parser.GetGlyphPaths('A')
	if err != nil {
		t.Error("DefaultFontParser.GetGlyphPaths should return nil error")
	}
	if paths != nil {
		t.Error("DefaultFontParser.GetGlyphPaths should return nil paths")
	}

	paths, width, err := parser.GetTextPaths("Test", 72.0)
	if err != nil {
		t.Error("DefaultFontParser.GetTextPaths should return nil error")
	}
	if paths != nil {
		t.Error("DefaultFontParser.GetTextPaths should return nil paths")
	}
	if width != 0 {
		t.Error("DefaultFontParser.GetTextPaths should return width 0")
	}
}

func TestFontCacheWithProjectFonts(t *testing.T) {
	fontsDir := SkipIfNoFontsDir(t)

	dejaVuNamer := func(fontData FontData) string {
		switch fontData.Family {
		case FontFamilySans:
			if fontData.Style&FontStyleBold != 0 {
				return "DejaVuSans-Bold.ttf"
			}
			return "DejaVuSans.ttf"
		case FontFamilyMono:
			if fontData.Style&FontStyleBold != 0 {
				if fontData.Style&FontStyleItalic != 0 {
					return "DejaVuSansMono-BoldOblique.ttf"
				}
				return "DejaVuSansMono-Bold.ttf"
			}
			if fontData.Style&FontStyleItalic != 0 {
				return "DejaVuSansMono-Oblique.ttf"
			}
			return "DejaVuSansMono.ttf"
		}
		return "DejaVuSans.ttf"
	}

	cache := NewFolderFontCache(fontsDir)
	cache.namer = dejaVuNamer

	fontData := FontData{
		Name:   "DejaVu",
		Family: FontFamilySans,
		Style:  FontStyleNormal,
	}

	font, err := cache.Load(fontData)
	if err != nil {
		t.Fatalf("Failed to load font: %v", err)
	}

	if font == nil {
		t.Fatal("Expected font to be loaded")
	}

	RegisterFont(fontData, font)

	parser, err := NewTTFFontParserFromCache(fontData)
	if err != nil {
		t.Fatalf("Failed to create font parser from cache: %v", err)
	}

	if parser.font == nil {
		t.Error("Font not loaded in parser")
	}

	paths, _, err := parser.GetTextPaths("Test", 48.0)
	if err != nil {
		t.Fatalf("Failed to get text paths: %v", err)
	}

	if len(paths) == 0 {
		t.Error("Expected at least one path")
	}

	t.Logf("Got %d paths from cached font", len(paths))
}

func TestFolderFontCacheWithProjectFonts(t *testing.T) {
	fontsDir := SkipIfNoFontsDir(t)

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

	font, err := cache.Load(fontData)
	if err != nil {
		t.Fatalf("Failed to load font: %v", err)
	}

	if font == nil {
		t.Error("Expected font to be loaded")
	}

	font2, err := cache.Load(fontData)
	if err != nil {
		t.Fatalf("Failed to load font from cache: %v", err)
	}

	if font != font2 {
		t.Error("Expected same font instance from cache")
	}

	t.Log("Font cache test passed")
}
