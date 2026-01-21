# go-static-mesh 架构设计文档

## 1. 项目概述

go-static-mesh 是一个用于生成 3D 模型的 Go 库，支持 3D 打印和展示用途。项目通过组合多种地理空间数据源生成高质量的 3D 模型。

### 1.1 目标

- 支持多种数据源：高程数据、卫星影像、TIN Mesh、3D 标记、地理矢量数据
- 输出多种格式：STL（3D 打印）、GLB/OBJ（展示）
- 灵活的数据获取接口设计
- 支持地理数据的不同展示方式（贴图或突出显示）
- TIN 地形模型作为所有输出的基础

### 1.2 核心数据流

```
数据源 Provider -> 数据获取 -> Raster 转 TIN -> TIN Mesh（面片）
                    -> 闭合处理 -> 闭合 Mesh（体积）-> 输出（STL/GLB/OBJ）
```

### 1.3 重要说明

- **TIN 地形模型是必须存在的**：所有模型输出都基于 TIN Mesh
- **Raster 和 TIN Mesh 源互斥**：RasterProvider 和 TinMeshProvider 只能选择其一
- **Raster 数据转换**：Raster 数据通过 TIN 算法转换为 TIN Mesh（算法由用户提供）
- **TIN Mesh 是面片**：单纯的 TIN 模型是面片（surface），不是闭合的体积
- **3D 打印需要闭合**：STL 输出需要将面片闭合为体积（volume），不是简化
- **展示模型流程**：OBJ 和 GLB 走完全相同的流程，仅输出格式不同
- **外部传入范围**：生成模型时需要外部传入地理范围（bounds + srs）
- **Tile 获取失败处理**：自动处理范围内部分 tile 获取不到的情况
- **Zoom 自动判断**：支持手动设置 zoom 或自动判断最优 zoom

## 2. 核心模块设计

### 2.1 数据源模块 (`static/`)

#### 2.1.1 基础接口

```go
type TileProvider interface {
    Attribution() string
    Grid() *geo.TileGrid
    Bounds() vec2d.Rect
    Srs() geo.Proj
}

type TileFetcher interface {
    Fetch(coord [3]int) (image.Image, error)
}
```

#### 2.1.2 Provider 类型

**RasterProvider** - 高程数据提供者（用于生成 TIN Mesh）
- 支持 image 格式的高程数据
- 支持 GeoTIFF 格式的高程数据
- 提供高度值访问接口
- **重要**: Raster 数据必须通过 TIN 算法转换为 TIN Mesh

```go
type RasterProvider interface {
    TileProvider
    GetElevation(lng, lat float64) float64
    GetElevationGrid() *ElevationGrid
}
```

**ImageryProvider** - 卫星影像提供者（用于生成纹理）
- 支持 TMS 瓦片服务
- 支持 XYZ 瓦片格式
- 支持 GeoTIFF 影像数据

```go
type ImageryProvider interface {
    TileProvider
    GetImageTile(coord [3]int) (image.Image, error)
}
```

**TinMeshProvider** - TIN 网格提供者（可直接使用或作为 TIN 转换输出）
- 支持 Cesium quantized-mesh 格式
- 提供三角形网格数据
- **重要**: TIN Mesh 是所有模型输出的基础

```go
type TinMeshProvider interface {
    TileProvider
    GetMeshTile(coord [3]int) (*TinMesh, error)
    GetMesh() (*TinMesh, error)
}

type TinMesh struct {
    Vertices    []vec3d.T
    Indices     []uint32
    EdgeIndices []uint32
    NorthEdge   []uint16
    SouthEdge   []uint16
    WestEdge    []uint16
    EastEdge    []uint16
    MinHeight   float64
    MaxHeight   float64
}
```

**Model3DProvider** - 3D 模型标记提供者
- 支持带位置的三维模型
- 支持多种 3D 格式

```go
type Model3DProvider interface {
    TileProvider
    GetModels(bounds vec2d.Rect) ([]Model3D, error)
}

type Model3D struct {
    Position vec2d.T
    Elevation float64
    Rotation vec3d.T
    Scale    vec3d.T
    Mesh     interface{} // 支持多种格式
}
```

**GeoDataProvider** - 地理数据提供者
- 支持 GPX 路径数据
- 支持矢量地图数据

```go
type GeoDataProvider interface {
    GetPaths() []Path
    GetAreas() []Area
    GetPoints() []Point
}
```

### 2.2 Mesh 构建模块 (`mesh/`)

#### 2.2.1 核心流程

```
1. RasterProvider -> TIN 算法 -> TinMesh（基础地形）
2. ImageryProvider -> 纹理图像（可选，用于展示）
3. GeoDataProvider -> 贴图绘制 / 突出显示
4. TinMesh + 纹理 + 3D模型 -> 完整 Mesh
5. Mesh -> STL/GLB/OBJ 输出
```

#### 2.2.2 数据源接口

```go
type Source interface {
    Type() SourceType
    Bounds() vec2d.Rect
    TransformTo(srs geo.Proj) Source
}

type SourceType int

const (
    SourceRaster SourceType = iota  // 高程数据，需转 TIN
    SourceTinMesh                    // TIN 网格，可直接使用
    SourceImage                      // 影像数据，用于纹理
    SourceModel3D                    // 3D 模型
)
```

#### 2.2.3 TIN 转换接口（使用 go-tin）

```go
import "github.com/flywave/go-tin"

// go-tin 提供了 TIN 生成算法
// TINGenerator 封装 go-tin 的接口

type TINGenerator interface {
    GenerateFromRaster(grid *ElevationGrid) (*TinMesh, error)
    GenerateFromPoints(points []vec2d.T, elevations []float64) (*TinMesh, error)
    Simplify(mesh *TinMesh, tolerance float64) (*TinMesh, error)
    Optimize(mesh *TinMesh) (*TinMesh, error)
}

// 使用示例
import (
    static "github.com/flywave/go-static-mesh"
    "github.com/flywave/go-static-mesh/mesh"
    "github.com/flywave/go-tin"
)

func main() {
    builder := mesh.NewBuilder()

    // 设置 TIN 生成算法（使用 go-tin）
    tinGenerator := &tin.Algorithm{}
    builder.SetTINGenerator(tinGenerator)

    // ...
}
```

#### 2.2.4 Mesh Builder

```go
type Builder struct {
    tinGenerator          TINGenerator           // TIN 生成算法
    rasterProvider        RasterProvider         // 高程数据源（与 tinMeshProvider 互斥）
    tinMeshProvider       TinMeshProvider        // TIN 网格源（与 rasterProvider 互斥）
    imageryProvider       ImageryProvider         // 影像数据源（可选）
    model3DProvider       Model3DProvider        // 3D 模型源（可选）
    geoData               []MapObject           // 地理数据
    bounds                vec2d.Rect            // 外部传入的范围
    srs                   geo.Proj              // 范围的坐标系
    zoom                  *int                  // 手动设置的 zoom（nil 则自动判断）
    autoZoomMin           int                   // 自动判断的最小 zoom
    autoZoomMax           int                   // 自动判断的最大 zoom
    resolution            float64               // 输出分辨率（影响 zoom 判断）
    verticalExaggeration float64
    baseElevation         float64
    extrudeGeoData        bool
    geoDataHeight         float64
    closeMesh             bool                   // 是否闭合模型（3D 打印需要）
    baseThickness         float64                // 基础厚度（闭合时使用）
    tileErrorHandler      TileErrorHandler       // Tile 获取失败处理策略
}

// Tile 获取失败处理策略
type TileErrorHandler struct {
    SkipMissing bool          // 跳过获取失败的 tile
    UseNoData   bool          // 使用 NoData 值填充
    NoDataValue float64       // NoData 值
    MaxRetries  int           // 最大重试次数
    Timeout     time.Duration // 超时时间
}

func NewBuilder() *Builder
func (b *Builder) SetTINGenerator(generator TINGenerator)

// 设置数据源（RasterProvider 和 TinMeshProvider 互斥）
func (b *Builder) SetRasterProvider(provider RasterProvider)
func (b *Builder) SetTinMeshProvider(provider TinMeshProvider)

func (b *Builder) AddImageryProvider(provider ImageryProvider)
func (b *Builder) AddModel3DProvider(provider Model3DProvider)
func (b *Builder) AddGeoData(obj MapObject)

// 设置范围（必须）
func (b *Builder) SetBounds(bounds vec2d.Rect, srs geo.Proj)

// 设置 zoom 级别（可选，不设置则自动判断）
func (b *Builder) SetZoom(zoom int)
func (b *Builder) SetAutoZoomRange(minZoom, maxZoom int)

// 设置其他参数
func (b *Builder) SetResolution(resolution float64)
func (b *Builder) SetVerticalExaggeration(scale float64)
func (b *Builder) SetBaseElevation(elevation float64)
func (b *Builder) SetExtrudeGeoData(extrude bool, height float64)

// 设置 Tile 获取失败处理策略
func (b *Builder) SetTileErrorHandler(handler TileErrorHandler)

// 设置闭合参数（3D 打印必须）
func (b *Builder) SetCloseMesh(close bool, thickness float64)

// 构建方法
func (b *Builder) BuildForPrint() (*Mesh, error)              // 构建 3D 打印模型（闭合）
func (b *Builder) BuildForDisplay() (*Mesh, error)            // 构建展示模型（面片）
func (b *Builder) BuildForDisplayWithTexture() (*Mesh, error) // 构建带纹理的展示模型

// 内部方法
func (b *Builder) determineZoom() (int, error)                // 自动判断最优 zoom
func (b *Builder) fetchTiles(zoom int) ([]*Tile, error)       // 获取范围内的所有 tile
func (b *Builder) handleTileError(tile [3]int, err error) error // 处理 tile 获取失败
```

#### 2.2.5 Mesh 数据结构

```go
type Mesh struct {
    Vertices    []vec3d.T
    Normals     []vec3d.T
    UVs         []vec2d.T
    Indices     []uint32
    Materials   []Material
    Texture     image.Image
    Bounds      vec2d.Rect
    Srs         geo.Proj
    TinMesh     *TinMesh              // 原始 TIN 数据
}

type Material struct {
    Name        string
    Diffuse     color.Color
    Specular    color.Color
    Shininess   float32
    Alpha       float32
}
```

### 2.3 绘制模块 (`draw/`)

#### 2.3.1 MapObject 接口（已存在）

```go
type MapObject interface {
    Bounds() vec2d.Rect
    SrsProj() geo.Proj
    ExtraMarginPixels() (float64, float64, float64, float64)
    Draw(dc *gg.Context, trans *Transformer)
}
```

#### 2.3.2 扩展接口

```go
type MeshObject interface {
    MapObject
    ExtrudeToMesh(mesh *Mesh, resolution float64) error
}

type TextureObject interface {
    MapObject
    DrawToTexture(dc *gg.Context, trans *Transformer) image.Image
}
```

### 2.4 范围和 Zoom 处理

#### 2.4.1 外部传入范围

**重要**: 所有模型生成都需要外部传入地理范围（bounds + srs）

```go
// 范围必须是外部传入
bounds := vec2d.Rect{{minLat, minLng}, {maxLat, maxLng}}
srs := geo.NewProj(4326) // WGS84
builder.SetBounds(bounds, srs)
```

#### 2.4.2 Zoom 设置和自动判断

**方式 1: 手动设置 Zoom**

```go
builder.SetZoom(15) // 强制使用 zoom 15
```

**方式 2: 自动判断 Zoom**

```go
// 设置自动判断范围
builder.SetAutoZoomRange(10, 18) // zoom 在 10-18 之间自动判断

// 或使用默认范围（13-17）
// builder.SetResolution(1.0) // 分辨率也会影响 zoom 判断
```

**自动判断 Zoom 算法**:

```go
func (b *Builder) determineZoom() (int, error) {
    // 如果手动设置了 zoom，直接使用
    if b.zoom != nil {
        return *b.zoom, nil
    }
    
    // 根据范围和分辨率自动判断
    // 考虑因素：
    // 1. 范围大小：范围越大，zoom 越小
    // 2. 分辨率要求：分辨率越高，zoom 越大
    // 3. Provider 的 grid 信息
    // 4. 内存限制：避免过大的数据量
    
    // 计算最优 zoom
    grid := b.rasterProvider.Grid()
    tileSize := float64(grid.TileSize[0])
    
    // 计算需要的像素数量
    widthPx := b.bounds.Width() * tileSize / (360.0 / math.Pow(2, float64(*zoom)))
    heightPx := b.bounds.Height() * tileSize / (180.0 / math.Pow(2, float64(*zoom)))
    
    // 根据 resolution 计算需要的 zoom
    // ...
    
    // 确保 zoom 在合理范围内
    if bestZoom < b.autoZoomMin {
        bestZoom = b.autoZoomMin
    }
    if bestZoom > b.autoZoomMax {
        bestZoom = b.autoZoomMax
    }
    
    return bestZoom, nil
}
```

#### 2.4.3 Tile 获取失败处理

**处理策略**:

```go
// 方式 1: 跳过获取失败的 tile（默认）
handler := TileErrorHandler{
    SkipMissing: true,
    MaxRetries:  3,
    Timeout:     30 * time.Second,
}
builder.SetTileErrorHandler(handler)

// 方式 2: 使用 NoData 值填充
handler := TileErrorHandler{
    UseNoData:   true,
    NoDataValue: -9999.0,
    MaxRetries:  3,
    Timeout:     30 * time.Second,
}
builder.SetTileErrorHandler
```

**Tile 获取流程**:

```go
func (b *Builder) fetchTiles(zoom int) ([]*Tile, error) {
    grid := b.rasterProvider.Grid()
    
    // 计算范围内的所有 tile 坐标
    tiles := b.calculateTileCoords(zoom, b.bounds)
    
    var wg sync.WaitGroup
    fetchedTiles := make(chan *Tile)
    errors := make(chan error)
    
    // 并发获取 tile
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
                // 处理错误
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
    
    // 收集结果
    result := []*Tile{}
    for tile := range fetchedTiles {
        result = append(result, tile)
    }
    
    return result, nil
}
```

### 2.5 输出模块 (`mesh/`)

#### 2.4.1 模型闭合处理（3D 打印必须）

**重要**: TIN Mesh 是面片（surface），3D 打印需要闭合为体积（volume）

```go
type MeshCloser interface {
    CloseSurfaceMesh(mesh *TinMesh, thickness float64) (*Mesh, error)
    CloseWithBase(mesh *TinMesh, baseHeight float64) (*Mesh, error)
}

// 默认实现：添加底部平面闭合模型
type SimpleCloser struct{}

func (c *SimpleCloser) CloseSurfaceMesh(mesh *TinMesh, thickness float64) (*Mesh, error) {
    // 将 TIN Mesh 面片闭合为体积
    // 1. 添加底部平面
    // 2. 添加侧面三角形
    // 3. 翻转法线方向
    return closedMesh, nil
}
```

#### 2.4.2 STL Writer（3D 打印）

**特点**:
- 必须使用闭合的模型（volume）
- 不支持纹理
- 专注于几何形状
- 适合 3D 打印需求

```go
type StlWriter struct {
    binary      bool
    mergeMesh   bool
}

func (w *StlWriter) Write(mesh *Mesh, writer io.Writer) error
func (w *StlWriter) WriteFile(mesh *Mesh, filepath string) error
```

#### 2.4.2 展示 Writer（GLB 和 OBJ）

**重要**: GLB 和 OBJ 走完全相同的流程，仅输出格式不同

**特点**:
- 支持纹理和材质
- 保留完整的 TIN 网格细节
- 支持多种材质
- 优化视觉效果

**GLTF/GLB Writer**:

```go
type GltfWriter struct {
    binary       bool      // true=GLB, false=GLTF
    embedImages  bool
    draco        bool
}

func (w *GltfWriter) Write(mesh *Mesh, writer io.Writer) error
func (w *GltfWriter) WriteFile(mesh *Mesh, filepath string) error
```

**OBJ Writer**:

```go
type ObjWriter struct {
    separateMtl bool
}

func (w *ObjWriter) Write(mesh *Mesh, writer io.Writer) error
func (w *ObjWriter) WriteFile(mesh *Mesh, filepath string) error
```

#### 2.4.3 统一的展示接口

由于 GLB 和 OBJ 流程相同，可以统一处理：

```go
type DisplayWriter interface {
    Write(mesh *Mesh, writer io.Writer) error
    WriteFile(mesh *Mesh, filepath string) error
}

// 根据扩展名选择 Writer
func NewDisplayWriter(filename string) DisplayWriter {
    if strings.HasSuffix(filename, ".glb") || strings.HasSuffix(filename, ".gltf") {
        return NewGltfWriter()
    }
    if strings.HasSuffix(filename, ".obj") {
        return NewObjWriter()
    }
    return nil
}
```

## 3. 包结构

```
go-static-mesh/
├── static/              # 核心接口和数据源
│   ├── provider.go      # Provider 基础接口定义
│   ├── raster.go        # RasterProvider 实现（高程数据）
│   ├── imagery.go       # ImageryProvider 实现（卫星影像）
│   ├── tinmesh.go       # TinMeshProvider 实现（TIN 网格）
│   ├── model3d.go       # Model3DProvider 实现（3D 模型）
│   ├── geo.go           # GeoDataProvider 实现（GPX/GeoJSON/KML）
│   └── tin_generator.go # TIN 生成接口（算法由用户提供）
├── mesh/               # Mesh 构建和输出
│   ├── builder.go       # Mesh Builder（核心）
│   ├── source.go        # Source 接口和实现
│   ├── mesh.go          # Mesh 数据结构
│   ├── tin.go           # TIN Mesh 数据结构
│   ├── stl.go           # STL Writer（3D 打印）
│   ├── gltf.go          # GLTF/GLB Writer（展示）
│   ├── obj.go           # OBJ Writer（展示）
│   ├── texture.go       # 纹理生成
│   └── writer.go        # 统一的 Writer 接口
├── draw/               # 地理数据绘制
│   ├── context.go       # 绘制上下文（已存在）
│   ├── map_object.go    # MapObject 接口（已存在）
│   ├── path.go          # 路径绘制（已存在）
│   ├── area.go          # 区域绘制（已存在）
│   ├── circle.go        # 圆形绘制（已存在）
│   └── marker.go        # 标记绘制（已存在）
├── raster/             # 栅格数据处理
│   └── raster.go        # 栅格数据工具（已存在）
└── utils/              # 工具函数
    └── utils.go         # 工具函数（已存在）
```

## 4. 使用场景

### 4.1 3D 打印场景

**流程**:
```
RasterProvider (高程) -> TIN 算法 -> TIN Mesh（面片）
                    + GeoDataProvider (GPX路径突出显示)
                    -> 闭合处理 -> 闭合 Mesh（体积）-> STL Writer -> STL 文件
```

**代码示例**:

```go
builder := mesh.NewBuilder()

// 设置 TIN 生成算法
builder.SetTINGenerator(tin.Algorithm)

// 添加高程数据（RasterProvider 和 TinMeshProvider 只能选一个）
builder.SetRasterProvider(rasterProvider)

// 或者使用现有的 TIN Mesh
// builder.SetTinMeshProvider(tinMeshProvider)

// 添加地理数据（可选，突出显示到模型）
builder.AddGeoData(gpxPath.AsMeshObject())

// 设置范围（必须）
builder.SetBounds(bounds, srs)

// 设置 zoom（可选）
// 方式 1: 手动设置
// builder.SetZoom(15)

// 方式 2: 自动判断（根据范围和分辨率）
builder.SetAutoZoomRange(10, 18)

// 设置 Tile 获取失败处理（可选）
handler := mesh.TileErrorHandler{
    SkipMissing: true,
    MaxRetries:  3,
    Timeout:     30 * time.Second,
}
builder.SetTileErrorHandler(handler)

// 设置闭合参数（3D 打印必须）
builder.SetCloseMesh(true, 5.0)  // 闭合模型，基础厚度 5mm

mesh, err := builder.BuildForPrint()  // 构建闭合的打印模型
if err != nil {
    panic(err)
}

writer := mesh.NewStlWriter()
writer.WriteFile(mesh, "model.stl")
```

### 4.2 展示场景（带纹理 - GLB/OBJ）

**重要**: GLB 和 OBJ 流程完全相同，仅输出格式不同
**注意**: 展示模型是面片（surface），不需要闭合

**流程**:
```
RasterProvider (高程) -> TIN 算法 -> TIN Mesh（面片）
                    + ImageryProvider (卫星图) -> 纹理
                    + GeoDataProvider (路径绘制到贴图)
                    -> MeshBuilder (完整细节) -> GLB/OBJ Writer -> GLB/OBJ 文件
```

**代码示例（GLB）**:

```go
builder := mesh.NewBuilder()

// 设置 TIN 生成算法
builder.SetTINGenerator(tin.Algorithm)

// 添加高程数据（RasterProvider 和 TinMeshProvider 只能选一个）
builder.SetRasterProvider(rasterProvider)

// 或者使用现有的 TIN Mesh
// builder.SetTinMeshProvider(tinMeshProvider)

// 添加卫星影像（可选，用于纹理）
builder.AddImageryProvider(imageryProvider)

// 添加地理数据（可选，绘制到纹理）
builder.AddGeoData(gpxPath.AsTextureObject())

// 设置范围（必须）
builder.SetBounds(bounds, srs)

// 设置 zoom（可选）
// 方式 1: 手动设置
// builder.SetZoom(15)

// 方式 2: 自动判断（根据范围和分辨率）
builder.SetAutoZoomRange(10, 18)

// 设置 Tile 获取失败处理（可选）
handler := mesh.TileErrorHandler{
    SkipMissing: true,
    MaxRetries:  3,
    Timeout:     30 * time.Second,
}
builder.SetTileErrorHandler(handler)

// 构建展示模型（面片，不闭合）
mesh, err := builder.BuildForDisplayWithTexture()
if err != nil {
    panic(err)
}

// GLB 输出
writer := mesh.NewGltfWriter()
writer.SetBinary(true)  // 输出 GLB
writer.WriteFile(mesh, "model.glb")
```

**代码示例（OBJ）**:

```go
// ... builder 设置同上 ...

mesh, err := builder.BuildForDisplayWithTexture()  // 构建带纹理的展示模型
if err != nil {
    panic(err)
}

// OBJ 输出
writer := mesh.NewObjWriter()
writer.WriteFile(mesh, "model.obj")
```

### 4.3 TIN Mesh 场景（已有 TIN 数据）

**流程**:
```
TinMeshProvider (已有 TIN) -> MeshBuilder
                    + Model3DProvider (3D 模型)
                    -> MeshBuilder -> GLB/OBJ Writer -> GLB/OBJ 文件
```

**代码示例**:

```go
builder := mesh.NewBuilder()

// 直接使用现有的 TIN Mesh（RasterProvider 和 TinMeshProvider 只能选一个）
builder.SetTinMeshProvider(tinMeshProvider)

// 添加 3D 模型（可选）
builder.AddModel3DProvider(model3DProvider)

builder.SetBounds(bounds, srs)

// 构建展示模型（面片）
mesh, err := builder.BuildForDisplay()
if err != nil {
    panic(err)
}

// 输出（GLB 或 OBJ 都可以）
writer := mesh.NewGltfWriter()
writer.WriteFile(mesh, "model.glb")
```

## 5. 扩展性设计

### 5.1 自定义 Provider

用户可以实现自己的 Provider：

```go
type MyCustomProvider struct {
    // 自定义实现
}

func (p *MyCustomProvider) Attribution() string { ... }
func (p *MyCustomProvider) Grid() *geo.TileGrid { ... }
func (p *MyCustomProvider) GetElevation(lng, lat float64) float64 { ... }
```

### 5.2 自定义 MapObject

用户可以扩展 MapObject 接口：

```go
type MyCustomObject struct {
    // 自定义实现
}

func (o *MyCustomObject) Bounds() vec2d.Rect { ... }
func (o *MyCustomObject) Draw(dc *gg.Context, trans *Transformer) { ... }
func (o *MyCustomObject) ExtrudeToMesh(mesh *Mesh, resolution float64) error { ... }
```

### 5.3 自定义 Writer

用户可以实现自己的输出格式：

```go
type MyCustomWriter struct {
    // 自定义实现
}

func (w *MyCustomWriter) Write(mesh *Mesh, writer io.Writer) error { ... }
```

## 6. 性能优化

### 6.1 并发处理

- Tile 获取使用 goroutine 并行下载
- Mesh 构建过程支持分块处理
- 纹理生成支持 GPU 加速（未来）

### 6.2 内存管理

- 流式处理大数据源
- 使用内存池减少 GC 压力
- 支持分块写入输出文件

## 7. 依赖库

### 7.1 地理空间处理

- `github.com/flywave/go-geo` - 地理坐标转换
- `github.com/flywave/go-proj` - 投影转换
- `github.com/flywave/gg` - 2D 绘图

### 7.2 TIN 和 Mesh 算法

- `github.com/flywave/go-tin` - TIN 三角网算法（生成 TIN Mesh）
- `github.com/flywave/go-quantized-mesh` - Cesium quantized-mesh 解码

### 7.3 栅格和影像处理

- `github.com/flywave/go-mapbox` - Raster 图片-高程解码（Mapbox Terrain RGB）
- `github.com/flywave/go-cog` - GeoTIFF 读取（Cloud Optimized GeoTIFF）

### 7.4 3D 格式

- `github.com/flywave/gltf` - GLTF 格式支持
- `github.com/flywave/go-stl` - STL 格式支持
- `github.com/flywave/go-obj` - OBJ 格式支持
- `github.com/flywave/go3d` - 3D 数学运算

### 7.5 其他

- `github.com/flywave/go-gpx` - GPX 文件解析

## 8. 开发计划

### 8.1 Phase 1 - 核心接口
- [ ] 完成 Provider 接口设计
- [ ] 定义 TIN 生成接口（待用户实现）
- [ ] 实现 Source 接口
- [ ] 完成 Mesh Builder 基础功能

### 8.2 Phase 2 - TIN 核心功能
- [ ] 实现 Raster 到 TIN 的集成
- [ ] 实现 TIN Mesh 数据结构
- [ ] 完成 TIN Mesh Builder 基础功能
- [ ] 实现 TIN 网格优化和简化

### 8.3 Phase 3 - 数据源实现
- [ ] 实现 RasterProvider
- [ ] 实现 ImageryProvider
- [ ] 实现 TinMeshProvider（Cesium quantized-mesh）
- [ ] 实现 Model3DProvider
- [ ] 实现 GeoDataProvider（GPX/GeoJSON/KML）

### 8.4 Phase 4 - 输出格式
- [ ] 完善 STL Writer（3D 打印，带简化）
- [ ] 完善 GLTF/GLB Writer（展示）
- [ ] 完善 OBJ Writer（展示）
- [ ] 实现统一的展示 Writer 接口

### 8.5 Phase 5 - 高级功能
- [ ] 纹理生成和映射
- [ ] 地理数据绘制到纹理（展示）
- [ ] 地理数据突出显示到 Mesh（3D 打印）
- [ ] 3D 模型集成
- [ ] 性能优化
- [ ] 测试和文档
