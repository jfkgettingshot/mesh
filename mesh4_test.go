package mesh_test

import (
	"github.com/MojaveMF/mesh"
	"os"
	"testing"
)

func TestConvertV4_V2(t *testing.T) {
	file, err := os.Open("./testdata/output.v4")
	if err != nil {
		t.Error(err)
		return
	}
	defer file.Close()
	bytes := make([]byte, 13)
	file.Read(bytes)

	meshStream := mesh.MeshStream4{file}
	meshData, err := meshStream.LoadMesh()
	if err != nil {
		t.Error(err)
		return
	}

	output, err := os.Create("./output.v4v2")
	if err != nil {
		t.Error(err)
		return
	}
	defer output.Close()

	meshData.ExportV2().Write(output)
}

func TestConvertV4_V3(t *testing.T) {
	file, err := os.Open("./testdata/output.v4")
	if err != nil {
		t.Error(err)
		return
	}
	defer file.Close()
	bytes := make([]byte, 13)
	file.Read(bytes)

	meshStream := mesh.MeshStream4{file}
	meshData, err := meshStream.LoadMesh()
	if err != nil {
		t.Error(err)
		return
	}

	output, err := os.Create("./output.v4v3")
	if err != nil {
		t.Error(err)
		return
	}
	defer output.Close()

	meshData.ExportV3().Write(output)
}

func TestConvertV4_V4(t *testing.T) {
	file, err := os.Open("./testdata/output.v4")
	if err != nil {
		t.Error(err)
		return
	}
	defer file.Close()
	bytes := make([]byte, 13)
	file.Read(bytes)

	meshStream := mesh.MeshStream4{file}
	meshData, err := meshStream.LoadMesh()
	if err != nil {
		t.Error(err)
		return
	}

	output, err := os.Create("./output.v4v4")
	if err != nil {
		t.Error(err)
		return
	}
	defer output.Close()

	meshData.ExportV4().Write(output)
}
