# go-static-mesh

用于生成 3D 模型的 Go 库，支持 3D 打印和展示用途。

## 特性

- 支持多种地理空间数据源：高程数据、卫星影像、TIN Mesh、3D 模型、地理矢量数据
- 输出多种格式：STL（3D 打印）、GLB/OBJ（展示）
- 基于 TIN Mesh 构建所有模型
- 灵活的数据获取接口设计
- 支持地理数据的不同展示方式（贴图或突出显示）
- Tile 缓存机制：减少重复网络请求，加速重复构建
- 可配置的缓存策略：支持 LRU 淘汰、TTL 过期
- GeoTIFF 支持：
  - 高程数据通过 GDAL 读取（`GeoTIFFRasterProvider`）
  - 影像数据通过 COG 读取（`GeoTIFFImageryProvider`）
  - 支持 `io.Reader` 接口，使用临时文件处理

## 核心概念

### TIN Mesh 是基础

所有模型输出都基于 TIN Mesh（不规则三角网）：

- Raster 高程数据 → TIN 算法 → TIN Mesh（面片）
- TIN Mesh → 闭合处理 → 闭合 Mesh（体积）→ STL（3D 打印）
- TIN Mesh → 完整细节 + 纹理 → GLB/OBJ（展示，面片）

**重要**: RasterProvider 和 TinMeshProvider 只能选择其一

### 数据源类型

- **RasterProvider**: 高程数据（Image/GeoTIFF）
- **ImageryProvider**: 卫星影像（TMS/XYZ/GeoTIFF）
- **TinMeshProvider**: TIN 网格（Cesium quantized-mesh）
- **Model3DProvider**: 3D 模型（带位置的三维模型）
- **GeoDataProvider**: 地理数据（GPX/GeoJSON/KML）

### 输出格式

- **STL**: 3D 打印格式（简化处理）
- **GLB/GLTF**: Web 展示格式（完整细节 + 纹理）
- **OBJ**: 通用 3D 模型格式（完整细节 + 纹理）

**重要**: GLB 和 OBJ 走完全相同的流程，仅输出格式不同

## 文档

- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - 整体架构设计
  - 项目概述和核心数据流
  - 5大模块设计
  - 包结构
  - 使用场景
  - 开发计划

- **[INTERFACES.md](./INTERFACES.md)** - 接口设计详解
  - 完整的接口层级和定义
  - Provider 接口实现
  - Mesh Builder 接口
  - Writer 接口
  - 完整使用示例

- **[TIN_MESH.md](./TIN_MESH.md)** - TIN Mesh 核心流程
  - TIN Mesh 数据结构
  - Raster 到 TIN 的转换流程
  - TIN 到不同输出格式的转换
  - 地理数据处理
  - TIN 优化
  - 范围和 Zoom 处理

- **[BOUNDS_AND_ZOOM.md](./BOUNDS_AND_ZOOM.md)** - 范围和 Zoom 处理指南
  - 外部传入范围
  - Zoom 级别设置和自动判断
  - Tile 获取失败处理策略
  - Tile 获取流程
  - 完整示例

- **[DEPENDENCIES.md](./DEPENDENCIES.md)** - 依赖库使用指南
  - go-tin: TIN 三角网算法
  - go-quantized-mesh: Cesium quantized-mesh 解码
  - go-mapbox: Raster 图片-高程解码
  - go-cog: GeoTIFF 读取
  - TileProvider: 瓦片提供者使用
  - 完整示例

- **[QUICKSTART.md](./QUICKSTART.md)** - 快速开始指南
  - 安装和基本概念
  - 实用示例
  - 常用参数说明
  - 性能优化建议
  - 常见问题解答

- **[AGENTS.md](./AGENTS.md)** - Agent 指南
  - 构建命令
  - 代码风格指南
  - 包结构约定

## 快速开始

### 安装

```bash
go get github.com/flywave/go-static-mesh
```

### 示例：使用 Tile 缓存加速构建

```go
package main

import (
    "github.com/flywave/go-static-mesh/mesh"
    "time"
)

func main() {
    builder := mesh.NewBuilder()

    // 设置 Tile 缓存（最多缓存 1000 个 tile，TTL 为 10 分钟）
    cache := mesh.NewMemoryTileCache(1000, 10*time.Minute)
    builder.SetTileCache(cache)

    // 第一次构建（会从网络获取 tiles）
    mesh1, err := builder.BuildForDisplay()

    // 查看缓存统计
    stats := builder.GetCacheStats()
    fmt.Printf("Cache hits: %d, misses: %d, evictions: %d\n",
        stats.Hits, stats.Misses, stats.Evictions)

    // 第二次构建相同区域（会从缓存读取，速度更快）
    mesh2, err := builder.BuildForDisplay()

    // 清除缓存
    builder.SetTileCache(nil)
}
```

### 示例：从 io.Reader 加载 GeoTIFF

```go
package main

import (
    "bytes"
    "github.com/flywave/go-static-mesh/mesh"
    "github.com/flywave/go-static-mesh/static"
    "github.com/flywave/go-tin"
    "io"
    "net/http"
)

func main() {
    // 从网络下载 GeoTIFF 数据
    resp, err := http.Get("https://example.com/elevation.tif")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    // 使用 io.Reader 创建 Provider
    provider, err := static.NewGeoTIFFRasterProviderFromReader(resp.Body)
    if err != nil {
        panic(err)
    }
    defer provider.Close()

    builder := mesh.NewBuilder()
    builder.SetTINGenerator(&tin.Algorithm{})
    builder.SetRasterProvider(provider)
    builder.SetBounds(provider.Bounds(), provider.Srs())

    mesh, err := builder.BuildForDisplay()
    // ...
}
```

### 示例：高程数据转 STL（3D 打印）

```go
package main

import (
    "github.com/flywave/go-static-mesh/static"
    "github.com/flywave/go-static-mesh/mesh"
    "github.com/flywave/go-tin" // 用户提供的 TIN 算法
)

func main() {
    builder := mesh.NewBuilder()
    
    // 设置 TIN 生成算法
    builder.SetTINGenerator(&tin.Algorithm{})
    
    // 添加高程数据（RasterProvider 和 TinMeshProvider 只能选一个）
    provider := static.NewGeoTIFFRasterProvider("elevation.tif")
    builder.SetRasterProvider(provider)
    
    // 设置边界
    builder.SetBounds(provider.Bounds(), provider.Srs())
    
    // 设置闭合参数（3D 打印必须）
    builder.SetCloseMesh(true, 5.0)  // 闭合模型，基础厚度 5mm
    
    // 构建 3D 打印模型（自动闭合为体积）
    mesh, err := builder.BuildForPrint()
    if err != nil {
        panic(err)
    }
    
    // 导出 STL
    writer := mesh.NewStlWriter()
    writer.WriteFile(mesh, "terrain.stl")
}
```

### 示例：高程 + 卫星图转 GLB（展示）

```go
package main

import (
    "github.com/flywave/go-static-mesh/static"
    "github.com/flywave/go-static-mesh/draw"
    "github.com/flywave/go-static-mesh/mesh"
    "github.com/flywave/go-tin"
)

func main() {
    builder := mesh.NewBuilder()
    
    // 设置 TIN 生成算法
    builder.SetTINGenerator(&tin.Algorithm{})
    
    // 添加高程数据（RasterProvider 和 TinMeshProvider 只能选一个）
    elevationProvider := static.NewGeoTIFFRasterProvider("elevation.tif")
    builder.SetRasterProvider(elevationProvider)
    
    // 添加卫星影像（使用 NewTileProvider）
    imageryProvider := static.NewTileProvider(
        "https://tile.openstreetmap.org/{z}/{x}/{y}.png",
        static.TileProviderModeXYZ,
    )
    builder.AddImageryProvider(imageryProvider)
    
    // 设置边界
    builder.SetBounds(elevationProvider.Bounds(), elevationProvider.Srs())
    
    // 构建带纹理的展示模型（面片，不闭合）
    mesh, err := builder.BuildForDisplayWithTexture()
    if err != nil {
        panic(err)
    }
    
    // 导出 GLB
    writer := mesh.NewGltfWriter()
    writer.WriteFile(mesh, "terrain.glb")
}
```

## 核心流程

### 3D 打印流程

```
Raster 高程数据 -> TIN 算法 -> TIN Mesh（面片）
                    -> 闭合处理 -> 闭合 Mesh（体积）-> STL Writer -> STL 文件
```

### 展示流程（GLB/OBJ）

```
Raster 高程数据 -> TIN 算法 -> TIN Mesh（面片）
                    + 卫星影像 -> 纹理
                    + 地理数据 -> 绘制到纹理
                    -> Mesh Builder -> GLB/OBJ Writer -> GLB/OBJ 文件
```

## 项目结构

```
go-static-mesh/
├── static/              # 核心接口和数据源
├── mesh/               # Mesh 构建和输出
├── draw/               # 地理数据绘制
├── raster/             # 栅格数据处理
├── utils/              # 工具函数
├── ARCHITECTURE.md     # 整体架构设计
├── INTERFACES.md       # 接口设计详解
├── TIN_MESH.md         # TIN Mesh 核心流程
├── QUICKSTART.md       # 快速开始指南
└── AGENTS.md           # Agent 指南
```

## 依赖

- Go 1.24+
- [go-geo](https://github.com/flywave/go-geo) - 地理坐标转换
- [go-tin](https://github.com/flywave/go-tin) - TIN 三角网算法
- [go-quantized-mesh](https://github.com/flywave/go-quantized-mesh) - Cesium quantized-mesh 解码
- [go-mapbox](https://github.com/flywave/go-mapbox) - Raster 图片-高程解码
- [go-cog](https://github.com/flywave/go-cog) - GeoTIFF 读取
- [gltf](https://github.com/flywave/gltf) - GLTF 格式支持
- [go-stl](https://github.com/flywave/go-stl) - STL 格式支持
- [go-obj](https://github.com/flywave/go-obj) - OBJ 格式支持

## 开发状态

✅ 核心功能已完成 (85%)

- [x] 架构设计
- [x] 接口实现
- [x] TIN 集成
- [x] 输出格式实现 (STL/GLTF/OBJ)
- [x] Tile 缓存机制
- [ ] 文档完善
- [ ] 大数据处理 (P5: 进行中)
- [ ] 更多数据源支持 (P7)

## 贡献

欢迎贡献！请阅读 [AGENTS.md](./AGENTS.md) 了解代码风格和开发指南。

## 许可证

MIT License
