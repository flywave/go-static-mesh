package writer

import (
	"fmt"
	"io"
	"os"

	"github.com/flywave/go-static-mesh/mesh"
	mst "github.com/flywave/go-mst"
	"github.com/flywave/go3d/vec2"
	"github.com/flywave/go3d/vec3"
)

type MSTWriter struct {
	IncludeUVs bool
}

func NewMSTWriter(includeUVs bool) *MSTWriter {
	return &MSTWriter{IncludeUVs: includeUVs}
}

func NewMstWriter() *MSTWriter {
	return &MSTWriter{IncludeUVs: true}
}

func (w *MSTWriter) Write(m *mesh.Mesh, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	return w.WriteTo(m, file)
}

func (w *MSTWriter) WriteFile(m *mesh.Mesh, path string) error {
	return w.Write(m, path)
}

func (w *MSTWriter) WriteTo(m *mesh.Mesh, wt io.Writer) error {
	mstMesh := w.convertToMST(m)
	return mst.MeshMarshal(wt, mstMesh)
}

func (w *MSTWriter) convertToMST(m *mesh.Mesh) *mst.Mesh {
	mstMesh := mst.NewMesh()

	node := &mst.MeshNode{
		Vertices:  make([]vec3.T, len(m.Vertices)),
		Normals:   make([]vec3.T, len(m.Normals)),
		FaceGroup: []*mst.MeshTriangle{},
	}
	for i, v := range m.Vertices {
		node.Vertices[i] = vec3.T{float32(v[0]), float32(v[1]), float32(v[2])}
	}
	for i, n := range m.Normals {
		node.Normals[i] = vec3.T{float32(n[0]), float32(n[1]), float32(n[2])}
	}
	if w.IncludeUVs && len(m.UVs) > 0 {
		node.TexCoords = make([]vec2.T, len(m.UVs))
		for i, uv := range m.UVs {
			node.TexCoords[i] = vec2.T{float32(uv[0]), float32(uv[1])}
		}
	}

	mstMesh.Materials = w.convertMaterials(m)

	if len(m.MaterialIndices) > 0 {
		buildFaceGroupsMST(node, m.Indices, m.MaterialIndices)
	} else {
		tri := &mst.MeshTriangle{Batchid: 0}
		for i := 0; i < len(m.Indices); i += 3 {
			if i+2 >= len(m.Indices) {
				break
			}
			i0, i1, i2 := m.Indices[i], m.Indices[i+1], m.Indices[i+2]
			if i0 >= uint32(len(m.Vertices)) || i1 >= uint32(len(m.Vertices)) || i2 >= uint32(len(m.Vertices)) {
				continue
			}
			face := &mst.Face{Vertex: [3]uint32{i0, i1, i2}}
			if len(node.TexCoords) > 0 {
				face.Uv = &[3]uint32{i0, i1, i2}
			}
			tri.Faces = append(tri.Faces, face)
		}
		if len(tri.Faces) > 0 {
			node.FaceGroup = append(node.FaceGroup, tri)
		}
	}

	if len(node.FaceGroup) > 0 {
		mstMesh.Nodes = append(mstMesh.Nodes, node)
	}
	return mstMesh
}

func buildFaceGroupsMST(node *mst.MeshNode, indices []uint32, matIndices []uint32) {
	type triRange struct {
		start  int
		count  int
		matIdx uint32
	}
	ranges := make([]triRange, 0)
	currentMat := matIndices[0]
	rangeStart := 0
	for i := 0; i < len(matIndices); i++ {
		if matIndices[i] != currentMat {
			ranges = append(ranges, triRange{rangeStart * 3, (i - rangeStart) * 3, currentMat})
			currentMat = matIndices[i]
			rangeStart = i
		}
	}
	ranges = append(ranges, triRange{rangeStart * 3, (len(matIndices) - rangeStart) * 3, currentMat})

	for _, r := range ranges {
		tri := &mst.MeshTriangle{Batchid: int32(r.matIdx)}
		end := r.start + r.count
		for idx := r.start; idx < end; idx += 3 {
			if idx+2 >= len(indices) {
				break
			}
			i0, i1, i2 := indices[idx], indices[idx+1], indices[idx+2]
			if i0 >= uint32(len(node.Vertices)) || i1 >= uint32(len(node.Vertices)) || i2 >= uint32(len(node.Vertices)) {
				continue
			}
			face := &mst.Face{Vertex: [3]uint32{i0, i1, i2}}
			if len(node.TexCoords) > 0 {
				face.Uv = &[3]uint32{i0, i1, i2}
			}
			tri.Faces = append(tri.Faces, face)
		}
		if len(tri.Faces) > 0 {
			node.FaceGroup = append(node.FaceGroup, tri)
		}
	}
}

func (w *MSTWriter) convertMaterials(m *mesh.Mesh) []mst.MeshMaterial {
	if len(m.Materials) == 0 {
		return nil
	}
	materials := make([]mst.MeshMaterial, 0, len(m.Materials))
	for _, mat := range m.Materials {
		r, g, b, _ := mat.Diffuse.RGBA()

		pbr := &mst.PbrMaterial{
			Metallic:        mat.Metalness,
			Roughness:       mat.Roughness,
		}
		pbr.Color = [3]byte{byte(r >> 8), byte(g >> 8), byte(b >> 8)}

		img := mat.Texture
		if img == nil {
			img = m.Texture
		}
		if img != nil {
			tex, err := mst.CreateTextureFromImage(img, "texture", true)
			if err == nil {
				pbr.Texture = tex
			}
		}
		materials = append(materials, pbr)
	}
	return materials
}
