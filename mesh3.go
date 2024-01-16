package mesh

import (
	"encoding/binary"
	"io"
)

type MeshHeader3 struct {
	MeshHeaderSize ushort
	VertexSize     uint8
	FaceSize       uint8
	SizeofLod      ushort

	NumLods  uint16
	NumVerts uint32
	NumFaces uint32
}

type Mesh3 struct {
	Header MeshHeader3
	Verts  []VertexModern
	Faces  []Face
	Lods   []uint32
}

type MeshStream3 struct {
	Stream io.Reader
}

func (S *MeshStream3) ReadHeader() (*MeshHeader3, error) {
	var Header MeshHeader3
	if err := binary.Read(S.Stream, binary.LittleEndian, &Header); err != nil {
		return nil, err
	}
	return &Header, nil
}

func (S *MeshStream3) ReadValue(ptr any) error {
	return binary.Read(S.Stream, binary.LittleEndian, ptr)
}

func (S *MeshStream3) LoadMesh() (*Mesh3, error) {
	header, err := S.ReadHeader()
	if err != nil {
		return nil, err
	}

	newMesh := Mesh3{
		Header: *header,
		Verts:  make([]VertexModern, header.NumVerts),
		Faces:  make([]Face, header.NumFaces),
		Lods:   make([]uint32, header.NumLods),
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

	for i := uint16(0); i < header.NumLods; i++ {
		if err := S.ReadValue(&newMesh.Lods[i]); err != nil {
			return nil, err
		}
	}
	return &newMesh, nil
}

func (M *Mesh3) Write(stream io.Writer) error {
	/* Write metadata bs */
	if _, err := stream.Write([]byte("version 3.00\n")); err != nil {
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
	for _, lodData := range M.Lods {
		if err := binary.Write(stream, binary.LittleEndian, lodData); err != nil {
			return err
		}
	}

	return nil
}

func (M *Mesh3) GetNormalFaces() []Face {
	if len(M.Lods) > 1 {
		return M.Faces[:M.Lods[1]]
	}
	return M.Faces
}

func (M *Mesh3) GetAllVerticies(faces []Face) []VertexModern {
	vertBuffer := []VertexModern{}
	for i := 0; i < len(faces); i++ {
		face := faces[i]
		vertBuffer = append(vertBuffer, M.Verts[face.A])
		vertBuffer = append(vertBuffer, M.Verts[face.B])
		vertBuffer = append(vertBuffer, M.Verts[face.C])
	}

	return vertBuffer
}

func (M *Mesh3) ExportV1() *Mesh1 {
	return M.ExportV2().ExportV1()
}

func (M *Mesh3) ExportV2() Mesh2 {
	faces := M.GetNormalFaces()

	mesh2Header := MeshHeader2{
		Header2Size,
		VertexModernSize,
		FaceSize,
		uint32(len(M.Verts)),
		uint32(len(faces)),
	}

	newMesh := Mesh2Rgba{
		Header: mesh2Header,
		Verts:  M.Verts,
		Faces:  faces,
	}

	return &newMesh
}

func (M *Mesh3) ExportV3() *Mesh3 {
	return M
}

func (M *Mesh3) ExportV4() *Mesh4 {
	newHeader := MeshHeader4{
		SizeOf_MeshHeader:        Header4Size,
		LodType:                  0,
		NumVerts:                 M.Header.NumVerts,
		NumFaces:                 M.Header.NumFaces,
		NumLods:                  M.Header.NumLods,
		NumBones:                 0,
		SizeOf_bone_names_Buffer: 0,
		NumSubsets:               0,
		NumHighQualityLods:       0,
		Unused:                   0,
	}
	newMesh := Mesh4{
		Header:      newHeader,
		Verts:       M.Verts,
		Envelopes:   make([]Envelope, 0),
		Faces:       M.Faces,
		Lods:        M.Lods,
		Bones:       make([]Bone, 0),
		NameTable:   make([]byte, 0),
		MeshSubsets: make([]MeshSubset, 0),
	}

	return &newMesh
}
