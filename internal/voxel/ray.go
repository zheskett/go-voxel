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
	Time   float32
	Color  [3]byte
	Normal mgl32.Vec3
}
