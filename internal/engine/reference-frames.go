package engine

import (
	"github.com/zheskett/go-voxel/internal/tensor"
)

type refframe uint8

const (
	FrameWorld refframe = iota
	FrameCamera
	FrameVoxel
)

type ReferenceFrame struct {
	b11 tensor.Vector3
	b22 tensor.Vector3
	b33 tensor.Vector3
}

type Basis interface {
	BasisFrame() ReferenceFrame
}

func (engine *Engine) BasisFrame() ReferenceFrame {
	return ReferenceFrame{tensor.Vec3X(), tensor.Vec3Y(), tensor.Vec3Z()}
}

type FrameChange struct {
	to   refframe
	from refframe
}
