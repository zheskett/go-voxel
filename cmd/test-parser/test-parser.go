package main

import (
	"fmt"

	"github.com/zheskett/go-voxel/internal/parser"
)

func main() {
	var path string
	fmt.Printf("Location of OBJ: ")
	fmt.Scanln(&path)
	obj, err := parser.ParseObj(path)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Obj: \n%v\n", obj)
	fmt.Printf("Vert Count: %v, Face Count: %v\n", len(obj.Vertices), len(obj.FaceVertices))
}
