// There are duplicates of all the motion functions, with an FPS version, that doesn't include rolling
package render

import (
	"sync"

	"github.com/chewxy/math32"
	"github.com/go-gl/glfw/v3.3/glfw"
	te "github.com/zheskett/go-voxel/internal/tensor"
	vxl "github.com/zheskett/go-voxel/internal/voxel"
)

// Holds all the required information to get a ray's offset from the cam's front vector at any given pixel
type CameraRayBasis struct {
	drdx       te.Vector3
	dudy       te.Vector3
	halfwidth  float32
	halfheight float32
}

func CameraRayBasisInit(cam *Camera, pix *Pixels) CameraRayBasis {
	scale := math32.Tan(cam.Fov * math32.Pi / 360.0)
	hh, hw := float32(pix.Height/2), float32(pix.Width/2)

	dcamrdx := cam.Rvec.Mul(scale * cam.Aspect)
	dcamudy := cam.Uvec.Mul(scale)

	return CameraRayBasis{dcamrdx, dcamudy, hw, hh}
}

type Camera struct {
	Fvec           te.Vector3
	Rvec           te.Vector3
	Uvec           te.Vector3
	Wupvec         te.Vector3
	Pos            te.Vector3
	Lookspeed      float32
	Movespeed      float32
	Fov            float32
	Aspect         float32
	RenderDistance float32

	Pitch float32
	Yaw   float32
}

func CameraInit() Camera {
	return Camera{
		Fvec:   te.Vec3Z(),
		Rvec:   te.Vec3X(),
		Uvec:   te.Vec3Y(),
		Wupvec: te.Vec3Y(),

		Yaw:   0.0,
		Pitch: math32.Pi / 4.0,
	}
}

func (cam *Camera) UpdateRotation(rx, ry, rz float32, frame *FrameData) {
	rot := cam.Fvec.Mul(rz).Add(cam.Uvec.Mul(ry)).Add(cam.Rvec.Mul(rx)).Mul(cam.Lookspeed).Mul(frame.Deltat)
	att := te.Matrix3x3FromCols(cam.Rvec, cam.Uvec, cam.Fvec)
	att = te.Rotate3DXYZ(rot.X, rot.Y, rot.Z).Mul(att)

	cam.Fvec = att.Col(2)
	cam.Uvec = att.Col(1)
	cam.Rvec = att.Col(0)
}

// These are really messy and will be cleaned up eventually I swear
func (cam *Camera) UpdateRotationFPS(pitch, yaw float32) {
	cam.Pitch += pitch * cam.Lookspeed
	cam.Yaw += yaw * cam.Lookspeed
	cam.Pitch = math32.Min(math32.Max(cam.Pitch, -math32.Pi/2*0.99), math32.Pi/2*0.99)

	front := te.Rotate3DY(cam.Yaw).Mul(te.Rotate3DX(cam.Pitch)).MulVec(te.Vec3Z())
	right := front.Cross(cam.Wupvec)
	up := right.Cross(front)

	cam.Fvec = front.Normalized()
	cam.Rvec = right.Normalized()
	cam.Uvec = up.Normalized()
}

func (cam *Camera) UpdatePosition(dx, dy, dz float32, frame *FrameData) {
	movement := cam.Movespeed * frame.Deltat
	forward := cam.Fvec.Mul(dz * movement)
	vertical := cam.Uvec.Mul(dy * movement)
	lateral := cam.Rvec.Mul(dx * movement)
	cam.Pos = cam.Pos.Add(forward).Add(vertical).Add(lateral)
}

func (cam *Camera) UpdatePositionFPS(dx, dy, dz float32, frame *FrameData) {
	dx, dy, dz = te.Vec3(dx, dy, dz).NormalizedOrZero().Mul(cam.Movespeed * frame.Deltat).Elms()
	clampedfront := te.Vec3(cam.Fvec.X, 0, cam.Fvec.Z).Normalized()
	forward := clampedfront.Mul(dz)
	vertical := cam.Wupvec.Mul(dy)
	lateral := cam.Rvec.Mul(dx)
	cam.Pos = cam.Pos.Add(forward).Add(vertical).Add(lateral)
}

func (cam *Camera) getPixelRay(column int, row int, basis CameraRayBasis) vxl.Ray {
	dx, dy := float32(column)+0.5, float32(row)+0.5

	ndcx := (dx - basis.halfwidth) / basis.halfwidth
	ndcy := -(dy - basis.halfheight) / basis.halfheight

	// This is effectively finding the ray that points to that specific pixel
	dcamr := basis.drdx.Mul(ndcx)
	dcamu := basis.dudy.Mul(ndcy)
	raydirec := (cam.Fvec.Add(dcamr).Add(dcamu)).Normalized()
	return vxl.Ray{
		Origin: cam.Pos,
		Dir:    raydirec,
		Tmax:   cam.RenderDistance, // Max distance a ray can travel before terminating
	}
}

func (cam *Camera) RenderVoxels(vox *vxl.Voxels, pix *Pixels) {
	basis := CameraRayBasisInit(cam, pix)
	// Shouldn't be here, but tbh light info shouldn't be in the Voxel struct at all probably
	vox.LightCached.Clear()

	// Iterate and spawn a thread for each row of the pixel buffer
	threads := sync.WaitGroup{}
	for row := 0; row < pix.Height; row++ {
		threads.Go(func() {
			// Iterate each column of the pixel row
			for column := 0; column < pix.Width; column++ {
				ray := cam.getPixelRay(column, row, basis)

				hit := vox.MarchRay(ray)
				if hit.Hit {
					color := te.Vec3(float32(hit.Color[0]), float32(hit.Color[1]), float32(hit.Color[2]))

					/* Two choices for lighting, doing it per pixel or per voxel */
					shadedintensity := GetPixelShading(vox, hit, cam.RenderDistance)
					// shadedintensity := GetVoxelShading(vox, hit, cam.RenderDistance)

					shadedcolor := shadedintensity.MulComponent(color).ComponentMin(255.0)
					pix.SetPixel(column, row, byte(shadedcolor.X), byte(shadedcolor.Y), byte(shadedcolor.Z))
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

func UpdateCamInputGLFWFPS(cam *Camera, window *glfw.Window, frame *FrameData) {
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
	if window.GetKey(glfw.KeyT) == glfw.Press {
		window.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	}
	dx, dy := frame.GetMouseDelta(window)
	cam.UpdateRotationFPS(dy, dx)
	cam.UpdatePositionFPS(float32(tx), float32(ty), float32(tz), frame)
}
