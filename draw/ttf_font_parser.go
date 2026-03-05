package draw

import (
	"math"
	"os"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/math/fixed"

	vec2d "github.com/flywave/go3d/float64/vec2"
)

var (
	ErrFontNotLoaded = NewFontError("font not loaded")
	ErrFontNotFound  = NewFontError("font not found")
	ErrGlyphNotFound = NewFontError("glyph not found")
)

type FontError struct {
	msg string
}

func NewFontError(msg string) *FontError {
	return &FontError{msg: msg}
}

func (e *FontError) Error() string {
	return e.msg
}

type TTFFontParser struct {
	font     *truetype.Font
	fontData FontData
}

func NewTTFFontParser() *TTFFontParser {
	return &TTFFontParser{}
}

func NewTTFFontParserFromData(ttfData []byte) (*TTFFontParser, error) {
	parser := &TTFFontParser{}
	if err := parser.ParseFont(ttfData); err != nil {
		return nil, err
	}
	return parser, nil
}

func NewTTFFontParserFromFile(fontPath string) (*TTFFontParser, error) {
	parser := &TTFFontParser{}
	data, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, err
	}
	if err := parser.ParseFont(data); err != nil {
		return nil, err
	}
	return parser, nil
}

func NewTTFFontParserFromCache(fontData FontData) (*TTFFontParser, error) {
	parser := &TTFFontParser{
		fontData: fontData,
	}

	fnt := GetFont(fontData)
	if fnt == nil {
		return nil, ErrFontNotFound
	}

	parser.font = fnt
	return parser, nil
}

func (p *TTFFontParser) ParseFont(ttfData []byte) error {
	fnt, err := truetype.Parse(ttfData)
	if err != nil {
		return err
	}

	p.font = fnt
	return nil
}

func (p *TTFFontParser) GetGlyphPaths(char rune) ([][]vec2d.T, error) {
	if p.font == nil {
		return nil, ErrFontNotLoaded
	}

	glyphIndex := p.font.Index(char)
	if glyphIndex == 0 {
		return nil, ErrGlyphNotFound
	}

	paths := p.extractGlyphPaths(glyphIndex)

	return paths, nil
}

func (p *TTFFontParser) GetTextPaths(text string, fontSize float64) ([][]vec2d.T, float64, error) {
	if p.font == nil {
		return nil, 0, ErrFontNotLoaded
	}

	scale := fontSize / float64(p.font.FUnitsPerEm())

	var allPaths [][]vec2d.T
	currentX := 0.0

	for _, char := range text {
		glyphIndex := p.font.Index(char)
		if glyphIndex == 0 {
			continue
		}

		paths := p.extractGlyphPaths(glyphIndex)

		for i := range paths {
			for j := range paths[i] {
				paths[i][j][0] = (paths[i][j][0] + currentX) * scale
				paths[i][j][1] = paths[i][j][1] * scale
			}
		}

		allPaths = append(allPaths, paths...)

		advanceWidth := p.font.HMetric(fixed.I(int(p.font.FUnitsPerEm())), glyphIndex).AdvanceWidth
		currentX += float64(advanceWidth)
	}

	totalWidth := currentX * scale

	return allPaths, totalWidth, nil
}

func (p *TTFFontParser) extractGlyphPaths(glyphIndex truetype.Index) [][]vec2d.T {
	b := p.font.Bounds(fixed.I(int(p.font.FUnitsPerEm())))
	_ = b

	var paths [][]vec2d.T
	var currentPath []vec2d.T

	var g truetype.GlyphBuf
	err := g.Load(p.font, fixed.I(int(p.font.FUnitsPerEm())), glyphIndex, 0)
	if err != nil {
		return paths
	}

	points := g.Points
	ends := g.Ends

	if len(ends) == 0 {
		return paths
	}

	pointIdx := 0
	for _, end := range ends {
		currentPath = nil

		for i := pointIdx; i < int(end); i++ {
			pt := points[i]

			x := float64(pt.X) / 64.0
			y := float64(pt.Y) / 64.0

			currentPath = append(currentPath, vec2d.T{x, y})
		}

		if len(currentPath) > 2 {
			paths = append(paths, currentPath)
		}

		pointIdx = int(end)
	}

	return paths
}

func (p *TTFFontParser) GetFont() *truetype.Font {
	return p.font
}

func (p *TTFFontParser) GetTextWidth(text string, fontSize float64) float64 {
	if p.font == nil {
		return 0
	}

	scale := fontSize / float64(p.font.FUnitsPerEm())
	width := 0.0

	for _, char := range text {
		glyphIndex := p.font.Index(char)
		if glyphIndex == 0 {
			continue
		}

		advanceWidth := p.font.HMetric(fixed.I(int(p.font.FUnitsPerEm())), glyphIndex).AdvanceWidth
		width += float64(advanceWidth)
	}

	return width * scale
}

func (p *TTFFontParser) GetTextHeight(fontSize float64) float64 {
	if p.font == nil {
		return 0
	}

	scale := fontSize / float64(p.font.FUnitsPerEm())
	bounds := p.font.Bounds(fixed.I(int(p.font.FUnitsPerEm())))

	height := float64(bounds.Max.Y-bounds.Min.Y) / 64.0
	return math.Abs(height * scale)
}
