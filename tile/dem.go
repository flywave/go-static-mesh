package tile

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"

	"github.com/flywave/go-geo"
	mraster "github.com/flywave/go-mapbox/raster"
	vec2d "github.com/flywave/go3d/float64/vec2"
)

type RasterDemMode uint32

const (
	ModeMapbox    RasterDemMode = mraster.DEM_ENCODING_MAPBOX
	ModeTerrarium RasterDemMode = mraster.DEM_ENCODING_TERRARIUM
)

type TileData struct {
	Size         [2]uint32
	Datas        []float64
	Border       BorderMode
	LeftBorder   []float64
	TopBorder    []float64
	RightBorder  []float64
	BottomBorder []float64
	NoData       float64
	Box          vec2d.Rect
	Boxsrs       geo.Proj
}

type BorderMode uint32

const (
	BORDER_NONE       BorderMode = 0
	BORDER_UNILATERAL BorderMode = 1
	BORDER_BILATERAL  BorderMode = 2
)

type BorderType uint32

const (
	BORDER_LEFT   BorderType = 0
	BORDER_TOP    BorderType = 1
	BORDER_RIGHT  BorderType = 2
	BORDER_BOTTOM BorderType = 3
)

func NewTileData(size [2]uint32, border BorderMode) *TileData {
	d := &TileData{Size: size, Datas: make([]float64, size[0]*size[1]), Border: border}
	if border == BORDER_UNILATERAL {
		d.LeftBorder = make([]float64, size[1])
		d.TopBorder = make([]float64, size[0]+1)
	} else if border == BORDER_BILATERAL {
		d.LeftBorder = make([]float64, size[1])
		d.TopBorder = make([]float64, size[0]+2)
		d.RightBorder = make([]float64, size[1])
		d.BottomBorder = make([]float64, size[0]+2)
	}
	return d
}

func (d *TileData) NoDataValue() float64 {
	return d.NoData
}

func (d *TileData) HasBorder() bool {
	return d.Border != BORDER_NONE
}

func (d *TileData) IsBilateral() bool {
	return d.Border == BORDER_BILATERAL
}

func (d *TileData) IsUnilateral() bool {
	return d.Border == BORDER_UNILATERAL
}

func (d *TileData) FillBorder(tp BorderType, i int, h float64) {
	switch tp {
	case BORDER_LEFT:
		if i < len(d.LeftBorder) {
			d.LeftBorder[i] = h
		}
	case BORDER_RIGHT:
		if i < len(d.RightBorder) {
			d.RightBorder[i] = h
		}
	case BORDER_TOP:
		if i < len(d.TopBorder) {
			d.TopBorder[i] = h
		}
	case BORDER_BOTTOM:
		if i < len(d.BottomBorder) {
			d.BottomBorder[i] = h
		}
	}
}

func (d *TileData) ToUnilateral() {
	d.Border = BORDER_UNILATERAL
}

func (d *TileData) Set(x, y int, h float64) {
	d.Datas[y*int(d.Size[0])+x] = h
}

func (d *TileData) Get(x, y int) float64 {
	return d.Datas[y*int(d.Size[0])+x]
}

func (d *TileData) GetExtend() ([]float64, [2]uint32, [6]float64) {
	pixelsize := calculatePixelSize(int(d.Size[0]), int(d.Size[1]), d.Box)

	if !d.HasBorder() {
		return d.Datas[:], d.Size, [6]float64{d.Box.Min[0], pixelsize[0], 0, d.Box.Max[1], 0, -pixelsize[1]}
	}
	if d.IsBilateral() {
		w, h := (d.Size[0] + 2), (d.Size[1] + 2)
		ret := make([]float64, w*h)
		copy(ret[:w], d.TopBorder)

		for y := 0; y < int(d.Size[1]); y++ {
			off := (y+1)*int(w) + 1
			ret[off-1] = d.LeftBorder[y]
			ret[off+int(d.Size[0])] = d.RightBorder[y]

			copy(ret[off:off+int(d.Size[0])], d.Datas[y*int(d.Size[0]):(y+1)*int(d.Size[0])])
		}

		off := (d.Size[1] + 1) * w
		copy(ret[off:int(off+w)], d.BottomBorder)

		return ret, [2]uint32{(d.Size[0] + 2), (d.Size[1] + 2)}, [6]float64{d.Box.Min[0], pixelsize[0], 0, d.Box.Max[1], 0, -pixelsize[1]}
	}
	if d.IsUnilateral() {
		w, h := (d.Size[0] + 1), (d.Size[1] + 1)
		ret := make([]float64, w*h)
		copy(ret[:w], d.TopBorder)

		for y := 0; y < int(d.Size[1]); y++ {
			off := (y+1)*int(w) + 1
			ret[off-1] = d.LeftBorder[y]
			copy(ret[off:off+int(d.Size[0])], d.Datas[y*int(d.Size[0]):(y+1)*int(d.Size[0])])
		}

		return ret, [2]uint32{(d.Size[0] + 1), (d.Size[1] + 1)}, [6]float64{d.Box.Min[0], pixelsize[0], 0, d.Box.Max[1], 0, -pixelsize[1]}
	}
	return nil, [2]uint32{}, [6]float64{}
}

func calculatePixelSize(width, height int, box vec2d.Rect) [2]float64 {
	if width <= 0 || height <= 0 {
		return [2]float64{0, 0}
	}
	return [2]float64{
		(box.Max[0] - box.Min[0]) / float64(width),
		(box.Max[1] - box.Min[1]) / float64(height),
	}
}

type DemRasterSource struct {
	Mode RasterDemMode
}

func NewDemRasterSource(mode RasterDemMode) *DemRasterSource {
	return &DemRasterSource{Mode: mode}
}

func LoadDEM(r io.Reader, mode RasterDemMode) (*mraster.DEMData, error) {
	return mraster.LoadDEMDataWithStream(r, int(mode))
}

type DemIO struct {
	Mode RasterDemMode
}

func (d *DemIO) Decode(r io.Reader) (*TileData, error) {
	data, err := mraster.LoadDEMDataWithStream(r, int(d.Mode))
	if err != nil {
		return nil, err
	}
	tiledata := NewTileData([2]uint32{uint32(data.Dim - 2), uint32(data.Dim - 2)}, BORDER_BILATERAL)
	for x := 0; x < data.Dim; x++ {
		for y := 0; y < data.Dim; y++ {
			if x > 0 && y > 0 && x < data.Dim-1 && y < data.Dim-1 {
				tiledata.Set(x-1, y-1, data.Get(x, y))
			}

			if x == 0 && y != 0 && y != data.Dim-1 {
				tiledata.FillBorder(BORDER_LEFT, y-1, data.Get(x, y))
			}

			if x == data.Dim-1 && y != 0 && y != data.Dim-1 {
				tiledata.FillBorder(BORDER_RIGHT, y-1, data.Get(x, y))
			}

			if y == 0 {
				tiledata.FillBorder(BORDER_TOP, x, data.Get(x, y))
			}

			if y == data.Dim-1 {
				tiledata.FillBorder(BORDER_BOTTOM, x, data.Get(x, y))
			}
		}
	}
	return tiledata, nil
}

func (d *DemIO) Encode(tile *TileData) ([]byte, error) {
	if !tile.IsBilateral() {
		return nil, fmt.Errorf("dem must sample bilateral border")
	}
	data, si, _ := tile.GetExtend()
	if si[0] != si[1] {
		return nil, fmt.Errorf("row === col")
	}

	var packer DemPacker
	if d.Mode == ModeMapbox {
		packer = &mapboxPacker{}
	} else {
		packer = &terrariumPacker{}
	}

	img := image.NewNRGBA(image.Rect(0, 0, int(si[0]), int(si[1])))
	for y := 0; y < int(si[1]); y++ {
		for x := 0; x < int(si[0]); x++ {
			rgba := packer.Pack(data[y*int(si[0])+x])
			img.SetNRGBA(x, y, color.NRGBA{
				R: rgba[0],
				G: rgba[1],
				B: rgba[2],
				A: rgba[3],
			})
		}
	}

	buf := &bytes.Buffer{}
	png.Encode(buf, img)
	return buf.Bytes(), nil
}

type DemPacker interface {
	Pack(h float64) [4]byte
}

type mapboxPacker struct{}

func (p *mapboxPacker) Pack(h float64) [4]byte {
	base := -10000.0
	interval := 0.1
	val := (h + base) / interval

	rVal := (math.Floor(math.Floor(val/256)/256)/256 - math.Floor(math.Floor(math.Floor(val/256)/256)/256)) * 256
	gVal := (math.Floor(val/256)/256 - math.Floor(math.Floor(val/256)/256)) * 256
	bVal := ((val / 256) - math.Floor(val/256)) * 256

	r := uint8(math.Round(rVal))
	g := uint8(math.Round(gVal))
	b := uint8(math.Round(bVal))
	return [4]byte{r, g, b, 255}
}

type terrariumPacker struct{}

func (p *terrariumPacker) Pack(h float64) [4]byte {
	base := 32768.0
	val := h + base

	r := uint8(math.Floor(val / 256))
	g := uint8(int(val) % 256)
	b := uint8(int(math.Mod(val*256, 25)))
	return [4]byte{r, g, b, 255}
}
