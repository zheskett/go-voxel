package voxel

import (
	te "github.com/zheskett/go-voxel/internal/tensor"
)

type Ray struct {
	Origin te.Vector3
	Dir    te.Vector3
	Tmax   float32
}

type RayHit struct {
	Hit    bool
	Color  [3]byte
	Normal te.Vector3
}
