# 地形感知拉伸 + 自动裁剪功能总结

## 核心问题解决

### 1. 地形感知拉伸（Terrain-Aware Extrusion）
**问题**：模型底部固定在 z=0，导致与地形间存在悬空间隙

**解决方案**：
- **Path** (`draw/path.go`): 底部顶点跟随地形高度
  - 修改 `ExtrudeToMeshWithTerrain()` 方法
  - 每个线段起点/终点采样地形高度
  - 底部顶点 Z = 地形高度，顶部顶点 Z = 地形高度 + 对象高度
  
- **Area** (`draw/area.go`): 每个顶点底部跟随地形
  - 修改 `ExtrudeToMeshWithTerrain()` 方法
  - 为每个顶点计算地形高度
  - 底部三角形直接使用地形高度，顶部 = 地形 + 高度

### 2. 自动裁剪（Auto-Clipping）
**问题**：矢量数据没有裁剪机制，可能超出处理范围

**解决方案**：
- **创建 Clipper 工具** (`draw/clipper.go`):
  - 使用 GEOS 库的 `ClipByRect()` 方法
  - 支持裁剪 Path 和 Area 对象
  - 返回裁剪后的对象（完全在 bounds 外返回 nil）
  
- **Builder 集成** (`mesh/builder/builder.go`, `mesh/builder/builder_extrusion.go`):
  - 移除 `clipGeoData` 标志和 `SetClipGeoData()` 方法
  - 在 `addGeoDataToMesh()` 中自动对所有 geo data 进行裁剪
  - 如果裁剪后为 nil，跳过该对象（记录日志）
  
- **自动 Bounds 解析** (`mesh/builder/builder.go`):
  - 新增 `resolveBounds()` 方法
  - 在 `build()` 中自动调用
  - 优先级：
    1. 使用 `SetBounds()` 传入的 bounds
    2. 如果未设置，尝试从 provider 的 `Bounds()` 方法获取
    3. 同时自动解析 SRS（如果 provider 支持）

### 3. 统一底面封闭（Unified Base Closure）
**问题**：只封闭地形网格，geo data 拉伸到 z=0，导致内部空洞

**解决方案**：
- **添加 CloseUnifiedMesh()** (`mesh/textured_closer.go`):
  - 对整个统一网格（地形 + geo data）进行底面封闭
  - 计算全局最小高度
  - 统一底面高度 = 全局最小 - 厚度
  
- **Builder 修改** (`mesh/builder/builder.go`):
  - 移动 mesh 封闭到 geo data 添加之后
  - 对 resultMesh（统一网格）调用 `CloseUnifiedMesh()`
  - 确保所有模型共享同一底面

## 文件变更

| 文件 | 变更 |
|------|------|
| `draw/clipper.go` | 新建 - GEOS 裁剪工具 |
| `draw/clipper_test.go` | 新建 - 裁剪单元测试 |
| `draw/path.go` | 修改 - 地形感知底部拉伸 |
| `draw/area.go` | 修改 - 地形感知底部拉伸 |
| `mesh/textured_closer.go` | 添加 - CloseUnifiedMesh() 方法 |
| `mesh/builder/builder.go` | 添加 resolveBounds()，修改 build() 流程，移除 clipGeoData |
| `mesh/builder/builder_extrusion.go` | 修改 - 自动裁剪 geo data |
| `examples/auto_clip_example.go` | 新建 - 使用示例 |

## 使用方法

### 基础用法（自动裁剪）
```go
builder := builder.NewBuilder()
builder.SetTINGenerator(generator)
builder.SetRasterProvider(rasterProvider)
builder.SetZoom(15)
builder.SetExtrudeGeoData(true, 10.0)
builder.SetCloseMesh(true, 5.0)

path := draw.NewPathWithHeight(positions, srs, color, weight, height)
builder.AddGeoData(path)

mesh, err := builder.BuildForPrint()
```

### 使用 TIFF 文件（自动获取 bounds）
```go
imageryProvider, err := tile.NewGeoTIFFImageryProvider("world.tif")
builder.AddImageryProvider(imageryProvider)
builder.SetZoom(15)
builder.SetExtrudeGeoData(true, 10.0)
builder.SetCloseMesh(true, 5.0)

mesh, err := builder.BuildForPrint()
```

### 限制范围（手动设置 bounds）
```go
builder := builder.NewBuilder()
builder.SetTINGenerator(generator)
builder.SetRasterProvider(rasterProvider)
builder.SetBounds(bounds, srs)
builder.SetZoom(15)

mesh, err := builder.BuildForPrint()
```

## 优势

✅ **地形贴合**：模型底部跟随地形表面，无悬空  
✅ **自动裁剪**：无需手动设置，支持无限大数据源  
✅ **统一底面**：所有模型 watertight，适合 3D 打印  
✅ **灵活使用**：支持手动或自动 bounds 设置  
✅ **向后兼容**：现有代码无需修改即可使用新功能  

## 技术实现

### Clipper
- 使用 `geos.CreatePolygon()` / `CreateLineString()` 创建几何对象
- 调用 `geom.ClipByRect(xmin, ymin, xmax, ymax)` 裁剪
- 处理多部分几何（MultiLineString, MultiPolygon）

### 地形感知拉伸
- Path: 对每段起点/终点采样地形高度，计算底面和顶面 Z 坐标
- Area: 为每个顶点预计算地形高度数组，拉伸时使用
- 侧面三角形连接地形底面到顶部面

### 统一底面封闭
- 遍历所有顶点找到全局最小高度
- 复制顶点到底面高度（统一值）
- 反转底部三角形法线，形成封闭底面
- 支持纹理和颜色设置

### 自动 Bounds 解析
- 检查 bounds 是否已有效设置
- 尝试从 provider 获取：`provider.Bounds()`
- 同时解析 SRS：`provider.Srs()`
- 记录 bounds 来源日志

## 测试

运行测试：
```bash
go test ./draw/... -run TestClip
go test ./mesh/builder/...
go build ./...
```

## 兼容性

- 支持所有现有 Provider
- TIFF/GeoTIFF 数据源自动解析 bounds
- Raster/TinMesh Provider 通过 Bounds() 方法暴露
- ImageryProvider 支持自动 bounds
