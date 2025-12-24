package voxel

import (
	"fmt"
	"image/color"
	"math/rand/v2"

	"github.com/chewxy/math32"
	"github.com/zheskett/go-voxel/internal/noise"
	te "github.com/zheskett/go-voxel/internal/tensor"
)

const (
	TerrainGrassIdx = iota
	TerrainStoneIdx
	TerrainSnowIdx
	terrainMaxIdx = TerrainSnowIdx + 1
)

type TerrainOpts struct {
	X, Y, Z              int16
	MinHeight, MaxHeight int16
	SnowStart            int16
	SnowForce            int16
	SampleDist           float32
	Offset               te.Vector3
	Source               noise.Noise3D
	ColorPalette         [terrainMaxIdx]color.RGBA
}

func DefaultTerrainNoise() (noise.Noise3D, error) {
	perlin, err := noise.GenPerlin3D(256)
	if err != nil {
		return nil, err
	}

	fbm := &noise.FBM3D{Source: &perlin, Octaves: 7, H: 1}
	fn := &noise.FnNoise3D{Source: fbm, Fn: noise.CuRat}
	fn = &noise.FnNoise3D{Source: fn, Fn: func(t float32) float32 {
		return 0.5 * (math32.Pow(t, 14) + math32.Pow(t, 0.75))
	}}
	return fn, nil
}

func GenTerrainObj(opts TerrainOpts) (VoxelObj, error) {
	// Error handling
	if opts.X < 1 || opts.Y < 1 || opts.Z < 1 {
		return VoxelObj{}, fmt.Errorf("values X,Y,Z must be positive")
	}
	heightDiff := opts.MaxHeight - opts.MinHeight
	if heightDiff <= 0 {
		return VoxelObj{}, fmt.Errorf("difference in terrain height must be positive")
	}
	snowDiff := opts.SnowForce - opts.SnowStart

	// Rough estimate of map size, may be to big or too small
	vObj := VoxelObj{
		X: opts.X, Y: opts.Y, Z: opts.Z,
		Voxels: make(map[[3]int16]byte,
			int(opts.X)*int(opts.Y)*int((heightDiff/2))),
		ColorPalette: opts.ColorPalette[:],
	}

	for x := range vObj.X {
		xStep := float32(x) * opts.SampleDist
		for z := range vObj.Z {
			// Use z as fixed direction in noise sample
			yStep := float32(z) * opts.SampleDist
			pos := opts.Offset.Add(te.Vec3(xStep, yStep, 0))
			samp := opts.Source.Sample(pos)
			height := int16(math32.Round(samp*float32(heightDiff))) + opts.MinHeight
			var topIdx byte = TerrainGrassIdx
			if height >= opts.SnowForce {
				topIdx = TerrainSnowIdx
			} else if height >= opts.SnowStart {
				distFromForce := opts.SnowForce - height
				chance := 1 - (float32(distFromForce-1) / float32(snowDiff))
				randF := rand.Float32()
				if randF <= chance {
					topIdx = TerrainSnowIdx
				}
			}

			// I guess i gotta invert Y
			vObj.Voxels[[3]int16{x, vObj.Y - height - 1, z}] = topIdx
			var belowIdx byte = TerrainStoneIdx
			for y := range height {
				vObj.Voxels[[3]int16{x, vObj.Y - y - 1, z}] = belowIdx
			}
		}
	}

	return vObj, nil
}
