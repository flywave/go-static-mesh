# 字体配置总结

## 完成的工作

### 1. 创建项目字体目录
- ✅ 创建 `fonts/` 目录
- ✅ 复制 DejaVuSans 字体系列到项目
  - DejaVuSans.ttf (742K)
  - DejaVuSans-Bold.ttf (693K)
  - DejaVuSansMono.ttf (336K)
  - DejaVuSansMono-Bold.ttf (327K)
  - DejaVuSansMono-Oblique.ttf (248K)
  - DejaVuSansMono-BoldOblique.ttf (249K)

### 2. 更新字体路径配置

#### font_cache.go
- ✅ 添加 `getFontsDir()` 函数，自动检测项目字体目录
- ✅ 默认字体缓存路径改为项目 `fonts/` 目录

#### billboard.go
- ✅ 更新 `findSystemFont()` 函数，优先查找项目字体
- ✅ 查找顺序：
  1. `fonts/DejaVuSans.ttf` (当前目录)
  2. `../fonts/DejaVuSans.ttf` (上级目录)
  3. `../../fonts/DejaVuSans.ttf` (上上级目录)
  4. 系统字体路径 (Linux/Windows/macOS)

### 3. 创建测试工具函数

#### test_utils.go
- ✅ `GetTestFontPath()` - 获取测试字体路径
- ✅ `GetTestFontsDir()` - 获取测试字体目录
- ✅ `SkipIfNoFont(t)` - 如果没有字体则跳过测试
- ✅ `SkipIfNoFontsDir(t)` - 如果没有字体目录则跳过测试

### 4. 更新所有测试文件

#### ttf_font_parser_test.go
- ✅ 使用 `SkipIfNoFont()` 替代硬编码路径
- ✅ 使用 `SkipIfNoFontsDir()` 替代硬编码目录
- ✅ 所有测试从项目字体目录加载字体

#### ttf_font_parser_bench_test.go
- ✅ 使用 `GetTestFontPath()` 获取字体路径
- ✅ 使用 `GetTestFontsDir()` 获取字体目录
- ✅ 所有基准测试从项目字体目录加载字体

#### billboard_texture_test.go
- ✅ 使用 `SkipIfNoFont()` 获取字体路径
- ✅ Billboard 纹理测试使用项目字体

### 5. 创建文档

#### fonts/README.md
- ✅ 说明包含的字体文件
- ✅ 提供使用示例
- ✅ 说明字体许可证
- ✅ 说明如何添加自定义字体

#### .gitignore
- ✅ 创建 .gitignore 文件
- ✅ DejaVu 字体保持提交（开源许可证）
- ✅ 提供排除大字体的示例

## 字体查找优先级

### 程序运行时
1. **项目字体目录** - `fonts/DejaVuSans.ttf`
2. **上级目录** - `../fonts/DejaVuSans.ttf`
3. **系统字体** - `/usr/share/fonts/...`

### 测试时
1. **项目字体** - `fonts/DejaVuSans.ttf`
2. **相对路径** - `../fonts/` 或 `../../fonts/`
3. **系统字体** - 作为后备选项

## 性能基准

### 字体解析性能
```
BenchmarkTTFFontParserParseFont-12            396364    5243 ns/op    3680 B/op    2 allocs/op
```

### 字形提取性能
```
BenchmarkTTFFontParserGetGlyphPaths-12        393330    3112 ns/op     824 B/op   16 allocs/op
```

### 文字路径性能
```
BenchmarkTTFFontParserGetTextPaths-12          28431   39118 ns/op   15128 B/op  154 allocs/op
BenchmarkTTFFontParserGetTextPathsLong-12       2557  419057 ns/op  173889 B/op 1406 allocs/op
```

### 缓存性能
```
BenchmarkFolderFontCacheLoad-12              74670808      15.84 ns/op      0 B/op    0 allocs/op
BenchmarkSyncFolderFontCacheLoad-12          41757931      28.20 ns/op      0 B/op    0 allocs/op
BenchmarkSyncFolderFontCacheConcurrent-12    21561508      49.15 ns/op      0 B/op    0 allocs/op
```

## 使用示例

### 基本使用
```go
import "github.com/flywave/go-static-mesh/draw"

// 方法1：直接从项目字体加载
parser, err := draw.NewTTFFontParserFromFile("fonts/DejaVuSans.ttf")

// 方法2：使用缓存自动加载
fontData := draw.FontData{
    Name:   "DejaVu",
    Family: draw.FontFamilySans,
    Style:  draw.FontStyleNormal,
}
parser, err := draw.NewTTFFontParserFromCache(fontData)

// 方法3：在 Billboard 中使用
billboard := draw.NewBillboardWithSize(pos, srs, "Hello", 10.0, 5.0, 0.5)
billboard.SetFontPath("fonts/DejaVuSans.ttf")
```

### 测试中使用
```go
func TestMyFeature(t *testing.T) {
    // 自动获取字体路径，如果没有则跳过测试
    fontPath := draw.SkipIfNoFont(t)
    
    parser, err := draw.NewTTFFontParserFromFile(fontPath)
    // ...
}
```

## 测试结果

所有测试通过：
```
PASS
ok  	github.com/flywave/go-static-mesh/draw	0.107s
```

所有基准测试通过：
```
PASS
ok  	github.com/flywave/go-static-mesh/draw	14.406s
```

## 文件结构

```
go-static-mesh/
├── fonts/
│   ├── DejaVuSans.ttf
│   ├── DejaVuSans-Bold.ttf
│   ├── DejaVuSansMono.ttf
│   ├── DejaVuSansMono-Bold.ttf
│   ├── DejaVuSansMono-Oblique.ttf
│   ├── DejaVuSansMono-BoldOblique.ttf
│   └── README.md
├── draw/
│   ├── font_cache.go           (已更新)
│   ├── billboard.go             (已更新)
│   ├── test_utils.go            (新增)
│   ├── ttf_font_parser_test.go  (已更新)
│   ├── ttf_font_parser_bench_test.go (已更新)
│   └── billboard_texture_test.go (已更新)
└── .gitignore                   (新增)
```

## 优势

1. **独立性** - 项目不依赖系统字体
2. **可移植性** - 可以在任何环境运行
3. **一致性** - 所有环境使用相同的字体
4. **测试稳定性** - 测试不因缺少字体而失败
5. **易于部署** - 无需额外安装字体

## 许可证

DejaVu 字体使用自由许可证，可以自由使用、修改和分发。
详见：https://dejavu-fonts.github.io/License.html
