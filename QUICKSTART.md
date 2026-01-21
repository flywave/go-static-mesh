# 快速开始指南

## 安装

```bash
go get github.com/flywave/go-static-mesh
```

## 基本概念

### 数据源（Providers）

项目提供多种数据源来获取地理空间数据：

- **RasterProvider**: 高程数据（Image/GeoTIFF）
- **ImageryProvider**: 卫星影像（TMS/XYZ/GeoTIFF）
- **TinMeshProvider**: TIN 网格（Cesium quantized-mesh）
- **Model3DProvider**: 3D 模型（带位置的三维模型）
- **GeoDataProvider**: 地理数据（GPX/GeoJSON/KML）

### 输出格式

- **STL**: 3D 打印格式
- **GLB/GLTF**: Web 展示格式
- **OBJ**: 通用 3D 模型格式

### 两种使用模式

1. **3D 打印模式**: 地理数据突出显示到模型表面
2. **展示模式**: 地理数据绘制到贴图上

## 示例 1: 高程数据转 STL（3D 打印）

最简单的例子：将高程数据转换为 STL 模型用于 3D 打印

**注意**: 3D 打印需要闭合模型，TIN Mesh 是面片，需要闭合为体积

```go
package main

import (
    "github.com/flywave/go-static-mesh/static"
    "github.com/flywave/go-static-mesh/mesh"
    "github.com/flywave/go-tin"
)

func main() {
    // 1. 创建 Mesh Builder
    builder := mesh.NewBuilder()
    
    // 2. 设置 TIN 生成算法
    builder.SetTINGenerator(&tin.Algorithm{})
    
    // 3. 加载高程数据（RasterProvider 和 TinMeshProvider 只能选一个）
    provider := static.NewGeoTIFFRasterProvider("elevation.tif")
    builder.SetRasterProvider(provider)
    
    // 4. 设置范围（必须）
    bounds := vec2d.Rect{{minLat, minLng}, {maxLat, maxLng}}
    srs := geo.NewProj(4326)
    builder.SetBounds(bounds, srs)
    builder.SetVerticalExaggeration(1.0) // 垂直放大倍数
    
    // 5. 设置 zoom（可选）
    // 方式 1: 手动设置
    // builder.SetZoom(15)
    
    // 方式 2: 自动判断（推荐）
    builder.SetAutoZoomRange(10, 18)
    
    // 6. 设置 Tile 获取失败处理（可选）
    handler := mesh.TileErrorHandler{
        SkipMissing: true,
        MaxRetries:  3,
        Timeout:     30 * time.Second,
    }
    builder.SetTileErrorHandler(handler)
    
    // 7. 设置闭合参数（3D 打印必须）
    builder.SetCloseMesh(true, 5.0)  // 闭合模型，基础厚度 5mm
    
    // 8. 构建闭合的 Mesh
    terrainMesh, err := builder.BuildForPrint()
    if err != nil {
        panic(err)
    }
    
    // 9. 导出 STL
    writer := mesh.NewStlWriter()
    writer.SetBinary(true) // 二进制 STL
    writer.WriteFile(terrainMesh, "terrain.stl")
}
```

## 示例 2: 高程 + 卫星图转 GLB（展示）

生成带卫星图纹理的 3D 地形模型

**注意**: 展示模型是面片，不需要闭合

```go
package main

import (
    "github.com/flywave/go-static-mesh/static"
    "github.com/flywave/go-static-mesh/draw"
    "github.com/flywave/go-static-mesh/mesh"
    "github.com/flywave/go-tin"
)

func main() {
    // 1. 创建 Builder
    builder := mesh.NewBuilder()
    
    // 2. 设置 TIN 生成算法
    builder.SetTINGenerator(&tin.Algorithm{})
    
    // 3. 添加高程数据（RasterProvider 和 TinMeshProvider 只能选一个）
    elevationProvider := static.NewGeoTIFFRasterProvider("elevation.tif")
    builder.SetRasterProvider(elevationProvider)
    
    // 4. 添加卫星影像（使用 NewTileProvider）
    imageryProvider := static.NewTileProvider(
        "https://tile.openstreetmap.org/{z}/{x}/{y}.png",
        static.TileProviderModeXYZ,
    )
    builder.AddImageryProvider(imageryProvider)
    
    // 5. 设置范围（必须）
    bounds := vec2d.Rect{{minLat, minLng}, {maxLat, maxLng}}
    srs := geo.NewProj(4326)
    builder.SetBounds(bounds, srs)
    
    // 6. 设置 zoom（可选）
    // 方式 1: 手动设置
    // builder.SetZoom(15)
    
    // 方式 2: 自动判断（推荐）
    builder.SetAutoZoomRange(10, 18)
    
    // 7. 设置 Tile 获取失败处理（可选）
    handler := mesh.TileErrorHandler{
        SkipMissing: true,
        MaxRetries:  3,
        Timeout:     30 * time.Second,
    }
    builder.SetTileErrorHandler(handler)
    
    // 8. 构建带纹理的 Mesh（面片，不闭合）
    texturedMesh, err := builder.BuildForDisplayWithTexture()
    if err != nil {
        panic(err)
    }
    
    // 9. 导出 GLB
    writer := mesh.NewGltfWriter()
    writer.SetBinary(true)
    writer.SetEmbedImages(true) // 嵌入纹理到文件
    writer.WriteFile(texturedMesh, "terrain.glb")
}
```

## 示例 3: GPX 路径突出显示到 3D 模型（3D 打印）

将 GPX 轨迹突出显示在地形模型上

**注意**: 3D 打印需要闭合模型

```go
package main

import (
    "github.com/flywave/go-static-mesh/static"
    "github.com/flywave/go-static-mesh/draw"
    "github.com/flywave/go-static-mesh/mesh"
    "github.com/flywave/go-tin"
)

func main() {
    // 1. 创建 Builder
    builder := mesh.NewBuilder()
    
    // 2. 设置 TIN 生成算法
    builder.SetTINGenerator(&tin.Algorithm{})
    
    // 3. 添加高程数据（RasterProvider 和 TinMeshProvider 只能选一个）
    elevationProvider := static.NewGeoTIFFRasterProvider("elevation.tif")
    builder.SetRasterProvider(elevationProvider)
    
    // 4. 加载 GPX 文件
    gpxProvider := static.NewGPXProvider("track.gpx")
    paths := gpxProvider.GetPaths()
    
    // 5. 将路径添加为 MeshObject（突出显示到模型上）
    for _, path := range paths {
        builder.AddPath(path)
    }
    
    // 6. 设置突出显示参数
    builder.SetExtrudeGeoData(true, 3.0) // 3mm 高度突出显示
    
    // 7. 设置边界
    bounds := elevationProvider.Bounds()
    builder.SetBounds(bounds, elevationProvider.Srs())
    
    // 8. 设置闭合参数（3D 打印必须）
    builder.SetCloseMesh(true, 5.0)  // 闭合模型，基础厚度 5mm
    
    // 9. 构建闭合的 Mesh
    mesh, err := builder.BuildForPrint()
    if err != nil {
        panic(err)
    }
    
    // 10. 导出 STL
    writer := mesh.NewStlWriter()
    writer.WriteFile(mesh, "terrain_with_track.stl")
}
```

## 示例 4: GPX 路径绘制到纹理（展示）

将 GPX 轨迹绘制到卫星图纹理上

**注意**: 展示模型是面片，不需要闭合

```go
package main

import (
    "github.com/flywave/go-static-mesh/static"
    "github.com/flywave/go-static-mesh/draw"
    "github.com/flywave/go-static-mesh/mesh"
    "github.com/flywave/go-tin"
)

func main() {
    // 1. 创建 Builder
    builder := mesh.NewBuilder()
    
    // 2. 设置 TIN 生成算法
    builder.SetTINGenerator(&tin.Algorithm{})
    
    // 3. 添加高程数据（RasterProvider 和 TinMeshProvider 只能选一个）
    elevationProvider := static.NewGeoTIFFRasterProvider("elevation.tif")
    builder.SetRasterProvider(elevationProvider)
    
    // 4. 添加卫星影像（使用 NewTileProvider）
    imageryProvider := static.NewTileProvider(
        "https://tile.openstreetmap.org/{z}/{x}/{y}.png",
        static.TileProviderModeXYZ,
    )
    
    // 5. 加载 GPX 文件
    gpxProvider := static.NewGPXProvider("track.gpx")
    paths := gpxProvider.GetPaths()
    
    // 6. 创建绘制上下文
    ctx := draw.NewContext()
    ctx.SetTileProvider(imageryProvider)
    
    // 添加路径到绘制上下文
    for _, path := range paths {
        ctx.AddPath(path)
    }
    
    // 7. 设置边界和生成纹理
    bounds := elevationProvider.Bounds()
    ctx.SetBoundingBox(bounds, elevationProvider.Srs())
    ctx.SetSize(4096, 4096)
    
    texture, err := ctx.Render()
    if err != nil {
        panic(err)
    }
    
    // 8. 使用纹理构建 Mesh（面片，不闭合）
    builder.SetBounds(bounds, elevationProvider.Srs())
    builder.SetTexture(texture)
    
    mesh, err := builder.BuildForDisplayWithTexture()
    if err != nil {
        panic(err)
    }
    
    // 9. 导出 GLB
    writer := mesh.NewGltfWriter()
    writer.WriteFile(mesh, "terrain_with_track_texture.glb")
}
```

## 示例 5: TIN Mesh 数据

使用 Cesium quantized-mesh 格式

**注意**: 使用 TinMeshProvider 时，不需要 RasterProvider

```go
package main

import (
    "github.com/flywave/go-static-mesh/static"
    "github.com/flywave/go-static-mesh/mesh"
)

func main() {
    // 1. 创建 Builder
    builder := mesh.NewBuilder()
    
    // 2. 添加 TIN Mesh 数据源（与 RasterProvider 互斥）
    tinProvider := static.NewCesiumQuantizedMeshProvider(
        "https://terrain.example.com/{z}/{x}/{y}.terrain",
    )
    builder.SetTinMeshProvider(tinProvider)
    
    // 3. 设置边界
    bounds := tinProvider.Bounds()
    builder.SetBounds(bounds, tinProvider.Srs())
    
    // 4. 构建展示 Mesh（面片）
    mesh, err := builder.BuildForDisplay()
    if err != nil {
        panic(err)
    }
    
    // 5. 导出 OBJ
    writer := mesh.NewObjWriter()
    writer.WriteFile(mesh, "terrain.obj")
}
```

## 示例 6: 自定义 Provider

实现自己的数据源

```go
package main

import (
    "image"
    vec2d "github.com/flywave/go3d/float64/vec2"
    "github.com/flywave/go-geo"
    static "github.com/flywave/go-static-mesh"
    "github.com/flywave/go-static-mesh/mesh"
)

type MyRasterProvider struct {
    grid    *geo.TileGrid
    data    [][]float64
}

func (p *MyRasterProvider) Attribution() string {
    return "Custom Data"
}

func (p *MyRasterProvider) Grid() *geo.TileGrid {
    return p.grid
}

func (p *MyRasterProvider) Bounds() vec2d.Rect {
    return vec2d.Rect{{0, 0}, {10, 10}}
}

func (p *MyRasterProvider) Srs() geo.Proj {
    return geo.NewProj(4326) // WGS84
}

func (p *MyRasterProvider) GetElevation(lng, lat float64) float64 {
    // 实现高程查询
    return 100.0
}

func main() {
    builder := mesh.NewBuilder()
    
    provider := &MyRasterProvider{
        grid: geo.NewTileGrid(4326, [2]int{256, 256}, [2]float64{0, 0}, [2]float64{256, 256}),
    }
    
    builder.AddRasterProvider(provider)
    builder.SetBounds(provider.Bounds(), provider.Srs())
    
    mesh, err := builder.Build()
    if err != nil {
        panic(err)
    }
    
    writer := mesh.NewStlWriter()
    writer.WriteFile(mesh, "custom.stl")
}
```

## 常用参数说明

### Builder 参数

```go
builder.SetResolution(0.5) // 分辨率（mm），越小越精细
builder.SetVerticalExaggeration(2.0) // 垂直放大倍数
builder.SetBaseElevation(0.0) // 基础高程偏移
builder.SetExtrudeGeoData(true, 3.0) // 地理数据突出显示（启用，高度3mm）
```

### 范围设置（必须）

```go
// 设置地理范围（必须）
bounds := vec2d.Rect{{minLat, minLng}, {maxLat, maxLng}}
srs := geo.NewProj(4326) // WGS84
builder.SetBounds(bounds, srs)
```

### Zoom 设置（可选）

```go
// 方式 1: 手动设置 zoom
builder.SetZoom(15) // 强制使用 zoom 15

// 方式 2: 自动判断 zoom（推荐）
builder.SetAutoZoomRange(10, 18) // 在 10-18 之间自动判断
```

### Tile 获取失败处理

```go
// 方式 1: 跳过失败的 tile（推荐用于展示）
handler := TileErrorHandler{
    SkipMissing: true,
    MaxRetries:  3,
    Timeout:     30 * time.Second,
}
builder.SetTileErrorHandler(handler)

// 方式 2: 使用 NoData 值填充（推荐用于 3D 打印）
handler := TileErrorHandler{
    UseNoData:   true,
    NoDataValue: 0.0,
    MaxRetries:  3,
    Timeout:     30 * time.Second,
}
builder.SetTileErrorHandler(handler)

// 方式 3: 严格模式（任何失败都报错）
handler := TileErrorHandler{
    SkipMissing: false,
    UseNoData:   false,
    MaxRetries:  5,
    Timeout:     60 * time.Second,
}
builder.SetTileErrorHandler(handler)
```

### STL Writer 参数

```go
writer.SetBinary(true) // 二进制格式（推荐）
writer.SetMergeMesh(true) // 合并网格
```

### GLTF Writer 参数

```go
writer.SetBinary(true) // 输出 GLB 而非 GLTF
writer.SetEmbedImages(true) // 嵌入纹理图像
writer.SetDraco(true) // 使用 Draco 压缩
```

### OBJ Writer 参数

```go
writer.SetSeparateMtl(true) // 生成独立的 .mtl 文件
```

## 性能优化建议

1. **设置合适的分辨率**: 
   - 3D 打印: 0.3 - 0.8mm
   - 展示: 1.0 - 2.0mm

2. **使用二进制格式**:
   - STL: `writer.SetBinary(true)`
   - GLTF: `writer.SetBinary(true)`

3. **限制区域大小**:
   - 避免处理过大的区域
   - 可以分块处理并合并

4. **使用缓存**:
   - 瓦片数据会被自动缓存
   - 避免重复下载相同数据

## 常见问题

### Q: 如何将多个高程文件合并？

```go
// 添加多个 RasterProvider
builder.AddRasterProvider(provider1)
builder.AddRasterProvider(provider2)
// Builder 会自动处理重叠区域
```

### Q: 如何调整模型的高度？

```go
builder.SetVerticalExaggeration(2.0) // 放大2倍
builder.SetBaseElevation(10.0) // 抬高10mm
```

### Q: 如何生成带颜色的 STL？

STL 不支持颜色，可以使用 GLB/OBJ 格式代替

### Q: 如何处理大型数据集？

```go
builder.SetResolution(2.0) // 降低分辨率
// 或者分块处理
```

## 下一步

- 查看 [ARCHITECTURE.md](./ARCHITECTURE.md) 了解整体架构
- 查看 [INTERFACES.md](./INTERFACES.md) 了解接口设计
- 查看 GoDoc 获取完整 API 文档
