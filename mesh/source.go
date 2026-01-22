package mesh

import (
	"github.com/flywave/go-geo"
	"github.com/flywave/go-static-mesh/tile"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

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
	provider tile.RasterProvider
	bounds   vec2d.Rect
	srs      geo.Proj
}

func NewRasterSource(provider tile.RasterProvider) *RasterSource {
	return &RasterSource{
		provider: provider,
		bounds:   provider.Bounds(),
		srs:      provider.Srs(),
	}
}

func (s *RasterSource) Type() SourceType {
	return SourceRaster
}

func (s *RasterSource) Bounds() vec2d.Rect {
	return s.bounds
}

func (s *RasterSource) Srs() geo.Proj {
	return s.srs
}

func (s *RasterSource) TransformTo(srs geo.Proj) (Source, error) {
	if s.srs.Eq(srs) {
		return s, nil
	}

	transformedBounds := s.srs.TransformRectTo(srs, s.bounds, 16)
	return &RasterSource{
		provider: s.provider,
		bounds:   transformedBounds,
		srs:      srs,
	}, nil
}

type TinMeshSource struct {
	provider tile.TinMeshProvider
	bounds   vec2d.Rect
	srs      geo.Proj
}

func NewTinMeshSource(provider tile.TinMeshProvider) *TinMeshSource {
	return &TinMeshSource{
		provider: provider,
		bounds:   provider.Bounds(),
		srs:      provider.Srs(),
	}
}

func (s *TinMeshSource) Type() SourceType {
	return SourceTinMesh
}

func (s *TinMeshSource) Bounds() vec2d.Rect {
	return s.bounds
}

func (s *TinMeshSource) Srs() geo.Proj {
	return s.srs
}

func (s *TinMeshSource) TransformTo(srs geo.Proj) (Source, error) {
	if s.srs.Eq(srs) {
		return s, nil
	}

	transformedBounds := s.srs.TransformRectTo(srs, s.bounds, 16)
	return &TinMeshSource{
		provider: s.provider,
		bounds:   transformedBounds,
		srs:      srs,
	}, nil
}

type ImageSource struct {
	provider tile.ImageryProvider
	bounds   vec2d.Rect
	srs      geo.Proj
}

func NewImageSource(provider tile.ImageryProvider) *ImageSource {
	return &ImageSource{
		provider: provider,
		bounds:   provider.GetImageBounds(),
		srs:      provider.Srs(),
	}
}

func (s *ImageSource) Type() SourceType {
	return SourceImage
}

func (s *ImageSource) Bounds() vec2d.Rect {
	return s.bounds
}

func (s *ImageSource) Srs() geo.Proj {
	return s.srs
}

func (s *ImageSource) TransformTo(srs geo.Proj) (Source, error) {
	if s.srs.Eq(srs) {
		return s, nil
	}

	transformedBounds := s.srs.TransformRectTo(srs, s.bounds, 16)
	return &ImageSource{
		provider: s.provider,
		bounds:   transformedBounds,
		srs:      srs,
	}, nil
}

type Model3DSource struct {
	provider tile.Model3DProvider
	bounds   vec2d.Rect
	srs      geo.Proj
}

func NewModel3DSource(provider tile.Model3DProvider) *Model3DSource {
	bounds := vec2d.Rect{
		Min: vec2d.T{-90.0, -180.0},
		Max: vec2d.T{90.0, 180.0},
	}

	return &Model3DSource{
		provider: provider,
		bounds:   bounds,
		srs:      geo.NewProj(4326),
	}
}

func (s *Model3DSource) Type() SourceType {
	return SourceModel3D
}

func (s *Model3DSource) Bounds() vec2d.Rect {
	return s.bounds
}

func (s *Model3DSource) Srs() geo.Proj {
	return s.srs
}

func (s *Model3DSource) TransformTo(srs geo.Proj) (Source, error) {
	if s.srs.Eq(srs) {
		return s, nil
	}

	transformedBounds := s.srs.TransformRectTo(srs, s.bounds, 16)
	return &Model3DSource{
		provider: s.provider,
		bounds:   transformedBounds,
		srs:      srs,
	}, nil
}
