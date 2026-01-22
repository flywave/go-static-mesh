package main

import (
	"fmt"
	"log"

	"github.com/flywave/go-geo"
	"github.com/flywave/go-static-mesh/mesh"
	"github.com/flywave/go-static-mesh/static"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

func main() {
	// 示例 1: 使用自定义配置的 CesiumQuantizedMeshProvider
	config := static.NewDecoderConfig()
	config.VertexNormals = true
	config.WaterMask = false
	config.Metadata = false
	config.MaxRetries = 5
	config.Timeout = 60

	provider := static.NewCesiumQuantizedMeshProviderWithConfig(
		"https://assets.ion.cesium.com/1/terrain/{z}/{x}/{y}.terrain",
		config,
	)

	// 示例 2: 使用 CesiumQuantizedMeshProvider 构建 Mesh
	builder := mesh.NewBuilder()
	builder.SetTinMeshProvider(provider)

	// 设置范围（旧金山湾区）
	bounds := vec2d.Rect{
		Min: vec2d.T{37.4, -122.5},
		Max: vec2d.T{37.8, -122.0},
	}
	srs := geo.NewProj(4326)
	builder.SetBounds(bounds, srs)

	// 构建 Mesh
	fmt.Println("Building mesh from Cesium Quantized Mesh...")
	terrainMesh, err := builder.BuildForDisplay()
	if err != nil {
		log.Fatalf("Failed to build mesh: %v", err)
	}

	fmt.Printf("Mesh built successfully:\n")
	fmt.Printf("  Vertices: %d\n", len(terrainMesh.Vertices))
	fmt.Printf("  Triangles: %d\n", len(terrainMesh.Indices)/3)
	fmt.Printf("  Bounds: %v\n", terrainMesh.Bounds)

	// 示例 3: 缓存管理
	fmt.Println("\nCache management:")
	fmt.Printf("  Cache size: %d\n", provider.GetCacheSize())

	provider.SetMaxCacheSize(50)
	fmt.Printf("  Max cache size set to: %d\n", 50)

	provider.ClearCache()
	fmt.Printf("  Cache cleared, size: %d\n", provider.GetCacheSize())

	// 示例 4: 查询高度和法线
	fmt.Println("\nQuerying elevation and normals:")
	lat := 37.7749
	lng := -122.4194

	height := provider.GetHeight(lng, lat)
	fmt.Printf("  Elevation at (%.4f, %.4f): %.2f meters\n", lat, lng, height)

	normal := provider.GetNormal(lng, lat)
	fmt.Printf("  Normal at (%.4f, %.4f): (%.2f, %.2f, %.2f)\n", lat, lng,
		normal[0], normal[1], normal[2])

	// 示例 5: 使用新的错误处理
	fmt.Println("\nError handling example:")
	errorHandler := mesh.NewErrorHandler(mesh.LenientErrorPolicy())

	// 模拟错误
	mockErr := mesh.NewNetworkError("FetchTile", [3]int{10, 20, 15},
		fmt.Errorf("connection timeout"))

	strategy := errorHandler.HandleError(mockErr, map[string]interface{}{
		"retry": 0,
	})

	fmt.Printf("  Error type: %s\n", mockErr.Type)
	fmt.Printf("  Recovery strategy: %s\n", strategy)
	fmt.Printf("  Total errors: %d\n", errorHandler.GetTotalErrorCount())
	fmt.Printf("  Error stats: %v\n", errorHandler.GetStats())

	fmt.Println("\nExamples completed successfully!")
}
