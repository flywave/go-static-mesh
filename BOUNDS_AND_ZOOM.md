# 范围和 Zoom 处理指南

## 概述

本文档详细说明如何处理地理范围设置、Zoom 级别选择，以及 Tile 获取失败的处理策略。

## 相关依赖库

- `github.com/flywave/go-cog` - GeoTIFF 读取（Cloud Optimized GeoTIFF）
- `github.com/flywave/go-mapbox` - Raster 图片-高程解码（Mapbox Terrain RGB）
- `github.com/flywave/go-tin` - TIN 三角网算法
- `github.com/flywave/go-quantized-mesh` - Cesium quantized-mesh 解码

## 1. 范围设置（必须）

### 1.1 外部传入范围

**重要**: 所有模型生成都需要外部传入地理范围（bounds + srs），这是必需参数。

```go
import (
    vec2d "github.com/flywave/go3d/float64/vec2"
    "github.com/flywave/go-geo"
)

// 定义地理范围
bounds := vec2d.Rect{
    Min: vec2d.T{30.0, 120.0}, // {minLat, minLng}
    Max: vec2d.T{31.0, 121.0}, // {maxLat, maxLng}
}

// 定义坐标系
srs := geo.NewProj(4326) // WGS84

// 设置到 Builder
builder.SetBounds(bounds, srs)
```

### 1.2 范围坐标系统

支持多种坐标系统：

```go
// WGS84 (经纬度）
srs := geo.NewProj(4326)
bounds := vec2d.Rect{{30.0, 120.0}, {31.0, 121.0}}

// Web Mercator (墨卡托投影）
srs := geo.NewProj(3857)
bounds := vec2d.Rect{{13358338, 3503549}, {13520000, 3700000}}

// UTM (通用横轴墨卡托投影）
srs := geo.NewProj(32650) // UTM Zone 50N
bounds := vec2d.Rect{{400000, 4400000}, {500000, 4500000}}
```

### 1.3 从数据源获取范围

虽然需要外部传入范围，但可以从数据源获取参考：

```go
// 从高程数据源获取
elevationProvider := static.NewGeoTIFFRasterProvider("elevation.tif")
referenceBounds := elevationProvider.Bounds()

// 调整范围（缩小或扩大）
adjustedBounds := vec2d.Rect{
    Min: vec2d.T{referenceBounds.Min[0] - 0.01, referenceBounds.Min[1] - 0.01},
    Max: vec2d.T{referenceBounds.Max[0] + 0.01, referenceBounds.Max[1] + 0.01},
}

builder.SetBounds(adjustedBounds, elevationProvider.Srs())
```

## 2. Zoom 级别设置

### 2.1 手动设置 Zoom

如果需要精确控制数据量，可以手动设置 Zoom 级别：

```go
// 设置固定 zoom
builder.SetZoom(15)

// Zoom 级别说明：
// - 较低的 zoom (10-12): 覆盖大范围，数据量小，分辨率低
// - 中等 zoom (13-15): 平衡范围和分辨率
// - 较高的 zoom (16-18): 覆盖小范围，数据量大，分辨率高
```

### 2.2 自动判断 Zoom（推荐）

自动判断 Zoom 会根据以下因素计算最优级别：

- **范围大小**: 范围越大，zoom 越小
- **分辨率要求**: 分辨率越高，zoom 越大
- **Provider 的 grid 信息**: 考虑 tile 大小
- **内存限制**: 避免过大的数据量

```go
// 设置自动判断范围
builder.SetAutoZoomRange(10, 18)

// 或使用默认范围（13-17）
// builder.SetResolution(1.0) // 分辨率也会影响 zoom 判断
```

### 2.3 自动判断算法

```go
func (b *Builder) determineZoom() (int, error) {
    // 如果手动设置了 zoom，直接使用
    if b.zoom != nil {
        return *b.zoom, nil
    }
    
    // 获取 grid 信息
    grid := b.rasterProvider.Grid()
    tileSize := float64(grid.TileSize[0])
    
    // 计算范围的大小
    width := b.bounds.Width()
    height := b.bounds.Height()
    
    // 根据范围大小估算初始 zoom
    initialZoom := int(math.Log2(360.0 / width))
    
    // 根据分辨率调整 zoom
    // 较高的分辨率需要较高的 zoom
    if b.resolution > 0 {
        resolutionFactor := math.Log2(1.0 / b.resolution)
        initialZoom += int(resolutionFactor)
    }
    
    // 确保 zoom 在合理范围内
    bestZoom := initialZoom
    if bestZoom < b.autoZoomMin {
        bestZoom = b.autoZoomMin
    }
    if bestZoom > b.autoZoomMax {
        bestZoom = b.autoZoomMax
    }
    
    return bestZoom, nil
}
```

### 2.4 Zoom 级别选择建议

| 用途 | Zoom 范围 | 说明 |
|------|----------|------|
| 3D 打印 - 大范围 | 10-12 | 减少数据量，提高打印速度 |
| 3D 打印 - 中等范围 | 13-14 | 平衡质量和数据量 |
| 3D 打印 - 小范围 | 15-16 | 高精度打印 |
| 展示 - 大范围 | 10-12 | 快速浏览 |
| 展示 - 中等范围 | 13-14 | 平衡质量和性能 |
| 展示 - 小范围 | 15-18 | 高精度展示 |

## 3. Tile 获取失败处理

### 3.1 处理策略

当范围内部分 tile 获取不到时，可以采用以下策略：

#### 策略 1: 跳过获取失败的 tile（推荐用于展示）

```go
handler := mesh.TileErrorHandler{
    SkipMissing: true,    // 跳过获取失败的 tile
    MaxRetries:  3,       // 最大重试次数
    Timeout:     30 * time.Second, // 超时时间
}
builder.SetTileErrorHandler(handler)
```

**适用场景**:
- Web 展示
- 部分数据缺失不影响整体效果
- 网络不稳定

#### 策略 2: 使用 NoData 值填充（推荐用于 3D 打印）

```go
handler := mesh.TileErrorHandler{
    UseNoData:   true,    // 使用 NoData 值填充
    NoDataValue: 0.0,     // NoData 值
    MaxRetries:  3,       // 最大重试次数
    Timeout:     30 * time.Second,
}
builder.SetTileErrorHandler(handler)
```

**适用场景**:
- 3D 打印（需要完整的体积）
- 可以用默认值填充缺失区域
- 保证模型连续性

#### 策略 3: 严格模式（任何失败都报错）

```go
handler := mesh.TileErrorHandler{
    SkipMissing: false,   // 不跳过
    UseNoData:   false,   // 不使用 NoData
    MaxRetries:  5,       // 更多重试次数
    Timeout:     60 * time.Second, // 更长超时
}
builder.SetTileErrorHandler(handler)
```

**适用场景**:
- 需要完整数据的场景
- 数据完整性至关重要
- 可以容忍较长的加载时间

### 3.2 Tile 获取流程

```go
func (b *Builder) fetchTiles(zoom int) ([]*Tile, error) {
    grid := b.rasterProvider.Grid()
    
    // 1. 计算范围内的所有 tile 坐标
    tiles := b.calculateTileCoords(zoom, b.bounds)
    
    var wg sync.WaitGroup
    fetchedTiles := make(chan *Tile)
    errors := make(chan error)
    
    // 2. 并发获取 tile
    for _, tileCoord := range tiles {
        wg.Add(1)
        go func(coord [3]int) {
            defer wg.Done()
            
            var tile *Tile
            var err error
            
            // 3. 重试机制
            for i := 0; i < b.tileErrorHandler.MaxRetries; i++ {
                tile, err = b.fetchSingleTile(coord, zoom)
                if err == nil {
                    break
                }
                time.Sleep(time.Second * time.Duration(i+1))
            }
            
            // 4. 处理失败
            if err != nil {
                if b.tileErrorHandler.SkipMissing {
                    // 跳过该 tile
                    return
                } else if b.tileErrorHandler.UseNoData {
                    // 创建 NoData tile
                    tile = b.createNoDataTile(coord, zoom)
                } else {
                    // 返回错误
                    errors <- err
                    return
                }
            }
            
            fetchedTiles <- tile
        }(tileCoord)
    }
    
    wg.Wait()
    close(fetchedTiles)
    
    // 5. 收集结果
    result := []*Tile{}
    for tile := range fetchedTiles {
        result = append(result, tile)
    }
    
    return result, nil
}
```

### 3.3 Tile 坐标计算

```go
func (b *Builder) calculateTileCoords(zoom int) [][3]int {
    grid := b.rasterProvider.Grid()
    
    // 将范围转换为网格坐标
    minX := int((b.bounds.Min[0] + 180.0) / 360.0 * float64(1<<uint(zoom)))
    maxX := int((b.bounds.Max[0] + 180.0) / 360.0 * float64(1<<uint(zoom)))
    
    minY := int((1.0 - math.Log(math.Tan(b.bounds.Min[1]*math.Pi/180.0) +
        1.0/math.Cos(b.bounds.Min[1]*math.Pi/180.0))/math.Pi) / 2.0 * float64(1<<uint(zoom)))
    maxY := int((1.0 - math.Log(math.Tan(b.bounds.Max[1]*math.Pi/180.0) +
        1.0/math.Cos(b.bounds.Max[1]*math.Pi/180.0))/math.Pi) / 2.0 * float64(1<<uint(zoom)))
    
    // 生成所有 tile 坐标
    var tiles [][3]int
    for x := minX; x <= maxX; x++ {
        for y := minY; y <= maxY; y++ {
            tiles = append(tiles, [3]int{x, y, zoom})
        }
    }
    
    return tiles
}
```

### 3.4 创建 NoData Tile

```go
func (b *Builder) createNoDataTile(coord [3]int, zoom int) *Tile {
    // 创建全 NoData 值的 tile
    tileSize := 256
    data := make([]float64, tileSize*tileSize)
    for i := range data {
        data[i] = b.tileErrorHandler.NoDataValue
    }
    
    return &Tile{
        Coord:  coord,
        Zoom:   zoom,
        Data:   data,
        Width:  tileSize,
        Height: tileSize,
    }
}
```

## 4. Tile 合并为网格

当获取到多个 tile 后，需要将它们合并为一个单一的网格：

```go
func (b *Builder) mergeTilesToGrid(tiles []*Tile) *ElevationGrid {
    if len(tiles) == 0 {
        return nil
    }
    
    // 1. 计算网格的总体范围
    minX := math.MaxFloat64
    minY := math.MaxFloat64
    maxX := math.MinFloat64
    maxY := math.MinFloat64
    
    for _, tile := range tiles {
        minX = math.Min(minX, tile.Bounds.Min[0])
        minY = math.Min(minY, tile.Bounds.Min[1])
        maxX = math.Max(maxX, tile.Bounds.Max[0])
        maxY = math.Max(maxY, tile.Bounds.Max[1])
    }
    
    // 2. 计算网格尺寸
    tileSize := float64(256)
    tilesX := int((maxX - minX) / tileSize)
    tilesY := int((maxY - minY) / tileSize)
    
    gridWidth := tilesX * 256
    gridHeight := tilesY * 256
    
    // 3. 创建合并后的网格
    grid := &ElevationGrid{
        Width:    gridWidth,
        Height:   gridHeight,
        MinX:     minX,
        MinY:     minY,
        CellSize: tileSize / 256.0,
        Data:     make([]float64, gridWidth*gridHeight),
    }
    
    // 4. 将每个 tile 的数据复制到网格中
    for _, tile := range tiles {
        // 计算该 tile 在网格中的起始位置
        startX := int((tile.Bounds.Min[0] - minX) / tileSize * 256)
        startY := int((tile.Bounds.Min[1] - minY) / tileSize * 256)
        
        // 复制数据
        for y := 0; y < 256; y++ {
            for x := 0; x < 256; x++ {
                if startX+x < gridWidth && startY+y < gridHeight {
                    gridIdx := (startY+y)*gridWidth + (startX+x)
                    tileIdx := y*256 + x
                    grid.Data[gridIdx] = tile.Data[tileIdx]
                }
            }
        }
    }
    
    return grid
}
```

## 5. 完整示例

### 5.1 3D 打印示例（使用 NoData 填充）

```go
package main

import (
    "time"
    "github.com/flywave/go-static-mesh/static"
    "github.com/flywave/go-static-mesh/mesh"
    "github.com/flywave/go-tin"
    vec2d "github.com/flywave/go3d/float64/vec2"
    "github.com/flywave/go-geo"
)

func main() {
    builder := mesh.NewBuilder()
    
    // 设置 TIN 生成算法
    builder.SetTINGenerator(&tin.Algorithm{})
    
    // 添加高程数据
    rasterProvider := static.NewGeoTIFFRasterProvider("elevation.tif")
    builder.SetRasterProvider(rasterProvider)
    
    // 设置范围（必须）
    bounds := vec2d.Rect{{30.0, 120.0}, {31.0, 121.0}}
    srs := geo.NewProj(4326)
    builder.SetBounds(bounds, srs)
    
    // 设置 zoom
    builder.SetAutoZoomRange(13, 15)
    
    // 设置 Tile 获取失败处理（使用 NoData 填充，保证 3D 打印完整性）
    handler := mesh.TileErrorHandler{
        UseNoData:   true,
        NoDataValue: 0.0,
        MaxRetries:  5,
        Timeout:     60 * time.Second,
    }
    builder.SetTileErrorHandler(handler)
    
    // 设置闭合参数
    builder.SetCloseMesh(true, 5.0)
    
    // 构建模型
    mesh, err := builder.BuildForPrint()
    if err != nil {
        panic(err)
    }
    
    // 导出 STL
    writer := mesh.NewStlWriter()
    writer.WriteFile(mesh, "model.stl")
}
```

### 5.2 展示示例（跳过失败的 tile）

```go
package main

import (
    "time"
    "github.com/flywave/go-static-mesh/static"
    "github.com/flywave/go-static-mesh/mesh"
    "github.com/flywave/go-tin"
    vec2d "github.com/flywave/go3d/float64/vec2"
    "github.com/flywave/go-geo"
)

func main() {
    builder := mesh.NewBuilder()
    
    // 设置 TIN 生成算法
    builder.SetTINGenerator(&tin.Algorithm{})
    
    // 添加高程数据
    rasterProvider := static.NewGeoTIFFRasterProvider("elevation.tif")
    builder.SetRasterProvider(rasterProvider)
    
    // 添加卫星影像（使用 NewTileProvider）
    imageryProvider := static.NewTileProvider(
        "https://tile.example.com/{z}/{x}/{y}.png",
        static.TileProviderModeXYZ,
    )
    builder.AddImageryProvider(imageryProvider)
    
    // 设置范围（必须）
    bounds := vec2d.Rect{{30.0, 120.0}, {31.0, 121.0}}
    srs := geo.NewProj(4326)
    builder.SetBounds(bounds, srs)
    
    // 设置 zoom
    builder.SetAutoZoomRange(15, 17)
    
    // 设置 Tile 获取失败处理（跳过失败的 tile）
    handler := mesh.TileErrorHandler{
        SkipMissing: true,
        MaxRetries:  3,
        Timeout:     30 * time.Second,
    }
    builder.SetTileErrorHandler(handler)
    
    // 构建模型
    mesh, err := builder.BuildForDisplayWithTexture()
    if err != nil {
        panic(err)
    }
    
    // 导出 GLB
    writer := mesh.NewGltfWriter()
    writer.WriteFile(mesh, "model.glb")
}
```

## 6. 最佳实践

### 6.1 范围设置

- ✅ 从数据源获取参考范围
- ✅ 根据实际需求调整范围
- ✅ 确保范围在数据源覆盖范围内
- ❌ 不要使用过大的范围（会导致数据量过大）

### 6.2 Zoom 设置

- ✅ 使用自动判断（推荐）
- ✅ 根据用途选择合适的 zoom 范围
- ✅ 3D 打印使用中等 zoom (13-15)
- ✅ 展示使用较高 zoom (15-17)
- ❌ 不要使用过高的 zoom (>=19) 除非必要

### 6.3 Tile 获取失败处理

- ✅ 3D 打印使用 NoData 填充
- ✅ 展示使用跳过模式
- ✅ 设置合理的重试次数 (3-5)
- ✅ 设置合理的超时时间 (30-60s)
- ❌ 不要在严格模式下处理网络不稳定的数据

### 6.4 性能优化

- ✅ 并发获取 tile
- ✅ 合理设置 zoom 范围
- ✅ 缓存已获取的 tile
- ✅ 使用流式处理大数据
- ❌ 不要一次性获取过多 tile
