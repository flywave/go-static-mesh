package static

import (
	"testing"

	"github.com/flywave/go-geo"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

func TestNewStaticModel3DProvider(t *testing.T) {
	provider := NewStaticModel3DProvider(nil)

	if provider == nil {
		t.Fatal("NewStaticModel3DProvider returned nil")
	}

	if provider.Grid() == nil {
		t.Error("Grid returned nil")
	}

	if provider.ModelCount() != 0 {
		t.Errorf("Expected model count 0, got %d", provider.ModelCount())
	}
}

func TestStaticModel3DProvider_AddModel(t *testing.T) {
	provider := NewStaticModel3DProvider(nil)

	model := &Model3D{
		ID:       "test1",
		Position: vec2d.T{37.7749, -122.4194},
		Rotation: vec3d.T{0, 0, 0},
		Scale:    vec3d.T{1, 1, 1},
	}

	provider.AddModel(model)

	if provider.ModelCount() != 1 {
		t.Errorf("Expected model count 1, got %d", provider.ModelCount())
	}

	retrieved, err := provider.GetModel("test1")
	if err != nil {
		t.Fatalf("Failed to get model: %v", err)
	}

	if retrieved.ID != "test1" {
		t.Errorf("Expected ID 'test1', got '%s'", retrieved.ID)
	}
}

func TestStaticModel3DProvider_GetModel(t *testing.T) {
	provider := NewStaticModel3DProvider(nil)

	model := &Model3D{
		ID:   "test1",
		Name: "Test Model",
	}

	provider.AddModel(model)

	_, err := provider.GetModel("test1")
	if err != nil {
		t.Errorf("Expected to find model, got error: %v", err)
	}

	_, err = provider.GetModel("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent model")
	}
}

func TestStaticModel3DProvider_GetModels(t *testing.T) {
	provider := NewStaticModel3DProvider(nil)

	model1 := &Model3D{
		ID:       "model1",
		Position: vec2d.T{37.5, -122.3},
	}

	model2 := &Model3D{
		ID:       "model2",
		Position: vec2d.T{38.0, -122.0},
	}

	model3 := &Model3D{
		ID:       "model3",
		Position: vec2d.T{40.0, -120.0},
	}

	provider.AddModel(model1)
	provider.AddModel(model2)
	provider.AddModel(model3)

	bounds := vec2d.Rect{
		Min: vec2d.T{37.0, -123.0},
		Max: vec2d.T{39.0, -121.0},
	}

	models, err := provider.GetModels(bounds)
	if err != nil {
		t.Fatalf("Failed to get models: %v", err)
	}

	if len(models) != 2 {
		t.Errorf("Expected 2 models in bounds, got %d", len(models))
	}
}

func TestStaticModel3DProvider_Clear(t *testing.T) {
	provider := NewStaticModel3DProvider(nil)

	model := &Model3D{
		ID: "test1",
	}

	provider.AddModel(model)

	if provider.ModelCount() != 1 {
		t.Errorf("Expected model count 1, got %d", provider.ModelCount())
	}

	provider.Clear()

	if provider.ModelCount() != 0 {
		t.Errorf("Expected model count 0 after clear, got %d", provider.ModelCount())
	}
}

func TestStaticModel3DProvider_ModelInBounds(t *testing.T) {
	provider := NewStaticModel3DProvider(geo.NewProj(4326))

	model := &Model3D{
		ID:       "test",
		Position: vec2d.T{37.7749, -122.4194},
	}

	provider.AddModel(model)

	bounds := vec2d.Rect{
		Min: vec2d.T{37.5, -122.5},
		Max: vec2d.T{38.0, -122.0},
	}

	if !provider.modelInBounds(model, bounds) {
		t.Error("Expected model to be in bounds")
	}

	bounds2 := vec2d.Rect{
		Min: vec2d.T{40.0, -120.0},
		Max: vec2d.T{41.0, -119.0},
	}

	if provider.modelInBounds(model, bounds2) {
		t.Error("Expected model to be outside bounds")
	}
}

func TestStaticModel3DProvider_Bounds(t *testing.T) {
	provider := NewStaticModel3DProvider(nil)
	bounds := provider.Bounds()

	if bounds.Min[0] != -85.0 {
		t.Errorf("Expected Min[0] -85.0, got %f", bounds.Min[0])
	}

	if bounds.Min[1] != -180.0 {
		t.Errorf("Expected Min[1] -180.0, got %f", bounds.Min[1])
	}

	if bounds.Max[0] != 85.0 {
		t.Errorf("Expected Max[0] 85.0, got %f", bounds.Max[0])
	}

	if bounds.Max[1] != 180.0 {
		t.Errorf("Expected Max[1] 180.0, got %f", bounds.Max[1])
	}
}

func TestStaticModel3DProvider_Srs(t *testing.T) {
	provider := NewStaticModel3DProvider(geo.NewProj(4326))
	srs := provider.Srs()

	if srs == nil {
		t.Fatal("Srs returned nil")
	}

	if !srs.Eq(geo.NewProj(4326)) {
		t.Error("Expected SRS to be EPSG:4326")
	}
}

func TestStaticModel3DProvider_Attribution(t *testing.T) {
	provider := NewStaticModel3DProvider(nil)
	attribution := provider.Attribution()

	if attribution != "" {
		t.Errorf("Expected empty attribution, got '%s'", attribution)
	}
}

func TestStaticModel3DProvider_triangulateFace(t *testing.T) {
	provider := NewStaticModel3DProvider(nil)

	t.Run("Triangle", func(t *testing.T) {
		indices := []int{0, 1, 2}
		tris := provider.triangulateFace(indices)

		if len(tris) != 1 {
			t.Errorf("Expected 1 triangle, got %d", len(tris))
		}

		if tris[0] != [3]int{0, 1, 2} {
			t.Errorf("Expected [0,1,2], got %v", tris[0])
		}
	})

	t.Run("Quad", func(t *testing.T) {
		indices := []int{0, 1, 2, 3}
		tris := provider.triangulateFace(indices)

		if len(tris) != 2 {
			t.Errorf("Expected 2 triangles, got %d", len(tris))
		}
	})

	t.Run("Pentagon", func(t *testing.T) {
		indices := []int{0, 1, 2, 3, 4}
		tris := provider.triangulateFace(indices)

		if len(tris) != 3 {
			t.Errorf("Expected 3 triangles, got %d", len(tris))
		}
	})
}

func TestStaticModel3DProvider_parseFaceIndices(t *testing.T) {
	provider := NewStaticModel3DProvider(nil)

	tests := []struct {
		input    []string
		expected []int
	}{
		{[]string{"1/1/1", "2/2/2", "3/3/3"}, []int{0, 1, 2}},
		{[]string{"1", "2", "3"}, []int{0, 1, 2}},
		{[]string{"0/1/2"}, []int{0}},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := provider.parseFaceIndices(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d indices, got %d", len(tt.expected), len(result))
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Index %d: expected %d, got %d", i, expected, result[i])
				}
			}
		})
	}
}

func TestStaticModel3DProvider_convertMSTToTinMesh(t *testing.T) {
	provider := NewStaticModel3DProvider(nil)

	t.Run("Nil mesh", func(t *testing.T) {
		tinMesh := provider.convertMSTToTinMesh(nil)

		if tinMesh == nil {
			t.Error("Expected non-nil TinMesh")
		}

		if len(tinMesh.Vertices) != 0 {
			t.Errorf("Expected 0 vertices, got %d", len(tinMesh.Vertices))
		}
	})
}
