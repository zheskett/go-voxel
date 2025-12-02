package voxel

import (
	te "github.com/zheskett/go-voxel/internal/tensor"
)

const (
	// 	All raymarched position data is ambiguous as it gives the face lies exactly on
	// the shared face of two neighbor voxels. This distance offset is used in:
	// vox = hit_position - hit_normal * VoxelRayDelta
	// to find the actual voxel the ray hit
	VoxelRayDelta = 0.25
)

type Ray struct {
	Origin te.Vector3
	Dir    te.Vector3
	Tmax   float32
}

type RayHit struct {
	Hit      bool
	Time     float32
	Color    [3]byte
	Normal   te.Vector3
	Position te.Vector3
}
