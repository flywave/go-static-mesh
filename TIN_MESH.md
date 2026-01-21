# TIN Mesh 核心流程

## 概述

**TIN Mesh 是所有模型输出的基础**。无论是 3D 打印（STL）还是展示模型（GLB/OBJ），都必须基于 TIN Mesh 构建。

## TIN Mesh 的作用

### 1. 地形表示

TIN（Triangulated Irregular Network）是不规则三角网，用于精确表示地形表面：

- 使用三角形面片表示地形
- 可以精确表示复杂的地形特征
- 在平坦区域使用较少三角形，在复杂区域使用较多三角形
- 比规则网格更高效
- **TIN Mesh 是面片（surface），不是闭合的体积（volume）**

### 2. 所有输出的基础

```
Raster 高程数据 -> TIN 算法 -> TIN Mesh（面片）
    ├─> 闭合处理 -> 闭合 Mesh（体积）-> STL（3D 打印）
    └─> 完整细节 + 纹理 -> GLB/OBJ（展示，面片）
```

### 3. 面片 vs 体积

- **面片（Surface）**: TIN Mesh，只有顶部表面，没有底部和侧面
- **体积（Volume）**: 闭合的 Mesh，有顶部、底部和侧面，适合 3D 打印
- **3D 打印需要体积**: 单纯的面片无法打印，需要闭合为体积
- **展示可以用面片**: Web 展示可以用面片，通过双面渲染显示

## TIN Mesh 数据结构

```go
package mesh

import vec3d "github.com/flywave/go3d/float64/vec3"

type TinMesh struct {
    Vertices    []vec3d.T      // 顶点坐标（x, y, z）
    Indices     []uint32       // 三角形索引（每3个一组）
    EdgeIndices []uint32       // 边索引
    NorthEdge   []uint16       // 北边（用于瓦片拼接）
    SouthEdge   []uint16       // 南边
    WestEdge    []uint16       // 西边
    EastEdge    []uint16       // 东边
    MinHeight   float64        // 最小高度
    MaxHeight   float64        // 最大高度
    Bounds      vec2d.Rect     // 边界范围
    Srs         geo.Proj       // 坐标系
}

// TinMesh 方法
func (m *TinMesh) VertexCount() int
func (m *TinMesh) TriangleCount() int
func (m *TinMesh) CalculateNormals()
func (m *TinMesh) GetElevation(x, y float64) float64
func (m *TinMesh) GetNormal(x, y float64) vec3d.T
func (m *TinMesh) Simplify(tolerance float64) (*TinMesh, error)
func (m *TinMesh) Optimize() error
func (m *TinMesh) ToMesh() (*Mesh, error)
```

## Raster 到 TIN 的转换流程

### 1. Raster 数据结构

```go
package static

type ElevationGrid struct {
    Width      int           // 网格宽度
    Height     int           // 网格高度
    Data       []float64     // 高程数据（一维数组）
    MinX       float64       // 最小 X 坐标
    MinY       float64       // 最小 Y 坐标
    CellSize   float64       // 单元格大小
    NoData     float64       // 无数据值
    Bounds     vec2d.Rect    // 边界
    Srs        geo.Proj      // 坐标系
}

// ElevationGrid 方法
func (g *ElevationGrid) GetElevation(x, y float64) float64
func (g *ElevationGrid) GetElevationByGrid(gridX, gridY int) float64
func (g *ElevationGrid) GetBounds() vec2d.Rect
```

### 2. TIN 生成接口（用户提供算法）

```go
package mesh

type TINGenerator interface {
    // 从 Raster 网格生成 TIN
    GenerateFromRaster(grid *static.ElevationGrid) (*TinMesh, error)
    
    // 从离散点生成 TIN
    GenerateFromPoints(points []vec2d.T, elevations []float64) (*TinMesh, error)
    
    // 简化 TIN（用于 3D 打印）
    Simplify(mesh *TinMesh, tolerance float64) (*TinMesh, error)
    
    // 优化 TIN（提高三角形质量）
    Optimize(mesh *TinMesh) (*TinMesh, error)
}
```

### 3. 转换流程

```go
package mesh

func (b *Builder) GenerateTINFromRaster() (*TinMesh, error) {
    // 1. 确定最优 zoom（如果未手动设置）
    zoom, err := b.determineZoom()
    if err != nil {
        return nil, err
    }
    
    // 2. 获取范围内的所有 tile（处理获取失败的情况）
    tiles, err := b.fetchTiles(zoom)
    if err != nil {
        return nil, err
    }
    
    // 3. 合并 tile 数据为单一网格
    grid := b.mergeTilesToGrid(tiles)
    
    // 4. 调用 TIN 生成算法
    tinMesh, err := b.tinGenerator.GenerateFromRaster(grid)
    if err != nil {
        return nil, err
    }
    
    // 5. 应用垂直夸张
    b.applyVerticalExaggeration(tinMesh)
    
    // 6. 应用基础高程偏移
    b.applyBaseElevation(tinMesh)
    
    return tinMesh, nil
}

func (b *Builder) applyVerticalExaggeration(mesh *TinMesh) {
    for i := range mesh.Vertices {
        mesh.Vertices[i][2] *= b.verticalExaggeration
    }
}

func (b *Builder) applyBaseElevation(mesh *TinMesh) {
    for i := range mesh.Vertices {
        mesh.Vertices[i][2] += b.baseElevation
    }
}
```

## 范围和 Zoom 处理

### 1. 外部传入范围

**重要**: 所有模型生成都需要外部传入地理范围

```go
// 范围必须是外部传入
bounds := vec2d.Rect{{minLat, minLng}, {maxLat, maxLng}}
srs := geo.NewProj(4326) // WGS84
builder.SetBounds(bounds, srs)
```

### 2. Zoom 设置和自动判断

**方式 1: 手动设置 Zoom**

```go
builder.SetZoom(15) // 强制使用 zoom 15
```

**方式 2: 自动判断 Zoom**

```go
// 设置自动判断范围
builder.SetAutoZoomRange(10, 18) // zoom 在 10-18 之间自动判断

// 自动判断算法会考虑：
// - 范围大小：范围越大，zoom 越小
// - 分辨率要求：分辨率越高，zoom 越大
// - Provider 的 grid 信息
// - 内存限制：避免过大的数据量
```

### 3. Tile 获取失败处理

```go
// 方式 1: 跳过获取失败的 tile（推荐用于展示）
handler := TileErrorHandler{
    SkipMissing: true,
    MaxRetries:  3,
    Timeout:     30 * time.Second,
}
builder.SetTileErrorHandler(handler)

// 方式 2: 使用 NoData 值填充（推荐用于 3D 打印）
handler := TileErrorHandler{
    UseNoData:   true,
    NoDataValue: 0.0, // 或其他合适的默认值
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

### 4. Tile 获取流程

```go
func (b *Builder) fetchTiles(zoom int) ([]*Tile, error) {
    // 1. 计算范围内的所有 tile 坐标
    tiles := b.calculateTileCoords(zoom, b.bounds)
    
    var wg sync.WaitGroup
    fetchedTiles := make(chan *Tile)
    errors := make(chan error)
    
    // 2. 并发获取 tile（带重试）
    for _, tileCoord := range tiles {
        wg.Add(1)
        go func(coord [3]int) {
            defer wg.Done()
            
            var tile *Tile
            var err error
            
            // 重试机制
            for i := 0; i < b.tileErrorHandler.MaxRetries; i++ {
                tile, err = b.fetchSingleTile(coord, zoom)
                if err == nil {
                    break
                }
                time.Sleep(time.Second * time.Duration(i+1))
            }
            
            if err != nil {
                // 3. 根据错误处理策略处理失败
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
    
    // 4. 收集结果
    result := []*Tile{}
    for tile := range fetchedTiles {
        result = append(result, tile)
    }
    
    return result, nil
}
```

### 5. Tile 合并为网格

```go
func (b *Builder) mergeTilesToGrid(tiles []*Tile) *ElevationGrid {
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
    tileSize := float64(256) // 假设 tile 为 256x256
    gridWidth := int(math.Ceil((maxX - minX) / tileSize))
    gridHeight := int(math.Ceil((maxY - minY) / tileSize))
    
    // 3. 创建合并后的网格
    grid := &ElevationGrid{
        Width:    gridWidth * 256,
        Height:   gridHeight * 256,
        MinX:     minX,
        MinY:     minY,
        CellSize: tileSize / 256.0,
        Data:     make([]float64, gridWidth*256*gridHeight*256),
    }
    
    // 4. 将每个 tile 的数据复制到网格中
    for _, tile := range tiles {
        startX := int((tile.Bounds.Min[0] - minX) / tileSize * 256)
        startY := int((tile.Bounds.Min[1] - minY) / tileSize * 256)
        
        for y := 0; y < 256; y++ {
            for x := 0; x < 256; x++ {
                gridIdx := (startY+y)*grid.Width + (startX+x)
                tileIdx := y*256 + x
                grid.Data[gridIdx] = tile.Data[tileIdx]
            }
        }
    }
    
    return grid
}
```

## 从 TinMesh 到不同输出格式

### 1. 面片闭合为体积（3D 打印必须）

```go
// MeshCloser 接口 - 将面片闭合为体积
type MeshCloser interface {
    CloseSurfaceMesh(mesh *TinMesh, thickness float64) (*Mesh, error)
    CloseWithBase(mesh *TinMesh, baseHeight float64) (*Mesh, error)
}

// 默认实现：添加底部平面和侧面
type SimpleCloser struct{}

func (c *SimpleCloser) CloseSurfaceMesh(mesh *TinMesh, thickness float64) (*Mesh, error) {
    // 1. 计算最小高度
    minHeight := mesh.MinHeight
    baseHeight := minHeight - thickness
    
    // 2. 复制顶部顶点
    vertices := make([]vec3d.T, len(mesh.Vertices)*2)
    copy(vertices, mesh.Vertices)
    
    // 3. 创建底部顶点
    for i := 0; i < len(mesh.Vertices); i++ {
        vertices[len(mesh.Vertices)+i] = vec3d.T{
            mesh.Vertices[i][0],
            mesh.Vertices[i][1],
            baseHeight,
        }
    }
    
    // 4. 复制顶部三角形
    indices := make([]uint32, len(mesh.Indices)*2)
    copy(indices, mesh.Indices)
    
    // 5. 创建底部三角形（翻转法线）
    offset := uint32(len(mesh.Vertices))
    for i := 0; i < len(mesh.Indices); i += 3 {
        idx := len(mesh.Indices) + i
        indices[idx+0] = mesh.Indices[i+2] + offset
        indices[idx+1] = mesh.Indices[i+1] + offset
        indices[idx+2] = mesh.Indices[i+0] + offset
    }
    
    // 6. 添加侧面三角形（连接顶部和底部）
    // 遍历边界边，添加侧面三角形
    sideIndices := c.generateSideTriangles(mesh, offset)
    indices = append(indices, sideIndices...)
    
    // 7. 创建 Mesh
    closedMesh := &Mesh{
        Vertices: vertices,
        Indices:  indices,
        Bounds:   mesh.Bounds,
        Srs:      mesh.Srs,
    }
    
    // 8. 计算法线
    closedMesh.CalculateNormals()
    
    return closedMesh, nil
}
```

### 2. TIN 到 Mesh（内部表示，面片）

```go
func (m *TinMesh) ToMesh() (*Mesh, error) {
    mesh := &Mesh{
        Vertices: make([]vec3d.T, len(m.Vertices)),
        Normals:  make([]vec3d.T, len(m.Vertices)),
        Indices:  m.Indices,
        Bounds:   m.Bounds,
        Srs:      m.Srs,
        TinMesh:  m,
    }
    
    // 复制顶点
    copy(mesh.Vertices, m.Vertices)
    
    // 计算法线
    mesh.CalculateNormals()
    
    return mesh, nil
}
```

### 2. TIN 到 STL（3D 打印 - 闭合体积）

```go
func (b *Builder) BuildForPrint() (*Mesh, error) {
    // 1. 生成 TIN Mesh（面片）
    tinMesh, err := b.GenerateTINFromRaster()
    if err != nil {
        return nil, err
    }
    
    // 2. 处理地理数据（突出显示到面片上）
    for _, geoData := range b.geoData {
        if meshObj, ok := geoData.(MeshObject); ok && b.extrudeGeoData {
            meshObj.ExtrudeToMesh(tinMesh, b.geoDataHeight)
        }
    }
    
    // 3. 闭合面片为体积（3D 打印必须）
    if b.closeMesh {
        closer := &SimpleCloser{}
        mesh, err := closer.CloseSurfaceMesh(tinMesh, b.baseThickness)
        if err != nil {
            return nil, err
        }
        return mesh, nil
    }
    
    // 4. 如果不需要闭合，返回面片 Mesh（不推荐用于 3D 打印）
    return tinMesh.ToMesh(), nil
}
```

### 3. TIN 到 GLB/OBJ（展示 - 面片）

```go
func (b *Builder) BuildForDisplay() (*Mesh, error) {
    // 1. 生成 TIN Mesh（面片）
    tinMesh, err := b.GenerateTINFromRaster()
    if err != nil {
        return nil, err
    }
    
    // 2. 处理地理数据（绘制到纹理）
    var texture image.Image
    if b.imageryProvider != nil {
        texture = b.generateTexture()
    }
    
    // 3. 添加 3D 模型
    if b.model3DProvider != nil {
        b.add3DModelsToMesh(tinMesh)
    }
    
    // 4. 转换为 Mesh（面片，不闭合）
    mesh := tinMesh.ToMesh()
    mesh.Texture = texture
    
    // 5. 计算 UV 坐标
    mesh.CalculateUVs(mesh.Bounds)
    
    return mesh, nil
}
```

### 4. TIN 到带纹理的 GLB/OBJ（展示 - 面片）

```go
func (b *Builder) BuildForDisplayWithTexture() (*Mesh, error) {
    mesh, err := b.BuildForDisplay()
    if err != nil {
        return nil, err
    }
    
    // 生成高分辨率纹理
    mesh.Texture = b.generateHighResTexture()
    
    return mesh, nil
}
```

## 地理数据处理

### 1. 突出显示到 Mesh（3D 打印）

```go
func (p *Path) ExtrudeToMesh(tinMesh *TinMesh, height float64) error {
    // 1. 将路径点投影到 TIN 表面
    surfacePoints := make([]vec3d.T, len(p.Points))
    for i, pt := range p.Points {
        surfacePoints[i] = tinMesh.GetElevationPoint(pt[0], pt[1])
    }
    
    // 2. 创建突起的三角形
    for i := 0; i < len(surfacePoints)-1; i++ {
        p1 := surfacePoints[i]
        p2 := surfacePoints[i+1]
        
        // 创建突起的高度
        p1Top := vec3d.T{p1[0], p1[1], p1[2] + height}
        p2Top := vec3d.T{p2[0], p2[1], p2[2] + height}
        
        // 添加三角形到 TIN Mesh
        tinMesh.AddTriangle(p1, p2, p1Top)
        tinMesh.AddTriangle(p2, p2Top, p1Top)
    }
    
    return nil
}
```

### 2. 绘制到纹理（展示）

```go
func (b *Builder) generateTexture() image.Image {
    // 1. 创建绘制上下文
    ctx := draw.NewContext()
    ctx.SetTileProvider(b.imageryProvider)
    
    // 2. 添加地理对象
    for _, geoData := range b.geoData {
        if texObj, ok := geoData.(TextureObject); ok {
            ctx.AddObject(geoData)
        }
    }
    
    // 3. 设置边界和渲染
    ctx.SetBoundingBox(b.bounds, b.srs)
    ctx.SetSize(b.textureWidth, b.textureHeight)
    
    texture, err := ctx.Render()
    if err != nil {
        return nil
    }
    
    return texture
}
```

## TIN Mesh 优化

### 1. 简化（可选，用于减少三角形数量）

**注意**: 简化不是为了 3D 打印，而是为了减少数据量

```go
// 简化算法示例（用户提供）
type SimpleSimplifier struct{}

func (s *SimpleSimplifier) Simplify(mesh *TinMesh, tolerance float64) (*TinMesh, error) {
    // 实现简化算法，如：
    // - Edge collapse
    // - Vertex decimation
    // - Quadric error metrics
    
    // 返回简化后的 TIN Mesh
    return simplifiedMesh, nil
}
```

### 2. 优化（提高三角形质量）

```go
type SimpleOptimizer struct{}

func (o *SimpleOptimizer) Optimize(mesh *TinMesh) (*TinMesh, error) {
    // 实现优化算法，如：
    // - Laplacian smoothing
    // - Triangle quality improvement
    // - Edge flipping
    
    return optimizedMesh, nil
}
```

## 使用示例

### 示例 1: 基础 TIN 生成

```go
package main

import (
    static "github.com/flywave/go-static-mesh"
    "github.com/flywave/go-static-mesh/mesh"
    "github.com/flywave/go-tin" // 用户提供的 TIN 算法
)

func main() {
    // 1. 加载高程数据
    rasterProvider := static.NewGeoTIFFRasterProvider("elevation.tif")
    
    // 2. 创建 Builder
    builder := mesh.NewBuilder()
    builder.AddRasterProvider(rasterProvider)
    
    // 3. 设置 TIN 生成算法（用户提供）
    builder.SetTINGenerator(&tin.Algorithm{})
    
    // 4. 设置边界
    builder.SetBounds(rasterProvider.Bounds(), rasterProvider.Srs())
    
    // 5. 生成 TIN Mesh
    tinMesh, err := builder.GenerateTINFromRaster()
    if err != nil {
        panic(err)
    }
    
    // 6. 使用 TIN Mesh
    mesh := tinMesh.ToMesh()
}
```

### 示例 2: TIN 到 STL（3D 打印 - 闭合体积）

```go
func main() {
    builder := mesh.NewBuilder()
    
    // 设置 TIN 生成算法
    builder.SetTINGenerator(&tin.Algorithm{})
    
    // 添加高程数据（RasterProvider 和 TinMeshProvider 只能选一个）
    rasterProvider := static.NewGeoTIFFRasterProvider("elevation.tif")
    builder.SetRasterProvider(rasterProvider)
    
    // 或者使用现有的 TIN Mesh
    // builder.SetTinMeshProvider(tinMeshProvider)
    
    // 设置边界
    builder.SetBounds(rasterProvider.Bounds(), rasterProvider.Srs())
    
    // 设置闭合参数（3D 打印必须）
    builder.SetCloseMesh(true, 5.0)  // 闭合模型，基础厚度 5mm
    
    // 构建 3D 打印模型（自动闭合为体积）
    mesh, err := builder.BuildForPrint()
    if err != nil {
        panic(err)
    }
    
    // 导出 STL
    writer := mesh.NewStlWriter()
    writer.WriteFile(mesh, "model.stl")
}
```

### 示例 3: TIN 到 GLB（展示）

```go
func main() {
    builder := mesh.NewBuilder()
    
    // 设置 TIN 生成算法
    builder.SetTINGenerator(&tin.Algorithm{})
    
    // 添加高程数据
    rasterProvider := static.NewGeoTIFFRasterProvider("elevation.tif")
    builder.AddRasterProvider(rasterProvider)
    
    // 添加卫星影像（使用 NewTileProvider）
    imageryProvider := static.NewTileProvider(
        "https://tile.example.com/{z}/{x}/{y}.png",
        static.TileProviderModeXYZ,
    )
    builder.AddImageryProvider(imageryProvider)
    
    // 设置边界
    builder.SetBounds(rasterProvider.Bounds(), rasterProvider.Srs())
    
    // 构建展示模型（保留完整细节 + 纹理）
    mesh, err := builder.BuildForDisplayWithTexture()
    if err != nil {
        panic(err)
    }
    
    // 导出 GLB
    writer := mesh.NewGltfWriter()
    writer.WriteFile(mesh, "model.glb")
}
```

## 注意事项

1. **TIN Mesh 是必须的**: 所有输出都基于 TIN Mesh，不能跳过
2. **Raster 和 TIN Mesh 源互斥**: RasterProvider 和 TinMeshProvider 只能选择其一
3. **Raster 必须转换**: Raster 数据必须通过 TIN 算法转换为 TIN Mesh
4. **TIN 算法由用户提供**: `TINGenerator` 接口需要用户实现
5. **展示模型流程统一**: GLB 和 OBJ 走完全相同的构建流程
6. **3D 打印需要闭合**: TIN Mesh 是面片，3D 打印需要闭合为体积，不是简化
7. **面片 vs 体积**: 
   - 面片（Surface）: 只有顶部表面，适合展示
   - 体积（Volume）: 有顶部、底部和侧面，适合 3D 打印

## 性能考虑

1. **TIN 生成**: 可能较慢，建议缓存结果
2. **闭合处理**: 3D 打印需要闭合，会增加三角形数量
3. **简化处理**: 可选的，用于减少数据量（不是为了打印）
4. **纹理生成**: 高分辨率纹理可能占用大量内存
5. **网格优化**: 对于大数据集，优化可能耗时
6. **内存管理**: 闭合后的体积模型比面片模型占用更多内存
