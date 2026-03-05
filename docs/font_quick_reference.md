# 字体系统快速参考

## 项目字体位置

```
go-static-mesh/fonts/
├── DejaVuSans.ttf              (742K) - Regular
├── DejaVuSans-Bold.ttf         (693K) - Bold
├── DejaVuSansMono.ttf          (336K) - Mono Regular
├── DejaVuSansMono-Bold.ttf     (327K) - Mono Bold
├── DejaVuSansMono-Oblique.ttf  (248K) - Mono Italic
└── DejaVuSansMono-BoldOblique.ttf (249K) - Mono Bold Italic
```

## 快速使用

### 1. 直接加载字体

```go
parser, err := draw.NewTTFFontParserFromFile("fonts/DejaVuSans.ttf")
```

### 2. 使用字体缓存（推荐）

```go
fontData := draw.FontData{
    Name:   "DejaVu",
    Family: draw.FontFamilySans,
    Style:  draw.FontStyleNormal,
}
parser, err := draw.NewTTFFontParserFromCache(fontData)
```

### 3. Billboard 纹理

```go
billboard := draw.NewBillboardWithSize(pos, srs, "Hello", 10.0, 5.0, 0.5)
billboard.SetFontPath("fonts/DejaVuSans.ttf")
billboard.SetMode(draw.BillboardModeDisplay)
```

### 4. Billboard 3D 文字

```go
parser, _ := draw.NewTTFFontParserFromFile("fonts/DejaVuSans.ttf")
billboard := draw.NewBillboardWithSize(pos, srs, "Hello", 10.0, 5.0, 0.5)
billboard.SetMode(draw.BillboardModePrint)
billboard.SetFontParser(parser)
```

## 测试工具

```go
// 在测试中自动获取字体路径
func TestMyFeature(t *testing.T) {
    fontPath := draw.SkipIfNoFont(t)
    fontsDir := draw.SkipIfNoFontsDir(t)
    // ...
}
```

## 字体样式

```go
// Sans 字体
FontFamilySans, FontStyleNormal      -> DejaVuSans.ttf
FontFamilySans, FontStyleBold        -> DejaVuSans-Bold.ttf

// Mono 字体
FontFamilyMono, FontStyleNormal      -> DejaVuSansMono.ttf
FontFamilyMono, FontStyleBold        -> DejaVuSansMono-Bold.ttf
FontFamilyMono, FontStyleItalic      -> DejaVuSansMono-Oblique.ttf
FontFamilyMono, FontStyleBold|Italic -> DejaVuSansMono-BoldOblique.ttf
```

## 字体缓存 API

```go
// 设置字体目录
draw.SetFontFolder("/path/to/fonts")

// 自定义字体命名
draw.SetFontNamer(func(fontData draw.FontData) string {
    return "MyCustomFont.ttf"
})

// 手动注册字体
draw.RegisterFont(fontData, font)

// 获取字体
font := draw.GetFont(fontData)

// 使用自定义缓存
draw.SetFontCache(myCustomCache)
```

## 性能参考

| 操作 | 性能 | 内存 |
|------|------|------|
| 解析字体 | 5,243 ns/op | 3,680 B/op |
| 获取字形 | 3,112 ns/op | 824 B/op |
| 获取文字路径 | 39,118 ns/op | 15,128 B/op |
| 缓存加载（命中） | 15.84 ns/op | 0 B/op |
| 缓存加载（并发） | 49.15 ns/op | 0 B/op |

## 常见问题

### Q: 如何添加自定义字体？

A: 将 TTF 文件放入 `fonts/` 目录，然后设置字体命名器：

```go
draw.SetFontNamer(func(fontData draw.FontData) string {
    if fontData.Name == "MyFont" {
        return "MyCustomFont.ttf"
    }
    return "DejaVuSans.ttf"
})
```

### Q: 测试时找不到字体？

A: 测试会自动查找字体，按以下顺序：
1. `fonts/DejaVuSans.ttf`
2. `../fonts/DejaVuSans.ttf`
3. `../../fonts/DejaVuSans.ttf`
4. 系统字体路径

### Q: 如何在不同环境使用？

A: 项目字体优先，系统字体作为后备。无需额外配置。

### Q: 支持哪些字体格式？

A: 目前仅支持 TrueType 字体（.ttf）。OpenType（.otf）可能工作但未测试。

## 文件位置

```
go-static-mesh/
├── fonts/              # 字体文件目录
│   ├── *.ttf          # 字体文件
│   └── README.md      # 字体说明
├── draw/
│   ├── font_cache.go  # 字体缓存
│   ├── test_utils.go  # 测试工具
│   └── ttf_font_parser.go # 字体解析器
└── docs/
    ├── font_setup_summary.md  # 详细总结
    └── font_quick_reference.md # 本文档
```

## 更多信息

- [详细配置总结](./font_setup_summary.md)
- [字体解析器文档](./ttf_font_parser.md)
- [Billboard 指南](./billboard_guide.md)
