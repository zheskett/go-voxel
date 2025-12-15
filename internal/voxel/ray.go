package voxel

import (
	te "github.com/zheskett/go-voxel/internal/tensor"
)

const (
	// All raymarched position data is ambiguous as it gives the face lies exactly on
	// the shared face of two neighbor voxels. This distance offset is used in:
	// vox = hit_position - hit_normal * VoxelRayDelta
	// to find the actual voxel the ray hit
	VoxelRayDelta = 1e-2
)

// Enum for axis
type axis uint8

const (
	axisX axis = iota
	axisY
	axisZ
	none
)

type Ray struct {
	Origin te.Vector3
	Dir    te.Vector3
	Tmax   float32
}

// Returned information after a ray is cast into the scene
type RayHit struct {
	Hit      bool
	Time     float32
	Color    [3]byte
	IntPos   [3]int
	Position te.Vector3
	Normal   te.Vector3
	Voxel    *Voxel
}

// The default method that should be used anytime a raycast is made
type Marchable interface {
	MarchRay(ray Ray) RayHit
}

type MarchData struct {
	Tmin float32
	Tmax float32
	Time float32
	Inv  te.Vector3
	Side axis
}

func MarchDataInit(tmin, tmax float32, ray Ray) MarchData {
	return MarchData{Tmin: tmin, Tmax: tmax, Time: tmin, Inv: ray.Dir.Inv(), Side: none}
}

// The sort of 'internal' marching method. Shouldn't be called directly for
// marching rays, rather any time that can be marched should implement this
type StateMachineMarch interface {
	StateMarchRay(ray Ray, data MarchData) RayHit
}
