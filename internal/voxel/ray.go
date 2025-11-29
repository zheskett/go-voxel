package voxel

import (
	"github.com/go-gl/mathgl/mgl32"
)

type Ray struct {
	Origin mgl32.Vec3
	Direc  mgl32.Vec3
	Tmax   float32
}

type RayHit struct {
	Hit    bool
	Color  [3]byte
	Normal mgl32.Vec3
}
