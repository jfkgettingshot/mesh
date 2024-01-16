package mesh

import (
	"errors"
	"fmt"
	"io"
	"strconv"
)

var (
	ErrInvalidOffset     = errors.New("expected [ while reading stream")
	ErrUnexpectedBracket = errors.New("while reading mesh got [")
)

type Mesh1 struct {
	FaceCount uint32
	Verts     []VertexV1
}

type MeshStream1 struct {
	Stream io.Reader
}

type VertexV1 struct {
	Px, Py, Pz float32
	Nx, Ny, Nz float32
	Tu, Tv     float32
}

type Vector3 struct {
	X, Y, Z float32
}

func writeBracketData(stream io.Writer, data string) error {
	if _, err := stream.Write([]byte{'['}); err != nil {
		return err
	}
	if _, err := stream.Write([]byte(data)); err != nil {
		return err
	}
	if _, err := stream.Write([]byte{']'}); err != nil {
		return err
	}
	return nil
}

func (V *VertexV1) WriteV1(stream io.Writer) error {
	if err := writeBracketData(stream, fmt.Sprintf("%f,%f,%f", V.Px, V.Py, V.Pz)); err != nil {
		return err
	}
	if err := writeBracketData(stream, fmt.Sprintf("%f,%f,%f", V.Nx, V.Ny, V.Nz)); err != nil {
		return err
	}
	if err := writeBracketData(stream, fmt.Sprintf("%f,%f,0", V.Tu, -V.Tv)); err != nil {
		return err
	}

	return nil
}

func (V *VertexV1) Modern() VertexModern {
	noRgba := VertexNoRgba{
		V.Px, V.Py, V.Pz,
		V.Nx, V.Ny, V.Nz,
		V.Tu, V.Tv,

		0, 0, 0, 0,
	}
	return noRgba.Modern()
}

func (V *VertexV1) Write(stream io.Writer) error {
	noRgba := V.NoColor()
	return noRgba.Write(stream)
}

func (V *VertexV1) NoColor() VertexNoRgba {
	return VertexNoRgba{
		V.Px, V.Py, V.Pz,
		V.Nx, V.Ny, V.Nz,
		V.Tu, V.Tv,

		0, 0, 0, 0,
	}
}

func (M *Mesh1) GetAllVerticies(faces []Face) []VertexV1 {
	vertBuffer := []VertexV1{}
	for i := 0; i < len(faces); i++ {
		face := faces[i]
		vertBuffer = append(vertBuffer, M.Verts[face.A])
		vertBuffer = append(vertBuffer, M.Verts[face.B])
		vertBuffer = append(vertBuffer, M.Verts[face.C])
	}

	return vertBuffer
}

func (V *Mesh1) Write(output io.Writer) error {
	headerOutput := fmt.Sprintf("version 1.01\n%d\n", len(V.Verts)/3)
	if _, err := output.Write([]byte(headerOutput)); err != nil {
		return err
	}
	for _, vertex := range V.Verts {
		if err := vertex.WriteV1(output); err != nil {
			return err
		}
	}
	return nil
}

func (V *Mesh1) NoColorVerts() []VertexNoRgba {
	newVerts := make([]VertexNoRgba, len(V.Verts))
	for i, vert := range V.Verts {
		newVerts[i] = vert.NoColor()
	}
	return newVerts
}
func (V *Mesh1) ModernVerts() []VertexModern {
	newVerts := make([]VertexModern, len(V.Verts))
	for i, vert := range V.Verts {
		newVerts[i] = vert.Modern()
	}
	return newVerts
}

func (M *Mesh1) ExportV1() *Mesh1 {
	return M
}

func (M *Mesh1) GenerateFaces(num_face uint32) []Face {
	faces := make([]Face, num_face)
	for i := uint32(0); i < num_face/3; i++ {
		faces[i].A = uint32((i * 3) + 0)
		faces[i].B = uint32((i * 3) + 1)
		faces[i].C = uint32((i * 3) + 2)
	}
	return faces
}

/* Converting V1 Forward is pointless and unsafe
func (M *Mesh1) ExportV2() Mesh2 {
	mesh2Header := MeshHeader2{
		Header2Size,
		VertexNoRgbaSize,
		FaceSize,
		uint32(len(M.Verts)),
		M.FaceCount,
	}

	newMesh := Mesh2NoRgba{
		Header: mesh2Header,
		Verts:  M.NoColorVerts(),
		Faces:  M.GenerateFaces(M.FaceCount),
	}

	return &newMesh
}

func (M *Mesh1) ExportV3() *Mesh3 {
	return M.ExportV2().ExportV3()
}

func (M *Mesh1) ExportV4() *Mesh4 {
	return M.ExportV3().ExportV4()
}
*/

func (S *MeshStream1) ReadNumber() (float32, error) {
	numberBytes := []byte{}
	for {
		currentByte := make([]byte, 1)
		if _, err := S.Stream.Read(currentByte); err != nil {
			return 0, err
		}
		if currentByte[0] == ']' || currentByte[0] == ',' {
			break
		} else if currentByte[0] == '[' {
			return 0, ErrUnexpectedBracket
		}
		numberBytes = append(numberBytes, currentByte[0])
	}
	if float, err := strconv.ParseFloat(string(numberBytes), 32); err == nil {
		return float32(float), nil
	} else {
		return 0, err
	}
}

func (S *MeshStream1) ReadLine() (string, error) {
	LineData := []byte{}
	for {
		currentByte := make([]byte, 1)
		if _, err := S.Stream.Read(currentByte); err != nil {
			return "", err
		}
		if currentByte[0] == '\n' {
			break
		}
		LineData = append(LineData, currentByte[0])
	}
	return string(LineData), nil
}

func (S *MeshStream1) ReadVector3() (*Vector3, error) {
	newData := make([]byte, 1)
	if _, err := S.Stream.Read(newData); err != nil {
		return nil, err
	} else if newData[0] != '[' {
		return nil, errors.Join(ErrInvalidOffset, fmt.Errorf("illegal character %s", string(newData)))
	}

	X, err := S.ReadNumber()
	if err != nil {
		return nil, err
	}
	Y, err := S.ReadNumber()
	if err != nil {
		return nil, err
	}
	Z, err := S.ReadNumber()
	if err != nil {
		return nil, err
	}

	return &Vector3{
		X, Y, Z,
	}, nil
}

func (S *MeshStream1) ReadVertex() (*VertexV1, error) {
	Position, err := S.ReadVector3()
	if err != nil {
		return nil, err
	}
	Normal, err := S.ReadVector3()
	if err != nil {
		return nil, err
	}
	Texture, err := S.ReadVector3()
	if err != nil {
		return nil, err
	}

	return &VertexV1{
		Position.X, Position.Y, Position.Z,
		Normal.X, Normal.Y, Normal.Z,
		Texture.X, Texture.Y,
	}, nil
}

/* This shit is so slow */
func (S *MeshStream1) LoadMesh() (*Mesh1, error) {
	FaceCountRaw, err := S.ReadLine()
	if err != nil {
		return nil, err
	}
	FaceCountI, err := strconv.Atoi(FaceCountRaw)
	if err != nil {
		return nil, err
	}
	FaceCount := uint32(FaceCountI)
	Verts := make([]VertexV1, FaceCount*3)
	for i := uint32(0); i < FaceCount*3; i++ {
		vert, err := S.ReadVertex()
		if err != nil {
			return nil, err
		}
		Verts[i] = *vert
	}

	return &Mesh1{FaceCount, Verts}, nil
}
