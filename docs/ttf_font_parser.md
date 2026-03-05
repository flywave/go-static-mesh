# TTF 字体解析系统

## 概述

基于 `github.com/golang/freetype/truetype` 实现的字体解析系统，支持：
- 从文件或字节数据加载 TTF 字体
- 提取字形轮廓路径
- 将文字转换为矢量路径
- 字体缓存系统

## 字体缓存系统

### 1. FontData 结构

```go
type FontData struct {
    Name   string      // 字体名称
    Family FontFamily  // 字体家族（Sans/Serif/Mono）
    Style  FontStyle   // 字体样式（Normal/Bold/Italic）
}
```

### 2. 字体缓存类型

#### FolderFontCache
非线程安全的字体缓存，适合单线程使用：

```go
cache := draw.NewFolderFontCache("/path/to/fonts")
fontData := draw.FontData{
    Name:   "DejaVu",
    Family: draw.FontFamilySans,
    Style:  draw.FontStyleNormal,
}

font, err := cache.Load(fontData)
```

#### SyncFolderFontCache
线程安全的字体缓存，适合并发使用：

```go
cache := draw.NewSyncFolderFontCache("/path/to/fonts")
font, err := cache.Load(fontData)
```

### 3. 全局字体缓存

```go
// 设置字体文件夹
draw.SetFontFolder("/path/to/fonts")

// 自定义字体文件命名器
draw.SetFontNamer(func(fontData draw.FontData) string {
    switch fontData.Family {
    case draw.FontFamilySans:
        return "DejaVuSans.ttf"
    case draw.FontFamilySerif:
        return "DejaVuSerif.ttf"
    case draw.FontFamilyMono:
        return "DejaVuSansMono.ttf"
    }
    return "DejaVuSans.ttf"
})

// 从全局缓存获取字体
font := draw.GetFont(fontData)
```

### 4. 自定义字体缓存

```go
type CustomFontCache struct {
    // ...
}

func (c *CustomFontCache) Load(fontData draw.FontData) (*truetype.Font, error) {
    // 实现字体加载逻辑
}

func (c *CustomFontCache) Store(fontData draw.FontData, font *truetype.Font) {
    // 实现字体存储逻辑
}

// 使用自定义缓存
draw.SetFontCache(&CustomFontCache{})
```

## TTFFontParser 使用

### 1. 从文件创建

```go
parser, err := draw.NewTTFFontParserFromFile("/path/to/font.ttf")
if err != nil {
    log.Fatal(err)
}
```

### 2. 从字节数据创建

```go
data, err := os.ReadFile("/path/to/font.ttf")
if err != nil {
    log.Fatal(err)
}

parser, err := draw.NewTTFFontParserFromData(data)
if err != nil {
    log.Fatal(err)
}
```

### 3. 从缓存创建

```go
fontData := draw.FontData{
    Name:   "DejaVu",
    Family: draw.FontFamilySans,
    Style:  draw.FontStyleNormal,
}

parser, err := draw.NewTTFFontParserFromCache(fontData)
if err != nil {
    log.Fatal(err)
}
```

### 4. 获取字形路径

```go
// 获取单个字符的路径
paths, err := parser.GetGlyphPaths('A')
if err != nil {
    log.Fatal(err)
}

// paths 是二维数组，每个元素代表一个闭合路径
for i, path := range paths {
    fmt.Printf("Path %d: %d points\n", i, len(path))
    for _, pt := range path {
        fmt.Printf("  (%.2f, %.2f)\n", pt[0], pt[1])
    }
}
```

### 5. 获取文字路径

```go
text := "Hello World"
fontSize := 72.0

paths, width, err := parser.GetTextPaths(text, fontSize)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Text width: %.2f\n", width)
fmt.Printf("Number of paths: %d\n", len(paths))
```

### 6. 获取文字尺寸

```go
text := "Hello World"
fontSize := 72.0

// 获取文字宽度
width := parser.GetTextWidth(text, fontSize)
fmt.Printf("Text width: %.2f\n", width)

// 获取文字高度
height := parser.GetTextHeight(fontSize)
fmt.Printf("Font height: %.2f\n", height)
```

## 与 Billboard 集成

### Display 模式（纹理）

```go
// 1. 创建 billboard
billboard := draw.NewBillboardWithSize(pos, srs, "Hello", 10.0, 5.0, 0.5)
billboard.SetMode(draw.BillboardModeDisplay)

// 2. 设置字体路径
billboard.SetFontPath("/path/to/font.ttf")

// 3. 生成纹理（在 extruder 中自动调用）
texture, err := billboard.CreateDisplayTexture()
```

### Print 模式（3D 文字）

```go
// 1. 创建字体解析器
parser, err := draw.NewTTFFontParserFromFile("/path/to/font.ttf")
if err != nil {
    log.Fatal(err)
}

// 2. 创建 billboard
billboard := draw.NewBillboardWithSize(pos, srs, "Hello", 10.0, 5.0, 0.5)
billboard.SetMode(draw.BillboardModePrint)
billboard.SetFontParser(parser)

// 3. 生成 3D mesh
extruder := extruder.NewBillboardExtruder()
err = extruder.ExtrudeBillboardToTerrain(billboard, options)
```

## 完整示例

### 示例 1：提取文字轮廓

```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "github.com/flywave/go-static-mesh/draw"
)

func main() {
    // 加载字体
    parser, err := draw.NewTTFFontParserFromFile("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf")
    if err != nil {
        log.Fatal(err)
    }
    
    // 获取文字路径
    text := "Hello"
    fontSize := 100.0
    
    paths, width, err := parser.GetTextPaths(text, fontSize)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Text: %s\n", text)
    fmt.Printf("Font size: %.1f\n", fontSize)
    fmt.Printf("Total width: %.2f\n", width)
    fmt.Printf("Number of paths: %d\n", len(paths))
    
    // 输出每个路径的信息
    for i, path := range paths {
        fmt.Printf("\nPath %d: %d points\n", i, len(path))
        if len(path) > 0 {
            fmt.Printf("  Start: (%.2f, %.2f)\n", path[0][0], path[0][1])
            fmt.Printf("  End: (%.2f, %.2f)\n", path[len(path)-1][0], path[len(path)-1][1])
        }
    }
}
```

### 示例 2：使用字体缓存

```go
package main

import (
    "log"
    
    "github.com/flywave/go-static-mesh/draw"
)

func main() {
    // 设置字体文件夹
    draw.SetFontFolder("/usr/share/fonts/truetype/dejavu")
    
    // 自定义字体命名器
    draw.SetFontNamer(func(fontData draw.FontData) string {
        baseName := "DejaVuSans"
        
        switch fontData.Family {
        case draw.FontFamilySerif:
            baseName = "DejaVuSerif"
        case draw.FontFamilyMono:
            baseName = "DejaVuSansMono"
        }
        
        if fontData.Style&draw.FontStyleBold != 0 {
            baseName += "-Bold"
        }
        
        return baseName + ".ttf"
    })
    
    // 创建字体解析器
    fontData := draw.FontData{
        Name:   "DejaVu",
        Family: draw.FontFamilySans,
        Style:  draw.FontStyleNormal,
    }
    
    parser, err := draw.NewTTFFontParserFromCache(fontData)
    if err != nil {
        log.Fatal(err)
    }
    
    // 使用解析器
    paths, width, err := parser.GetTextPaths("Test", 72.0)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Got %d paths, width %.2f", len(paths), width)
}
```

### 示例 3：创建 Billboard

```go
package main

import (
    "image/color"
    "log"
    
    "github.com/flywave/go-geo"
    "github.com/flywave/go-static-mesh/draw"
    "github.com/flywave/go-static-mesh/mesh/extruder"
    vec2d "github.com/flywave/go3d/float64/vec2"
)

func main() {
    // 1. 创建字体解析器
    parser, err := draw.NewTTFFontParserFromFile("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf")
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. 创建 billboard (Print 模式)
    pos := vec2d.T{30.0, 120.0}
    srs := geo.NewProj(4326)
    
    billboard := draw.NewBillboardWithSize(pos, srs, "Welcome", 10.0, 5.0, 0.5)
    billboard.SetMode(draw.BillboardModePrint)
    billboard.SetFontParser(parser)
    billboard.SetColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
    billboard.SetBackground(color.RGBA{0xff, 0xff, 0xff, 0xff})
    
    // 3. 生成 mesh
    ext := extruder.NewBillboardExtruder()
    options := &extruder.BillboardExtrusionOptions{
        SampleRadius:  100.0,
        EnsureContact: true,
    }
    
    if err := ext.ExtrudeBillboardToTerrain(billboard, options); err != nil {
        log.Fatal(err)
    }
    
    mesh := ext.GetResult()
    log.Printf("Mesh created: %d vertices, %d indices", 
        len(mesh.Vertices), len(mesh.Indices))
}
```

## 路径数据格式

`GetGlyphPaths` 和 `GetTextPaths` 返回的路径是二维点数组：

```go
[][]vec2d.T  // 外层：多个闭合路径，内层：路径上的点
```

每个路径代表一个闭合的轮廓，由一系列二维点组成。这些点可以直接用于：
- 2D 绘图
- 3D 挤出
- 转换为 SVG 路径
- 其他矢量图形处理

## 注意事项

1. **字体格式**：仅支持 TrueType 字体（.ttf）
2. **字体缓存**：推荐使用 `SyncFolderFontCache` 处理并发访问
3. **字形索引**：如果字符在字体中不存在，返回 glyphIndex = 0，会报错 `ErrGlyphNotFound`
4. **单位**：
   - 路径坐标单位与字体设计单位相关
   - `fontSize` 参数会缩放路径到所需大小
5. **性能**：首次加载字体会解析整个字体文件，后续从缓存获取会更快

## 错误处理

```go
parser, err := draw.NewTTFFontParserFromFile(fontPath)
if err != nil {
    // 可能的错误：
    // - 文件不存在
    // - 字体文件格式错误
    // - 字体文件损坏
}

paths, err := parser.GetGlyphPaths('A')
if err != nil {
    // 可能的错误：
    // - ErrFontNotLoaded: 字体未加载
    // - ErrGlyphNotFound: 字符在字体中不存在
}
```
