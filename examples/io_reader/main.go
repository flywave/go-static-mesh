package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/flywave/go-static-mesh/tile"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <geotiff-file.tif>")
		os.Exit(1)
	}

	filename := os.Args[1]

	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	r := bytes.NewReader(data)

	provider, err := tile.NewGeoTIFFRasterProviderFromReader(r)
	if err != nil {
		fmt.Printf("Error creating provider from reader: %v\n", err)
		os.Exit(1)
	}
	defer provider.Close()

	grid := provider.GetElevationGrid()
	if grid == nil {
		fmt.Println("Error: elevation grid is nil")
		os.Exit(1)
	}

	fmt.Printf("Grid size: %d x %d\n", grid.Width, grid.Height)
	fmt.Printf("Bounds: [%.6f, %.6f] to [%.6f, %.6f]\n",
		grid.Bounds.Min[0], grid.Bounds.Min[1],
		grid.Bounds.Max[0], grid.Bounds.Max[1])
	fmt.Printf("Cell size: %.6f\n", grid.CellSize)
	fmt.Printf("No data value: %.2f\n", grid.NoData)

	fmt.Printf("\nMin height: %.2f\n", grid.GetMinHeight())
	fmt.Printf("Max height: %.2f\n", grid.GetMaxHeight())
}
