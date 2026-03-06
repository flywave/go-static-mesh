# Mapbox DEM 数据下载工具

## 概述

这个工具用于从 Mapbox 下载地形 DEM (Digital Elevation Model) 瓦片数据。

## 使用方法

### 基本用法（跳过已存在的文件）

```bash
go run download_mapbox_dem.go
```

### 强制重新下载（覆盖已存在的文件）

```bash
go run download_mapbox_dem.go -force
```

## 配置

### 当前配置

- **Token**: `pk.eyJ1IjoiYW5pbmdnbyIsImEiOiJjbW1ld2VpZHgwMXE3MnFxem1hZ2o4cTE3In0.dHx9oXT7vdHnWMlpkscYTg`
- **SKU**: `101XxiLvoFYxL`
- **瓦片范围**:
  - Zoom: 14
  - X: 13565 - 13568 (4 列)
  - Y: 6403 - 6406 (4 行)
  - 总计: 16 个瓦片

### 修改配置

如需下载其他区域的瓦片，请在 `download_mapbox_dem.go` 中修改以下参数：

```go
const (
    newToken = "你的 Mapbox Token"
    sku      = "你的 SKU"
)

// 在 main() 函数中修改
zoom := 14
minX, maxX := 13565, 13568
minY, maxY := 6403, 6406
```

## 输出

- **保存位置**: `data/dem/`
- **文件命名**: `{zoom}_{x}_{y}.webp`
- **示例**: `14_13565_6403.webp`

## 下载速度

- 每个瓦片下载间隔: 200ms（避免触发 Mapbox API 限制）
- 16 个瓦片预计耗时: ~4 秒

## 验证下载

下载完成后，可以运行地形生成程序验证：

```bash
cd examples/terrain_with_texture
go run main.go
```

检查输出的高程范围是否符合预期：
```
DEM merged: size=1024x1024
✓ Merged DEM saved to: output/merged_dem.png (elevation range: 215.60 - 608.70)
```

## 文件大小

每个瓦片大小约 90-120 KB，总计约 1.6 MB。

## 注意事项

1. **Token 有效期**: Mapbox Token 可能会过期，需要定期更新
2. **API 限制**: 遵守 Mapbox API 使用限制，避免频繁下载
3. **数据格式**: Mapbox DEM 使用 WebP 格式，需要相应的解码器支持

## 相关文件

- `download_mapbox_dem.go` - 下载程序
- `data/dem/` - DEM 数据存储目录
- `examples/terrain_with_texture/` - 地形生成示例
- `DEM_FIX_EXPLANATION.md` - DEM 瓦片合并修复说明
