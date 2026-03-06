# 地形网格生成示例

本示例演示如何使用本地 DEM（数字高程模型）数据构建地形网格。

## 数据准备

示例使用 `data/dem/` 目录中的 GeoTIFF 格式的高程数据。数据文件命名格式：`{z}_{x}_{y}.tif`

- `z`: 缩放级别（zoom level）
- `x`: 瓦片 X 坐标
- `y`: 瓦片 Y 坐标

当前测试数据包含 16 个瓦片（4×4 网格）：
- Zoom Level: 14
- 范围：山东省济南市附近
- 坐标：118.059°E - 118.147°E, 36.474°N - 36.545°N

## 运行示例

```bash
go run examples/terrain_from_local_data/main.go
```

## 程序流程

1. **加载 DEM 数据**：从 GeoTIFF 文件读取高程数据
2. **创建 Provider**：构建本地瓦片提供者
3. **初始化 Builder**：配置地形网格构建器
4. **设置参数**：边界、垂直夸张、基础高程等
5. **生成网格**：构建 TIN（不规则三角网）
6. **导出文件**：保存为 GLTF 格式

## 架构说明

### LocalTileProvider

本地 DEM 瓦片提供者，实现了以下功能：
- 从 GeoTIFF 文件加载高程数据
- 管理多个瓦片
- 提供地理范围和坐标系统信息

### Builder

地形网格构建器，支持：
- 多种数据源（DEM、TIN mesh、影像）
- 垂直夸张和基础高程设置
- 网格闭合选项
- 进度回调

## 后续开发

当前示例提供了基础框架，需要完善以下部分：

1. **TIN 生成算法**
   - 集成 `github.com/flywave/go-tin` 库
   - 实现高程数据采样
   - 生成优化三角网格

2. **数据采样**
   - 从 GeoTIFF 读取高程值
   - 支持不同分辨率
   - 处理 NoData 值

3. **纹理映射**
   - 集成卫星影像（`data/satellite/`）
   - 生成 UV 坐标
   - 支持多种影像格式

4. **性能优化**
   - 实现瓦片缓存
   - 支持并行处理
   - LOD（细节层次）支持

## 参考实现

参考 `github.com/flywave/go-tileproxy/terrain` 模块：
- `dem.go`: DEM 数据编码/解码
- `raster.go`: 栅格数据处理
- `terrain.go`: Quantized Mesh 格式支持

## 输出格式

支持以下 3D 格式：
- **GLTF/GLB**: 推荐，支持 PBR 材质
- **STL**: 简单三角形格式
- **OBJ**: Wavefront OBJ 格式

## 示例代码结构

```
examples/terrain_from_local_data/
├── main.go          # 主程序
└── README.md        # 本文档

data/
├── dem/             # DEM 数据（GeoTIFF）
│   └── *.tif
└── satellite/       # 卫星影像（WebP）
    └── *.webp

output/
└── terrain.gltf     # 输出的 3D 模型
```
