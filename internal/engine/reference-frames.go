package engine

import "github.com/zheskett/go-voxel/internal/tensor"

type refframe uint8

const (
	FrameWorld refframe = iota
	FrameCamera
	FrameVoxel
)

// We need something like this but unfortunately this isn't valid in Go
//
// const (
// 	GlobalFrame ReferenceFrame = ReferenceFrame{
// 		tensor.Vec3X(), tensor.Vec3Y(), tensor.Vec3Z(), tensor.Vec3Zero()
// 	}
// 	VoxelFrame ReferenceFrame = ReferenceFrame{
// 		...
// 	}
// )

type ReferenceFrame struct {
	// Canonical basis vectors
	b11 tensor.Vector3
	b22 tensor.Vector3
	b33 tensor.Vector3
	// Origin location
	o tensor.Vector3
}

func (f ReferenceFrame) toGlobal(v tensor.Vector3) tensor.Vector3 {
	return f.o.Add(f.b11.Mul(v.X)).Add(f.b22.Mul(v.Y)).Add(f.b33.Mul(v.Z))
}

func (f ReferenceFrame) fromGlobal(v tensor.Vector3) tensor.Vector3 {
	rel := v.Sub(f.o)
	return tensor.Vec3(
		rel.Dot(f.b11),
		rel.Dot(f.b22),
		rel.Dot(f.b33),
	)
}

type Basis interface {
	BasisFrame() ReferenceFrame
}

// The global reference frame is exactly what you would expect
func (engine *Engine) BasisFrame() ReferenceFrame {
	return ReferenceFrame{tensor.Vec3X(), tensor.Vec3Y(), tensor.Vec3Z(), tensor.Vec3Zero()}
}

// This ergonomics of this interface need to be improved, but I don't know how to
// make const structs in Go. So, in order to change between frames you need to have
// a type that implements Basis in scope
//
// This is the only function that should be public
func Convert(v tensor.Vector3, from Basis, to Basis) tensor.Vector3 {
	fromframe := from.BasisFrame()
	toframe := to.BasisFrame()

	glob := fromframe.toGlobal(v)
	dest := toframe.fromGlobal(glob)

	return dest
}
