package main

import (
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/zheskett/go-voxel/internal/parser"
	vxl "github.com/zheskett/go-voxel/internal/voxel"
)

func main() {
	f, err := os.Create("cpu.pprof")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	var path string = "assets/bunny.obj"
	//fmt.Printf("Location of OBJ: ")
	//fmt.Scanln(&path)
	obj, err := parser.ParseObj(path, false, true, false)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("Obj: \n%v\n", obj)
	// fmt.Printf("Vert Count: %v, Face Count: %v\n", len(obj.Vertices), len(obj.Faces))

	vxl.Voxelize(obj, vxl.T26, 2048, [3]byte{255, 255, 255})
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("vObj: \n%v\n", vObj)
	//vObj.Flip(false, true, false)
	fmt.Println("Done")
}
