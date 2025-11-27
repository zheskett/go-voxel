package main

import (
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
	cam.Movespeed = 0.25
	cam.Lookspeed = 0.05
	cam.Fov = 90
	cam.Aspect = float32(rm.Pixels.Width) / float32(rm.Pixels.Height)

	for {
		RenderDebugRasterTri(&rm.Pixels, &cam)
		rm.Render()
		rm.CheckExit()
		render.UpdateCamInputGLFW(&cam, rm.Window)
	}
}

func RenderDebugRasterTri(pix *render.Pixels, cam *render.Camera) {
	pix.FillPixels(15, 25, 40)
	vpos := []mgl32.Vec3{
		{-0.5, 0.0, 3.0},
		{0.5, 0.0, 3.0},
		{0.0, 1.0, 3.0},
	}
	vcol := []mgl32.Vec3{
		{1.0, 0.7, 0.0},
		{0.0, 1.0, 0.7},
		{0.7, 0.0, 1.0},
	}
	scale := 1.0 / math.Tan(float64(cam.Fov/2.0))
	hw, hh := float32(pix.Width/2), float32(pix.Height/2)

	for i := range vpos {
		// translate
		vpos[i] = vpos[i].Sub(cam.Pos)
		// relative to camera
		vpos[i] = mgl32.Vec3{
			vpos[i].Dot(cam.Rvec),
			vpos[i].Dot(cam.Uvec),
			vpos[i].Dot(cam.Fvec),
		}
		// perspective projection
		vpos[i][0] /= vpos[i][2] * float32(scale) * cam.Aspect
		vpos[i][1] /= vpos[i][2] * float32(scale)

		// get in ndc
		vpos[i][0] = vpos[i][0]*hw + hw
		vpos[i][1] = -vpos[i][1]*hh + hh
	}

	minx, maxx := 1e9, -1e9
	miny, maxy := 1e9, -1e9
	for _, vert := range vpos {
		minx = math.Min(minx, float64(vert[0]))
		miny = math.Min(miny, float64(vert[1]))
		maxx = math.Max(maxx, float64(vert[0]))
		maxy = math.Max(maxy, float64(vert[1]))
	}
	minx, maxx, miny, maxy = math.Max(minx, 0.0), math.Min(maxx, float64(pix.Width)), math.Max(miny, 0.0), math.Min(maxy, float64(pix.Height))

	a, b, c := mgl32.Vec2{vpos[0][0], vpos[0][1]}, mgl32.Vec2{vpos[1][0], vpos[1][1]}, mgl32.Vec2{vpos[2][0], vpos[2][1]}
	ba, cb, ac := b.Sub(a), c.Sub(b), a.Sub(c)
	for i := miny; i < maxy; i++ {
		for j := minx; j < maxx; j++ {
			p := mgl32.Vec2{float32(j), float32(i)}
			ap, bp, cp := p.Sub(a), p.Sub(b), p.Sub(c)

			apb := mgl32.Mat2FromRows(ba, ap).Det()
			bpc := mgl32.Mat2FromRows(cb, bp).Det()
			cpa := mgl32.Mat2FromRows(ac, cp).Det()
			total := apb + bpc + cpa
			weights := mgl32.Vec3{bpc, cpa, apb}.Mul(1.0 / total)

			if weights[0] > 0.0 && weights[1] > 0.0 && weights[2] > 0.0 {
				x, y := int(j), int(i)
				color := vcol[0].Mul(weights[0]).Add(vcol[1].Mul(weights[1])).Add(vcol[2].Mul(weights[2])).Mul(255.0)
				if pix.Surrounds(x, y) {
					pix.SetPixel(x, y, byte(color[0]), byte(color[1]), byte(color[2]))
				}
			}
		}
	}
}
