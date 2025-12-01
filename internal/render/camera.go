package render

import (
	"sync"

	"github.com/chewxy/math32"
	"github.com/go-gl/glfw/v3.3/glfw"
	te "github.com/zheskett/go-voxel/internal/tensor"
	vxl "github.com/zheskett/go-voxel/internal/voxel"
)

type Camera struct {
	Fvec           te.Vector3
	Rvec           te.Vector3
	Uvec           te.Vector3
	Pos            te.Vector3
	Lookspeed      float32
	Movespeed      float32
	Fov            float32
	Aspect         float32
	RenderDistance float32
}

func CameraInit() Camera {
	return Camera{
		Fvec: te.Vec3(0, 0, 1),
		Rvec: te.Vec3(1, 0, 0),
		Uvec: te.Vec3(0, 1, 0),
	}
}

func (cam *Camera) UpdateRotation(rx, ry, rz float32, frame *FrameData) {
	// Doesn't actually make sense to have dt if camera wasn't controlled with arrow keys
	// Once we switch to mouse this needs to be removed
	rot := cam.Fvec.Mul(rz).Add(cam.Uvec.Mul(ry)).Add(cam.Rvec.Mul(rx)).Mul(cam.Lookspeed).Mul(frame.Deltat)
	att := te.Matrix3x3FromCols(cam.Rvec, cam.Uvec, cam.Fvec)
	att = te.Rotate3DXYZ(rot.X, rot.Y, rot.Z).Mul(att)

	cam.Fvec = att.Col(2)
	cam.Uvec = att.Col(1)
	cam.Rvec = att.Col(0)
}

func (cam *Camera) UpdatePosition(dx, dy, dz float32, frame *FrameData) {
	movement := cam.Movespeed * frame.Deltat
	forward := cam.Fvec.Mul(dz * movement)
	vertical := cam.Uvec.Mul(dy * movement)
	lateral := cam.Rvec.Mul(dx * movement)
	cam.Pos = cam.Pos.Add(forward).Add(vertical).Add(lateral)
}

func (cam *Camera) RenderVoxels(vox *vxl.Voxels, pix *Pixels) {
	scale := math32.Tan(cam.Fov * math32.Pi / 360.0)
	hh, hw := float32(pix.Height/2), float32(pix.Width/2)

	// These kinds of things and some stuff in 'vox.MarchRay' can be pre-computed
	// instead of doing it again every time for every pixel
	// Need to find a nice way to package all that
	dcamrdx := cam.Rvec.Mul(scale * cam.Aspect)
	dcamudy := cam.Uvec.Mul(scale)

	// Iterate and spawn a thread for each row of the pixel buffer
	threads := sync.WaitGroup{}
	for row := 0; row < pix.Height; row++ {
		threads.Go(func() {
			// Iterate each column of the pixel row
			for column := 0; column < pix.Width; column++ {
				dx, dy := float32(column)+0.5, float32(row)+0.5

				ndcx := (dx - hw) / hw
				ndcy := -(dy - hh) / hh

				// This is effectively finding the ray that points to that specific pixel
				dcamr := dcamrdx.Mul(ndcx)
				dcamu := dcamudy.Mul(ndcy)
				raydirec := (cam.Fvec.Add(dcamr).Add(dcamu)).Normalized()
				ray := vxl.Ray{
					Origin: cam.Pos,
					Dir:    raydirec,
					Tmax:   cam.RenderDistance, // Max distance a ray can travel before terminating
				}

				rayhit := vox.MarchRay(ray)
				if rayhit.Hit {
					// All this lighting will be pulled into its own module
					lightvec := rayhit.Position.Sub(vox.Light)
					length := lightvec.Len()
					intensity := lightvec.Div(length).Dot(rayhit.Normal) * vox.LightIntensity / length
					if intensity < 0.025 {
						intensity = 0.025 // Limit dimness to 2.5 %
					}
					color := rayhit.Color
					floatcolor := te.Vec3(float32(color[0]), float32(color[1]), float32(color[2])).Mul(intensity)
					if floatcolor.X > 255 {
						floatcolor.X = 255
					}
					if floatcolor.Y > 255 {
						floatcolor.Y = 255
					}
					if floatcolor.Z > 255 {
						floatcolor.Z = 255
					}
					pix.SetPixel(column, row, byte(floatcolor.X), byte(floatcolor.Y), byte(floatcolor.Z))
				}
			}
		})
	}
	threads.Wait()
}

func UpdateCamInputGLFW(cam *Camera, window *glfw.Window, frame *FrameData) {
	rx, ry, rz := 0, 0, 0
	tx, ty, tz := 0, 0, 0
	if window.GetKey(glfw.KeyW) == glfw.Press {
		tz++
	}
	if window.GetKey(glfw.KeyS) == glfw.Press {
		tz--
	}
	if window.GetKey(glfw.KeyA) == glfw.Press {
		tx--
	}
	if window.GetKey(glfw.KeyD) == glfw.Press {
		tx++
	}
	if window.GetKey(glfw.KeySpace) == glfw.Press {
		ty--
	}
	if window.GetKey(glfw.KeyLeftShift) == glfw.Press {
		ty++
	}
	if window.GetKey(glfw.KeyUp) == glfw.Press {
		rx++
	}
	if window.GetKey(glfw.KeyDown) == glfw.Press {
		rx--
	}
	if window.GetKey(glfw.KeyRight) == glfw.Press {
		ry++
	}
	if window.GetKey(glfw.KeyLeft) == glfw.Press {
		ry--
	}
	if window.GetKey(glfw.KeyQ) == glfw.Press {
		rz--
	}
	if window.GetKey(glfw.KeyE) == glfw.Press {
		rz++
	}
	cam.UpdateRotation(float32(rx), float32(ry), float32(rz), frame)
	cam.UpdatePosition(float32(tx), float32(ty), float32(tz), frame)
}
