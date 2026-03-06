# 使用本地 DEM 数据构建地形网格 - 实现说明

## 概述

本文档说明如何使用 `go-static-mesh` 项目中 data 目录的测试数据构建真实的地形网格。

## 数据分析

### 数据位置
- **DEM 数据**: `data/dem/` - GeoTIFF 格式的高程数据
- **卫星影像**: `data/satellite/` - WebP 格式的卫星影像

### 数据范围
基于墨卡托投影的瓦片数据：
- **Zoom Level**: 14
- **瓦片数量**: 16 个（4×4 网格）
- **瓦片坐标**: X: 13565-13568, Y: 6403-6406
- **地理位置**: 中国山东省济南市
- **坐标范围**: 
  - 经度: 118.059°E - 118.147°E
  - 纬度: 36.474°N - 36.545°N
- **覆盖面积**: 约 7.86 km × 7.86 km

### 文件格式
```
data/dem/
├── 14_13565_6403.tif  # 高程数据
├── 14_13565_6403.webp # 预览影像
├── ...
└── 14_13568_6406.tif

data/satellite/
├── satellite_14_13565_6403.webp  # Mapbox 卫星影像
├── ...
└── satellite_14_13568_6406.webp
```

## Provider 实现方案

参考 `github.com/flywave/go-tileproxy/terrain` 模块的实现方式，特别是 `dem.go` 中对 Web 格式 DEM 数据的读取。

### 关键组件

#### 1. GeoTIFFRasterProvider
位置: `tile/raster.go`

用于读取 GeoTIFF 格式的 DEM 数据：
```go
provider, err := tile.NewGeoTIFFRasterProvider("data/dem/14_13565_6403.tif")
if err != nil {
    log.Fatal(err)
}
defer provider.Close()

// 获取高程网格
grid := provider.GetElevationGrid()
```

#### 2. DemIO
位置: `tile/dem.go`

处理 DEM 数据的编码/解码，支持 Mapbox 和 Terrarium 两种编码格式：
```go
// 从 PNG/WebP 图像解码 DEM 数据
demIO := &tile.DemIO{
    Mode:   tile.ModeMapbox,
    Format: tile.TileFormat("webp"),
}

tileData, err := demIO.Decode(reader)
```

#### 3. LocalTileProvider
位置: `examples/terrain_from_local_data/main.go`

自定义的本地瓦片提供者，管理多个 GeoTIFF 文件：
```go
type LocalTileProvider struct {
    providers map[[3]int]*tile.GeoTIFFRasterProvider
    grid      *geo.TileGrid
    bounds    vec2d.Rect
    srs       geo.Proj
}
```

## 构建流程

### 1. 加载 DEM 数据
```go
provider, err := NewLocalTileProvider("data/dem")
// 加载所有 GeoTIFF 文件
// 解析文件名获取瓦片坐标
// 计算整体边界范围
```

### 2. 配置 Builder
```go
b := builder.NewBuilder()
b.SetBounds(provider.Bounds(), provider.Srs())
b.SetVerticalExaggeration(1.0)
b.SetBaseElevation(0.0)
b.SetCloseMesh(true, 100.0)
```

### 3. 生成 TIN 网格（待实现）
```go
// 需要集成 TIN 生成算法
// 参考: github.com/flywave/go-tin
tinGenerator := tin.NewTINGenerator()
b.SetTINGenerator(tinGenerator)

// 设置 DEM provider
b.SetRasterProvider(provider)

// 构建网格
mesh, err := b.Build()
```

### 4. 添加纹理（可选）
```go
// 加载卫星影像
imageryProvider := tile.NewGeoTIFFImageryProvider("data/satellite/...")
b.AddImageryProvider(imageryProvider)
```

### 5. 导出 GLTF
```go
w := writer.NewGltfWriter()
err := w.Write(mesh, "output/terrain.gltf")
```

## 数据读取方式对比

### go-tileproxy/terrain 模块
```go
// DemIO 解码 Web 格式 DEM
type DemIO struct {
    Mode   RasterDemMode  // Mapbox 或 Terrarium
    Format tile.TileFormat
}

func (d *DemIO) Decode(r io.Reader) (*TileData, error) {
    // 使用 go-mapbox/raster 库解码
    data, err := mraster.LoadDEMDataWithStream(r, int(d.Mode))
    // 转换为 TileData 结构
    // 处理边界数据
}
```

### go-static-mesh 实现
```go
// GeoTIFFRasterProvider 直接读取 GeoTIFF
provider, err := tile.NewGeoTIFFRasterProvider(filename)

// 使用 GDAL 库读取
ds, err := gdal.Open(filename, gdal.ReadOnly)

// 获取高程数据
grid := provider.GetElevationGrid()
```

## 完整示例

运行示例程序：
```bash
go run examples/terrain_from_local_data/main.go
```

输出：
```
=== 地形网格生成示例 ===

加载瓦片: 14/13565/6403 from 14_13565_6403.tif
加载瓦片: 14/13565/6404 from 14_13565_6404.tif
...
总共加载 16 个瓦片
边界范围: [118.059082, 36.474307] - [118.146973, 36.544949]

步骤 1: 创建 Builder
步骤 2: 设置边界
步骤 3: 设置垂直夸张
步骤 4: 设置基础高程
步骤 5: 配置网格闭合
步骤 6: 构建地形网格
步骤 7: 保存为 GLTF 格式

=== 示例完成 ===
```

## 下一步开发

### 1. 完善 TIN 生成
- [ ] 集成 `github.com/flywave/go-tin` 库
- [ ] 实现高程数据采样
- [ ] 优化三角形网格

### 2. 数据处理优化
- [ ] 实现瓦片缓存机制
- [ ] 支持并行瓦片加载
- [ ] 处理 NoData 值

### 3. 纹理支持
- [ ] 集成卫星影像
- [ ] 生成 UV 坐标
- [ ] 支持多分辨率纹理

### 4. 性能优化
- [ ] LOD（细节层次）支持
- [ ] 空间索引
- [ ] 增量更新

## 技术栈

- **go-geo**: 地理坐标系统
- **go-tin**: TIN 网格生成
- **go-mapbox/raster**: DEM 编码/解码
- **go-cog**: Cloud Optimized GeoTIFF
- **gdal**: GeoTIFF 读写
- **gltf**: 3D 模型格式

## 参考资料

- [Mapbox Terrain RGB](https://docs.mapbox.com/data/tilesets/guides/access-elevation-data/)
- [Terrarium Format](https://github.com/tilezen/joerd/blob/master/docs/formats.md#terrarium)
- [Quantized Mesh](https://github.com/CesiumGS/quantized-mesh)
- [go-tileproxy/terrain](/home/aninggo/work/go-tileproxy/terrain)
