package mesh_test

import (
	"fmt"
	"mesh"
	"os"
	"testing"
)

func TestDecode(t *testing.T) {
	stream2, err := os.Open("./testdata/output.v2")
	defer stream2.Close()
	if err != nil {
		t.Error(err)
		return
	}
	stream3, err := os.Open("./testdata/output.v3")
	defer stream3.Close()
	if err != nil {
		t.Error(err)
		return
	}
	stream4, err := os.Open("./testdata/output.v4")
	defer stream4.Close()
	if err != nil {
		t.Error(err)
		return
	}

	mesh2, err := mesh.DecodeMesh(stream2)
	if err != nil {
		t.Error(err)
		return
	}
	mesh3, err := mesh.DecodeMesh(stream3)
	if err != nil {
		t.Error(err)
		return
	}
	mesh4, err := mesh.DecodeMesh(stream4)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(mesh2, mesh3, mesh4)
}
