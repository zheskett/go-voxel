package voxel

import (
	_ "github.com/zheskett/go-voxel/internal/tensor"
)

type TerrainOpts struct {
	X, Y, Z                            int32
	minTerrainHeight, maxTerrainHeight int32
}

func GenTerrainObj() VoxelObj {
	return VoxelObj{}
}
