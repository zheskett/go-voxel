package main

import (
	_ "fmt"

	"github.com/chewxy/math32"
	"github.com/zheskett/go-voxel/internal/noise"
)

func main() {
	perlin, err := noise.GenPerlin3D(256)
	if err != nil {
		panic(err)
	}
	fbm := &noise.FBM3D{Source: &perlin, Octaves: 7, H: 1}
	fn := &noise.FnNoise3D{Source: fbm, Fn: noise.CuRat}
	fn = &noise.FnNoise3D{Source: fn, Fn: func(t float32) float32 {
		return 0.5 * (math32.Pow(t, 14) + math32.Pow(t, 0.75))
	}}
	noise.Draw(fn, "applied.png", 2048, 0, 0.01)
}
