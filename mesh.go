package mesh

import (
	"encoding/binary"
	"errors"
	"io"
	"unsafe"
)

const (
	MeshVersion1 uint8 = 1 << iota
	MeshVersion1_01
	MeshVersion2
	MeshVersion3
	MeshVersion4
	MeshVersion4_1
)

type Mesh interface {
	ExportV1() *Mesh1
	ExportV2() Mesh2
	ExportV3() *Mesh3
	ExportV4() *Mesh4
	Write(io.Writer) error
}

var (
	/* Idk why this is so unsafe :rofl: */
	VertexNoRgbaSize = uint8(unsafe.Sizeof(VertexNoRgba{}))
	VertexModernSize = uint8(unsafe.Sizeof(VertexModern{}))
	FaceSize         = uint8(unsafe.Sizeof(Face{}))
	Header2Size      = uint16(unsafe.Sizeof(MeshHeader2{}))
	Header3Size      = uint16(unsafe.Sizeof(MeshHeader3{}))
	Header4Size      = uint16(unsafe.Sizeof(MeshHeader4{}))
)

var (
	ErrUnkownMeshVersion = errors.New("mesh version read from buffer is unkown")
	ErrMeshVersion1      = errors.New("mesh is version 1 this cant be parsed safely")
	ErrBadMeshVersion    = errors.New("mesh version is not known")
)

type Vertex interface {
	WriteV1(io.Writer) error
	Write(io.Writer) error
	Modern() VertexModern
	NoColor() VertexNoRgba
	Legacy() VertexV1
}

/* same from 2-4 */
type VertexModern struct {
	Px, Py, Pz float32
	Nx, Ny, Nz float32
	Tu, Tv     float32

	Tx, Ty, Tz, Ts int8
	R, G, B, A     byte
}

/* same from 2-4 */
type VertexNoRgba struct {
	Px, Py, Pz float32
	Nx, Ny, Nz float32
	Tu, Tv     float32

	Tx, Ty, Tz, Ts int8
}

/* Same from 2-4 */
type Face struct {
	A uint32
	B uint32
	C uint32
}

func (V *VertexModern) NoColor() VertexNoRgba {
	return VertexNoRgba{
		V.Px, V.Py, V.Pz,
		V.Nx, V.Ny, V.Nz,
		V.Tu, V.Tv,

		V.Tx, V.Ty, V.Tz, V.Ts,
	}
}

func (V *VertexModern) Legacy() VertexV1 {
	return VertexV1{
		V.Px, V.Py, V.Pz,
		V.Nx, V.Ny, V.Nz,
		V.Tu, V.Tv,
	}
}

func (V *VertexModern) WriteV1(stream io.Writer) error {
	legacy := V.Legacy()
	return legacy.WriteV1(stream)
}

func (V *VertexModern) Write(stream io.Writer) error {
	return binary.Write(stream, binary.LittleEndian, V)
}

func (V *VertexModern) Modern() VertexModern {
	clone := *V
	return clone
}

func (V *VertexNoRgba) Legacy() VertexV1 {
	return VertexV1{
		V.Px, V.Py, V.Pz,
		V.Nx, V.Ny, V.Nz,
		V.Tu, V.Tv,
	}
}

func (V *VertexNoRgba) NoColor() VertexNoRgba {
	clone := *V
	return clone
}

func (V *VertexNoRgba) Modern() VertexModern {
	return VertexModern{
		V.Px, V.Py, V.Pz,
		V.Nx, V.Ny, V.Nz,
		V.Tu, V.Tv,

		V.Tx, V.Ty, V.Tz, V.Ts,
		255, 255, 255, 0,
	}
}

func (V *VertexNoRgba) WriteV1(stream io.Writer) error {
	legacy := V.Legacy()
	return legacy.WriteV1(stream)
}

func (V *VertexNoRgba) Write(stream io.Writer) error {
	return binary.Write(stream, binary.LittleEndian, V)
}

func WriteValues(stream io.Writer, values ...any) error {
	for _, value := range values {
		if err := binary.Write(stream, binary.LittleEndian, value); err != nil {
			return err
		}
	}
	return nil
}

func ReadLine(stream io.Reader) (string, error) {
	LineData := []byte{}
	for {
		currentByte := make([]byte, 1)
		if _, err := stream.Read(currentByte); err != nil {
			return "", err
		}
		if currentByte[0] == '\n' {
			break
		}
		LineData = append(LineData, currentByte[0])
	}
	return string(LineData), nil
}

func MeshDecodeLayer(MaxVersion uint8, ExportVersion uint8) func(io.Reader, io.Writer) error {
	return func(rc io.Reader, wc io.Writer) error {
		meshVersion, err := MeshVersion(rc)
		if err != nil {
			return err
		}
		if meshVersion > MaxVersion {
			parsedMesh, err := decodeMesh(rc, meshVersion)
			if err != nil {
				return err
			}
			newMesh := EncodeMeshVersion(parsedMesh, ExportVersion)
			if newMesh == nil {
				return ErrBadMeshVersion
			}
			return newMesh.Write(wc)
		} else {
			meshHeader, err := MeshHeader(meshVersion)
			if err != nil {
				return err
			}
			if _, err := wc.Write([]byte(meshHeader + "\n")); err != nil {
				return err
			}
			_, err = io.Copy(wc, rc)
			return err
		}

	}
}
func MeshHeader(meshVersion uint8) (string, error) {
	switch meshVersion {
	case MeshVersion1:
		return "version 1.00", nil
	case MeshVersion1_01:
		return "version 1.01", nil
	case MeshVersion2:
		return "version 2.00", nil
	case MeshVersion3:
		return "version 3.00", nil
	case MeshVersion4:
		return "version 4.00", nil
	case MeshVersion4_1:
		return "version 4.01", nil
	default:
		return "", ErrUnkownMeshVersion
	}
}

func MeshVersion(stream io.Reader) (uint8, error) {
	meshVersion, err := ReadLine(stream)
	if err != nil {
		return 0, err
	}

	switch meshVersion {
	case "version 1.00":
		return MeshVersion1, nil
	case "version 1.01":
		return MeshVersion1_01, nil
	case "version 2.00":
		return MeshVersion2, nil
	case "version 3.00":
		return MeshVersion3, nil
	case "version 4.00":
		return MeshVersion4, nil
	case "version 4.01":
		return MeshVersion4_1, nil
	default:
		return 0, ErrUnkownMeshVersion
	}

}

func DecodeMesh(stream io.Reader) (Mesh, error) {
	version, err := MeshVersion(stream)
	if err != nil {
		return nil, err
	}

	return decodeMesh(stream, version)
}

func decodeMesh(stream io.Reader, version uint8) (Mesh, error) {
	switch version {
	case MeshVersion1, MeshVersion1_01:
		return nil, ErrMeshVersion1
	case MeshVersion2:
		stream2 := MeshStream2{stream}
		return stream2.LoadMesh()
	case MeshVersion3:
		stream3 := MeshStream3{stream}
		return stream3.LoadMesh()
	case MeshVersion4, MeshVersion4_1:
		stream4 := MeshStream4{stream}
		return stream4.LoadMesh()
	default:
		return nil, ErrUnkownMeshVersion
	}
}

func EncodeMeshVersion(mesh Mesh, version uint8) Mesh {
	switch version {
	case MeshVersion2:
		return mesh.ExportV2()
	case MeshVersion3:
		return mesh.ExportV3()
	case MeshVersion4, MeshVersion4_1:
		return mesh.ExportV4()
	default:
		return nil
	}
}
