# Fonts Directory

This directory contains font files used by the go-static-mesh project for text rendering.

## Included Fonts

### DejaVu Sans Family

- **DejaVuSans.ttf** - Regular sans-serif font
- **DejaVuSans-Bold.ttf** - Bold sans-serif font

### DejaVu Sans Mono Family

- **DejaVuSansMono.ttf** - Regular monospace font
- **DejaVuSansMono-Bold.ttf** - Bold monospace font
- **DejaVuSansMono-Oblique.ttf** - Italic monospace font
- **DejaVuSansMono-BoldOblique.ttf** - Bold italic monospace font

## Usage

### In Code

```go
import "github.com/flywave/go-static-mesh/draw"

// Method 1: Load from project fonts directory (automatic)
parser, err := draw.NewTTFFontParserFromFile("fonts/DejaVuSans.ttf")

// Method 2: Use font cache with default directory
fontData := draw.FontData{
    Name:   "DejaVu",
    Family: draw.FontFamilySans,
    Style:  draw.FontStyleNormal,
}
parser, err := draw.NewTTFFontParserFromCache(fontData)

// Method 3: Set custom fonts directory
draw.SetFontFolder("/path/to/fonts")
parser, err := draw.NewTTFFontParserFromCache(fontData)
```

### In Tests

```go
import "github.com/flywave/go-static-mesh/draw"

func TestMyFeature(t *testing.T) {
    fontPath := draw.SkipIfNoFont(t)
    // Use fontPath...
    
    fontsDir := draw.SkipIfNoFontsDir(t)
    // Use fontsDir...
}
```

## Font License

The DejaVu fonts are distributed under the Free License. For more information, see:
https://dejavu-fonts.github.io/License.html

## Adding Custom Fonts

To add your own fonts:

1. Place the TTF files in this directory
2. Update the font namer function if needed:

```go
draw.SetFontNamer(func(fontData draw.FontData) string {
    switch fontData.Family {
    case draw.FontFamilySans:
        return "MyCustomSans.ttf"
    case draw.FontFamilySerif:
        return "MyCustomSerif.ttf"
    case draw.FontFamilyMono:
        return "MyCustomMono.ttf"
    }
    return "DejaVuSans.ttf"
})
```

## Font Format Support

Currently, only TrueType fonts (TTF) are supported. OpenType fonts (OTF) may work but are not officially tested.
