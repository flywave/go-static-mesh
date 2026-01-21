package mesh

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/flywave/gltf"
	stl "github.com/flywave/go-stl"
	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"
	vec3 "github.com/flywave/go3d/vec3"
)

type Writer interface {
	Write(mesh *Mesh, path string) error
	WriteTo(mesh *Mesh, w io.Writer) error
}

type STLWriter struct {
	ASCII bool
}

func NewSTLWriter(ascii bool) *STLWriter {
	return &STLWriter{ASCII: ascii}
}

func (w *STLWriter) Write(mesh *Mesh, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return w.WriteTo(mesh, file)
}

func (w *STLWriter) WriteTo(mesh *Mesh, writer io.Writer) error {
	solid := &stl.Solid{
		Name:    "mesh",
		IsAscii: w.ASCII,
	}

	triangles := make([]stl.Triangle, len(mesh.Indices)/3)

	for i := 0; i < len(mesh.Indices); i += 3 {
		v0 := mesh.Vertices[mesh.Indices[i]]
		v1 := mesh.Vertices[mesh.Indices[i+1]]
		v2 := mesh.Vertices[mesh.Indices[i+2]]

		normal := w.calculateNormal(v0, v1, v2)

		triangles[i/3] = stl.Triangle{
			Normal: normal,
			Vertices: [3]vec3.T{
				{float32(v0[0]), float32(v0[1]), float32(v0[2])},
				{float32(v1[0]), float32(v1[1]), float32(v1[2])},
				{float32(v2[0]), float32(v2[1]), float32(v2[2])},
			},
		}
	}

	solid.Triangles = triangles

	return solid.WriteAll(writer)
}

func (w *STLWriter) calculateNormal(v0, v1, v2 vec3d.T) vec3.T {
	edge1 := vec3d.T{v1[0] - v0[0], v1[1] - v0[1], v1[2] - v0[2]}
	edge2 := vec3d.T{v2[0] - v0[0], v2[1] - v0[1], v2[2] - v0[2]}

	cross := vec3d.T{
		edge1[1]*edge2[2] - edge1[2]*edge2[1],
		edge1[2]*edge2[0] - edge1[0]*edge2[2],
		edge1[0]*edge2[1] - edge1[1]*edge2[0],
	}

	length := w.vectorLength(cross)

	if length > 0 {
		return vec3.T{
			float32(cross[0] / length),
			float32(cross[1] / length),
			float32(cross[2] / length),
		}
	}
	return vec3.T{0, 0, 1}
}

func (w *STLWriter) vectorLength(v vec3d.T) float64 {
	return v[0]*v[0] + v[1]*v[1] + v[2]*v[2]
}

type OBJWriter struct {
	IncludeNormals bool
	IncludeUVs     bool
}

func NewOBJWriter(includeNormals, includeUVs bool) *OBJWriter {
	return &OBJWriter{
		IncludeNormals: includeNormals,
		IncludeUVs:     includeUVs,
	}
}

func (w *OBJWriter) Write(mesh *Mesh, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return w.WriteTo(mesh, file)
}

func (w *OBJWriter) WriteTo(mesh *Mesh, writer io.Writer) error {
	_, err := io.WriteString(writer, "# Exported using go-static-mesh\n")
	if err != nil {
		return err
	}

	_, err = io.WriteString(writer, fmt.Sprintf("# %d vertices, %d faces\n", len(mesh.Vertices), len(mesh.Indices)/3))
	if err != nil {
		return err
	}

	for _, v := range mesh.Vertices {
		_, err = io.WriteString(writer, fmt.Sprintf("v %g %g %g\n", v[0], v[1], v[2]))
		if err != nil {
			return err
		}
	}

	if w.IncludeNormals && len(mesh.Normals) > 0 {
		for _, n := range mesh.Normals {
			_, err = io.WriteString(writer, fmt.Sprintf("vn %g %g %g\n", n[0], n[1], n[2]))
			if err != nil {
				return err
			}
		}
	}

	if w.IncludeUVs && len(mesh.UVs) > 0 {
		for _, uv := range mesh.UVs {
			_, err = io.WriteString(writer, fmt.Sprintf("vt %g %g\n", uv[0], uv[1]))
			if err != nil {
				return err
			}
		}
	}

	_, err = io.WriteString(writer, "g mesh\n")
	if err != nil {
		return err
	}

	for i := 0; i < len(mesh.Indices); i += 3 {
		line := "f"
		if w.IncludeNormals && len(mesh.Normals) > 0 {
			if w.IncludeUVs && len(mesh.UVs) > 0 {
				line += fmt.Sprintf(" %d/%d/%d %d/%d/%d %d/%d/%d",
					mesh.Indices[i]+1, mesh.Indices[i]+1, mesh.Indices[i]+1,
					mesh.Indices[i+1]+1, mesh.Indices[i+1]+1, mesh.Indices[i+1]+1,
					mesh.Indices[i+2]+1, mesh.Indices[i+2]+1, mesh.Indices[i+2]+1)
			} else {
				line += fmt.Sprintf(" %d//%d %d//%d %d//%d",
					mesh.Indices[i]+1, mesh.Indices[i]+1,
					mesh.Indices[i+1]+1, mesh.Indices[i+1]+1,
					mesh.Indices[i+2]+1, mesh.Indices[i+2]+1)
			}
		} else if w.IncludeUVs && len(mesh.UVs) > 0 {
			line += fmt.Sprintf(" %d/%d %d/%d %d/%d",
				mesh.Indices[i]+1, mesh.Indices[i]+1,
				mesh.Indices[i+1]+1, mesh.Indices[i+1]+1,
				mesh.Indices[i+2]+1, mesh.Indices[i+2]+1)
		} else {
			line += fmt.Sprintf(" %d %d %d",
				mesh.Indices[i]+1,
				mesh.Indices[i+1]+1,
				mesh.Indices[i+2]+1)
		}
		_, err = io.WriteString(writer, line+"\n")
		if err != nil {
			return err
		}
	}

	return nil
}

type GLTFWriter struct {
	Binary     bool
	IncludeUVs bool
}

func NewGLTFWriter(binary bool, includeUVs bool) *GLTFWriter {
	return &GLTFWriter{
		Binary:     binary,
		IncludeUVs: includeUVs,
	}
}

func (w *GLTFWriter) Write(mesh *Mesh, path string) error {
	doc := gltf.NewDocument()

	positionFloats := w.convertVerticesToFloat32(mesh.Vertices)
	normalFloats := w.convertNormalsToFloat32(mesh.Normals)
	indexBytes := w.convertIndicesToBytes(mesh.Indices)

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
		Count:         uint32(len(mesh.Vertices)),
		Max:           w.findMax(positionFloats),
		Min:           w.findMin(positionFloats),
	})

	accessors = append(accessors, &gltf.Accessor{
		BufferView:    gltf.Index(1),
		ByteOffset:    0,
		ComponentType: gltf.ComponentFloat,
		Type:          gltf.AccessorVec3,
		Count:         uint32(len(mesh.Normals)),
	})

	accessors = append(accessors, &gltf.Accessor{
		BufferView:    gltf.Index(2),
		ByteOffset:    0,
		ComponentType: gltf.ComponentUint,
		Type:          gltf.AccessorScalar,
		Count:         uint32(len(mesh.Indices)),
	})

	if w.IncludeUVs && len(mesh.UVs) > 0 {
		uvFloats := w.convertUVsToFloat32(mesh.UVs)
		uvBuf := &gltf.Buffer{
			Data:       w.makeFloat32Buffer(uvFloats),
			ByteLength: uint32(len(uvFloats)) * 4,
		}
		buffers = append(buffers, uvBuf)

		attributes["TEXCOORD_0"] = 3

		accessors = append(accessors, &gltf.Accessor{
			BufferView:    gltf.Index(3),
			ByteOffset:    0,
			ComponentType: gltf.ComponentFloat,
			Type:          gltf.AccessorVec2,
			Count:         uint32(len(mesh.UVs)),
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

func (w *GLTFWriter) WriteTo(mesh *Mesh, writer io.Writer) error {
	return w.Write(mesh, "")
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
