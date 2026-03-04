# go-static-mesh - Agent Guidelines

## Build Commands

```bash
# Build all packages
go build ./...

# Build specific package
go build ./mesh
go build ./draw
go build ./tile

# Run all tests
go test ./...

# Run single test
go test -run TestFunctionName ./package
go test -run TestBuilderSetBounds ./mesh/builder

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...

# Run tests in specific package
go test ./mesh/builder -v
go test ./tile -run TestGPX

# Format code
go fmt ./...

# Vet code
go vet ./...

# Tidy dependencies
go mod tidy

# Static analysis
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...
```

## Code Style Guidelines

### Package Structure
- `draw/`: Drawing operations and map rendering (markers, paths, areas, circles)
  - `context.go`: Main rendering context with tile providers
  - `marker.go`, `path.go`, `area.go`, `circle.go`: Map object types
- `mesh/`: Mesh building and writing (GLTF, STL, OBJ formats)
  - `mesh.go`: Core mesh data structures (vertices, normals, UVs, indices)
  - `cache.go`: Tile caching with LRU eviction and TTL
  - `errors.go`: Custom error types for mesh operations
  - `builder/`: Mesh builder with TIN generation
  - `writer/`: Format-specific writers (gltf.go, stl.go, obj.go)
  - `bsp/`: Binary space partitioning for mesh operations
  - `extruder/`: Extrusion operations for geo data (GPX, 3D models)
- `tile/`: Tile providers and data sources
  - `provider.go`: Tile provider interfaces and configuration
  - `geo.go`: GPX provider for geographic data
  - `dem.go`: Digital elevation model providers
  - `imagery.go`: Imagery tile providers
  - `model3d_provider.go`: 3D model tile providers
  - `tinmesh.go`: TIN mesh generation from elevation data
- `utils/`: Shared utility functions

### Import Organization
```go
import (
    "errors"           // Standard library imports first, alphabetically
    "image"
    "io"
    "math"
    "sync"

    "github.com/flywave/go-geo"                    // External packages second, alphabetically
    "github.com/flywave/gg"
    draw "github.com/flywave/go-static-mesh/draw" // Local package imports with aliases
    "github.com/flywave/go-static-mesh/mesh"
    tile "github.com/flywave/go-static-mesh/tile"
    vec2d "github.com/flywave/go3d/float64/vec2"  // Aliased imports for clarity
    vec3d "github.com/flywave/go3d/float64/vec3"
    _ "github.com/flywave/gltf"                   // Blank imports for side effects only
)
```

### Naming Conventions
- **Exported types/functions**: PascalCase (`Context`, `NewBuilder`, `TileProvider`)
- **Private types/functions**: camelCase (`determineBounds`, `renderLayer`, `newTransformer`)
- **Interfaces**: Descriptive names, often ending in "er" (`TileProvider`, `TileFetcher`, `MapObject`, `TINGenerator`)
- **Constructors**: Prefix with "New" (`NewContext`, `NewBuilder`, `NewMemoryTileCache`)
- **Constants**: PascalCase (`TileProviderModeXYZ`, `TileProviderModeTMS`)
- **Receiver names**: Single lowercase letter (`m`, `t`, `b`, `p`)
- **Error types**: Suffix with "Error" (`MeshError`, `BuildError`)
- **Error constructors**: Prefix with "New" + error type (`NewNetworkError`, `NewTileError`)

### Type Definitions
```go
type Context struct {
    width               int                 // Private fields: camelCase
    height              int
    tileProvider        ImageTileProvider   // Interface types
    boundingBox         *vec2d.Rect         // Pointer to struct
    objects             []MapObject         // Slice of interfaces
    grid                *geo.TileGrid
}

type Mesh struct {
    Vertices  []vec3d.T
    Normals   []vec3d.T
    UVs       []vec2d.T
    Indices   []uint32
    Materials []Material
    Texture   image.Image
    Bounds    vec2d.Rect
    Srs       geo.Proj
    TinMesh   interface{}
}

type Material struct {
    Name              string
    Diffuse           color.Color
    Specular          color.Color
    Shininess         float32
    Alpha             float32
    Metalness         float32
    Roughness         float32
    Ao                float32
    Emissive          color.Color
    EmissiveIntensity float32
    NormalStrength    float32
    Displacement      float32
}

// Use pointer receivers for methods that modify state
func (m *Context) SetSize(width, height int) {
    m.width = width
    m.height = height
}

func (b *Builder) SetBounds(bounds vec2d.Rect, srs geo.Proj) {
    b.bounds = bounds
    b.srs = srs
}

// Use value receivers for methods that don't modify state
func (t *Tile) GetImage() image.Image {
    return t.image
}

func (m *Mesh) VertexCount() int {
    return len(m.Vertices)
}
```

### Interface Definitions
```go
type TileProvider interface {
    Attribution() string
    Grid() *geo.TileGrid
    Bounds() vec2d.Rect
    Srs() geo.Proj
}

type ImageTileProvider interface {
    TileProvider
    GetImageTile(coord [3]int) (image.Image, error)
}

type TileFetcher interface {
    Fetch(coord [3]int) ([]byte, error)
}

type Writer interface {
    Write(mesh *mesh.Mesh, writer io.Writer) error
}

type MapObject interface {
    Bounds() vec2d.Rect
    SrsProj() geo.Proj
    ExtraMarginPixels() (float64, float64, float64, float64)
    Draw(gc *gg.Context, trans *Transformer)
}
```

### Error Handling
- Return errors explicitly as last return value
- Use custom error types for domain-specific errors
- Use `fmt.Errorf` for error messages with context
- Check errors immediately after function calls
- Use `log.Printf` for non-fatal errors, not panics
- Support error wrapping with `Unwrap()` method

```go
// Custom error types with context
type MeshError struct {
    Type       ErrorType
    Operation  string
    TileCoord  [3]int
    Underlying error
    Context    map[string]interface{}
}

func (e *MeshError) Error() string {
    msg := fmt.Sprintf("[%s] %s", e.type, e.Operation)
    if e.TileCoord != [3]int{} {
        msg += fmt.Sprintf(" (tile %d/%d/%d)", e.TileCoord[2], e.TileCoord[0], e.TileCoord[1])
    }
    if e.Underlying != nil {
        msg += fmt.Sprintf(": %v", e.Underlying)
    }
    return msg
}

func (e *MeshError) Unwrap() error {
    return e.Underlying
}

// Error constructors
func NewNetworkError(op string, tile [3]int, err error) *MeshError {
    return &MeshError{
        Type:       ErrTypeNetwork,
        Operation:  op,
        TileCoord:  tile,
        Underlying: err,
    }
}

// Usage in functions
func (m *Context) Render() (image.Image, error) {
    zoom, center, srs, err := m.determineZoomCenter()
    if err != nil {
        return nil, err
    }
    
    trans := newTransformer(m.width, m.height, zoom, center, srs, m.grid)
    if err := m.renderLayer(gc, zoom, trans, provider); err != nil {
        return nil, err
    }
    
    return croppedImg, nil
}

func (b *Builder) Build() (*mesh.Mesh, error) {
    if b.rasterProvider == nil && b.tinMeshProvider == nil {
        return nil, ErrNoProviderSet
    }
    
    if b.bounds.Area() == 0 {
        return nil, ErrBoundsNotSet
    }
    
    tinMesh, err := b.generateTIN()
    if err != nil {
        return nil, NewTINError("generate", err)
    }
    
    return tinMesh, nil
}
```

### Concurrency
- Use `sync.WaitGroup` for parallel operations
- Use channels for communication between goroutines
- Close channels when done sending
- Use `defer wg.Done()` for WaitGroup cleanup
- Protect shared state with mutexes

```go
func (m *Context) renderLayer(gc *gg.Context, zoom int, trans *Transformer, provider ImageTileProvider) error {
    var wg sync.WaitGroup
    tiles := (1 << uint(zoom))
    fetchedTiles := make(chan *tile)
    
    go func() {
        for xx := 0; xx < trans.tCountX; xx++ {
            for yy := 0; yy < trans.tCountY; yy++ {
                wg.Add(1)
                coord := [3]int{x, y, zoom}
                go func(wg *sync.WaitGroup, c [3]int, xx, yy int) {
                    defer wg.Done()
                    if img, err := provider.GetImageTile(c); err == nil {
                        t := &tile{image: img, offx: xx * tileSize, offy: yy * tileSize}
                        fetchedTiles <- t
                    } else {
                        log.Printf("Error downloading tile: %s (Ignored)", err)
                    }
                }(&wg, coord, xx, yy)
            }
        }
        wg.Wait()
        close(fetchedTiles)
    }()
    
    for tile := range fetchedTiles {
        gc.DrawImage(tile.GetImage(), tile.offx, tile.offy)
    }
    
    return nil
}

// Thread-safe cache implementation
type MemoryTileCache struct {
    mu       sync.RWMutex
    tiles    map[[3]int]*cacheEntry
    maxSize  int
    ttl      time.Duration
}

func (c *MemoryTileCache) Get(key [3]int) (*Tile, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    entry, exists := c.tiles[key]
    if !exists {
        return nil, false
    }
    return entry.tile, true
}

func (c *MemoryTileCache) Put(key [3]int, tile *Tile) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.tiles[key] = &cacheEntry{tile: tile, accessed: time.Now()}
}
```

### Testing
- Place tests in same package with `_test.go` suffix
- Use table-driven tests for multiple test cases
- Test both success and error paths
- Use `t.Fatal` for setup failures, `t.Error` for test failures
- Test concurrent operations with goroutines

```go
func TestBuilderSetBounds(t *testing.T) {
    builder := NewBuilder()
    bounds := vec2d.Rect{Min: vec2d.T{30.0, 120.0}, Max: vec2d.T{31.0, 121.0}}
    srs := geo.NewProj(4326)
    
    builder.SetBounds(bounds, srs)
    
    if builder.bounds != bounds {
        t.Errorf("bounds not set correctly")
    }
}

func TestMemoryTileCache_LRUEviction(t *testing.T) {
    cache := NewMemoryTileCache(3, 0)
    
    for i := 0; i < 3; i++ {
        tile := &Tile{Coord: [3]int{i, i, 0}, Zoom: 0, Data: []float64{float64(i)}}
        cache.Put([3]int{i, i, 0}, tile)
    }
    
    stats := cache.Stats()
    if stats.Size != 3 {
        t.Fatalf("expected cache size 3, got %d", stats.Size)
    }
    
    tile4 := &Tile{Coord: [3]int{3, 3, 0}, Zoom: 0, Data: []float64{4.0}}
    cache.Put([3]int{3, 3, 0}, tile4)
    
    if stats.Evictions != 1 {
        t.Fatalf("expected 1 eviction, got %d", stats.Evictions)
    }
}

func TestGPXProviderGetPaths(t *testing.T) {
    gpxData := `<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1">
  <trk>
    <trkseg>
      <trkpt lat="35.6895" lon="139.6917"></trkpt>
      <trkpt lat="35.6896" lon="139.6918"></trkpt>
    </trkseg>
  </trk>
</gpx>`
    
    provider, err := NewGPXProviderFromReader(strings.NewReader(gpxData))
    if err != nil {
        t.Fatalf("Failed to create GPX provider: %v", err)
    }
    
    paths := provider.GetPaths()
    if len(paths) != 1 {
        t.Errorf("Expected 1 path, got %d", len(paths))
    }
}

func TestMemoryTileCache_ConcurrentAccess(t *testing.T) {
    cache := NewMemoryTileCache(1000, 0)
    var wg sync.WaitGroup
    
    numGoroutines := 10
    numOperations := 20
    
    for g := 0; g < numGoroutines; g++ {
        wg.Add(1)
        go func(goroutineID int) {
            defer wg.Done()
            for i := 0; i < numOperations; i++ {
                key := [3]int{goroutineID, i, 0}
                tile := &Tile{Coord: key, Zoom: 0, Data: []float64{float64(i)}}
                cache.Put(key, tile)
                _, found := cache.Get(key)
                if !found {
                    t.Errorf("expected to find tile for key %v", key)
                }
            }
        }(g)
    }
    
    wg.Wait()
}
```

### Formatting
- Use `gofmt` for formatting (standard Go style)
- Use tabs for indentation
- Limit line length to ~100-120 characters where practical
- Blank line between top-level declarations
- No comments unless necessary for clarity or public API documentation
- Group related constants and variables

```go
type TileProviderMode int

const (
    TileProviderModeXYZ TileProviderMode = iota
    TileProviderModeTMS
)

const (
    Radian = math.Pi / 180.0
    Degree = 180.0 / math.Pi
)

var (
    ErrNoProviderSet           = errors.New("no raster or tin mesh provider set")
    ErrBoundsNotSet            = errors.New("bounds not set properly")
    ErrNoTINGenerated          = errors.New("failed to generate TIN mesh")
)
```

### Package Documentation
- Add package doc comments only when package has complex behavior
- Keep documentation concise and focused
- Exported types/functions should be self-documenting via names
- Use example functions for complex APIs

### Dependencies
- All external dependencies listed in go.mod
- Use local replacements for development (see go.mod replace directives)
- Keep dependencies minimal and purposeful
- Primary dependencies:
  - `github.com/flywave/go-geo`: Geographic coordinate systems and tile grids
  - `github.com/flywave/go3d`: 3D vector math (vec2d, vec3d)
  - `github.com/flywave/gg`: 2D graphics rendering
  - `github.com/flywave/gltf`: GLTF format support
  - `github.com/flywave/go-tin`: TIN mesh generation
  - `github.com/flywave/go-stl`: STL format support

### Architecture Patterns
- **Builder Pattern**: Use `NewBuilder()` constructor with fluent setter methods
- **Provider Pattern**: Use interfaces for different data sources (TileProvider, ImageTileProvider)
- **Cache Pattern**: LRU cache with TTL support for tile data
- **Writer Pattern**: Interface-based writers for different output formats (GLTF, STL, OBJ)
- **Error Wrapping**: Use custom error types with `Unwrap()` for error chain

### Mesh Generation Pipeline
1. **Tile Fetching**: Retrieve elevation/imagery tiles from providers
2. **TIN Generation**: Create triangulated irregular network from elevation data
3. **Mesh Building**: Convert TIN to mesh with vertices, normals, UVs
4. **Texture Mapping**: Apply imagery textures to mesh
5. **Extrusion**: Optionally extrude geo data (GPX paths, 3D models)
6. **Mesh Closing**: Add base and sides to create solid mesh
7. **Format Export**: Write to GLTF, STL, or OBJ format

### Coordinate Systems
- Use `geo.Proj` for coordinate reference systems
- Common SRS: EPSG:4326 (WGS84 lat/lon), EPSG:3857 (Web Mercator)
- Transform coordinates between SRS as needed
- Tile coordinates: [x, y, zoom] where zoom is the level
