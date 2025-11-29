package main

import (
	"runtime"

	"github.com/chewxy/math32"
	ml "github.com/go-gl/mathgl/mgl32"
	ren "github.com/zheskett/go-voxel/internal/render"
	vxl "github.com/zheskett/go-voxel/internal/voxel"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// This is meantioned in the usage example on github.
	runtime.LockOSThread()
}

func main() {
	rm := ren.RenderManagerInit()
	cam := ren.CameraInit()
	cam.Movespeed = 15 // 15 voxels/s second walking
	cam.Lookspeed = 2  // 2 rad/s rotation
	cam.Fov = 90
	cam.Aspect = float32(rm.Pixels.Width) / float32(rm.Pixels.Height)
	cam.Pos = ml.Vec3{16, 4, 16}
	vox := vxl.VoxelsInit(128, 128, 128)
	fdata := ren.FrameDataInit()
	voxelDebugScene(&vox)

	for {
		renderDebugTri(&rm.Pixels, &cam)
		cam.RenderVoxels(&vox, &rm.Pixels)
		ren.UpdateCamInputGLFW(&cam, rm.Window, &fdata)
		fdata.Update()
		rm.Render()
		rm.CheckExit()
	}
}

func voxelDebugScene(vox *vxl.Voxels) {
	// Make a teal and purple checkerboard "ground"
	for i := 0; i < vox.Z; i++ {
		for j := 0; j < vox.X; j++ {
			for k := 0; k < 1+(i+j)/16; k++ {
				if (i+j)%2 == 0 {
					vox.SetVoxel(i, k, j, 70, 200, 200)
				} else {
					vox.SetVoxel(i, k, j, 200, 30, 200)
				}
			}
		}
	}
	// Same thing as above but on the roof with a more extreme slope
	for i := 0; i < vox.Z; i++ {
		for j := 0; j < vox.X; j++ {
			for k := vox.Y - 1; k > vox.Y-(i+j)/4; k-- {
				if (i+j)%2 == 0 {
					vox.SetVoxel(i, k, j, 200, 3, 180)
				} else {
					vox.SetVoxel(i, k, j, 150, 200, 20)
				}
			}
		}
	}
	// Floating red cube
	for i := 10; i < 15; i++ {
		for j := 10; j < 15; j++ {
			for k := 10; k < 15; k++ {
				vox.SetVoxel(i, j, k, 180, 50, 50)
			}
		}
	}
	// Floating orange cube
	for i := 32; i < 40; i++ {
		for j := 32; j < 40; j++ {
			for k := 32; k < 40; k++ {
				vox.SetVoxel(i, j, k, 200, 100, 100)
			}
		}
	}
	// A larger checkerboard wall one one side
	for i := 0; i < vox.X; i++ {
		for j := 0; j < vox.Y; j++ {
			if (i%2+j%2)%2 == 0 {
				vox.SetVoxel(0, i, j, 30, 30, 30)
			} else {
				vox.SetVoxel(0, i, j, 200, 200, 200)
			}
		}
	}
	// Some green pillar
	for i := 1; i < 16; i++ {
		vox.SetVoxel(5, i, 5, 30, 255, 30)
		vox.SetVoxel(vox.X-1, i, 0, 30, 255, 30)
		vox.SetVoxel(vox.X-1, i, vox.Z-1, 30, 255, 30)
		vox.SetVoxel(50, i, 50, 30, 255, 30)
		vox.SetVoxel(25, i, 9, 30, 255, 30)
	}
}

func renderDebugTri(pix *ren.Pixels, cam *ren.Camera) {
	pix.FillPixels(15, 25, 40)
	vpos := []ml.Vec3{
		{2.5, 32.0, 6.0},
		{18.5, 16.0, 7.0},
		{0.0, 32.0, 15.0},
	}
	vcol := []ml.Vec3{
		{1.0, 0.7, 0.0},
		{0.0, 1.0, 0.7},
		{0.7, 0.0, 1.0},
	}
	scale := 1.0 / math32.Tan(cam.Fov/2.0)
	hw, hh := float32(pix.Width/2), float32(pix.Height/2)

	for i := range vpos {
		// translate
		vpos[i] = vpos[i].Sub(cam.Pos)
		// relative to camera
		vpos[i] = ml.Vec3{
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

	var minx, maxx float32 = 1e9, -1e9
	var miny, maxy float32 = 1e9, -1e9
	for _, vert := range vpos {
		minx = math32.Min(minx, vert[0])
		miny = math32.Min(miny, vert[1])
		maxx = math32.Max(maxx, vert[0])
		maxy = math32.Max(maxy, vert[1])
	}
	minx, maxx, miny, maxy = math32.Max(minx, 0.0), math32.Min(maxx, float32(pix.Width)), math32.Max(miny, 0.0), math32.Min(maxy, float32(pix.Height))

	a, b, c := ml.Vec2{vpos[0][0], vpos[0][1]}, ml.Vec2{vpos[1][0], vpos[1][1]}, ml.Vec2{vpos[2][0], vpos[2][1]}
	ba, cb, ac := b.Sub(a), c.Sub(b), a.Sub(c)
	for i := miny; i < maxy; i++ {
		for j := minx; j < maxx; j++ {
			p := ml.Vec2{float32(j), float32(i)}
			ap, bp, cp := p.Sub(a), p.Sub(b), p.Sub(c)

			apb := ml.Mat2FromRows(ba, ap).Det()
			bpc := ml.Mat2FromRows(cb, bp).Det()
			cpa := ml.Mat2FromRows(ac, cp).Det()
			total := apb + bpc + cpa
			weights := ml.Vec3{bpc, cpa, apb}.Mul(1.0 / total)

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
