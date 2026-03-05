package draw

import (
	"os"
	"path/filepath"
)

func GetTestFontPath() string {
	fonts := []string{
		"fonts/DejaVuSans.ttf",
		"../fonts/DejaVuSans.ttf",
		"../../fonts/DejaVuSans.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
	}

	for _, font := range fonts {
		if _, err := os.Stat(font); err == nil {
			absPath, _ := filepath.Abs(font)
			return absPath
		}
	}

	return ""
}

func GetTestFontsDir() string {
	dirs := []string{
		"fonts",
		"../fonts",
		"../../fonts",
		"/usr/share/fonts/truetype/dejavu",
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); err == nil {
			absPath, _ := filepath.Abs(dir)
			return absPath
		}
	}

	return ""
}

func SkipIfNoFont(t interface{ Skip(args ...interface{}) }) string {
	fontPath := GetTestFontPath()
	if fontPath == "" {
		t.Skip("Font file not found, skipping test")
	}
	return fontPath
}

func SkipIfNoFontsDir(t interface{ Skip(args ...interface{}) }) string {
	fontsDir := GetTestFontsDir()
	if fontsDir == "" {
		t.Skip("Fonts directory not found, skipping test")
	}
	return fontsDir
}
