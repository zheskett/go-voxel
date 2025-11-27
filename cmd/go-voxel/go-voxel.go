package main

import (
	"fmt"
	"math"
	"runtime"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/zheskett/go-voxel/internal/render"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// This is meantioned in the usage example on github.
	runtime.LockOSThread()
}

func main() {
	rm := render.RenderManagerInit()
	cam := render.CameraInit()
	cam.Movespeed = 0.1
	cam.Lookspeed = 0.01

	for {
		RenderDebugPixel3D(&rm.Pixels, &cam, mgl32.Vec3{0, 0, 3})
		rm.Render()
		rm.CheckExit()
		render.UpdateCamInputGLFW(&cam, rm.Window)

		fmt.Printf("cam vectors: %+v\n", cam)
	}
}

func RenderDebugPixel3D(pix *render.Pixels, cam *render.Camera, point mgl32.Vec3) {
	fov := 90.0
	scale := 1.0 / math.Tan(fov/2.0)
	pix.FillPixels(0, 0, 0)

	point = point.Sub(cam.Pos)
	rel := mgl32.Vec3{
		point.Dot(cam.Rvec),
		point.Dot(cam.Uvec),
		point.Dot(cam.Fvec),
	}
	hh := float32(pix.Height / 2)
	hw := float32(pix.Width / 2)

	sx := rel[0] / rel[2] * float32(scale)
	sy := -rel[1] / rel[2] * float32(scale)

	ssx := int(sx*hw + hw)
	ssy := int(sy*hh + hh)

	if ssx > 0 && ssx < pix.Width && ssy > 0 && ssy < pix.Height {
		pix.SetPixel(ssx, ssy, 0, 255, 255)
	}
}
