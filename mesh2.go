package mesh

import (
	"encoding/binary"
	"io"
	"unsafe"
)

type MeshHeader2 struct {
	MeshHeaderSize uint16
	VertexSize     uint8
	FaceSize       uint8
	NumVerts       uint32
	NumFaces       uint32
}

type Mesh2 = Mesh

type Mesh2Rgba struct {
	Header MeshHeader2
	Verts  []VertexModern
	Faces  []Face
}
type Mesh2NoRgba struct {
	Header MeshHeader2
	Verts  []VertexNoRgba
	Faces  []Face
}

type MeshStream2 struct {
	Stream io.Reader
}

func (S *MeshStream2) ReadHeader() (*MeshHeader2, error) {
	var Header MeshHeader2
	if err := binary.Read(S.Stream, binary.LittleEndian, &Header); err != nil {
		return nil, err
	}
	return &Header, nil
}

func (S *MeshStream2) ReadValue(ptr any) error {
	return binary.Read(S.Stream, binary.LittleEndian, ptr)
}

func (S *MeshStream2) loadMeshNoRgba(header MeshHeader2) (*Mesh2NoRgba, error) {
	newMesh := Mesh2NoRgba{
		Header: header,
		Verts:  make([]VertexNoRgba, header.NumVerts),
		Faces:  make([]Face, header.NumFaces),
	}

	for i := uint32(0); i < header.NumVerts; i++ {
		if err := S.ReadValue(&newMesh.Verts[i]); err != nil {
			return nil, err
		}
	}
	for i := uint32(0); i < header.NumFaces; i++ {
		if err := S.ReadValue(&newMesh.Faces[i]); err != nil {
			return nil, err
		}
	}

	return &newMesh, nil
}
func (S *MeshStream2) loadMeshRgba(header MeshHeader2) (*Mesh2Rgba, error) {
	newMesh := Mesh2Rgba{
		Header: header,
		Verts:  make([]VertexModern, header.NumVerts),
		Faces:  make([]Face, header.NumFaces),
	}

	for i := uint32(0); i < header.NumVerts; i++ {
		if err := S.ReadValue(&newMesh.Verts[i]); err != nil {
			return nil, err
		}
	}
	for i := uint32(0); i < header.NumFaces; i++ {
		if err := S.ReadValue(&newMesh.Faces[i]); err != nil {
			return nil, err
		}
	}

	return &newMesh, nil
}

func (S *MeshStream2) LoadMesh() (Mesh2, error) {
	header, err := S.ReadHeader()
	if err != nil {
		return nil, err
	}

	switch header.VertexSize {
	case 36:
		mesh2, err := S.loadMeshNoRgba(*header)
		return mesh2, err
	default:
		mesh2, err := S.loadMeshRgba(*header)
		return mesh2, err
	}
}

func (M *Mesh2NoRgba) GetAllVerticies(faces []Face) []Vertex {
	vertBuffer := []Vertex{}
	for i := 0; i < len(faces); i++ {
		face := faces[i]
		vertBuffer = append(vertBuffer, &M.Verts[face.A])
		vertBuffer = append(vertBuffer, &M.Verts[face.B])
		vertBuffer = append(vertBuffer, &M.Verts[face.C])
	}

	return vertBuffer
}

func (S *Mesh2NoRgba) ExportV1() *Mesh1 {
	mesh1 := Mesh1{
		Verts: make([]VertexV1, len(S.Faces)*3),
	}
	for i, vertex := range S.GetAllVerticies(S.Faces) {
		mesh1.Verts[i] = vertex.Legacy()
	}

	return &mesh1
}

func (S *Mesh2NoRgba) ExportV2() Mesh2 {
	return S
}

func (S *Mesh2NoRgba) ExportV3() *Mesh3 {
	newHeader := MeshHeader3{
		MeshHeaderSize: uint16(Header3Size),
		VertexSize:     uint8(VertexModernSize),
		FaceSize:       uint8(FaceSize),
		SizeofLod:      uint16(unsafe.Sizeof(uint32(0))),

		NumLods:  0,
		NumVerts: S.Header.NumVerts,
		NumFaces: S.Header.NumFaces,
	}
	newMesh := Mesh3{
		Header: newHeader,
		Verts:  S.ConvertVerts(),
		Faces:  S.Faces,
		Lods:   make([]uint32, 0),
	}
	return &newMesh
}
func (S *Mesh2NoRgba) ConvertVerts() []VertexModern {
	newVerts := make([]VertexModern, S.Header.NumVerts)
	for i, value := range S.Verts {
		newVerts[i] = value.Modern()
	}
	return newVerts
}

func (S *Mesh2NoRgba) ExportV4() *Mesh4 {
	return S.ExportV3().ExportV4()
}

func (M *Mesh2Rgba) GetAllVerticies(faces []Face) []Vertex {
	vertBuffer := []Vertex{}
	for i := 0; i < len(faces); i++ {
		face := faces[i]
		vertBuffer = append(vertBuffer, &M.Verts[face.A])
		vertBuffer = append(vertBuffer, &M.Verts[face.B])
		vertBuffer = append(vertBuffer, &M.Verts[face.C])
	}

	return vertBuffer
}

func (S *Mesh2Rgba) ExportV1() *Mesh1 {
	mesh1 := Mesh1{
		FaceCount: S.Header.NumFaces,
		Verts:     make([]VertexV1, S.Header.NumFaces),
	}
	for i, vertex := range S.GetAllVerticies(S.Faces) {
		mesh1.Verts[i] = vertex.Legacy()
	}
	return &mesh1
}

func (S *Mesh2Rgba) ExportV2() Mesh2 {
	return S
}

func (S *Mesh2Rgba) ExportV3() *Mesh3 {
	newHeader := MeshHeader3{
		MeshHeaderSize: uint16(Header3Size),
		VertexSize:     uint8(VertexModernSize),
		FaceSize:       uint8(FaceSize),
		SizeofLod:      uint16(unsafe.Sizeof(uint32(0))),

		NumLods:  2,
		NumVerts: S.Header.NumVerts,
		NumFaces: S.Header.NumFaces,
	}
	newMesh := Mesh3{
		Header: newHeader,
		Verts:  S.Verts,
		Faces:  S.Faces,
		Lods:   []uint32{0, uint32(len(S.Faces))},
	}
	return &newMesh
}

func (S *Mesh2Rgba) ExportV4() *Mesh4 {
	return S.ExportV3().ExportV4()
}

func (M *Mesh2Rgba) Write(stream io.Writer) error {
	/* Write metadata bs */
	if _, err := stream.Write([]byte("version 2.00\n")); err != nil {
		return err
	} else if err := binary.Write(stream, binary.LittleEndian, M.Header); err != nil {
		return err
	}

	for _, vertex := range M.Verts {
		if err := binary.Write(stream, binary.LittleEndian, vertex); err != nil {
			return err
		}
	}
	for _, face := range M.Faces {
		if err := binary.Write(stream, binary.LittleEndian, face); err != nil {
			return err
		}
	}
	return nil
}

func (M *Mesh2NoRgba) Write(stream io.Writer) error {
	/* Write metadata bs */
	if _, err := stream.Write([]byte("version 2.00\n")); err != nil {
		return err
	} else if err := binary.Write(stream, binary.LittleEndian, M.Header); err != nil {
		return err
	}

	for _, vertex := range M.Verts {
		if err := binary.Write(stream, binary.LittleEndian, vertex); err != nil {
			return err
		}
	}
	for _, face := range M.Faces {
		if err := binary.Write(stream, binary.LittleEndian, face); err != nil {
			return err
		}
	}
	return nil
}
