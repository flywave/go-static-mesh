package writer

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/flywave/gltf"
	"github.com/flywave/go-static-mesh/mesh"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
)

type GLTFWriter struct {
	Binary      bool
	IncludeUVs  bool
	EmbedImages bool
	Draco       bool
}

func NewGLTFWriter(binary bool, includeUVs bool) *GLTFWriter {
	return &GLTFWriter{
		Binary:     binary,
		IncludeUVs: includeUVs,
	}
}

func NewGltfWriter() *GLTFWriter {
	return &GLTFWriter{
		Binary:     true,
		IncludeUVs: true,
	}
}

func (w *GLTFWriter) Write(m *mesh.Mesh, path string) error {
	doc := gltf.NewDocument()

	positionFloats := w.convertVerticesToFloat32(m.Vertices)
	normalFloats := w.convertNormalsToFloat32(m.Normals)
	indexBytes := w.convertIndicesToBytes(m.Indices)

	var buffers []*gltf.Buffer
	var bufferViews []*gltf.BufferView
	var accessors []*gltf.Accessor

	positionBuf := &gltf.Buffer{
		Data:       w.makeFloat32Buffer(positionFloats),
		ByteLength: uint32(len(positionFloats) * 4),
	}
	buffers = append(buffers, positionBuf)

	normalBuf := &gltf.Buffer{
		Data:       w.makeFloat32Buffer(normalFloats),
		ByteLength: uint32(len(normalFloats) * 4),
	}
	buffers = append(buffers, normalBuf)

	indexBuf := &gltf.Buffer{
		Data:       indexBytes,
		ByteLength: uint32(len(indexBytes)),
	}
	buffers = append(buffers, indexBuf)

	attributes := map[string]uint32{
		"POSITION": 0,
		"NORMAL":   1,
	}

	accessors = append(accessors, &gltf.Accessor{
		BufferView:    gltf.Index(0),
		ByteOffset:    0,
		ComponentType: gltf.ComponentFloat,
		Type:          gltf.AccessorVec3,
		Count:         uint32(len(m.Vertices)),
		Max:           w.findMax(positionFloats),
		Min:           w.findMin(positionFloats),
	})

	accessors = append(accessors, &gltf.Accessor{
		BufferView:    gltf.Index(1),
		ByteOffset:    0,
		ComponentType: gltf.ComponentFloat,
		Type:          gltf.AccessorVec3,
		Count:         uint32(len(m.Normals)),
	})

	accessors = append(accessors, &gltf.Accessor{
		BufferView:    gltf.Index(2),
		ByteOffset:    0,
		ComponentType: gltf.ComponentUint,
		Type:          gltf.AccessorScalar,
		Count:         uint32(len(m.Indices)),
	})

	uvAccessorIndex := uint32(3)
	if w.IncludeUVs && len(m.UVs) > 0 {
		uvFloats := w.convertUVsToFloat32(m.UVs)
		uvBuf := &gltf.Buffer{
			Data:       w.makeFloat32Buffer(uvFloats),
			ByteLength: uint32(len(uvFloats)) * 4,
		}
		buffers = append(buffers, uvBuf)

		attributes["TEXCOORD_0"] = uvAccessorIndex

		accessors = append(accessors, &gltf.Accessor{
			BufferView:    gltf.Index(uvAccessorIndex),
			ByteOffset:    0,
			ComponentType: gltf.ComponentFloat,
			Type:          gltf.AccessorVec2,
			Count:         uint32(len(m.UVs)),
		})
	}

	positionView := &gltf.BufferView{
		Buffer:     uint32(0),
		ByteOffset: 0,
		ByteLength: positionBuf.ByteLength,
		Target:     gltf.TargetArrayBuffer,
	}
	bufferViews = append(bufferViews, positionView)

	normalView := &gltf.BufferView{
		Buffer:     uint32(1),
		ByteOffset: 0,
		ByteLength: normalBuf.ByteLength,
		Target:     gltf.TargetArrayBuffer,
	}
	bufferViews = append(bufferViews, normalView)

	indexView := &gltf.BufferView{
		Buffer:     uint32(2),
		ByteOffset: 0,
		ByteLength: indexBuf.ByteLength,
		Target:     gltf.TargetElementArrayBuffer,
	}
	bufferViews = append(bufferViews, indexView)

	if w.IncludeUVs && len(m.UVs) > 0 {
		uvView := &gltf.BufferView{
			Buffer:     uvAccessorIndex,
			ByteOffset: 0,
			ByteLength: buffers[uvAccessorIndex].ByteLength,
			Target:     gltf.TargetArrayBuffer,
		}
		bufferViews = append(bufferViews, uvView)
	}

	doc.Meshes = []*gltf.Mesh{{
		Primitives: []*gltf.Primitive{
			{
				Indices:    gltf.Index(2),
				Attributes: attributes,
				Mode:       gltf.PrimitiveTriangles,
			},
		},
	}}

	doc.Nodes = []*gltf.Node{
		{
			Name: "mesh",
			Mesh: gltf.Index(0),
		},
	}

	doc.Scenes[0].Nodes = []uint32{0}

	if w.Binary {
		doc.Buffers = []*gltf.Buffer{{Data: w.combineBinaryData(buffers), ByteLength: 0, URI: ""}}
		doc.BufferViews = bufferViews
		doc.Accessors = accessors

		return gltf.SaveBinary(doc, path)
	}

	doc.Buffers = buffers
	doc.BufferViews = bufferViews
	doc.Accessors = accessors

	return gltf.Save(doc, path)
}

func (w *GLTFWriter) WriteTo(m *mesh.Mesh, writer io.Writer) error {
	return w.WriteFile(m, "")
}

func (w *GLTFWriter) WriteFile(m *mesh.Mesh, path string) error {
	return w.Write(m, path)
}

func (w *GLTFWriter) SetBinary(binary bool) {
	w.Binary = binary
}

func (w *GLTFWriter) SetIncludeUVs(include bool) {
	w.IncludeUVs = include
}

func (w *GLTFWriter) SetEmbedImages(embed bool) {
	w.EmbedImages = embed
}

func (w *GLTFWriter) SetDraco(enabled bool) {
	w.Draco = enabled
}

func (w *GLTFWriter) convertVerticesToFloat32(vertices []vec3d.T) []float32 {
	result := make([]float32, len(vertices)*3)
	for i, v := range vertices {
		result[i*3] = float32(v[0])
		result[i*3+1] = float32(v[1])
		result[i*3+2] = float32(v[2])
	}
	return result
}

func (w *GLTFWriter) convertNormalsToFloat32(normals []vec3d.T) []float32 {
	result := make([]float32, len(normals)*3)
	for i, n := range normals {
		result[i*3] = float32(n[0])
		result[i*3+1] = float32(n[1])
		result[i*3+2] = float32(n[2])
	}
	return result
}

func (w *GLTFWriter) convertUVsToFloat32(uvs []vec2d.T) []float32 {
	result := make([]float32, len(uvs)*2)
	for i, uv := range uvs {
		result[i*2] = float32(uv[0])
		result[i*2+1] = float32(uv[1])
	}
	return result
}

func (w *GLTFWriter) convertIndicesToBytes(indices []uint32) []byte {
	buf := new(bytes.Buffer)
	for _, idx := range indices {
		binary.Write(buf, binary.LittleEndian, idx)
	}
	return buf.Bytes()
}

func (w *GLTFWriter) makeFloat32Buffer(data []float32) []byte {
	buf := new(bytes.Buffer)
	for _, item := range data {
		binary.Write(buf, binary.LittleEndian, item)
	}
	return buf.Bytes()
}

func (w *GLTFWriter) combineBinaryData(buffers []*gltf.Buffer) []byte {
	var totalSize uint32
	for _, buf := range buffers {
		totalSize += buf.ByteLength
	}

	result := make([]byte, totalSize)
	offset := 0
	for _, buf := range buffers {
		copy(result[offset:], buf.Data)
		offset += int(buf.ByteLength)
	}

	return result
}

func (w *GLTFWriter) findMax(data []float32) []float32 {
	if len(data) == 0 {
		return []float32{0, 0, 0}
	}
	maxX := data[0]
	maxY := data[1]
	maxZ := data[2]
	for i := 3; i < len(data); i += 3 {
		if data[i] > maxX {
			maxX = data[i]
		}
		if data[i+1] > maxY {
			maxY = data[i+1]
		}
		if data[i+2] > maxZ {
			maxZ = data[i+2]
		}
	}
	return []float32{maxX, maxY, maxZ}
}

func (w *GLTFWriter) findMin(data []float32) []float32 {
	if len(data) == 0 {
		return []float32{0, 0, 0}
	}
	minX := data[0]
	minY := data[1]
	minZ := data[2]
	for i := 3; i < len(data); i += 3 {
		if data[i] < minX {
			minX = data[i]
		}
		if data[i+1] < minY {
			minY = data[i+1]
		}
		if data[i+2] < minZ {
			minZ = data[i+2]
		}
	}
	return []float32{minX, minY, minZ}
}
