# DEM 瓦片合并修复说明

## 问题描述

在 Mapbox XYZ 瓦片系统中，瓦片坐标的 Y 轴方向与地理坐标的纬度方向是相反的：

- **瓦片坐标**: Y=0 在顶部，Y 值向下递增
- **地理坐标**: 纬度向北递增（Y 轴正方向向上）

## 瓦片布局示例

对于 zoom=14 的瓦片：
- **Y=6403**: 纬度 36.5273-36.5449（最北边，纬度最高）
- **Y=6404**: 纬度 36.5096-36.5273
- **Y=6405**: 纬度 36.4920-36.5096
- **Y=6406**: 纬度 36.4743-36.4920（最南边，纬度最低）

## 修复方案

在 `examples/terrain_with_texture/main.go` 中，修改了 DEM 瓦片的加载顺序：

### 修改前（错误）
```go
for y := minY; y <= maxY; y++ {
    for x := minX; x <= maxX; x++ {
        coords[idx] = [3]int{x, y, zoom}
        idx++
    }
}
```

这会导致：
- 图像顶部（Y=0）→ Y=6403（北边）
- 图像底部（Y=768）→ Y=6406（南边）
- **问题**: 图像坐标系和地理坐标系不匹配

### 修改后（正确）
```go
for y := maxY; y >= minY; y-- {
    for x := minX; x <= maxX; x++ {
        coords[idx] = [3]int{x, y, zoom}
        idx++
    }
}
```

这会导致：
- 图像顶部（Y=0）→ Y=6406（南边）
- 图像底部（Y=768）→ Y=6403（北边）
- **结果**: 符合地理坐标系统

## 为什么卫星图不需要修改？

卫星图使用 `LayerMerger`，它根据瓦片的地理边界（Bounds）自动定位每个瓦片的位置，不依赖于瓦片的加载顺序。

## 验证方法

1. 运行测试程序生成瓦片布局图：
   ```bash
   go run test_tile_layout.go
   ```
   
2. 查看生成的 `tile_layout_test.png`：
   - 红色边框: Y=6403（北边）应该在图像底部
   - 黄色边框: Y=6406（南边）应该在图像顶部

3. 查看 `merged_dem.png`：
   - 地形特征应该与卫星图对齐
   - 不应该出现明显的断裂或错位

## 文件输出

修复后会生成以下文件：

- `examples/terrain_with_texture/output/merged_dem.png` - 合并后的 DEM 高程图
- `examples/terrain_with_texture/output/merged_texture.png` - 合并后的卫星纹理
- `examples/terrain_with_texture/output/terrain_textured.gltf` - 最终的 3D 地形网格
- `tile_layout_test.png` - 瓦片布局验证图

## 总结

- ✅ DEM 合并：需要反转 Y 坐标顺序以匹配地理坐标系统
- ✅ 卫星图合并：使用地理边界自动定位，无需修改
- ✅ 地形连续性：瓦片边界无缝拼接，形成连续地形
