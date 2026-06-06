package tile

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"

	vec2d "github.com/flywave/go3d/float64/vec2"
	vec3d "github.com/flywave/go3d/float64/vec3"

	"github.com/flywave/gltf"
	"github.com/flywave/go-geo"
	"github.com/flywave/go-mst"
)

type StaticModel3DProvider struct {
	models map[string]*Model3D
	bounds vec2d.Rect
	srs    geo.Proj
	grid   *geo.TileGrid
	mutex  sync.RWMutex
}

func NewStaticModel3DProvider(srs geo.Proj) *StaticModel3DProvider {
	if srs == nil {
		srs = geo.NewProj(4326)
	}

	return &StaticModel3DProvider{
		models: make(map[string]*Model3D),
		srs:    srs,
		bounds: vec2d.Rect{
			Min: vec2d.T{-85.0, -180.0},
			Max: vec2d.T{85.0, 180.0},
		},
		grid: &geo.TileGrid{},
	}
}

func (p *StaticModel3DProvider) AddModel(model *Model3D) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.models[model.ID] = model
}

func (p *StaticModel3DProvider) GetModels(bounds vec2d.Rect) ([]Model3D, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var result []Model3D
	for _, model := range p.models {
		if p.modelInBounds(model, bounds) {
			result = append(result, *model)
		}
	}

	return result, nil
}

func (p *StaticModel3DProvider) GetModel(id string) (*Model3D, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	model, ok := p.models[id]
	if !ok {
		return nil, fmt.Errorf("model not found: %s", id)
	}

	return model, nil
}

func (p *StaticModel3DProvider) LoadModelFromFile(id, filepath string) error {
	ext := getFileExt(filepath)
	switch ext {
	case ".mst":
		return p.loadMSTFromFile(id, filepath)
	case ".obj":
		return p.loadOBJFromFile(id, filepath)
	case ".stl":
		return p.loadSTLFromFile(id, filepath)
	case ".glb", ".gltf":
		return p.loadGLBFromFile(id, filepath)
	default:
		return fmt.Errorf("unsupported format: %s (supported: .mst, .obj, .stl, .glb, .gltf)", ext)
	}
}

func (p *StaticModel3DProvider) LoadModelFromData(id string, data []byte, format string) error {
	switch strings.ToLower(format) {
	case "mst":
		return p.loadMSTFromData(id, data)
	case "obj":
		return p.loadOBJFromData(id, data)
	case "stl":
		return p.loadSTLFromData(id, data)
	case "glb", "gltf":
		return p.loadGLBFromData(id, data)
	default:
		return fmt.Errorf("unsupported format: %s (supported: mst, obj, stl, glb, gltf)", format)
	}
}

func (p *StaticModel3DProvider) loadMSTFromFile(id, filepath string) error {
	mstMesh, err := mst.MeshReadFrom(filepath)
	if err != nil {
		return fmt.Errorf("failed to read MST file: %w", err)
	}

	model := &Model3D{
		ID:     id,
		Name:   getFileName(filepath),
		Format: "mst",
		Mesh:   p.convertMSTToTinMesh(mstMesh),
	}

	p.AddModel(model)
	return nil
}

func (p *StaticModel3DProvider) loadMSTFromData(id string, data []byte) error {
	mstMesh := mst.MeshUnMarshal(bytes.NewReader(data))

	model := &Model3D{
		ID:     id,
		Name:   id,
		Format: "mst",
		Mesh:   p.convertMSTToTinMesh(mstMesh),
	}

	p.AddModel(model)
	return nil
}

func (p *StaticModel3DProvider) convertMSTToTinMesh(mstMesh *mst.Mesh) *TinMesh {
	if mstMesh == nil || len(mstMesh.Nodes) == 0 {
		return &TinMesh{
			Vertices:  []vec3d.T{},
			Indices:   []uint32{},
			MinHeight: 0,
			MaxHeight: 0,
		}
	}

	var vertices []vec3d.T
	var indices []uint32

	for _, node := range mstMesh.Nodes {
		for i := 0; i < len(node.Vertices); i++ {
			v := node.Vertices[i]
			vertices = append(vertices, vec3d.T{
				float64(v[0]),
				float64(v[1]),
				float64(v[2]),
			})
		}

		for _, triangle := range node.FaceGroup {
			for _, face := range triangle.Faces {
				indices = append(indices,
					uint32(face.Vertex[0]),
					uint32(face.Vertex[1]),
					uint32(face.Vertex[2]),
				)
			}
		}
	}

	minHeight := 0.0
	maxHeight := 0.0
	if len(vertices) > 0 {
		minHeight = vertices[0][2]
		maxHeight = vertices[0][2]
		for _, v := range vertices {
			if v[2] < minHeight {
				minHeight = v[2]
			}
			if v[2] > maxHeight {
				maxHeight = v[2]
			}
		}
	}

	return &TinMesh{
		Vertices:  vertices,
		Indices:   indices,
		MinHeight: minHeight,
		MaxHeight: maxHeight,
	}
}

func (p *StaticModel3DProvider) loadOBJFromFile(id, filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read OBJ file: %w", err)
	}

	return p.loadOBJFromData(id, data)
}

func (p *StaticModel3DProvider) loadOBJFromData(id string, data []byte) error {
	lines := strings.Split(string(data), "\n")

	var vertices []vec3d.T
	var indices []uint32
	vertexOffset := uint32(0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "v":
			if len(parts) >= 4 {
				x, _ := strconv.ParseFloat(parts[1], 64)
				y, _ := strconv.ParseFloat(parts[2], 64)
				z, _ := strconv.ParseFloat(parts[3], 64)
				vertices = append(vertices, vec3d.T{x, y, z})
			}
		case "f":
			faceIndices := p.parseFaceIndices(parts[1:])
			tris := p.triangulateFace(faceIndices)
			for _, tri := range tris {
				indices = append(indices,
					vertexOffset+uint32(tri[0]),
					vertexOffset+uint32(tri[1]),
					vertexOffset+uint32(tri[2]),
				)
			}
		}
	}

	mesh := &TinMesh{
		Vertices:  vertices,
		Indices:   indices,
		MinHeight: 0,
		MaxHeight: 0,
	}

	if len(vertices) > 0 {
		minZ := vertices[0][2]
		maxZ := vertices[0][2]
		for _, v := range vertices {
			if v[2] < minZ {
				minZ = v[2]
			}
			if v[2] > maxZ {
				maxZ = v[2]
			}
		}
		mesh.MinHeight = minZ
		mesh.MaxHeight = maxZ
	}

	model := &Model3D{
		ID:     id,
		Name:   id,
		Format: "obj",
		Mesh:   mesh,
	}

	p.AddModel(model)
	return nil
}

func (p *StaticModel3DProvider) parseFaceIndices(parts []string) []int {
	var indices []int
	for _, part := range parts {
		parts2 := strings.Split(part, "/")
		idx, _ := strconv.Atoi(parts2[0])
		if idx < 1 {
			indices = append(indices, 0)
		} else {
			indices = append(indices, idx-1)
		}
	}
	return indices
}

func (p *StaticModel3DProvider) triangulateFace(indices []int) [][3]int {
	if len(indices) == 3 {
		return [][3]int{{indices[0], indices[1], indices[2]}}
	}
	if len(indices) == 4 {
		return [][3]int{
			{indices[0], indices[1], indices[2]},
			{indices[0], indices[2], indices[3]},
		}
	}

	var tris [][3]int
	for i := 1; i < len(indices)-1; i++ {
		tris = append(tris, [3]int{indices[0], indices[i], indices[i+1]})
	}
	return tris
}

func (p *StaticModel3DProvider) loadSTLFromFile(id, filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read STL file: %w", err)
	}

	return p.loadSTLFromData(id, data)
}

func (p *StaticModel3DProvider) loadSTLFromData(id string, data []byte) error {
	var vertices []vec3d.T
	var indices []uint32

	if len(data) < 84 {
		return fmt.Errorf("invalid STL file: too short")
	}

	if string(data[:5]) == "solid" {
		p.loadASCIISTL(data, &vertices, &indices)
	} else {
		p.loadBinarySTL(data, &vertices, &indices)
	}

	minZ := 0.0
	maxZ := 0.0
	if len(vertices) > 0 {
		minZ = vertices[0][2]
		maxZ = vertices[0][2]
		for _, v := range vertices {
			if v[2] < minZ {
				minZ = v[2]
			}
			if v[2] > maxZ {
				maxZ = v[2]
			}
		}
	}

	mesh := &TinMesh{
		Vertices:  vertices,
		Indices:   indices,
		MinHeight: minZ,
		MaxHeight: maxZ,
	}

	model := &Model3D{
		ID:     id,
		Name:   id,
		Format: "stl",
		Mesh:   mesh,
	}

	p.AddModel(model)
	return nil
}

func (p *StaticModel3DProvider) loadASCIISTL(data []byte, vertices *[]vec3d.T, indices *[]uint32) {
	lines := strings.Split(string(data), "\n")
	vertexIndex := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "facet normal") {
			continue
		} else if strings.HasPrefix(line, "vertex") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				x, _ := strconv.ParseFloat(parts[1], 64)
				y, _ := strconv.ParseFloat(parts[2], 64)
				z, _ := strconv.ParseFloat(parts[3], 64)
				*vertices = append(*vertices, vec3d.T{x, y, z})
				vertexIndex++
			}
		} else if strings.HasPrefix(line, "endfacet") {
			if vertexIndex >= 3 {
				*indices = append(*indices, uint32(vertexIndex-3))
				*indices = append(*indices, uint32(vertexIndex-2))
				*indices = append(*indices, uint32(vertexIndex-1))
				vertexIndex = 0
			}
		}
	}
}

func (p *StaticModel3DProvider) loadBinarySTL(data []byte, vertices *[]vec3d.T, indices *[]uint32) {
	triangleCount := binary.LittleEndian.Uint32(data[80:84])

	offset := 84
	vertexOffset := uint32(0)

	for i := uint32(0); i < triangleCount; i++ {
		if offset+50 > len(data) {
			break
		}

		offset += 12

		for j := 0; j < 3; j++ {
			var v [3]float64
			for k := 0; k < 3; k++ {
				bits := binary.LittleEndian.Uint32(data[offset+k*4 : offset+(k+1)*4])
				v[k] = float64(math.Float32frombits(bits))
			}
			*vertices = append(*vertices, vec3d.T{v[0], v[1], v[2]})
			offset += 12
		}

		*indices = append(*indices, vertexOffset)
		*indices = append(*indices, vertexOffset+1)
		*indices = append(*indices, vertexOffset+2)
		vertexOffset += 3

		offset += 2
	}
}

func (p *StaticModel3DProvider) loadGLBFromFile(id, filepath string) error {
	doc, err := gltf.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open GLB/GLTF file: %w", err)
	}

	mesh, err := p.extractGLTFMesh(doc)
	if err != nil {
		return fmt.Errorf("failed to extract GLTF mesh: %w", err)
	}

	format := "gltf"
	if strings.HasSuffix(strings.ToLower(filepath), ".glb") {
		format = "glb"
	}

	model := &Model3D{
		ID:     id,
		Name:   getFileName(filepath),
		Format: format,
		Mesh:   mesh,
	}

	p.AddModel(model)
	return nil
}

func (p *StaticModel3DProvider) loadGLBFromData(id string, data []byte) error {
	decoder := gltf.NewDecoder(bytes.NewReader(data))
	doc := new(gltf.Document)
	if err := decoder.Decode(doc); err != nil {
		return fmt.Errorf("failed to decode GLB/GLTF data: %w", err)
	}

	mesh, err := p.extractGLTFMesh(doc)
	if err != nil {
		return fmt.Errorf("failed to extract GLTF mesh: %w", err)
	}

	model := &Model3D{
		ID:     id,
		Name:   id,
		Format: "gltf",
		Mesh:   mesh,
	}

	p.AddModel(model)
	return nil
}

func (p *StaticModel3DProvider) extractGLTFMesh(doc *gltf.Document) (*TinMesh, error) {
	if len(doc.Meshes) == 0 {
		return &TinMesh{
			Vertices:  []vec3d.T{},
			Indices:   []uint32{},
			MinHeight: 0,
			MaxHeight: 0,
		}, nil
	}

	var allVertices []vec3d.T
	var allIndices []uint32
	vertexOffset := uint32(0)

	for _, mesh := range doc.Meshes {
		for _, primitive := range mesh.Primitives {
			vertices, indices, err := p.extractGLTFPrimitive(doc, primitive)
			if err != nil {
				return nil, err
			}

			allVertices = append(allVertices, vertices...)

			for i := range indices {
				allIndices = append(allIndices, vertexOffset+indices[i])
			}

			vertexOffset += uint32(len(vertices))
		}
	}

	minHeight := 0.0
	maxHeight := 0.0
	if len(allVertices) > 0 {
		minHeight = allVertices[0][2]
		maxHeight = allVertices[0][2]
		for _, v := range allVertices {
			if v[2] < minHeight {
				minHeight = v[2]
			}
			if v[2] > maxHeight {
				maxHeight = v[2]
			}
		}
	}

	return &TinMesh{
		Vertices:  allVertices,
		Indices:   allIndices,
		MinHeight: minHeight,
		MaxHeight: maxHeight,
	}, nil
}

func (p *StaticModel3DProvider) extractGLTFPrimitive(doc *gltf.Document, primitive *gltf.Primitive) ([]vec3d.T, []uint32, error) {
	positionIdx, ok := primitive.Attributes[gltf.POSITION]
	if !ok {
		return nil, nil, fmt.Errorf("primitive missing POSITION attribute")
	}

	positionAccessor := doc.Accessors[positionIdx]
	if positionAccessor == nil {
		return nil, nil, fmt.Errorf("POSITION accessor not found")
	}

	vertices, err := p.extractGLTFVertices(doc, positionAccessor)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract vertices: %w", err)
	}

	var indices []uint32
	if primitive.Indices != nil {
		indicesAccessor := doc.Accessors[*primitive.Indices]
		if indicesAccessor != nil {
			indices, err = p.extractGLTFIndices(doc, indicesAccessor)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to extract indices: %w", err)
			}
		}
	} else {
		indices = make([]uint32, len(vertices))
		for i := range indices {
			indices[i] = uint32(i)
		}
	}

	mode := primitive.Mode
	if mode == 0 {
		mode = gltf.PrimitiveTriangles
	}

	if mode != gltf.PrimitiveTriangles {
		indices = p.triangulateGLTFPrimitive(indices, int(mode))
	}

	return vertices, indices, nil
}

func (p *StaticModel3DProvider) extractGLTFVertices(doc *gltf.Document, accessor *gltf.Accessor) ([]vec3d.T, error) {
	if accessor.Type != gltf.AccessorVec3 {
		return nil, fmt.Errorf("expected VEC3 accessor for POSITION, got %v", accessor.Type)
	}

	data, err := p.readAccessorData(doc, accessor)
	if err != nil {
		return nil, err
	}

	vertices := make([]vec3d.T, accessor.Count)
	for i := uint32(0); i < accessor.Count; i++ {
		offset := i * 12
		if offset+12 > uint32(len(data)) {
			break
		}
		bits := binary.LittleEndian.Uint32(data[offset : offset+4])
		x := float64(math.Float32frombits(bits))
		bits = binary.LittleEndian.Uint32(data[offset+4 : offset+8])
		y := float64(math.Float32frombits(bits))
		bits = binary.LittleEndian.Uint32(data[offset+8 : offset+12])
		z := float64(math.Float32frombits(bits))
		vertices[i] = vec3d.T{x, y, z}
	}

	return vertices, nil
}

func (p *StaticModel3DProvider) extractGLTFIndices(doc *gltf.Document, accessor *gltf.Accessor) ([]uint32, error) {
	data, err := p.readAccessorData(doc, accessor)
	if err != nil {
		return nil, err
	}

	indices := make([]uint32, accessor.Count)

	switch accessor.ComponentType {
	case gltf.ComponentUbyte:
		for i := uint32(0); i < accessor.Count; i++ {
			if uint32(i) >= uint32(len(data)) {
				break
			}
			indices[i] = uint32(data[i])
		}
	case gltf.ComponentUshort:
		for i := uint32(0); i < accessor.Count; i++ {
			offset := i * 2
			if offset+2 > uint32(len(data)) {
				break
			}
			indices[i] = uint32(binary.LittleEndian.Uint16(data[offset : offset+2]))
		}
	case gltf.ComponentUint:
		for i := uint32(0); i < accessor.Count; i++ {
			offset := i * 4
			if offset+4 > uint32(len(data)) {
				break
			}
			indices[i] = binary.LittleEndian.Uint32(data[offset : offset+4])
		}
	default:
		return nil, fmt.Errorf("unsupported component type for indices: %v", accessor.ComponentType)
	}

	return indices, nil
}

func (p *StaticModel3DProvider) readAccessorData(doc *gltf.Document, accessor *gltf.Accessor) ([]byte, error) {
	if accessor.BufferView == nil {
		return nil, fmt.Errorf("accessor has no buffer view")
	}

	bufferView := doc.BufferViews[*accessor.BufferView]
	if bufferView == nil {
		return nil, fmt.Errorf("buffer view not found")
	}

	buffer := doc.Buffers[bufferView.Buffer]
	if buffer == nil {
		return nil, fmt.Errorf("buffer not found")
	}

	if buffer.Data == nil {
		return nil, fmt.Errorf("buffer data not loaded")
	}

	offset := bufferView.ByteOffset + accessor.ByteOffset
	stride := bufferView.ByteStride
	if stride == 0 {
		stride = p.getComponentSize(accessor.ComponentType, accessor.Type)
	}

	elementSize := p.getComponentSize(accessor.ComponentType, accessor.Type)
	totalSize := uint32(len(buffer.Data))

	data := make([]byte, 0, accessor.Count*elementSize)
	for i := uint32(0); i < accessor.Count; i++ {
		start := offset + i*uint32(stride)
		end := start + elementSize
		if end > totalSize {
			break
		}
		data = append(data, buffer.Data[start:end]...)
	}

	return data, nil
}

func (p *StaticModel3DProvider) getComponentSize(componentType gltf.ComponentType, accessorType gltf.AccessorType) uint32 {
	componentSize := uint32(4)
	switch componentType {
	case gltf.ComponentByte, gltf.ComponentUbyte:
		componentSize = 1
	case gltf.ComponentShort, gltf.ComponentUshort:
		componentSize = 2
	case gltf.ComponentUint, gltf.ComponentFloat:
		componentSize = 4
	}

	numComponents := uint32(1)
	switch accessorType {
	case gltf.AccessorVec2:
		numComponents = 2
	case gltf.AccessorVec3:
		numComponents = 3
	case gltf.AccessorVec4:
		numComponents = 4
	case gltf.AccessorMat2:
		numComponents = 4
	case gltf.AccessorMat3:
		numComponents = 9
	case gltf.AccessorMat4:
		numComponents = 16
	}

	return componentSize * numComponents
}

func (p *StaticModel3DProvider) triangulateGLTFPrimitive(indices []uint32, mode int) []uint32 {
	switch mode {
	case 1:
		return p.triangulateGLTFLines(indices)
	case 4:
		return indices
	case 5:
		return p.triangulateGLTFStrip(indices)
	case 6:
		return p.triangulateGLTFFan(indices)
	default:
		return indices
	}
}

func (p *StaticModel3DProvider) triangulateGLTFStrip(indices []uint32) []uint32 {
	if len(indices) < 3 {
		return indices
	}

	tris := make([]uint32, 0, (len(indices)-2)*3)
	for i := 0; i < len(indices)-2; i++ {
		if i%2 == 0 {
			tris = append(tris, indices[i], indices[i+1], indices[i+2])
		} else {
			tris = append(tris, indices[i], indices[i+2], indices[i+1])
		}
	}
	return tris
}

func (p *StaticModel3DProvider) triangulateGLTFFan(indices []uint32) []uint32 {
	if len(indices) < 3 {
		return indices
	}

	tris := make([]uint32, 0, (len(indices)-2)*3)
	for i := 0; i < len(indices)-2; i++ {
		tris = append(tris, indices[0], indices[i+1], indices[i+2])
	}
	return tris
}

func (p *StaticModel3DProvider) triangulateGLTFLines(indices []uint32) []uint32 {
	tris := make([]uint32, 0)
	for i := 0; i < len(indices)-1; i += 2 {
		p1 := indices[i]
		p2 := indices[i+1]
		tris = append(tris, p1, p2, p1)
	}
	return tris
}

func (p *StaticModel3DProvider) modelInBounds(model *Model3D, bounds vec2d.Rect) bool {
	if bounds.Min[0] >= bounds.Max[0] || bounds.Min[1] >= bounds.Max[1] {
		return true
	}

	return model.Position[0] >= bounds.Min[0] &&
		model.Position[0] <= bounds.Max[0] &&
		model.Position[1] >= bounds.Min[1] &&
		model.Position[1] <= bounds.Max[1]
}

func (p *StaticModel3DProvider) Attribution() string {
	return ""
}

func (p *StaticModel3DProvider) Grid() *geo.TileGrid {
	return p.grid
}

func (p *StaticModel3DProvider) Bounds() vec2d.Rect {
	return p.bounds
}

func (p *StaticModel3DProvider) Srs() geo.Proj {
	return p.srs
}

func (p *StaticModel3DProvider) Clear() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.models = make(map[string]*Model3D)
}

func (p *StaticModel3DProvider) ModelCount() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return len(p.models)
}

func getFileExt(filepath string) string {
	idx := strings.LastIndex(filepath, ".")
	if idx == -1 {
		return ""
	}
	return strings.ToLower(filepath[idx:])
}

func getFileName(filepath string) string {
	idx := strings.LastIndex(filepath, "/")
	if idx == -1 {
		return filepath
	}
	return filepath[idx+1:]
}
