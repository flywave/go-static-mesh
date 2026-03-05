# Billboard 纹理绘制流程

## 概述

Billboard 支持两种模式：
1. **Display 模式** (`BillboardModeDisplay`): 将文字绘制到纹理上，贴到 box 表面
2. **Print 模式** (`BillboardModePrint`): 将文字转换为 3D 几何体，与 box 进行布尔融合

## Display 模式纹理绘制流程

### 1. 创建 Billboard

```go
pos := vec2d.T{30.0, 120.0}  // 经纬度坐标
srs := geo.NewProj(4326)      // 坐标系统
text := "Hello World"

billboard := draw.NewBillboardWithSize(pos, srs, text, 10.0, 5.0, 0.5)
// 参数：位置、坐标系统、文字、宽度(米)、高度(米)、厚度(米)
```

### 2. 配置字体和纹理

```go
// 设置字体路径（可选，不设置会使用系统默认字体）
billboard.SetFontPath("/path/to/font.ttf")

// 设置颜色
billboard.SetColor(color.RGBA{0x00, 0x00, 0x00, 0xff})      // 文字颜色：黑色
billboard.SetBackground(color.RGBA{0xff, 0xff, 0xff, 0xff}) // 背景颜色：白色

// 设置纹理参数
billboard.SetTextureDPI(300)     // 纹理 DPI（影响纹理质量）
billboard.SetTextureScale(1.0)   // 纹理缩放因子

// 设置字体大小（可选，默认为高度的 60%）
billboard.SetFontSize(3.0)  // 字体高度：3米
```

### 3. 生成纹理

```go
texture, err := billboard.CreateDisplayTexture()
if err != nil {
    log.Fatal(err)
}
```

纹理生成算法：
- **纹理尺寸**: 基于宽高比自动选择合适的分辨率
  - 默认: 512x256 像素
  - 宽高比 > 2.0: 1024x512 像素
  - 宽高比 < 0.5: 256x512 像素
  - 受 `TextureScale` 缩放
  - 限制范围: 64-4096 像素

- **字体大小**: 基于纹理高度计算
  - 初始大小 = 纹理高度 × 50%
  - 如果文字超出纹理，自动缩放字体大小以适应

- **文字布局**: 居中对齐
  - 保留 10% 边距
  - 自动缩放以确保文字完整显示

### 4. 生成 3D Mesh

```go
import "github.com/flywave/go-static-mesh/mesh/extruder"

// 创建 extruder
extruder := extruder.NewBillboardExtruder()

// 设置选项
options := &extruder.BillboardExtrusionOptions{
    TerrainMesh:   terrainMesh,  // 地形 mesh（可选）
    SampleRadius:  100.0,        // 采样半径
    EnsureContact: true,         // 确保与地形接触
}

// 挤出 billboard
err := extruder.ExtrudeBillboardToTerrain(billboard, options)
if err != nil {
    log.Fatal(err)
}

// 获取结果 mesh
resultMesh := extruder.GetResult()
```

对于 **Display 模式**：
- 创建 box mesh
- 生成纹理并应用
- 自动生成 UV 坐标
- **不需要布尔融合**

对于 **Print 模式**：
- 创建 box mesh
- 使用 `FontParser` 将文字转换为 3D 几何
- 与 box 进行布尔融合（union）

## 坐标换算关系

### 物理尺寸到纹理像素

```
纹理宽度(像素) = (物理宽度(米) / 0.0254) × DPI
纹理高度(像素) = (物理高度(米) / 0.0254) × DPI
```

其中：
- 0.0254 = 1 英寸（米）
- DPI = 每英寸像素数

### 字体大小换算

```
字体大小(points) = (字体物理高度(米) / 0.0254) × 72
```

其中：
- 72 = 1 英寸的 points 数
- 1 point = 1/72 英寸

## 系统字体查找

如果没有指定字体路径，系统会自动查找以下路径：

**Linux**:
- `/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf`
- `/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf`
- `/usr/share/fonts/TTF/DejaVuSans.ttf`

**Windows**:
- `C:\Windows\Fonts\arial.ttf`

**macOS**:
- `/System/Library/Fonts/Helvetica.ttc`

如果找不到系统字体，需要手动设置字体路径。

## 完整示例

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
    // 1. 创建 billboard
    pos := vec2d.T{30.0, 120.0}
    srs := geo.NewProj(4326)
    
    billboard := draw.NewBillboardWithSize(pos, srs, "Welcome", 10.0, 5.0, 0.5)
    billboard.SetMode(draw.BillboardModeDisplay)
    
    // 2. 配置外观
    billboard.SetFontPath("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf")
    billboard.SetColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
    billboard.SetBackground(color.RGBA{0xff, 0xff, 0xff, 0xff})
    billboard.SetTextureDPI(300)
    
    // 3. 生成 mesh
    extruder := extruder.NewBillboardExtruder()
    options := &extruder.BillboardExtrusionOptions{
        SampleRadius:  100.0,
        EnsureContact: true,
    }
    
    if err := extruder.ExtrudeBillboardToTerrain(billboard, options); err != nil {
        log.Fatal(err)
    }
    
    mesh := extruder.GetResult()
    
    // 4. 使用 mesh (导出 GLTF, STL 等)
    // ...
}
```

## 注意事项

1. **Display 模式适合**：
   - 简单文字显示
   - 性能要求高
   - 不需要复杂的 3D 文字效果

2. **Print 模式适合**：
   - 需要凸出的 3D 文字效果
   - 需要与 box 进行复杂的布尔运算
   - 需要自定义字体解析器

3. **纹理质量**：
   - 提高 DPI 可以增加纹理质量，但会增加内存占用
   - 使用 TextureScale 可以在不改变 DPI 的情况下调整纹理大小
   - 推荐范围：DPI 150-600, Scale 0.5-2.0

4. **字体选择**：
   - 建议使用矢量字体（TTF/OTF）
   - 对于中文，需要确保字体支持中文字符
   - 字体文件路径需要正确
