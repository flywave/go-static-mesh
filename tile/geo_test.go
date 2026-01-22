package tile

import (
	"strings"
	"testing"
)

func TestNewGPXProviderFromReader(t *testing.T) {
	gpxData := `<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1">
  <trk>
    <name>Test Track</name>
    <trkseg>
      <trkpt lat="35.6895" lon="139.6917"></trkpt>
      <trkpt lat="35.6896" lon="139.6918"></trkpt>
      <trkpt lat="35.6897" lon="139.6919"></trkpt>
    </trkseg>
  </trk>
</gpx>`

	provider, err := NewGPXProviderFromReader(strings.NewReader(gpxData))
	if err != nil {
		t.Fatalf("Failed to create GPX provider: %v", err)
	}

	if provider == nil {
		t.Fatal("Expected non-nil provider")
	}
}

func TestGPXProviderGetPaths(t *testing.T) {
	gpxData := `<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1">
  <trk>
    <name>Test Track</name>
    <trkseg>
      <trkpt lat="35.6895" lon="139.6917"></trkpt>
      <trkpt lat="35.6896" lon="139.6918"></trkpt>
      <trkpt lat="35.6897" lon="139.6919"></trkpt>
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

	path := paths[0]
	if len(path.Positions) != 3 {
		t.Errorf("Expected 3 positions, got %d", len(path.Positions))
	}

	if path.Positions[0][0] != 35.6895 || path.Positions[0][1] != 139.6917 {
		t.Errorf("Unexpected first position: %v", path.Positions[0])
	}
}

func TestGPXProviderGetMultipleTracks(t *testing.T) {
	gpxData := `<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1">
  <trk>
    <name>Track 1</name>
    <trkseg>
      <trkpt lat="35.6895" lon="139.6917"></trkpt>
      <trkpt lat="35.6896" lon="139.6918"></trkpt>
    </trkseg>
  </trk>
  <trk>
    <name>Track 2</name>
    <trkseg>
      <trkpt lat="35.7000" lon="139.7000"></trkpt>
      <trkpt lat="35.7001" lon="139.7001"></trkpt>
    </trkseg>
  </trk>
</gpx>`

	provider, err := NewGPXProviderFromReader(strings.NewReader(gpxData))
	if err != nil {
		t.Fatalf("Failed to create GPX provider: %v", err)
	}

	paths := provider.GetPaths()
	if len(paths) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(paths))
	}
}

func TestGPXProviderGetAreas(t *testing.T) {
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

	areas := provider.GetAreas()
	if areas != nil {
		t.Error("Expected nil areas for GPX provider")
	}
}

func TestGPXProviderGetPoints(t *testing.T) {
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

	points := provider.GetPoints()
	if points != nil {
		t.Error("Expected nil points for GPX provider")
	}
}

func TestGPXProviderEmptyGPX(t *testing.T) {
	gpxData := `<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1">
</gpx>`

	provider, err := NewGPXProviderFromReader(strings.NewReader(gpxData))
	if err != nil {
		t.Fatalf("Failed to create GPX provider: %v", err)
	}

	paths := provider.GetPaths()
	if len(paths) != 0 {
		t.Errorf("Expected 0 paths, got %d", len(paths))
	}
}

func TestGPXProviderMultipleSegments(t *testing.T) {
	gpxData := `<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1">
  <trk>
    <trkseg>
      <trkpt lat="35.6895" lon="139.6917"></trkpt>
      <trkpt lat="35.6896" lon="139.6918"></trkpt>
    </trkseg>
    <trkseg>
      <trkpt lat="35.7000" lon="139.7000"></trkpt>
      <trkpt lat="35.7001" lon="139.7001"></trkpt>
    </trkseg>
  </trk>
</gpx>`

	provider, err := NewGPXProviderFromReader(strings.NewReader(gpxData))
	if err != nil {
		t.Fatalf("Failed to create GPX provider: %v", err)
	}

	paths := provider.GetPaths()
	if len(paths) != 2 {
		t.Errorf("Expected 2 paths (one per segment), got %d", len(paths))
	}
}
