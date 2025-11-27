package render

import "github.com/go-gl/mathgl/mgl32"

type Ray struct {
	origin mgl32.Vec3
	direc  mgl32.Vec3
	tmax   float32
}

type RayHit struct {
	hit    bool
	color  Color
	normal mgl32.Vec3
}

// need to learn how these work
type CastRay interface {
	march() RayHit
}
