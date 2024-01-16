package mesh

import (
	"encoding/binary"
	"io"
	"unsafe"
)

type ushort = uint16

type Mesh4 struct {
	Header      MeshHeader4
	Verts       []VertexModern
	Envelopes   []Envelope
	Faces       []Face
	Lods        []uint32
	Bones       []Bone
	NameTable   []byte
	MeshSubsets []MeshSubset
}

type MeshHeader4 struct {
	SizeOf_MeshHeader        ushort
	LodType                  ushort
	NumVerts                 uint32
	NumFaces                 uint32
	NumLods                  ushort
	NumBones                 ushort
	SizeOf_bone_names_Buffer uint32
	NumSubsets               ushort
	NumHighQualityLods       byte
	Unused                   byte
}

type MeshStream4 struct {
	Stream io.Reader
}

type Envelope struct {
	Bones   [4]byte
	Weights [4]byte
}

type Bone struct {
	BoneNameIndex  uint32
	ParentIndex    ushort
	LodParentIndex ushort
	Culling        float32

	R00 float32
	R01 float32
	R02 float32
	R10 float32
	R11 float32
	R12 float32
	R20 float32
	R21 float32
	R22 float32

	X float32
	Y float32
	Z float32
}

type MeshSubset struct {
	FacesBegin       uint32
	FacesLength      uint32
	VertsBegin       uint32
	VertsLength      uint32
	NumBonesIndicies uint32
	BoneIndicies     [26]ushort
}

func (S *MeshStream4) ReadHeader() (*MeshHeader4, error) {
	var Header MeshHeader4
	if err := binary.Read(S.Stream, binary.LittleEndian, &Header); err != nil {
		return nil, err
	}
	return &Header, nil
}

func (S *MeshStream4) ReadValue(ptr any) error {
	return binary.Read(S.Stream, binary.LittleEndian, ptr)
}

func (S *MeshStream4) LoadMesh() (*Mesh4, error) {
	header, err := S.ReadHeader()
	if err != nil {
		return nil, err
	}

	newMesh := Mesh4{
		Header:      *header,
		Verts:       make([]VertexModern, header.NumVerts),
		Envelopes:   make([]Envelope, header.NumVerts),
		Faces:       make([]Face, header.NumFaces),
		Lods:        make([]uint32, header.NumLods),
		Bones:       make([]Bone, header.NumBones),
		NameTable:   make([]byte, header.SizeOf_bone_names_Buffer),
		MeshSubsets: make([]MeshSubset, header.NumSubsets),
	}

	for i := uint32(0); i < header.NumVerts; i++ {
		if err := S.ReadValue(&newMesh.Verts[i]); err != nil {
			return nil, err
		}
	}
	if header.NumBones > 0 {
		for i := uint32(0); i < header.NumVerts; i++ {
			if err := S.ReadValue(&newMesh.Envelopes[i]); err != nil {
				return nil, err
			}
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

	for i := uint16(0); i < header.NumBones; i++ {
		if err := S.ReadValue(&newMesh.Bones[i]); err != nil {
			return nil, err
		}

	}

	if _, err := S.Stream.Read(newMesh.NameTable); err != nil {

	}

	for i := uint16(0); i < header.NumSubsets; i++ {
		if err := S.ReadValue(&newMesh.MeshSubsets[i]); err != nil {
			return nil, err
		}
	}

	return &newMesh, nil
}

func (M *Mesh4) GetNormalFaces() []Face {
	if len(M.Lods) > 0 {
		return M.Faces[:M.Lods[1]]
	}
	return M.Faces
}

func (M *Mesh4) GetAllVerticies(faces []Face) []VertexModern {
	vertBuffer := []VertexModern{}
	for i := 0; i < len(faces); i++ {
		face := faces[i]
		vertBuffer = append(vertBuffer, M.Verts[face.A])
		vertBuffer = append(vertBuffer, M.Verts[face.B])
		vertBuffer = append(vertBuffer, M.Verts[face.C])
	}

	return vertBuffer
}

func (M *Mesh4) Write(stream io.Writer) error {
	if _, err := stream.Write([]byte("version 4.00\n")); err != nil {
		return err
	} else if err := binary.Write(stream, binary.LittleEndian, M.Header); err != nil {
		return err
	}

	for i := uint32(0); i < M.Header.NumVerts; i++ {
		if err := binary.Write(stream, binary.LittleEndian, M.Verts[i]); err != nil {
			return err
		}
	}
	if M.Header.NumBones > 0 {
		for i := uint32(0); i < M.Header.NumVerts; i++ {
			if err := binary.Write(stream, binary.LittleEndian, M.Envelopes[i]); err != nil {
				return err
			}
		}
	}

	for i := uint32(0); i < M.Header.NumFaces; i++ {
		if err := binary.Write(stream, binary.LittleEndian, M.Faces[i]); err != nil {
			return err
		}
	}

	for i := uint16(0); i < M.Header.NumLods; i++ {
		if err := binary.Write(stream, binary.LittleEndian, M.Lods[i]); err != nil {
			return err
		}
	}

	if M.Header.NumBones > 0 {
		for i := uint16(0); i < M.Header.NumBones; i++ {
			if err := binary.Write(stream, binary.LittleEndian, M.Bones[i]); err != nil {
				return err
			}
		}
		if err := binary.Write(stream, binary.LittleEndian, M.NameTable); err != nil {
			return err
		}
	}

	for i := uint16(0); i < M.Header.NumSubsets; i++ {
		if err := binary.Write(stream, binary.LittleEndian, M.MeshSubsets[i]); err != nil {
			return err
		}
	}

	return nil
}

/* Using mesh v1.01 since i dont feel like descaling */
func (M *Mesh4) ExportV1() *Mesh1 {
	return M.ExportV2().ExportV1()
}

func (M *Mesh4) ExportV2() Mesh2 {
	return M.ExportV3().ExportV2()
}

func (M *Mesh4) ExportV3() *Mesh3 {
	mesh3Header := MeshHeader3{
		Header3Size,
		VertexModernSize,
		FaceSize,
		uint16(unsafe.Sizeof(uint32(0))),
		M.Header.NumLods,
		uint32(len(M.Verts)),
		uint32(len(M.Faces)),
	}

	newMesh := Mesh3{
		Header: mesh3Header,
		Verts:  M.Verts,
		Faces:  M.Faces,
		Lods:   M.Lods,
	}

	return &newMesh
}

func (M *Mesh4) ExportV4() *Mesh4 {
	return M
}
