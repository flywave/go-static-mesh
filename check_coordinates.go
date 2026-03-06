package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"

	"github.com/flywave/gltf"
)

func main() {
	filePath := "examples/terrain_with_texture/output/terrain_textured.gltf"

	doc, err := gltf.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open GLTF file: %v", err)
	}

	fmt.Println("========================================")
	fmt.Println("Vertex Coordinate Analysis")
	fmt.Println("========================================")
	fmt.Println()

	// Get position accessor
	positionAccessor := doc.Accessors[0]
	positionBufferView := doc.BufferViews[*positionAccessor.BufferView]
	buffer := doc.Buffers[positionBufferView.Buffer]

	fmt.Printf("Position Accessor:\n")
	fmt.Printf("  Count: %d vertices\n", positionAccessor.Count)
	fmt.Printf("  Min: %v\n", positionAccessor.Min)
	fmt.Printf("  Max: %v\n", positionAccessor.Max)
	fmt.Println()

	// Read first 20 vertices
	offset := uint64(positionBufferView.ByteOffset + positionAccessor.ByteOffset)

	fmt.Println("First 20 vertices (X, Y, Z):")
	fmt.Println("----------------------------------------")

	minX, minY, minZ := float32(math.Inf(1)), float32(math.Inf(1)), float32(math.Inf(1))
	maxX, maxY, maxZ := float32(math.Inf(-1)), float32(math.Inf(-1)), float32(math.Inf(-1))

	for i := uint32(0); i < 20 && i < positionAccessor.Count; i++ {
		x := math.Float32frombits(binary.LittleEndian.Uint32(buffer.Data[offset+uint64(i)*12:]))
		y := math.Float32frombits(binary.LittleEndian.Uint32(buffer.Data[offset+uint64(i)*12+4:]))
		z := math.Float32frombits(binary.LittleEndian.Uint32(buffer.Data[offset+uint64(i)*12+8:]))

		fmt.Printf("Vertex %4d: (%.6f, %.6f, %.6f)\n", i, x, y, z)

		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if z < minZ {
			minZ = z
		}
		if x > maxX {
			maxX = x
		}
		if y > maxY {
			maxY = y
		}
		if z > maxZ {
			maxZ = z
		}
	}

	fmt.Println()
	fmt.Println("Analyzing coordinate ranges...")

	// Analyze all vertices
	for i := uint32(0); i < positionAccessor.Count; i++ {
		x := math.Float32frombits(binary.LittleEndian.Uint32(buffer.Data[offset+uint64(i)*12:]))
		y := math.Float32frombits(binary.LittleEndian.Uint32(buffer.Data[offset+uint64(i)*12+4:]))
		z := math.Float32frombits(binary.LittleEndian.Uint32(buffer.Data[offset+uint64(i)*12+8:]))

		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if z < minZ {
			minZ = z
		}
		if x > maxX {
			maxX = x
		}
		if y > maxY {
			maxY = y
		}
		if z > maxZ {
			maxZ = z
		}
	}

	fmt.Println()
	fmt.Println("Coordinate Ranges:")
	fmt.Printf("  X: [%.6f, %.6f] (range: %.6f)\n", minX, maxX, maxX-minX)
	fmt.Printf("  Y: [%.6f, %.6f] (range: %.6f)\n", minY, maxY, maxY-minY)
	fmt.Printf("  Z: [%.6f, %.6f] (range: %.6f)\n", minZ, maxZ, maxZ-minZ)
	fmt.Println()

	// Check if coordinates look like lat/lon
	if minX >= -180 && maxX <= 180 && minY >= -90 && maxY <= 90 {
		fmt.Println("⚠️  WARNING: Coordinates appear to be in geographic (lat/lon) format!")
		fmt.Println("   This will cause issues in 3D viewers.")
		fmt.Println()
		fmt.Println("Diagnosis:")
		if maxX-minX < 1.0 {
			fmt.Printf("  X range is very small (%.6f), likely longitude\n", maxX-minX)
		}
		if maxY-minY < 1.0 {
			fmt.Printf("  Y range is very small (%.6f), likely latitude\n", maxY-minY)
		}
		fmt.Println()
		fmt.Println("Solution:")
		fmt.Println("  1. Convert geographic coordinates to local Cartesian coordinates")
		fmt.Println("  2. Center the mesh around origin (0, 0, 0)")
		fmt.Println("  3. Scale appropriately for the scene")
	}
}
