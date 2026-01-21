# go-static-mesh - Agent Guidelines

## Build Commands

```bash
# Build all packages
go build ./...

# Build specific package
go build ./mesh

# Run all tests
go test ./...

# Run single test
go test -run TestFunctionName ./package

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...

# Format code
go fmt ./...

# Vet code
go vet ./...

# Tidy dependencies
go mod tidy
```

## Code Style Guidelines

### Package Structure
- `static/`: Root package with tile provider interfaces
- `mesh/`: Mesh building and writing (GLTF, STL, OBJ formats)
- `draw/`: Drawing operations and map rendering
- `raster/`: Raster operations
- `utils/`: Shared utility functions

### Import Organization
```go
import (
    "errors"           // Standard library imports first
    "image"
    "math"

    vec2d "github.com/flywave/go3d/float64/vec2"  // External packages
    "github.com/flywave/gg"
    _ "github.com/flywave/gltf"  // Blank imports for side effects
)
```
- Standard library imports first, grouped alphabetically
- Third-party imports second, grouped alphabetically
- Use alias imports for clarity (e.g., `vec2d`)
- Blank imports for side effects only

### Naming Conventions
- **Exported types/functions**: PascalCase (`Context`, `NewMarker`, `Draw`)
- **Private types/functions**: camelCase (`determineBounds`, `renderLayer`)
- **Interfaces**: Descriptive names, often ending in "er" (`TileProvider`, `TileFetcher`, `MapObject`)
- **Constructors**: Prefix with "New" (`NewContext`, `NewMarker`, `NewTileFetcher`)
- **Constants**: PascalCase (`Radian`, `Degree`)
- **Receiver names**: Single lowercase letter (`m`, `t`, `f`)

### Type Definitions
```go
type Context struct {
    width      int      // Private fields: camelCase
    tileProvider static.TileProvider  // Exported: PascalCase

// Use pointer receivers for methods that modify state
func (m *Context) SetSize(width, height int) {
    m.width = width
    m.height = height
}

// Use value receivers for methods that don't modify state
func (t *Tile) GetImage() image.Image {
    return t.image
}
```

### Error Handling
- Return errors explicitly as last return value
- Use `fmt.Errorf` for error messages with context
- Check errors immediately after function calls
- Use `log.Printf` for non-fatal errors, not panics

```go
func (m *Context) Render() (image.Image, error) {
    zoom, center, srs, err := m.determineZoomCenter()
    if err != nil {
        return nil, err  // Return error with context
    }
    // ... rest of implementation
}
```

### Concurrency
- Use `sync.WaitGroup` for parallel operations
- Use channels for communication between goroutines
- Close channels when done sending
- Use `defer wg.Done()` for WaitGroup cleanup

```go
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func(item T) {
        defer wg.Done()
        // Process item
    }(item)
}
wg.Wait()
```

### Formatting
- Use `gofmt` for formatting (standard Go style)
- No tabs/spaces debate - use tabs for indentation
- Limit line length to ~100-120 characters where practical
- Blank line between top-level declarations
- No comments unless necessary for clarity

### Package Documentation
- Add package doc comments only when package has complex behavior
- Keep documentation concise and focused
- Exported types/functions should be self-documenting via names

### Dependencies
- All external dependencies listed in go.mod
- Use local replacements for development (see go.mod replace directives)
- Keep dependencies minimal and purposeful
