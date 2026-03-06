# DEM 数据格式处理修复总结

## 问题分析

### 原始问题
地形瓦片合并时出现排列错误，导致地形不连续。

### 根本原因
程序使用了两种不同格式的 DEM 数据，它们的地理坐标排列方式不同：
1. **GeoTIFF 文件**：包含完整的地理坐标信息，文件内部像素排列与地理坐标系统可能不同
2. **Mapbox WebP 文件**：纯图像格式，像素排列遵循 XYZ 瓦片坐标系统

## 解决方案

### 1. WebP DEM 数据支持

添加了对 Mapbox WebP DEM 格式的支持：
- 使用 `MapboxRasterProvider` 解码 WebP DEM 数据
- WebP 文件的高程数据编码在 RGB 像素值中
- 每个瓦片尺寸：512×512（包含边界数据）

### 2. 正确的坐标计算

实现了 XYZ 瓦片坐标到地理坐标的转换：

```go
func calculateTileBounds(coord [3]int) vec2d.Rect {
    x, y, z := coord[0], coord[1], coord[2]
    n := 1 << uint(z)
    
    // 经度计算
    lon1 := 360.0*float64(x)/float64(n) - 180.0
    lon2 := 360.0*float64(x+1)/float64(n) - 180.0
    
    // 纬度计算（墨卡托投影反算）
    lat1 := mercatorToLat(math.Pi * (1 - 2*float64(y)/float64(n)))
    lat2 := mercatorToLat(math.Pi * (1 - 2*float64(y+1)/float64(n)))
    
    return vec2d.Rect{
        Min: vec2d.T{minLon, minLat},
        Max: vec2d.T{maxLon, maxLat},
    }
}

func mercatorToLat(rad float64) float64 {
    lat := math.Atan(math.Sinh(rad))
    return lat * 180.0 / math.Pi
}
```

### 3. Y 坐标反转

Mapbox XYZ 瓦片系统中，Y 坐标从上往下递增，与地理纬度方向相反：

```go
// 正确的瓦片加载顺序
for y := maxY; y >= minY; y-- {
    for x := minX; x <= maxX; x++ {
        coords[idx] = [3]int{x, y, zoom}
        idx++
    }
}
```

### 4. 自动格式检测

修改了 `loadAndMergeDEM` 函数，自动检测数据格式：
- 优先使用 WebP 文件（`.webp`）
- 如果没有 WebP，则使用 GeoTIFF 文件（`.tif`）

## 测试结果

### WebP DEM 数据
- 瓦片数量：16（4×4 网格）
- 瓦片尺寸：512×512 像素
- 合并后尺寸：1024×1024
- 高程范围：0.00 - 587.70 米
- 坐标范围：[118.059082, 36.474307] - [118.146973, 36.544949]

### 生成的 3D 地形
- 顶点数：20,478
- 三角形数：40,580
- 文件大小：2.06 MB
- 纹理：已正确应用

## 文件更新

### 修改的文件
1. `examples/terrain_with_texture/main.go`
   - 添加 WebP DEM 数据加载支持
   - 实现 XYZ 瓦片坐标转换
   - 自动检测数据格式

2. `examples/terrain_with_merge/main.go`
   - 修正 Y 坐标反转逻辑

### 新增文件
1. `download_mapbox_dem.go` - Mapbox DEM 数据下载工具
2. `test_webp_dem.go` - WebP DEM 合并测试（已删除）

### 文档文件
1. `DEM_FIX_EXPLANATION.md` - DEM 瓦片合并修复说明
2. `DOWNLOAD_MAPBOX_DEM.md` - Mapbox DEM 下载工具说明

## 使用方法

### 下载最新 DEM 数据
```bash
go run download_mapbox_dem.go
```

### 生成地形网格
```bash
cd examples/terrain_with_texture
go run main.go
```

### 验证输出
- `output/merged_dem.png` - DEM 高程可视化
- `output/merged_texture.png` - 卫星纹理
- `output/terrain_textured.gltf` - 3D 地形网格

## 技术要点

### Mapbox XYZ 瓦片坐标系统
- 原点：左上角（0,0）
- X 轴：向右递增（东）
- Y 轴：向下递增（南）
- Zoom：缩放级别

### 地理坐标系统
- 原点：赤道和本初子午线交点
- 纬度：向北递增（+90° 到 -90°）
- 经度：向东递增（-180° 到 +180°）

### 坐标转换
XYZ 瓦片坐标 → 墨卡托投影 → 地理坐标（WGS84）

## 结论

✅ WebP DEM 数据现在可以正确加载和合并
✅ 地形瓦片排列正确，形成连续地形
✅ 纹理与 DEM 对齐
✅ 支持自动格式检测（WebP/TIFF）
