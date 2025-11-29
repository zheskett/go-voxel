package render

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	te "github.com/zheskett/go-voxel/internal/tensor"
)

type Camera struct {
	Fvec      te.Vector3
	Rvec      te.Vector3
	Uvec      te.Vector3
	Pos       te.Vector3
	Lookspeed float32
	Movespeed float32
	Fov       float32
	Aspect    float32
}

func CameraInit() Camera {
	return Camera{
		Fvec: te.Vector3{X: 0, Y: 0, Z: 1},
		Rvec: te.Vector3{X: 1, Y: 0, Z: 0},
		Uvec: te.Vector3{X: 0, Y: 1, Z: 0},
	}
}

func (cam *Camera) UpdateRotation(rx, ry, rz float32) {
	rot := cam.Fvec.Mul(rz).Add(cam.Uvec.Mul(ry)).Add(cam.Rvec.Mul(rx)).Mul(cam.Lookspeed)
	att := te.Mat3FromCols(cam.Rvec, cam.Uvec, cam.Fvec)
	rox := te.Rotate3DX(rot[0])
	roy := te.Rotate3DY(rot[1])
	roz := te.Rotate3DZ(rot[2])

	att = roz.Mul3(roy).Mul3(rox).Mul3(att)
	cam.Fvec = att.Col(2)
	cam.Uvec = att.Col(1)
	cam.Rvec = att.Col(0)
}

func (cam *Camera) UpdatePosition(dx, dy, dz float32) {
	forward := cam.Fvec.Mul(dz * cam.Movespeed)
	vertial := cam.Uvec.Mul(dy * cam.Movespeed)
	lateral := cam.Rvec.Mul(dx * cam.Movespeed)
	cam.Pos = cam.Pos.Add(forward).Add(vertial).Add(lateral)
}

func UpdateCamInputGLFW(cam *Camera, window *glfw.Window) {
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
	cam.UpdateRotation(float32(rx), float32(ry), float32(rz))
	cam.UpdatePosition(float32(tx), float32(ty), float32(tz))
}
