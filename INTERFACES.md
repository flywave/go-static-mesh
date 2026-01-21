# 接口设计详解

## 1. 核心接口层级

```
TileProvider (基础)
    ├─> RasterProvider (高程数据)
    ├─> ImageryProvider (影像数据)
    ├─> TinMeshProvider (TIN网格)
    └─> Model3DProvider (3D模型)
```

## 2. Provider 接口详细定义

### 2.1 TileProvider - 基础瓦片提供者

位置: `static/provider.go`

```go
package static

import (
    vec2d "github.com/flywave/go3d/float64/vec2"
    "github.com/flywave/go-geo"
)

type TileProvider interface {
    Attribution() string
    Grid() *geo.TileGrid
    Bounds() vec2d.Rect
    Srs() geo.Proj
}
```

**用途**: 所有数据源的基础接口，提供瓦片网格信息和属性信息

### 2.2 TileFetcher - 瓦片获取接口

位置: `static/provider.go`

```go
type TileFetcher interface {
    Fetch(coord [3]int) (image.Image, error)
}
```

**用途**: 获取瓦片图像数据

## 3. 专用 Provider 接口

### 3.1 RasterProvider - 高程数据提供者

位置: `static/raster.go`

```go
package static

type ElevationGrid struct {
    Width      int
    Height     int
    Data       []float64
    MinX       float64
    MinY       float64
    CellSize   float64
    NoData     float64
}

type RasterProvider interface {
    TileProvider
    
    GetElevation(lng, lat float64) float64
    GetElevationGrid() *ElevationGrid
    GetElevationTile(coord [3]int) (*ElevationGrid, error)
}
```

**实现示例**:
- GeoTIFFRasterProvider: 从 GeoTIFF 文件读取高程数据（使用 go-cog）
- MapboxTerrainRGBProvider: 从 Mapbox Terrain RGB 读取高程数据（使用 go-mapbox）
- ImageRasterProvider: 从 image.Image 读取高程数据

```go
// 使用 GeoTIFFRasterProvider（使用 go-cog）
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
    
    // 或使用 COG（Cloud Optimized GeoTIFF）
    provider, err := static.NewCOGRasterProvider("elevation.cog.tif")
    if err != nil {
        panic(err)
    }
}

// 使用 MapboxTerrainRGBProvider（使用 go-mapbox）
import (
    static "github.com/flywave/go-static-mesh"
    "github.com/flywave/go-mapbox"
)

func main() {
    // 读取 Mapbox Terrain RGB 数据
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

### 3.2 ImageryProvider - 卫星影像提供者

位置: `static/imagery.go`

```go
package static

type ImageryProvider interface {
    TileProvider
    TileFetcher
    
    GetImageTile(coord [3]int) (image.Image, error)
    GetImageBounds() vec2d.Rect
}
```

**实现示例**:
- TileProvider: 通用的瓦片提供者，通过参数设置 xyz 或 tms
- GeoTIFFImageryProvider: 从 GeoTIFF 读取影像（使用 go-cog）
- MapboxTerrainProvider: Mapbox Terrain RGB 高程数据（使用 go-mapbox）

```go
// 使用 NewTileProvider 创建瓦片提供者
// 支持 xyz 和 tms 模式

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
```

### 3.3 TinMeshProvider - TIN 网格提供者

位置: `static/tinmesh.go`

```go
package static

import vec3d "github.com/flywave/go3d/float64/vec3"

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

type TinMeshProvider interface {
    TileProvider
    
    GetMeshTile(coord [3]int) (*TinMesh, error)
    GetHeight(lng, lat float64) float64
    GetNormal(lng, lat float64) vec3d.T
}
```

**实现示例**:
- CesiumQuantizedMeshProvider: Cesium quantized-mesh 格式（使用 go-quantized-mesh）
- CustomTinMeshProvider: 自定义 TIN 格式

```go
// 使用 CesiumQuantizedMeshProvider
import (
    static "github.com/flywave/go-static-mesh"
    "github.com/flywave/go-quantized-mesh"
)

func main() {
    // 创建 Cesium quantized-mesh 提供者
    provider := static.NewCesiumQuantizedMeshProvider(
        "https://example.com/terrain/{z}/{x}/{y}.terrain",
    )
    
    // 或者使用自定义配置
    config := &quantizedmesh.DecoderConfig{
        // 配置选项
    }
    provider := static.NewCesiumQuantizedMeshProviderWithConfig(
        "https://example.com/terrain/{z}/{x}/{y}.terrain",
        config,
    )
}
```

### 3.4 Model3DProvider - 3D 模型提供者

位置: `static/model3d.go`

```go
package static

import vec3d "github.com/flywave/go3d/float64/vec3"

type Model3D struct {
    ID          string
    Position    vec2d.T
    Elevation   float64
    Rotation    vec3d.T
    Scale       vec3d.T
    MeshData    []byte
    Format      string // glb, gltf, obj, stl, etc.
    Metadata    map[string]interface{}
}

type Model3DProvider interface {
    TileProvider
    
    GetModels(bounds vec2d.Rect) ([]Model3D, error)
    GetModel(id string) (*Model3D, error)
}
```

**实现示例**:
- StaticModelProvider: 静态 3D 模型文件
- GeoJSON3DProvider: 从 GeoJSON 读取 3D 建筑物
- I3SProvider: Esri I3S 服务

### 3.5 GeoDataProvider - 地理数据提供者

位置: `static/geo.go`

```go
package static

type GeoDataProvider interface {
    GetPaths() []Path
    GetAreas() []Area
    GetPoints() []Point
}
```

**实现示例**:
- GPXProvider: 从 GPX 文件读取路径
- GeoJSONProvider: 从 GeoJSON 读取矢量数据
- KMLProvider: 从 KML 文件读取地理数据
- ShapefileProvider: 从 Shapefile 读取矢量数据

## 4. Mesh Source 接口

位置: `mesh/source.go`

```go
package mesh

import vec2d "github.com/flywave/go3d/float64/vec2"
import "github.com/flywave/go-geo"

type SourceType int

const (
    SourceRaster SourceType = iota
    SourceTinMesh
    SourceImage
    SourceModel3D
)

type Source interface {
    Type() SourceType
    Bounds() vec2d.Rect
    Srs() geo.Proj
    TransformTo(srs geo.Proj) (Source, error)
}

type RasterSource struct {
    provider static.RasterProvider
    bounds   vec2d.Rect
    srs      geo.Proj
}

func NewRasterSource(provider static.RasterProvider) *RasterSource

type TinMeshSource struct {
    provider static.TinMeshProvider
    bounds   vec2d.Rect
    srs      geo.Proj
}

func NewTinMeshSource(provider static.TinMeshProvider) *TinMeshSource

type Model3DSource struct {
    provider static.Model3DProvider
    bounds   vec2d.Rect
    srs      geo.Proj
}

func NewModel3DSource(provider static.Model3DProvider) *Model3DSource
```

## 5. MapObject 扩展接口

位置: `draw/map_object.go`

```go
package draw

import vec3d "github.com/flywave/go3d/float64/vec3"
import "github.com/flywave/go-static-mesh/mesh"

type MeshObject interface {
    MapObject
    ExtrudeToMesh(meshBuilder *mesh.Builder, height float64) error
}

type TextureObject interface {
    MapObject
    DrawToTexture(dc *gg.Context, trans *Transformer) image.Image
}

// Path 实现 MeshObject - 用于 3D 打印
func (p *Path) ExtrudeToMesh(meshBuilder *mesh.Builder, height float64) error

// Path 实现 TextureObject - 用于展示
func (p *Path) DrawToTexture(dc *gg.Context, trans *Transformer) image.Image

// Area 实现 MeshObject
func (a *Area) ExtrudeToMesh(meshBuilder *mesh.Builder, height float64) error

// Area 实现 TextureObject
func (a *Area) DrawToTexture(dc *gg.Context, trans *Transformer) image.Image
```

## 6. Mesh Builder 接口

位置: `mesh/builder.go`

```go
package mesh

import (
    vec2d "github.com/flywave/go3d/float64/vec2"
    "github.com/flywave/go-geo"
    static "github.com/flywave/go-static-mesh"
    draw "github.com/flywave/go-static-mesh/draw"
)

type Builder struct {
    tinGenerator          TINGenerator           // TIN 生成算法
    rasterProvider        static.RasterProvider  // 高程数据源（与 tinMeshProvider 互斥）
    tinMeshProvider       static.TinMeshProvider // TIN 网格源（与 rasterProvider 互斥）
    imageryProvider       static.ImageryProvider // 影像数据源（可选）
    model3DProvider       static.Model3DProvider // 3D 模型源（可选）
    geoData               []draw.MapObject      // 地理数据
    bounds                vec2d.Rect            // 外部传入的范围
    srs                   geo.Proj              // 范围的坐标系
    zoom                  *int                  // 手动设置的 zoom（nil 则自动判断）
    autoZoomMin           int                   // 自动判断的最小 zoom
    autoZoomMax           int                   // 自动判断的最大 zoom
    resolution            float64               // 输出分辨率（影响 zoom 判断）
    verticalExaggeration  float64
    baseElevation         float64
    extrudeGeoData        bool
    geoDataHeight         float64
    closeMesh             bool                   // 是否闭合模型（3D 打印需要）
    baseThickness         float64                // 基础厚度（闭合时使用）
    tileErrorHandler      TileErrorHandler       // Tile 获取失败处理策略
}

func NewBuilder() *Builder

func (b *Builder) SetTINGenerator(generator TINGenerator)

// 设置数据源（RasterProvider 和 TinMeshProvider 互斥）
func (b *Builder) SetRasterProvider(provider static.RasterProvider)
func (b *Builder) SetTinMeshProvider(provider static.TinMeshProvider)

// 设置范围（必须）
func (b *Builder) SetBounds(bounds vec2d.Rect, srs geo.Proj)

// 设置 zoom 级别（可选，不设置则自动判断）
func (b *Builder) SetZoom(zoom int)
func (b *Builder) SetAutoZoomRange(minZoom, maxZoom int)

// 设置 Tile 获取失败处理策略
func (b *Builder) SetTileErrorHandler(handler TileErrorHandler)

type TileErrorHandler struct {
    SkipMissing bool          // 跳过获取失败的 tile
    UseNoData   bool          // 使用 NoData 值填充
    NoDataValue float64       // NoData 值
    MaxRetries  int           // 最大重试次数
    Timeout     time.Duration // 超时时间
}

func (b *Builder) AddImageryProvider(provider static.ImageryProvider)
func (b *Builder) AddModel3DProvider(provider static.Model3DProvider)

func (b *Builder) AddGeoData(obj draw.MapObject)
func (b *Builder) AddPath(path *draw.Path)
func (b *Builder) AddArea(area *draw.Area)

func (b *Builder) SetBounds(bounds vec2d.Rect, srs geo.Proj)
func (b *Builder) SetVerticalExaggeration(scale float64)
func (b *Builder) SetBaseElevation(elevation float64)
func (b *Builder) SetExtrudeGeoData(extrude bool, height float64)

// 设置闭合参数（3D 打印必须）
func (b *Builder) SetCloseMesh(close bool, thickness float64)

// 构建方法
func (b *Builder) BuildForPrint() (*Mesh, error)              // 构建 3D 打印模型（闭合体积）
func (b *Builder) BuildForDisplay() (*Mesh, error)            // 构建展示模型（面片）
func (b *Builder) BuildForDisplayWithTexture() (*Mesh, error) // 构建带纹理的展示模型

// 内部方法
func (b *Builder) determineZoom() (int, error)                // 自动判断最优 zoom
func (b *Builder) fetchTiles(zoom int) ([]*Tile, error)       // 获取范围内的所有 tile
func (b *Builder) handleTileError(tile [3]int, err error) error // 处理 tile 获取失败
func (b *Builder) mergeTilesToGrid(tiles []*Tile) *ElevationGrid // 合并 tile 为网格
```

## 7. Mesh 数据结构

位置: `mesh/mesh.go`

```go
package mesh

import (
    "image/color"
    vec2d "github.com/flywave/go3d/float64/vec2"
    vec3d "github.com/flywave/go3d/float64/vec3"
)

type Mesh struct {
    Vertices    []vec3d.T
    Normals     []vec3d.T
    UVs         []vec2d.T
    Indices     []uint32
    Materials   []Material
    Texture     image.Image
    Bounds      vec2d.Rect
    Srs         geo.Proj
}

type Material struct {
    Name        string
    Diffuse     color.Color
    Specular    color.Color
    Shininess   float32
    Alpha       float32
}

func (m *Mesh) VertexCount() int
func (m *Mesh) TriangleCount() int
func (m *Mesh) CalculateNormals()
func (m *Mesh) CalculateUVs(bounds vec2d.Rect)
func (m *Mesh) Optimize() error
func (m *Mesh) Merge(other *Mesh) error
```

## 8. Writer 接口

位置: `mesh/writer.go`

```go
package mesh

import "io"

type Writer interface {
    Write(mesh *Mesh, writer io.Writer) error
    WriteFile(mesh *Mesh, filepath string) error
}

// STL Writer
type StlWriter struct {
    binary      bool
    mergeMesh   bool
}

func NewStlWriter() *StlWriter
func (w *StlWriter) SetBinary(binary bool)
func (w *StlWriter) Write(mesh *Mesh, writer io.Writer) error
func (w *StlWriter) WriteFile(mesh *Mesh, filepath string) error

// GLTF Writer
type GltfWriter struct {
    binary       bool
    embedImages  bool
    draco        bool
}

func NewGltfWriter() *GltfWriter
func (w *GltfWriter) SetBinary(binary bool)
func (w *GltfWriter) SetEmbedImages(embed bool)
func (w *GltfWriter) Write(mesh *Mesh, writer io.Writer) error
func (w *GltfWriter) WriteFile(mesh *Mesh, filepath string) error

// OBJ Writer
type ObjWriter struct {
    separateMtl bool
}

func NewObjWriter() *ObjWriter
func (w *ObjWriter) Write(mesh *Mesh, writer io.Writer) error
func (w *ObjWriter) WriteFile(mesh *Mesh, filepath string) error
```

## 9. 使用示例

### 9.1 3D 打印 - 高程 + GPX 路径突出显示

```go
package main

import (
    static "github.com/flywave/go-static-mesh"
    "github.com/flywave/go-static-mesh/draw"
    "github.com/flywave/go-static-mesh/mesh"
    "github.com/flywave/go-tin"
)

func main() {
    // 创建 Builder
    builder := mesh.NewBuilder()
    
    // 设置 TIN 生成算法
    builder.SetTINGenerator(&tin.Algorithm{})
    
    // 添加高程数据（RasterProvider 和 TinMeshProvider 只能选一个）
    rasterProvider := static.NewGeoTIFFRasterProvider("elevation.tif")
    builder.SetRasterProvider(rasterProvider)
    
    // 加载 GPX 文件
    gpxProvider := static.NewGPXProvider("track.gpx")
    paths := gpxProvider.GetPaths()
    
    // 将路径添加为 MeshObject（突出显示）
    for _, path := range paths {
        builder.AddPath(path)
    }
    
    // 设置突出显示参数
    builder.SetExtrudeGeoData(true, 5.0) // 5mm 高度
    
    // 设置范围（必须）
    bounds := vec2d.Rect{{minLat, minLng}, {maxLat, maxLng}}
    srs := geo.NewProj(4326)
    builder.SetBounds(bounds, srs)
    
    // 设置 zoom（可选）
    // 方式 1: 手动设置
    // builder.SetZoom(15)
    
    // 方式 2: 自动判断（根据范围和分辨率）
    builder.SetAutoZoomRange(10, 18)
    
    // 设置 Tile 获取失败处理（可选）
    handler := TileErrorHandler{
        SkipMissing: true,
        MaxRetries:  3,
        Timeout:     30 * time.Second,
    }
    builder.SetTileErrorHandler(handler)
    
    // 设置闭合参数（3D 打印必须）
    builder.SetCloseMesh(true, 5.0)  // 闭合模型，基础厚度 5mm
    
    // 构建 3D 打印模型（闭合体积）
    mesh, err := builder.BuildForPrint()
    if err != nil {
        panic(err)
    }
    
    // 导出 STL
    writer := mesh.NewStlWriter()
    writer.SetBinary(true)
    writer.WriteFile(mesh, "output.stl")
}
```

### 9.2 展示模型 - 高程 + 卫星图 + 路径贴图

```go
package main

import (
    static "github.com/flywave/go-static-mesh"
    "github.com/flywave/go-static-mesh/draw"
    "github.com/flywave/go-static-mesh/mesh"
    "github.com/flywave/go-tin"
)

func main() {
    // 创建 Builder
    builder := mesh.NewBuilder()
    
    // 设置 TIN 生成算法
    builder.SetTINGenerator(&tin.Algorithm{})
    
    // 添加高程数据（RasterProvider 和 TinMeshProvider 只能选一个）
    elevationProvider := static.NewGeoTIFFRasterProvider("elevation.tif")
    builder.SetRasterProvider(elevationProvider)
    
    // 添加卫星影像（使用 NewTileProvider）
    imageryProvider := static.NewTileProvider(
        "https://tile.openstreetmap.org/{z}/{x}/{y}.png",
        static.TileProviderModeXYZ,
    )
    builder.AddImageryProvider(imageryProvider)
    
    // 加载 GPX 文件
    gpxProvider := static.NewGPXProvider("track.gpx")
    paths := gpxProvider.GetPaths()
    
    // 创建绘制上下文生成纹理
    ctx := draw.NewContext()
    ctx.SetTileProvider(imageryProvider)
    for _, path := range paths {
        ctx.AddPath(path)
    }
    
    // 设置范围（必须）
    bounds := vec2d.Rect{{minLat, minLng}, {maxLat, maxLng}}
    srs := geo.NewProj(4326)
    
    // 设置边界和生成纹理
    ctx.SetBoundingBox(bounds, srs)
    ctx.SetSize(4096, 4096)
    
    texture, err := ctx.Render()
    if err != nil {
        panic(err)
    }
    
    // 设置范围和构建带纹理的 Mesh
    builder.SetBounds(bounds, srs)
    builder.SetTexture(texture)
    
    // 设置 zoom（可选）
    // 方式 1: 手动设置
    // builder.SetZoom(15)
    
    // 方式 2: 自动判断（根据范围和分辨率）
    builder.SetAutoZoomRange(10, 18)
    
    // 设置 Tile 获取失败处理（可选）
    handler := TileErrorHandler{
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
    
    // 导出 GLB
    writer := mesh.NewGltfWriter()
    writer.SetBinary(true)
    writer.SetEmbedImages(true)
    writer.WriteFile(mesh, "output.glb")
}
```

### 9.3 TIN Mesh + 3D 模型

```go
package main

import (
    static "github.com/flywave/go-static-mesh"
    "github.com/flywave/go-static-mesh/mesh"
)

func main() {
    builder := mesh.NewBuilder()
    
    // 添加 TIN Mesh 数据源（与 RasterProvider 互斥）
    tinProvider := static.NewCesiumQuantizedMeshProvider("https://terrain.example.com")
    builder.SetTinMeshProvider(tinProvider)
    
    // 添加 3D 模型
    modelProvider := static.NewGeoJSON3DProvider("buildings.geojson")
    builder.AddModel3DProvider(modelProvider)
    
    // 设置边界
    bounds := tinProvider.Bounds()
    builder.SetBounds(bounds, tinProvider.Srs())
    
    // 构建展示模型（面片，不闭合）
    mesh, err := builder.BuildForDisplay()
    if err != nil {
        panic(err)
    }
    
    // 导出 OBJ
    writer := mesh.NewObjWriter()
    writer.WriteFile(mesh, "output.obj")
}
```
