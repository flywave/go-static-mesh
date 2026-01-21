# 依赖库使用指南

本文档详细说明如何使用项目依赖的核心库。

## 1. go-tin - TIN 三角网算法

`github.com/flywave/go-tin` 提供了从高程数据生成 TIN 网格的算法。

### 1.1 基本使用

```go
import (
    static "github.com/flywave/go-static-mesh"
    "github.com/flywave/go-static-mesh/mesh"
    "github.com/flywave/go-tin"
)

func main() {
    builder := mesh.NewBuilder()

    // 设置 TIN 生成算法
    tinGenerator := &tin.Algorithm{}
    builder.SetTINGenerator(tinGenerator)

    // ...
}
```

### 1.2 TIN 算法配置

```go
import "github.com/flywave/go-tin"

// 使用默认配置
tinGenerator := &tin.Algorithm{}

// 使用自定义配置
tinGenerator := &tin.Algorithm{
    MaxTriangleArea:    1000.0,  // 最大三角形面积
    MinTriangleAngle:   15.0,    // 最小三角形角度（度）
    MaxVertexCount:     100000,  // 最大顶点数量
    Simplify:           true,     // 是否简化
    SimplifyTolerance:  0.1,     // 简化容差
}

builder.SetTINGenerator(tinGenerator)
```

### 1.3 从 Raster 生成 TIN

```go
import (
    "github.com/flywave/go-tin"
    vec2d "github.com/flywave/go3d/float64/vec2"
)

func generateTINFromRaster(grid *static.ElevationGrid) (*mesh.TinMesh, error) {
    // 准备点数据
    var points []vec2d.T
    var elevations []float64

    for y := 0; y < grid.Height; y++ {
        for x := 0; x < grid.Width; x++ {
            elevation := grid.Data[y*grid.Width+x]
            if elevation == grid.NoData {
                continue
            }

            lon := grid.MinX + float64(x)*grid.CellSize
            lat := grid.MinY + float64(y)*grid.CellSize

            points = append(points, vec2d.T{lat, lon})
            elevations = append(elevations, elevation)
        }
    }

    // 使用 go-tin 生成 TIN
    tin := tin.NewTIN()
    err := tin.Triangulate(points, elevations)
    if err != nil {
        return nil, err
    }

    // 转换为 TinMesh
    tinMesh := &mesh.TinMesh{
        Vertices:  tin.Vertices(),
        Indices:   tin.Indices(),
        MinHeight: grid.MinHeight(),
        MaxHeight: grid.MaxHeight(),
    }

    return tinMesh, nil
}
```

## 2. go-quantized-mesh - Cesium quantized-mesh 解码

`github.com/flywave/go-quantized-mesh` 提供了 Cesium quantized-mesh 格式的解码。

### 2.1 创建 CesiumQuantizedMeshProvider

```go
import (
    static "github.com/flywave/go-static-mesh"
    "github.com/flywave/go-quantized-mesh"
)

func main() {
    // 创建 Cesium quantized-mesh 提供者
    provider := static.NewCesiumQuantizedMeshProvider(
        "https://example.com/terrain/{z}/{x}/{y}.terrain",
    )

    // 使用自定义配置
    config := &quantizedmesh.DecoderConfig{
        ExtensionHeader: true,  // 是否解析扩展头
        WaterMask:       false, // 是否解析水体掩码
        VertexNormals:   true,  // 是否计算顶点法线
    }

    provider := static.NewCesiumQuantizedMeshProviderWithConfig(
        "https://example.com/terrain/{z}/{x}/{y}.terrain",
        config,
    )
}
```

### 2.2 使用 Cesium Quantized Mesh

```go
func main() {
    builder := mesh.NewBuilder()

    // 设置 TIN Mesh 提供者
    provider := static.NewCesiumQuantizedMeshProvider(
        "https://example.com/terrain/{z}/{x}/{y}.terrain",
    )
    builder.SetTinMeshProvider(provider)

    // 设置范围
    bounds := vec2d.Rect{{minLat, minLng}, {maxLat, maxLng}}
    srs := geo.NewProj(4326)
    builder.SetBounds(bounds, srs)

    // 构建 Mesh
    mesh, err := builder.BuildForDisplay()
    if err != nil {
        panic(err)
    }

    // 导出
    writer := mesh.NewGltfWriter()
    writer.WriteFile(mesh, "model.glb")
}
```

### 2.3 解码 Quantized Mesh

```go
import (
    "github.com/flywave/go-quantized-mesh"
    "github.com/flywave/go-static-mesh/mesh"
    vec3d "github.com/flywave/go3d/float64/vec3"
)

func decodeQuantizedMesh(data []byte) (*mesh.TinMesh, error) {
    decoder := quantizedmesh.NewDecoder()

    // 解码 mesh
    qm, err := decoder.Decode(data)
    if err != nil {
        return nil, err
    }

    // 转换为 TinMesh
    tinMesh := &mesh.TinMesh{
        Vertices:    make([]vec3d.T, len(qm.Vertices)),
        Indices:     qm.Indices,
        EdgeIndices: qm.EdgeIndices,
        NorthEdge:   qm.NorthEdge,
        SouthEdge:   qm.SouthEdge,
        WestEdge:    qm.WestEdge,
        EastEdge:    qm.EastEdge,
        MinHeight:   qm.MinHeight,
        MaxHeight:   qm.MaxHeight,
    }

    // 复制顶点
    for i, v := range qm.Vertices {
        tinMesh.Vertices[i] = vec3d.T{
            qm.Center[0] + v[0]*qm.Scale,
            qm.Center[1] + v[1]*qm.Scale,
            qm.Center[2] + v[2]*qm.Scale,
        }
    }

    return tinMesh, nil
}
```

## 3. go-mapbox - Raster 图片-高程解码

`github.com/flywave/go-mapbox` 提供了 Mapbox Terrain RGB 格式的解码。

### 3.1 创建 MapboxTerrainRGBProvider

```go
import (
    static "github.com/flywave/go-static-mesh"
    "github.com/flywave/go-mapbox"
)

func main() {
    // 创建 Mapbox Terrain RGB 提供者
    provider := static.NewMapboxTerrainRGBProvider(
        "mapbox.terrain-rgb",
        "your-mapbox-access-token",
    )

    // 或使用自定义 URL
    provider := static.NewMapboxTerrainRGBProviderWithURL(
        "https://api.mapbox.com/v4/mapbox.terrain-rgb/{z}/{x}/{y}.pngraw?access_token={token}",
    )
}
```

### 3.2 解码 Terrain RGB

```go
import (
    "image"
    "github.com/flywave/go-mapbox"
)

func decodeTerrainRGB(img image.Image) ([]float64, int, int, error) {
    // 使用 go-mapbox 解码
    elevations, width, height, err := mapbox.DecodeTerrainRGB(img)
    if err != nil {
        return nil, 0, 0, err
    }

    return elevations, width, height, nil
}
```

### 3.3 使用 Mapbox Terrain RGB

```go
func main() {
    builder := mesh.NewBuilder()

    // 设置 Mapbox Terrain RGB 提供者
    provider := static.NewMapboxTerrainRGBProvider(
        "mapbox.terrain-rgb",
        "your-mapbox-access-token",
    )
    builder.SetRasterProvider(provider)

    // 设置 TIN 生成算法
    tinGenerator := &tin.Algorithm{}
    builder.SetTINGenerator(tinGenerator)

    // 设置范围
    bounds := vec2d.Rect{{minLat, minLng}, {maxLat, maxLng}}
    srs := geo.NewProj(4326)
    builder.SetBounds(bounds, srs)

    // 构建 Mesh
    mesh, err := builder.BuildForDisplay()
    if err != nil {
        panic(err)
    }

    // 导出
    writer := mesh.NewGltfWriter()
    writer.WriteFile(mesh, "model.glb")
}
```

## 4. go-cog - GeoTIFF 读取

`github.com/flywave/go-cog` 提供了 GeoTIFF 文件的读取支持。

### 4.1 创建 GeoTIFFRasterProvider

```go
import (
    static "github.com/flywave/go-static-mesh"
    "github.com/flywave/go-cog"
)

func main() {
    // 读取 GeoTIFF 文件
    provider, err := static.NewGeoTIFFRasterProvider("elevation.tif")
    if err != nil {
        panic(err)
    }

    // 读取 COG（Cloud Optimized GeoTIFF）
    provider, err := static.NewCOGRasterProvider("elevation.cog.tif")
    if err != nil {
        panic(err)
    }

    // 使用自定义配置
    config := &cog.ReaderConfig{
        MaxConnections: 4,           // 最大连接数
        Timeout:        30 * time.Second, // 超时时间
        Cache:          true,         // 是否缓存
    }

    provider, err := static.NewCOGRasterProviderWithConfig(
        "elevation.cog.tif",
        config,
    )
}
```

### 4.2 读取 GeoTIFF

```go
import (
    "github.com/flywave/go-cog"
    "github.com/flywave/go-static-mesh/static"
)

func readGeoTIFF(filename string) (*static.ElevationGrid, error) {
    // 使用 go-cog 读取
    reader, err := cog.NewReader(filename)
    if err != nil {
        return nil, err
    }
    defer reader.Close()

    // 获取高程数据
    data, err := reader.Read()
    if err != nil {
        return nil, err
    }

    // 创建 ElevationGrid
    grid := &static.ElevationGrid{
        Width:    reader.Width(),
        Height:   reader.Height(),
        Data:     data,
        MinX:     reader.MinX(),
        MinY:     reader.MinY(),
        CellSize: reader.CellSize(),
        NoData:   reader.NoData(),
    }

    return grid, nil
}
```

### 4.3 使用 GeoTIFF

```go
func main() {
    builder := mesh.NewBuilder()

    // 读取 GeoTIFF 文件
    provider, err := static.NewGeoTIFFRasterProvider("elevation.tif")
    if err != nil {
        panic(err)
    }
    builder.SetRasterProvider(provider)

    // 设置 TIN 生成算法
    tinGenerator := &tin.Algorithm{}
    builder.SetTINGenerator(tinGenerator)

    // 设置范围（从 provider 获取参考）
    bounds := provider.Bounds()
    srs := provider.Srs()
    builder.SetBounds(bounds, srs)

    // 构建 Mesh
    mesh, err := builder.BuildForDisplay()
    if err != nil {
        panic(err)
    }

    // 导出
    writer := mesh.NewGltfWriter()
    writer.WriteFile(mesh, "model.glb")
}
```

## 5. TileProvider - 瓦片提供者

统一的瓦片提供者，支持 XYZ 和 TMS 两种模式。

### 5.1 创建 TileProvider

```go
import (
    static "github.com/flywave/go-static-mesh"
)

// XYZ 模式（默认）
provider := static.NewTileProvider(
    "https://tile.openstreetmap.org/{z}/{x}/{y}.png",
    static.TileProviderModeXYZ,
)

// TMS 模式
provider := static.NewTileProvider(
    "https://tile.openstreetmap.org/{z}/{y}/{x}.png",
    static.TileProviderModeTMS,
)

// 自定义模板
provider := static.NewTileProvider(
    "https://example.com/tiles/{x}/{y}/{z}.png",
    static.TileProviderModeXYZ,
)

// 使用自定义配置
config := &static.TileProviderConfig{
    Mode:         static.TileProviderModeXYZ,
    MaxRetries:   3,
    Timeout:      30 * time.Second,
    UserAgent:    "MyApp/1.0",
    Headers:      map[string]string{"Authorization": "Bearer token"},
}

provider := static.NewTileProviderWithConfig(
    "https://example.com/tiles/{z}/{x}/{y}.png",
    config,
)
```

### 5.2 常见 TileProvider 示例

```go
// OpenStreetMap
osmProvider := static.NewTileProvider(
    "https://tile.openstreetmap.org/{z}/{x}/{y}.png",
    static.TileProviderModeXYZ,
)

// Mapbox Satellite
mapboxProvider := static.NewTileProvider(
    "https://api.mapbox.com/v4/mapbox.satellite/{z}/{x}/{y}.pngraw?access_token={token}",
    static.TileProviderModeXYZ,
)

// Google Satellite
googleProvider := static.NewTileProvider(
    "https://mt1.google.com/vt/lyrs=s&x={x}&y={y}&z={z}",
    static.TileProviderModeXYZ,
)

// Stamen Terrain
stamenProvider := static.NewTileProvider(
    "https://stamen-tiles.a.ssl.fastly.net/terrain/{z}/{x}/{y}.png",
    static.TileProviderModeXYZ,
)
```

### 5.3 使用 TileProvider

```go
func main() {
    builder := mesh.NewBuilder()

    // 添加卫星影像
    imageryProvider := static.NewTileProvider(
        "https://tile.openstreetmap.org/{z}/{x}/{y}.png",
        static.TileProviderModeXYZ,
    )
    builder.AddImageryProvider(imageryProvider)

    // 添加高程数据
    rasterProvider := static.NewGeoTIFFRasterProvider("elevation.tif")
    builder.SetRasterProvider(rasterProvider)

    // 设置 TIN 生成算法
    tinGenerator := &tin.Algorithm{}
    builder.SetTINGenerator(tinGenerator)

    // 设置范围
    bounds := rasterProvider.Bounds()
    builder.SetBounds(bounds, rasterProvider.Srs())

    // 构建带纹理的 Mesh
    mesh, err := builder.BuildForDisplayWithTexture()
    if err != nil {
        panic(err)
    }

    // 导出
    writer := mesh.NewGltfWriter()
    writer.WriteFile(mesh, "model.glb")
}
```

## 6. 完整示例

### 6.1 使用 GeoTIFF + TIN 算法 + GLB 输出

```go
package main

import (
    "github.com/flywave/go-static-mesh/static"
    "github.com/flywave/go-static-mesh/mesh"
    "github.com/flywave/go-tin"
)

func main() {
    builder := mesh.NewBuilder()

    // 设置 TIN 生成算法（使用 go-tin）
    tinGenerator := &tin.Algorithm{}
    builder.SetTINGenerator(tinGenerator)

    // 读取 GeoTIFF 文件（使用 go-cog）
    rasterProvider, err := static.NewGeoTIFFRasterProvider("elevation.tif")
    if err != nil {
        panic(err)
    }
    builder.SetRasterProvider(rasterProvider)

    // 设置范围
    bounds := rasterProvider.Bounds()
    builder.SetBounds(bounds, rasterProvider.Srs())

    // 构建 Mesh
    mesh, err := builder.BuildForDisplay()
    if err != nil {
        panic(err)
    }

    // 导出 GLB
    writer := mesh.NewGltfWriter()
    writer.WriteFile(mesh, "model.glb")
}
```

### 6.2 使用 Cesium Quantized Mesh

```go
package main

import (
    "github.com/flywave/go-static-mesh/static"
    "github.com/flywave/go-static-mesh/mesh"
    "github.com/flywave/go-quantized-mesh"
)

func main() {
    builder := mesh.NewBuilder()

    // 创建 Cesium quantized-mesh 提供者（使用 go-quantized-mesh）
    config := &quantizedmesh.DecoderConfig{
        ExtensionHeader: true,
        WaterMask:       false,
        VertexNormals:   true,
    }
    tinMeshProvider := static.NewCesiumQuantizedMeshProviderWithConfig(
        "https://example.com/terrain/{z}/{x}/{y}.terrain",
        config,
    )
    builder.SetTinMeshProvider(tinMeshProvider)

    // 设置范围
    bounds := vec2d.Rect{{minLat, minLng}, {maxLat, maxLng}}
    srs := geo.NewProj(4326)
    builder.SetBounds(bounds, srs)

    // 构建 Mesh
    mesh, err := builder.BuildForDisplay()
    if err != nil {
        panic(err)
    }

    // 导出 GLB
    writer := mesh.NewGltfWriter()
    writer.WriteFile(mesh, "model.glb")
}
```

### 6.3 使用 Mapbox Terrain RGB + 卫星影像

```go
package main

import (
    "github.com/flywave/go-static-mesh/static"
    "github.com/flywave/go-static-mesh/mesh"
    "github.com/flywave/go-tin"
)

func main() {
    builder := mesh.NewBuilder()

    // 设置 TIN 生成算法
    tinGenerator := &tin.Algorithm{}
    builder.SetTINGenerator(tinGenerator)

    // 添加高程数据（使用 go-mapbox）
    rasterProvider := static.NewMapboxTerrainRGBProvider(
        "mapbox.terrain-rgb",
        "your-mapbox-access-token",
    )
    builder.SetRasterProvider(rasterProvider)

    // 添加卫星影像（使用 TileProvider）
    imageryProvider := static.NewTileProvider(
        "https://tile.openstreetmap.org/{z}/{x}/{y}.png",
        static.TileProviderModeXYZ,
    )
    builder.AddImageryProvider(imageryProvider)

    // 设置范围
    bounds := vec2d.Rect{{minLat, minLng}, {maxLat, maxLng}}
    srs := geo.NewProj(4326)
    builder.SetBounds(bounds, srs)

    // 构建带纹理的 Mesh
    mesh, err := builder.BuildForDisplayWithTexture()
    if err != nil {
        panic(err)
    }

    // 导出 GLB
    writer := mesh.NewGltfWriter()
    writer.WriteFile(mesh, "model.glb")
}
```

## 7. 最佳实践

### 7.1 TIN 算法配置

- ✅ 根据数据特点调整 `MaxTriangleArea`
- ✅ 设置合理的 `MaxVertexCount` 避免内存溢出
- ✅ 对于 3D 打印，启用简化
- ❌ 不要设置过大的 `MaxVertexCount`

### 7.2 Quantized Mesh 解码

- ✅ 启用 `ExtensionHeader` 获取更多元数据
- ✅ 启用 `VertexNormals` 提高质量
- ✅ 根据需要启用 `WaterMask`
- ❌ 不要在不需要时解析所有扩展数据

### 7.3 GeoTIFF 读取

- ✅ 使用 COG 格式提高性能
- ✅ 设置合理的缓存策略
- ✅ 处理 NoData 值
- ❌ 不要一次性读取过大的文件

### 7.4 TileProvider 配置

- ✅ 设置合理的重试次数 (3-5)
- ✅ 设置合理的超时时间 (30-60s)
- ✅ 添加 User-Agent 标识
- ❌ 不要设置过短的超时时间
