# Billboard 和字体系统完整指南

## 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                        Billboard                             │
│                                                               │
│  ┌──────────────────┐          ┌──────────────────┐        │
│  │  Display Mode    │          │   Print Mode     │        │
│  │  (纹理贴图)       │          │  (3D 凸出文字)   │        │
│  └──────────────────┘          └──────────────────┘        │
│           │                              │                  │
│           │                              │                  │
│  ┌────────▼─────────┐          ┌────────▼─────────┐        │
│  │ CreateDisplay    │          │ TTFFontParser    │        │
│  │ Texture()        │          │ GetTextPaths()   │        │
│  └──────────────────┘          └──────────────────┘        │
│           │                              │                  │
│           │                              │                  │
│  ┌────────▼─────────┐          ┌────────▼─────────┐        │
│  │  gg.Context      │          │ FontParser       │        │
│  │  (2D 绘图)       │          │ (字体轮廓提取)    │        │
│  └──────────────────┘          └──────────────────┘        │
│                                                               │
└─────────────────────────────────────────────────────────────┘
                              │
                              │
                              ▼
                    ┌──────────────────┐
                    │ BillboardExtruder│
                    │ (生成 3D Mesh)   │
                    └──────────────────┘
                              │
                ┌─────────────┴─────────────┐
                │                           │
        ┌───────▼────────┐        ┌────────▼───────┐
        │ Box + Texture  │        │ Box + 3D Text  │
        │ (Display Mode) │        │ (Print Mode)   │
        └────────────────┘        └────────────────┘
```

## 核心组件

### 1. Billboard (`draw/billboard.go`)

支持两种模式：

**Display Mode（显示模式）**：
- 文字绘制到纹理
- 纹理贴到 box 表面
- 性能好，适合简单显示

**Print Mode（打印模式）**：
- 文字转换为 3D 几何
- 与 box 布尔融合
- 效果好，适合高端展示

### 2. 字体缓存系统 (`draw/font_cache.go`)

- `FolderFontCache`：单线程字体缓存
- `SyncFolderFontCache`：线程安全字体缓存
- 支持自定义字体命名规则

### 3. TTF 字体解析器 (`draw/ttf_font_parser.go`)

- 从文件/数据/缓存加载字体
- 提取字形轮廓为矢量路径
- 支持文字到路径的转换

### 4. Billboard 挤出器 (`mesh/extruder/billboard_extruder.go`)

- 生成 billboard 的 3D mesh
- 支持地形贴合
- 自动处理布尔运算

## 快速开始

### Display 模式示例

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
    // 创建 billboard
    pos := vec2d.T{30.0, 120.0}
    srs := geo.NewProj(4326)
    
    billboard := draw.NewBillboardWithSize(pos, srs, "Hello World", 10.0, 5.0, 0.5)
    billboard.SetMode(draw.BillboardModeDisplay)
    billboard.SetFontPath("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf")
    billboard.SetColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
    billboard.SetBackground(color.RGBA{0xff, 0xff, 0xff, 0xff})
    billboard.SetTextureDPI(300)
    
    // 生成 mesh
    ext := extruder.NewBillboardExtruder()
    options := &extruder.BillboardExtrusionOptions{
        SampleRadius:  100.0,
        EnsureContact: true,
    }
    
    if err := ext.ExtrudeBillboardToTerrain(billboard, options); err != nil {
        log.Fatal(err)
    }
    
    mesh := ext.GetResult()
    log.Printf("Created mesh with texture: %d vertices", len(mesh.Vertices))
}
```

### Print 模式示例

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
    // 创建字体解析器
    parser, err := draw.NewTTFFontParserFromFile("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf")
    if err != nil {
        log.Fatal(err)
    }
    
    // 创建 billboard
    pos := vec2d.T{30.0, 120.0}
    srs := geo.NewProj(4326)
    
    billboard := draw.NewBillboardWithSize(pos, srs, "Hello", 10.0, 5.0, 0.5)
    billboard.SetMode(draw.BillboardModePrint)
    billboard.SetFontParser(parser)
    billboard.SetTextDepth(0.3)
    
    // 生成 mesh
    ext := extruder.NewBillboardExtruder()
    options := &extruder.BillboardExtrusionOptions{
        SampleRadius:  100.0,
        EnsureContact: true,
    }
    
    if err := ext.ExtrudeBillboardToTerrain(billboard, options); err != nil {
        log.Fatal(err)
    }
    
    mesh := ext.GetResult()
    log.Printf("Created mesh with 3D text: %d vertices", len(mesh.Vertices))
}
```

## 配置选项

### Billboard 配置

| 参数 | 说明 | 默认值 |
|------|------|--------|
| Width | 物理宽度（米） | 10.0 |
| Height | 物理高度（米） | 5.0 |
| Thickness | 厚度（米） | 0.5 |
| Rotation | 旋转角度（度） | 0 |
| TextDepth | 文字深度（米） | 0.3 |
| FontSize | 字体大小（米） | Height * 0.6 |
| FontPath | 字体文件路径 | 系统默认字体 |
| TextureDPI | 纹理 DPI | 300 |
| TextureScale | 纹理缩放 | 1.0 |
| Color | 文字颜色 | 黑色 |
| Background | 背景颜色 | 白色 |

### Extruder 配置

| 参数 | 说明 | 默认值 |
|------|------|--------|
| TerrainMesh | 地形 mesh | nil |
| SampleRadius | 采样半径（米） | 100.0 |
| EnsureContact | 确保与地形接触 | true |

## 性能优化

### Display 模式优化

1. **纹理分辨率**：
   - 低质量：DPI 150, Scale 0.5
   - 中等质量：DPI 300, Scale 1.0（推荐）
   - 高质量：DPI 600, Scale 2.0

2. **字体选择**：
   - 使用矢量字体（TTF）
   - 避免使用过大的字体文件

### Print 模式优化

1. **字体解析器**：
   - 使用缓存避免重复加载
   - 预加载常用字体

2. **布尔运算**：
   - 简化文字路径
   - 减少顶点数量

## 文件结构

```
draw/
├── billboard.go              # Billboard 核心实现
├── billboard_texture_test.go # 纹理测试
├── font_cache.go            # 字体缓存系统
├── font_parser.go           # FontParser 接口
├── ttf_font_parser.go       # TTF 字体解析器实现
└── ttf_font_parser_test.go  # 解析器测试

mesh/extruder/
└── billboard_extruder.go    # Billboard 挤出器

docs/
├── billboard_texture.md     # 纹理绘制文档
├── ttf_font_parser.md       # 字体解析器文档
└── billboard_guide.md       # 本文档
```

## 测试

```bash
# 运行所有 billboard 测试
go test -v ./draw -run TestBillboard

# 运行字体解析器测试
go test -v ./draw -run TestTTFFontParser

# 运行字体缓存测试
go test -v ./draw -run TestFontCache

# 运行所有测试
go test -v ./draw
```

## 常见问题

### 1. 字体加载失败

**问题**：`ErrFontNotFound` 或字体文件打开失败

**解决**：
- 检查字体文件路径是否正确
- 使用绝对路径
- 确保字体文件存在且可读

### 2. 纹理质量差

**问题**：文字模糊或像素化

**解决**：
- 提高 `TextureDPI`（例如从 300 提高到 600）
- 提高 `TextureScale`
- 使用更高质量的字体

### 3. 3D 文字布尔运算失败

**问题**：布尔融合报错

**解决**：
- 简化文字内容
- 减少 `TextDepth`
- 检查字体是否有复杂的字形

### 4. 内存占用过高

**问题**：生成 mesh 时内存占用大

**解决**：
- 降低纹理 DPI
- 使用 Display 模式代替 Print 模式
- 分批处理多个 billboard

## 相关文档

- [Billboard 纹理绘制详细文档](./billboard_texture.md)
- [TTF 字体解析器详细文档](./ttf_font_parser.md)
- [AGENTS.md - 代码风格和约定](../AGENTS.md)

## 示例项目

完整的示例代码位于：
- `draw/billboard_texture_test.go` - 纹理绘制测试
- `draw/ttf_font_parser_test.go` - 字体解析测试

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

[根据项目许可证]
