package builder

import "fmt"

type BuilderConfig struct {
	Resolution           float64
	VerticalExaggeration float64
	BaseElevation        float64
	AutoZoomMin          int
	AutoZoomMax          int
	ExtrudeGeoData       bool
	GeoDataHeight        float64
	CloseMesh            bool
	BaseThickness        float64
	TextureSize          int
}

func (c *BuilderConfig) Validate() error {
	if c.Resolution <= 0 {
		return fmt.Errorf("resolution must be positive, got %f", c.Resolution)
	}

	if c.VerticalExaggeration <= 0 {
		return fmt.Errorf("vertical exaggeration must be positive, got %f", c.VerticalExaggeration)
	}

	if c.AutoZoomMin < 0 || c.AutoZoomMax < c.AutoZoomMin {
		return fmt.Errorf("invalid auto zoom range: [%d, %d]", c.AutoZoomMin, c.AutoZoomMax)
	}

	if c.CloseMesh && c.BaseThickness <= 0 {
		return fmt.Errorf("base thickness must be positive when close mesh is enabled, got %f", c.BaseThickness)
	}

	if c.TextureSize < 256 {
		return fmt.Errorf("texture size must be at least 256, got %d", c.TextureSize)
	}

	return nil
}

func (b *Builder) GetConfig() *BuilderConfig {
	return &BuilderConfig{
		Resolution:           b.resolution,
		VerticalExaggeration: b.verticalExaggeration,
		BaseElevation:        b.baseElevation,
		AutoZoomMin:          b.autoZoomMin,
		AutoZoomMax:          b.autoZoomMax,
		ExtrudeGeoData:       b.extrudeGeoData,
		GeoDataHeight:        b.geoDataHeight,
		CloseMesh:            b.closeMesh,
		BaseThickness:        b.baseThickness,
		TextureSize:          2048,
	}
}

func (b *Builder) SetConfig(config *BuilderConfig) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	b.resolution = config.Resolution
	b.verticalExaggeration = config.VerticalExaggeration
	b.baseElevation = config.BaseElevation
	b.autoZoomMin = config.AutoZoomMin
	b.autoZoomMax = config.AutoZoomMax
	b.extrudeGeoData = config.ExtrudeGeoData
	b.geoDataHeight = config.GeoDataHeight
	b.closeMesh = config.CloseMesh
	b.baseThickness = config.BaseThickness

	return nil
}
