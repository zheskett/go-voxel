package main

import (
	"fmt"

	"github.com/zheskett/go-voxel/internal/voxel"
)

func main() {
	tree := voxel.BrickTreeInit(100, 100, 100)
	for i := range 100 {
		tree.Insert(i, i, i, 0, 0, 0)
	}

	fmt.Printf("done")
}
