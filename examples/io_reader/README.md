# io.Reader Examples

This directory contains examples demonstrating how to use `io.Reader` interface with GeoTIFF providers.

## Examples

### GeoTIFF Raster from io.Reader

```bash
cd examples/io_reader
go run main.go /path/to/elevation.tif
```

This example shows how to:
1. Read a GeoTIFF file into memory
2. Create a `GeoTIFFRasterProvider` from `io.Reader`
3. Access elevation data
4. Automatically clean up temporary files

## API Reference

### NewGeoTIFFRasterProviderFromReader

```go
func NewGeoTIFFRasterProviderFromReader(r io.Reader) (*GeoTIFFRasterProvider, error)
```

Creates a `GeoTIFFRasterProvider` from an `io.Reader`. The data is written to a temporary file and then opened with GDAL.

**Parameters:**
- `r`: A reader containing GeoTIFF elevation data

**Returns:**
- `*GeoTIFFRasterProvider`: The raster provider
- `error`: An error if creation fails

**Important:** Always call `Close()` on the provider to clean up the temporary file.

### NewGeoTIFFImageryProviderFromReader

```go
func NewGeoTIFFImageryProviderFromReader(r io.Reader) (*GeoTIFFImageryProvider, error)
```

Creates a `GeoTIFFImageryProvider` from an `io.Reader`. The data is written to a temporary file and then opened with COG.

**Parameters:**
- `r`: A reader containing GeoTIFF imagery data

**Returns:**
- `*GeoTIFFImageryProvider`: The imagery provider
- `error`: An error if creation fails

**Important:** Always call `Close()` on the provider to clean up the temporary file.

## Use Cases

### HTTP Download

```go
resp, err := http.Get("https://example.com/elevation.tif")
if err != nil {
    panic(err)
}
defer resp.Body.Close()

provider, err := static.NewGeoTIFFRasterProviderFromReader(resp.Body)
if err != nil {
    panic(err)
}
defer provider.Close()

// Use provider...
```

### AWS S3

```go
resp, err := s3Client.GetObject(&s3.GetObjectInput{
    Bucket: aws.String("my-bucket"),
    Key:    aws.String("elevation.tif"),
})
if err != nil {
    panic(err)
}
defer resp.Body.Close()

provider, err := static.NewGeoTIFFRasterProviderFromReader(resp.Body)
if err != nil {
    panic(err)
}
defer provider.Close()

// Use provider...
```

### Memory Buffer

```go
data, err := os.ReadFile("elevation.tif")
if err != nil {
    panic(err)
}

r := bytes.NewReader(data)
provider, err := static.NewGeoTIFFRasterProviderFromReader(r)
if err != nil {
    panic(err)
}
defer provider.Close()

// Use provider...
```
