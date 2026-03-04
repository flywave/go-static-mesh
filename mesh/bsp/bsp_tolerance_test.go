package bsp

import (
	"math"
	"testing"

	vec3d "github.com/flywave/go3d/float64/vec3"
)

func TestBSPPlane_Classify_WithEpsilon(t *testing.T) {
	plane := &BSPPlane{
		Normal:   vec3d.T{0, 0, 1},
		Distance: 0,
	}

	testCases := []struct {
		name     string
		point    vec3d.T
		expected int
	}{
		{"Clearly in front", vec3d.T{0, 0, 1}, 1},
		{"Clearly behind", vec3d.T{0, 0, -1}, -1},
		{"On plane (exact)", vec3d.T{0, 0, 0}, 0},
		{"Near plane (+epsilon/2)", vec3d.T{0, 0, 5e-7}, 0},
		{"Near plane (-epsilon/2)", vec3d.T{0, 0, -5e-7}, 0},
		{"Just above epsilon", vec3d.T{0, 0, 2e-6}, 1},
		{"Just below epsilon", vec3d.T{0, 0, -2e-6}, -1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := plane.Classify(tc.point)
			if result != tc.expected {
				t.Errorf("Classify(%v) = %d, expected %d", tc.point, result, tc.expected)
			}
		})
	}
}

func TestBSPPlane_LineIntersection_Epsilon(t *testing.T) {
	plane := &BSPPlane{
		Normal:   vec3d.T{0, 0, 1},
		Distance: 0,
	}

	t.Run("Parallel line (near zero dot product)", func(t *testing.T) {
		a := vec3d.T{0, 0, 1}
		b := vec3d.T{1, 0, 1}

		intersection := plane.LineIntersection(a, b)

		if math.IsNaN(intersection[0]) || math.IsInf(intersection[0], 0) {
			t.Errorf("Intersection should not be NaN or Inf, got %v", intersection)
		}

		if intersection[0] != a[0] || intersection[1] != a[1] || intersection[2] != a[2] {
			t.Errorf("For parallel line, should return point a, got %v", intersection)
		}
	})

	t.Run("Nearly parallel line", func(t *testing.T) {
		a := vec3d.T{0, 0, 1}
		b := vec3d.T{1, 0, 1 + 1e-11}

		intersection := plane.LineIntersection(a, b)

		if math.IsNaN(intersection[0]) || math.IsInf(intersection[0], 0) {
			t.Errorf("Intersection should not be NaN or Inf, got %v", intersection)
		}
	})

	t.Run("Normal intersection", func(t *testing.T) {
		a := vec3d.T{0, 0, 1}
		b := vec3d.T{0, 0, -1}

		intersection := plane.LineIntersection(a, b)

		expected := vec3d.T{0, 0, 0}
		tolerance := 1e-6

		for i := 0; i < 3; i++ {
			if math.Abs(intersection[i]-expected[i]) > tolerance {
				t.Errorf("Intersection[%d] = %f, expected %f (tolerance %f)",
					i, intersection[i], expected[i], tolerance)
			}
		}
	})

	t.Run("Clamped intersection", func(t *testing.T) {
		a := vec3d.T{0, 0, 1}
		b := vec3d.T{0, 0, 2}

		intersection := plane.LineIntersection(a, b)

		for i := 0; i < 3; i++ {
			if math.IsNaN(intersection[i]) || math.IsInf(intersection[i], 0) {
				t.Errorf("Intersection[%d] should not be NaN or Inf", i)
			}
		}
	})
}

func TestBSPPlane_SplitPolygon_CoplanarHandling(t *testing.T) {
	plane := &BSPPlane{
		Normal:   vec3d.T{0, 0, 1},
		Distance: 0,
	}

	vertices := []*BSPVertex{
		{Position: vec3d.T{0, 0, 1e-7}},
		{Position: vec3d.T{1, 0, 1e-7}},
		{Position: vec3d.T{0, 1, 1e-7}},
	}

	poly := NewBSPPolygon(vertices)
	if poly == nil {
		t.Fatal("Failed to create polygon")
	}

	var coplanarFront, coplanarBack, front, back BSPPolygons

	plane.SplitPolygon(poly, &coplanarFront, &coplanarBack, &front, &back)

	totalPolygons := len(coplanarFront) + len(coplanarBack) + len(front) + len(back)
	if totalPolygons == 0 {
		t.Error("At least one polygon should be generated from split")
	}
}

func TestBSPPlane_SplitPolygon_NearPlanarVertices(t *testing.T) {
	plane := &BSPPlane{
		Normal:   vec3d.T{0, 0, 1},
		Distance: 0,
	}

	vertices := []*BSPVertex{
		{Position: vec3d.T{0, 0, -1e-7}},
		{Position: vec3d.T{1, 0, 1e-7}},
		{Position: vec3d.T{0.5, 1, -1e-7}},
	}

	poly := NewBSPPolygon(vertices)
	if poly == nil {
		t.Fatal("Failed to create polygon")
	}

	var coplanarFront, coplanarBack, front, back BSPPolygons

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SplitPolygon panicked: %v", r)
		}
	}()

	plane.SplitPolygon(poly, &coplanarFront, &coplanarBack, &front, &back)

	totalPolygons := len(coplanarFront) + len(coplanarBack) + len(front) + len(back)
	if totalPolygons == 0 {
		t.Error("At least one polygon should be generated")
	}
}

func TestBSPPlane_LineIntersection_NumericalStability(t *testing.T) {
	testCases := []struct {
		name      string
		normal    vec3d.T
		distance  float64
		pointA    vec3d.T
		pointB    vec3d.T
		expectNaN bool
	}{
		{
			"Normal case",
			vec3d.T{0, 0, 1},
			0,
			vec3d.T{0, 0, 1},
			vec3d.T{0, 0, -1},
			false,
		},
		{
			"Parallel to plane",
			vec3d.T{0, 0, 1},
			0,
			vec3d.T{0, 0, 1},
			vec3d.T{1, 1, 1},
			false,
		},
		{
			"Nearly parallel (very small angle)",
			vec3d.T{0, 0, 1},
			0,
			vec3d.T{0, 0, 1},
			vec3d.T{1e10, 1e10, 1 + 1e-10},
			false,
		},
		{
			"Large coordinates",
			vec3d.T{0, 0, 1},
			0,
			vec3d.T{1e6, 1e6, 1e6},
			vec3d.T{1e6, 1e6, -1e6},
			false,
		},
		{
			"Small coordinates",
			vec3d.T{0, 0, 1},
			0,
			vec3d.T{1e-6, 1e-6, 1e-6},
			vec3d.T{1e-6, 1e-6, -1e-6},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plane := &BSPPlane{
				Normal:   tc.normal,
				Distance: tc.distance,
			}

			intersection := plane.LineIntersection(tc.pointA, tc.pointB)

			for i := 0; i < 3; i++ {
				if math.IsNaN(intersection[i]) {
					if !tc.expectNaN {
						t.Errorf("Unexpected NaN at component %d", i)
					}
				} else if math.IsInf(intersection[i], 0) {
					t.Errorf("Unexpected Inf at component %d", i)
				}
			}
		})
	}
}

func TestBSPPlane_Classify_BoundaryConditions(t *testing.T) {
	plane := &BSPPlane{
		Normal:   vec3d.T{1, 0, 0},
		Distance: -5,
	}

	testCases := []struct {
		name     string
		point    vec3d.T
		expected int
	}{
		{"At plane boundary +epsilon/2", vec3d.T{5 - 5e-7, 0, 0}, 0},
		{"At plane boundary -epsilon/2", vec3d.T{5 + 5e-7, 0, 0}, 0},
		{"Just inside front", vec3d.T{5.000001, 0, 0}, 1},
		{"Just inside back", vec3d.T{4.999999, 0, 0}, -1},
		{"Far front", vec3d.T{100, 0, 0}, 1},
		{"Far back", vec3d.T{-100, 0, 0}, -1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := plane.Classify(tc.point)
			if result != tc.expected {
				t.Errorf("Classify(%v) = %d, expected %d", tc.point, result, tc.expected)
			}
		})
	}
}

func BenchmarkBSPPlane_LineIntersection(b *testing.B) {
	plane := &BSPPlane{
		Normal:   vec3d.T{0, 0, 1},
		Distance: 0,
	}

	a := vec3d.T{0, 0, 1}
	bb := vec3d.T{0, 0, -1}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		plane.LineIntersection(a, bb)
	}
}

func BenchmarkBSPPlane_Classify(b *testing.B) {
	plane := &BSPPlane{
		Normal:   vec3d.T{0, 0, 1},
		Distance: 0,
	}

	point := vec3d.T{1, 2, 3}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		plane.Classify(point)
	}
}
