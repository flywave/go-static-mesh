module github.com/flywave/go-static-mesh

go 1.24

toolchain go1.24.4

require (
	github.com/flywave/flywave-gdal v0.0.0-00010101000000-000000000000
	github.com/flywave/gg v1.3.1-0.20210910115449-fe7fd154baa2
	github.com/flywave/gltf v0.20.4-0.20250828104044-ebb99e75f3cc
	github.com/flywave/go-cog v0.0.0-20250314092301-4673589220b8
	github.com/flywave/go-geo v0.0.0-20250314091853-e818cb9de299
	github.com/flywave/go-geoid v0.0.0-20220306024153-21126c4758a2
	github.com/flywave/go-geos v0.0.0-20220312005430-b3e54ee96ed7
	github.com/flywave/go-gpx v1.2.2-0.20211027141055-7fa376dde073
	github.com/flywave/go-mapbox v0.0.0-00010101000000-000000000000
	github.com/flywave/go-mst v0.0.0-20260112101636-2c1059fcfa2a
	github.com/flywave/go-quantized-mesh v0.0.0-20210525134750-cb854922974d
	github.com/flywave/go-stl v0.0.0-20250818070638-f2c3dee7ad76
	github.com/flywave/go-tesselator v0.0.0-20250908022608-9af2b7bde38f
	github.com/flywave/go-tin v0.0.0-20250611104340-f25bca1483ad
	github.com/flywave/go3d v0.0.0-20250816053852-aed5d825659f
	github.com/flywave/imaging v1.6.5
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	golang.org/x/image v0.28.0
)

require (
	github.com/flywave/go-geom v0.0.0-20250607125323-f685bf20f12c // indirect
	github.com/flywave/go-proj v0.0.0-20250607132305-d70d32f5ad2d // indirect
	github.com/flywave/webp v1.1.2 // indirect
	github.com/go-test/deep v1.0.7 // indirect
	github.com/google/tiff v0.0.0-20161109161721-4b31f3041d9a // indirect
	github.com/google/uuid v1.5.0 // indirect
	github.com/hhrutter/lzw v0.0.0-20190829144645-6f07a24e8650 // indirect
	github.com/pborman/uuid v1.2.1 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/text v0.26.0 // indirect
)

replace github.com/flywave/go-mbgeom => ../go-mbgeom

replace github.com/flywave/go-geom => ../go-geom

replace github.com/flywave/go-geo => ../go-geo

replace github.com/flywave/go-geoid => ../go-geoid

replace github.com/flywave/go-proj => ../go-proj

replace github.com/flywave/go-xslt => ../go-xslt

replace github.com/flywave/go-geos => ../go-geos

replace github.com/flywave/go-tin => ../go-tin

replace github.com/flywave/go-mapbox => ../go-mapbox

replace github.com/flywave/go-quantized-mesh => ../go-quantized-mesh

replace github.com/flywave/flywave-gdal => ../flywave-gdal

replace github.com/flywave/go-mst => ../go-mst

replace github.com/flywave/go-tesselator => ../go-tesselator
