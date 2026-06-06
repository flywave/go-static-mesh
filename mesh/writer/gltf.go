package writer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"strings"

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
		Binary:      true,
		IncludeUVs:  true,
		EmbedImages: true,
	}
}

func (w *GLTFWriter) Write(m *mesh.Mesh, path string) error {
	doc := gltf.NewDocument()

	if len(m.Vertices) == 0 {
		return fmt.Errorf("mesh has no vertices")
	}

	numVertices := uint32(len(m.Vertices))
	for _, idx := range m.Indices {
		if idx >= numVertices {
			return fmt.Errorf("index %d out of bounds: mesh has %d vertices", idx, numVertices)
		}
	}

	positionFloats := w.convertVerticesToFloat32(m.Vertices)
	normalFloats := w.convertNormalsToFloat32(m.Normals)
	indexBytes := w.convertIndicesToBytes(m.Indices)

	var bufferViews []*gltf.BufferView
	var accessors []*gltf.Accessor

	var allBufferData []byte
	var currentOffset uint32

	positionData := w.makeFloat32Buffer(positionFloats)
	normalData := w.makeFloat32Buffer(normalFloats)

	positionView := &gltf.BufferView{
		Buffer:     0,
		ByteOffset: currentOffset,
		ByteLength: uint32(len(positionData)),
		Target:     gltf.TargetArrayBuffer,
	}
	bufferViews = append(bufferViews, positionView)
	allBufferData = append(allBufferData, positionData...)
	currentOffset += uint32(len(positionData))

	normalView := &gltf.BufferView{
		Buffer:     0,
		ByteOffset: currentOffset,
		ByteLength: uint32(len(normalData)),
		Target:     gltf.TargetArrayBuffer,
	}
	bufferViews = append(bufferViews, normalView)
	allBufferData = append(allBufferData, normalData...)
	currentOffset += uint32(len(normalData))

	indexDataOffset := currentOffset
	indexView := &gltf.BufferView{
		Buffer:     0,
		ByteOffset: currentOffset,
		ByteLength: uint32(len(indexBytes)),
		Target:     gltf.TargetElementArrayBuffer,
	}
	bufferViews = append(bufferViews, indexView)
	allBufferData = append(allBufferData, indexBytes...)
	currentOffset += uint32(len(indexBytes))

	var uvBufferViewIdx uint32
	var uvData []byte
	if w.IncludeUVs && len(m.UVs) > 0 {
		uvFloats := w.convertUVsToFloat32(m.UVs)
		uvData = w.makeFloat32Buffer(uvFloats)

		uvView := &gltf.BufferView{
			Buffer:     0,
			ByteOffset: currentOffset,
			ByteLength: uint32(len(uvData)),
			Target:     gltf.TargetArrayBuffer,
		}
		bufferViews = append(bufferViews, uvView)
		allBufferData = append(allBufferData, uvData...)
		currentOffset += uint32(len(uvData))
		uvBufferViewIdx = uint32(len(bufferViews) - 1)
	}

	textureBufferViewIdx := uint32(len(bufferViews))
	var textureData []byte
	if m.Texture != nil && w.EmbedImages {
		var err error
		textureData, _, err = w.encodeTexture(m.Texture)
		if err == nil && len(textureData) > 0 {
			textureView := &gltf.BufferView{
				Buffer:     0,
				ByteOffset: currentOffset,
				ByteLength: uint32(len(textureData)),
			}
			bufferViews = append(bufferViews, textureView)
			allBufferData = append(allBufferData, textureData...)
			currentOffset += uint32(len(textureData))
		}
	}

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

	if w.IncludeUVs && len(m.UVs) > 0 {
		attributes["TEXCOORD_0"] = uint32(len(accessors))
		accessors = append(accessors, &gltf.Accessor{
			BufferView:    gltf.Index(uvBufferViewIdx),
			ByteOffset:    0,
			ComponentType: gltf.ComponentFloat,
			Type:          gltf.AccessorVec2,
			Count:         uint32(len(m.UVs)),
		})
	}

	if len(m.MaterialIndices) > 0 {
		type triRange struct {
			start    uint32
			count    uint32
			matIdx   uint32
		}
		ranges := make([]triRange, 0)
		currentMat := m.MaterialIndices[0]
		rangeStart := uint32(0)
		for i := uint32(0); i < uint32(len(m.MaterialIndices)); i++ {
			if m.MaterialIndices[i] != currentMat {
				ranges = append(ranges, triRange{rangeStart * 3, (i - rangeStart) * 3, currentMat})
				currentMat = m.MaterialIndices[i]
				rangeStart = i
			}
		}
		ranges = append(ranges, triRange{rangeStart * 3, (uint32(len(m.MaterialIndices)) - rangeStart) * 3, currentMat})

		for _, r := range ranges {
			indexView := &gltf.BufferView{
				Buffer:     0,
				ByteOffset: indexDataOffset + r.start*4,
				ByteLength: r.count * 4,
				Target:     gltf.TargetElementArrayBuffer,
			}
			bufferViews = append(bufferViews, indexView)
			accessors = append(accessors, &gltf.Accessor{
				BufferView:    gltf.Index(uint32(len(bufferViews) - 1)),
				ByteOffset:    0,
				ComponentType: gltf.ComponentUint,
				Type:          gltf.AccessorScalar,
				Count:         r.count,
			})
		}

		for mi := range m.Materials {
			mat := &m.Materials[mi]
			r, g, b, a := mat.Diffuse.RGBA()
			alpha := float32(a) / 65535.0
			gltfMat := &gltf.Material{
				Name: mat.Name,
				PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
					BaseColorFactor: &[4]float32{
						float32(r) / 65535.0,
						float32(g) / 65535.0,
						float32(b) / 65535.0,
						alpha,
					},
					MetallicFactor:  gltf.Float(mat.Metalness),
					RoughnessFactor: gltf.Float(mat.Roughness),
				},
			}
			if mi == 0 && m.Texture != nil && w.EmbedImages && len(textureData) > 0 {
				textureIdx := uint32(len(doc.Textures))
				imageIdx := uint32(len(doc.Images))
				doc.Images = append(doc.Images, &gltf.Image{
					BufferView: gltf.Index(textureBufferViewIdx),
					MimeType:   "image/png",
				})
				doc.Textures = append(doc.Textures, &gltf.Texture{
					Source: gltf.Index(imageIdx),
				})
				gltfMat.PBRMetallicRoughness.BaseColorTexture = &gltf.TextureInfo{
					Index: textureIdx,
				}
			}
			doc.Materials = append(doc.Materials, gltfMat)
		}

		indexAccessorOffset := uint32(len(accessors) - len(ranges))
		primitives := make([]*gltf.Primitive, len(ranges))
		for i, r := range ranges {
			prim := &gltf.Primitive{
				Indices:    gltf.Index(indexAccessorOffset + uint32(i)),
				Attributes: attributes,
				Mode:       gltf.PrimitiveTriangles,
				Material:   gltf.Index(r.matIdx),
			}
			primitives[i] = prim
		}
		doc.Meshes = []*gltf.Mesh{{Primitives: primitives}}
	} else {
		accessors = append(accessors, &gltf.Accessor{
			BufferView:    gltf.Index(2),
			ByteOffset:    0,
			ComponentType: gltf.ComponentUint,
			Type:          gltf.AccessorScalar,
			Count:         uint32(len(m.Indices)),
		})

		primitive := &gltf.Primitive{
			Indices:    gltf.Index(2),
			Attributes: attributes,
			Mode:       gltf.PrimitiveTriangles,
		}

		if m.Texture != nil && w.EmbedImages && len(textureData) > 0 {
			textureIdx := uint32(len(doc.Textures))
			imageIdx := uint32(len(doc.Images))
			doc.Images = append(doc.Images, &gltf.Image{
				BufferView: gltf.Index(textureBufferViewIdx),
				MimeType:   "image/png",
			})
			doc.Textures = append(doc.Textures, &gltf.Texture{
				Source: gltf.Index(imageIdx),
			})
			doc.Materials = append(doc.Materials, &gltf.Material{
				Name: "textured_material",
				PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
					BaseColorTexture: &gltf.TextureInfo{
						Index: textureIdx,
					},
					MetallicFactor:  gltf.Float(0.0),
					RoughnessFactor: gltf.Float(0.5),
				},
			})
			primitive.Material = gltf.Index(0)
		}

		doc.Meshes = []*gltf.Mesh{{
			Primitives: []*gltf.Primitive{primitive},
		}}
	}

	doc.Nodes = []*gltf.Node{
		{
			Name: "mesh",
			Mesh: gltf.Index(0),
		},
	}

	doc.Scenes[0].Nodes = []uint32{0}

	doc.Buffers = []*gltf.Buffer{{
		Data:       allBufferData,
		ByteLength: uint32(len(allBufferData)),
		URI:        "",
	}}
	doc.BufferViews = bufferViews
	doc.Accessors = accessors

	shouldBeBinary := w.Binary
	if strings.HasSuffix(path, ".gltf") {
		shouldBeBinary = false
	} else if strings.HasSuffix(path, ".glb") {
		shouldBeBinary = true
	}

	if shouldBeBinary {
		return gltf.SaveBinary(doc, path)
	}

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

func (w *GLTFWriter) encodeTexture(img image.Image) ([]byte, string, error) {
	var buf bytes.Buffer
	var mimeType string

	if w.EmbedImages {
		if err := png.Encode(&buf, img); err != nil {
			return nil, "", err
		}
		mimeType = "image/png"
	} else {
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
			return nil, "", err
		}
		mimeType = "image/jpeg"
	}

	return buf.Bytes(), mimeType, nil
}
