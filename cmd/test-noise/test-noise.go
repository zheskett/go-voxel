package main

import (
	_ "fmt"

	"github.com/zheskett/go-voxel/internal/noise"
)

func main() {
	per, err := noise.GenPerlin3D(256)
	if err != nil {
		panic(err)
	}
	noise.Draw(&per, "noise.png", 2048, 0, 0.01)
	fbm := noise.FBM3D{Source: &per, Octaves: 8, H: 0.5}
	noise.Draw(&fbm, "FBM.png", 2048, 0, 0.01)
	applied := noise.FnNoise3D{Source: &fbm, Fn: noise.CuRat}
	noise.Draw(&applied, "applied.png", 2048, 0, 0.01)
}
