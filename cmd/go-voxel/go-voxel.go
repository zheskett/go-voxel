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
	vox := render.VoxelsInit(64, 64, 64)
	voxelDebugScene(&vox)

	for {
		renderDebugTri(&rm.Pixels, &cam)
		voxelRaymarchRender(&rm.Pixels, &cam, &vox)
		rm.Render()
		rm.CheckExit()
		render.UpdateCamInputGLFW(&cam, rm.Window)
	}
}

// This currently isn't doing anything, unsure why
func voxelRaymarchRender(pix *render.Pixels, cam *render.Camera, vox *render.Voxels) {
	scale := float32(1.0 / math.Tan(float64(cam.Fov/2.0)))
	hh, hw := float32(pix.Height), float32(pix.Width)

	dcamrdx := cam.Rvec.Mul(scale * cam.Aspect)
	dcamudy := cam.Uvec.Mul(scale)
	// Cast out a ray for each pixel on the screen
	for i := 0; i < pix.Height; i++ {
		for j := 0; j < pix.Width; j++ {
			dx, dy := float32(j)+0.5, float32(i)+0.5

			ndcx, ndcy := (dx-hw)/hw, (dy-hh)/hh
			// Does Go not have assert?
			if ndcx > 1 || ndcx < -1 {
				panic("math mistake")
			}
			if ndcy > 1 || ndcy < -1 {
				panic("math mistake")
			}

			dcamr := dcamrdx.Mul(-ndcx)
			dcamu := dcamudy.Mul(ndcy)
			// This is effectively finding the ray that points to that specific pixel
			raydirec := (cam.Fvec.Add(dcamr).Add(dcamu)).Normalize()
			ray := render.Ray{
				Origin: cam.Pos,
				Direc:  raydirec,
				Tmax:   32.0,
			}

			rayhit := vox.MarchRay(ray)
			if rayhit.Hit {
				color := rayhit.Color
				pix.SetPixel(j, i, color[0], color[1], color[2])
			}
		}
	}
}

func voxelDebugScene(vox *render.Voxels) {
	// Make a teal "ground"
	for i := 0; i < vox.Z; i++ {
		for j := 0; j < vox.X; j++ {
			vox.SetVoxel(i, 0, j, 0, 255, 255)
		}
	}
	// floating red cube
	for i := 5; i < 10; i++ {
		for j := 5; j < 10; j++ {
			for k := 5; k < 10; k++ {
				vox.SetVoxel(i, j, k, 180, 50, 50)
			}
		}
	}
	// some green pillars
	for i := 1; i < 5; i++ {
		vox.SetVoxel(0, i, 0, 30, 255, 30)
		vox.SetVoxel(vox.X-1, i, 0, 30, 255, 30)
		vox.SetVoxel(vox.X-1, i, vox.Z-1, 30, 255, 30)
		vox.SetVoxel(5, i, 5, 30, 255, 30)
	}
}

func renderDebugTri(pix *render.Pixels, cam *render.Camera) {
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
	scale := float32(1.0 / math.Tan(float64(cam.Fov/2.0)))
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
		vpos[i][0] /= vpos[i][2] * scale * cam.Aspect
		vpos[i][1] /= vpos[i][2] * scale

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
