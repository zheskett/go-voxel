package main

import (
	"fmt"

	"github.com/zheskett/go-voxel/internal/voxel"
	"github.com/zheskett/go-voxel/pkg/voxparse"
)

func main() {
	var path string = "assets/bunny.obj"
	fmt.Printf("Location of VOX: ")
	fmt.Scanln(&path)
	vox, err := voxparse.Parse(path)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Vox: \n%v\n", vox)

	vObj, err := voxel.ConvertVox(vox, false, false, false)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", vObj)
}
