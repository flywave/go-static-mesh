package draw

import (
	"github.com/flywave/go-geos"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

type Clipper struct{}

func NewClipper() *Clipper {
	return &Clipper{}
}

func (c *Clipper) ClipPathToBounds(path *Path, bounds vec2d.Rect) *Path {
	if len(path.Positions) == 0 {
		return nil
	}

	coords := make([]geos.Coord, len(path.Positions))
	for i, pos := range path.Positions {
		coords[i] = geos.Coord{X: pos[0], Y: pos[1]}
	}

	lineString := geos.CreateLineString(coords)
	if lineString == nil {
		return nil
	}

	clipped := lineString.ClipByRect(bounds.Min[0], bounds.Min[1], bounds.Max[0], bounds.Max[1])
	if clipped == nil {
		return nil
	}

	return c.geosToPath(clipped, path)
}

func (c *Clipper) ClipAreaToBounds(area *Area, bounds vec2d.Rect) *Area {
	if len(area.Positions) < 3 {
		return nil
	}

	coords := make([]geos.Coord, len(area.Positions))
	for i, pos := range area.Positions {
		coords[i] = geos.Coord{X: pos[0], Y: pos[1]}
	}

	polygon := geos.CreatePolygon(coords)
	if polygon == nil {
		return nil
	}

	clipped := polygon.ClipByRect(bounds.Min[0], bounds.Min[1], bounds.Max[0], bounds.Max[1])
	if clipped == nil {
		return nil
	}

	return c.geosToArea(clipped, area)
}

func (c *Clipper) geosToPath(geom *geos.Geometry, original *Path) *Path {
	if geom == nil {
		return nil
	}

	geomType := geom.GetType()
	if geomType != geos.LINESTRING && geomType != geos.MULTILINESTRING {
		return nil
	}

	var coords []geos.Coord

	if geomType == geos.LINESTRING {
		coords = geom.GetCoords()
	} else {
		numGeometries := geom.GetNumGeometries()
		for i := 0; i < numGeometries; i++ {
			subGeom := geom.GetGeometryN(i)
			if subGeom != nil && subGeom.GetType() == geos.LINESTRING {
				coords = append(coords, subGeom.GetCoords()...)
			}
		}
	}

	if len(coords) < 2 {
		return nil
	}

	positions := make([]vec2d.T, len(coords))
	for i, coord := range coords {
		positions[i] = vec2d.T{coord.X, coord.Y}
	}

	clippedPath := &Path{
		Positions: positions,
		Srs:       original.Srs,
		Color:     original.Color,
		Weight:    original.Weight,
		Height:    original.Height,
	}

	return clippedPath
}

func (c *Clipper) geosToArea(geom *geos.Geometry, original *Area) *Area {
	if geom == nil {
		return nil
	}

	geomType := geom.GetType()
	if geomType != geos.POLYGON && geomType != geos.MULTIPOLYGON {
		return nil
	}

	var shellCoords []geos.Coord

	if geomType == geos.POLYGON {
		if geom.GetNumGeometries() > 0 {
			ring := geom.GetGeometryN(0)
			if ring != nil {
				shellCoords = ring.GetCoords()
			}
		}
	} else {
		numGeometries := geom.GetNumGeometries()
		for i := 0; i < numGeometries; i++ {
			subGeom := geom.GetGeometryN(i)
			if subGeom != nil && subGeom.GetType() == geos.POLYGON {
				if subGeom.GetNumGeometries() > 0 {
					ring := subGeom.GetGeometryN(0)
					if ring != nil && len(shellCoords) == 0 {
						shellCoords = ring.GetCoords()
					}
				}
				break
			}
		}
	}

	if len(shellCoords) < 3 {
		return nil
	}

	positions := make([]vec2d.T, len(shellCoords))
	for i, coord := range shellCoords {
		positions[i] = vec2d.T{coord.X, coord.Y}
	}

	clippedArea := &Area{
		Positions: positions,
		Srs:       original.Srs,
		Color:     original.Color,
		Fill:      original.Fill,
		Weight:    original.Weight,
		Height:    original.Height,
	}

	return clippedArea
}
