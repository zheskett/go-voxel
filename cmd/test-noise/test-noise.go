package main

import (
	_ "fmt"

	"github.com/zheskett/go-voxel/internal/noise"
)

func main() {
	noise, err := noise.GenPerlin2D(256)
	if err != nil {
		panic(err)
	}
	noise.Draw("noise.png", 2048, 0, 0.01)
	noise.DrawFBM("FMB.png", 2048, 0, 5, 1, 0.01)
}
