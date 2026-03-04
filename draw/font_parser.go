package draw

import vec2d "github.com/flywave/go3d/float64/vec2"

type FontParser interface {
	ParseFont(ttfData []byte) error
	GetGlyphPaths(char rune) ([][]vec2d.T, error)
	GetTextPaths(text string, fontSize float64) ([][]vec2d.T, float64, error)
}

type DefaultFontParser struct{}

func NewDefaultFontParser() *DefaultFontParser {
	return &DefaultFontParser{}
}

func (p *DefaultFontParser) ParseFont(ttfData []byte) error {
	return nil
}

func (p *DefaultFontParser) GetGlyphPaths(char rune) ([][]vec2d.T, error) {
	return nil, nil
}

func (p *DefaultFontParser) GetTextPaths(text string, fontSize float64) ([][]vec2d.T, float64, error) {
	return nil, 0, nil
}
