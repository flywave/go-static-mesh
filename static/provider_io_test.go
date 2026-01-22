package static

import (
	"bytes"
	"os"
	"testing"
)

func TestGeoTIFFRasterProviderFromReader(t *testing.T) {
	testData, err := os.ReadFile("test_data/elevation.tif")
	if err != nil {
		t.Skip("test data file not found")
	}

	r := bytes.NewReader(testData)
	provider, err := NewGeoTIFFRasterProviderFromReader(r)
	if err != nil {
		t.Fatalf("failed to create provider from reader: %v", err)
	}
	defer provider.Close()

	if provider.tempFile == nil {
		t.Fatal("temp file not created")
	}

	grid := provider.GetElevationGrid()
	if grid == nil {
		t.Fatal("failed to get elevation grid")
	}

	if grid.Width <= 0 || grid.Height <= 0 {
		t.Fatalf("invalid grid size: %dx%d", grid.Width, grid.Height)
	}
}

func TestGeoTIFFImageryProviderFromReader(t *testing.T) {
	testData, err := os.ReadFile("test_data/imagery.tif")
	if err != nil {
		t.Skip("test data file not found")
	}

	r := bytes.NewReader(testData)
	provider, err := NewGeoTIFFImageryProviderFromReader(r)
	if err != nil {
		t.Fatalf("failed to create provider from reader: %v", err)
	}
	defer provider.Close()

	if provider.tempFile == nil {
		t.Fatal("temp file not created")
	}

	img, err := provider.GetImageTile([3]int{0, 0, 0})
	if err != nil {
		t.Fatalf("failed to get image tile: %v", err)
	}

	if img == nil {
		t.Fatal("image is nil")
	}

	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		t.Fatalf("invalid image bounds: %v", bounds)
	}
}

func TestGeoTIFFRasterProviderClose(t *testing.T) {
	testData, err := os.ReadFile("test_data/elevation.tif")
	if err != nil {
		t.Skip("test data file not found")
	}

	r := bytes.NewReader(testData)
	provider, err := NewGeoTIFFRasterProviderFromReader(r)
	if err != nil {
		t.Fatalf("failed to create provider from reader: %v", err)
	}

	tempFileName := provider.tempFile.Name()
	err = provider.Close()
	if err != nil {
		t.Fatalf("failed to close provider: %v", err)
	}

	_, err = os.Stat(tempFileName)
	if err == nil {
		t.Fatal("temp file should be deleted after close")
	}
}

func TestGeoTIFFImageryProviderClose(t *testing.T) {
	testData, err := os.ReadFile("test_data/imagery.tif")
	if err != nil {
		t.Skip("test data file not found")
	}

	r := bytes.NewReader(testData)
	provider, err := NewGeoTIFFImageryProviderFromReader(r)
	if err != nil {
		t.Fatalf("failed to create provider from reader: %v", err)
	}

	tempFileName := provider.tempFile.Name()
	err = provider.Close()
	if err != nil {
		t.Fatalf("failed to close provider: %v", err)
	}

	_, err = os.Stat(tempFileName)
	if err == nil {
		t.Fatal("temp file should be deleted after close")
	}
}
