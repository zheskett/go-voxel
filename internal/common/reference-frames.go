package common

import (
	"github.com/zheskett/go-voxel/internal/tensor"
)

type RefFrame uint8

const (
	FrameWorld RefFrame = iota
	FrameCamera
	FrameVoxel
)

type ReferenceFramef struct {
	// Canonical basis vectors
	b11 tensor.Vector3
	b22 tensor.Vector3
	b33 tensor.Vector3
	// Origin location
	o tensor.Vector3
}

type ReferenceFramei struct {
	// Canonical basis vectors
	b11 tensor.Vector3i
	b22 tensor.Vector3i
	b33 tensor.Vector3i
	// Origin location
	o tensor.Vector3i
}

func RefFramefInit(b11, b22, b33 tensor.Vector3, o tensor.Vector3) ReferenceFramef {
	return ReferenceFramef{b11, b22, b33, o}
}

func RefFrameiInit(b11, b22, b33 tensor.Vector3i, o tensor.Vector3i) ReferenceFramei {
	return ReferenceFramei{b11, b22, b33, o}
}

func (f ReferenceFramef) toGlobal(v tensor.Vector3) tensor.Vector3 {
	return f.o.Add(f.b11.Mul(v.X)).Add(f.b22.Mul(v.Y)).Add(f.b33.Mul(v.Z))
}

func (f ReferenceFramei) toGlobal(v tensor.Vector3i) tensor.Vector3i {
	return f.o.Add(f.b11.Mul(v.X)).Add(f.b22.Mul(v.Y)).Add(f.b33.Mul(v.Z))
}

func (f ReferenceFramef) fromGlobal(v tensor.Vector3) tensor.Vector3 {
	rel := v.Sub(f.o)
	return tensor.Vec3(
		rel.Dot(f.b11),
		rel.Dot(f.b22),
		rel.Dot(f.b33),
	)
}

func (f ReferenceFramei) fromGlobal(v tensor.Vector3i) tensor.Vector3i {
	rel := v.Sub(f.o)
	return tensor.Vec3i(
		rel.Dot(f.b11),
		rel.Dot(f.b22),
		rel.Dot(f.b33),
	)
}

type Basisf interface {
	BasisFramef() ReferenceFramef
}

type Basisi interface {
	BasisFramei() ReferenceFramei
}

func GlobalFramef() ReferenceFramef {
	return ReferenceFramef{tensor.Vec3X(), tensor.Vec3Y(), tensor.Vec3Z(), tensor.Vec3Zero()}
}

func GlobalFramei() ReferenceFramei {
	return ReferenceFramei{tensor.Vec3iX(), tensor.Vec3iY(), tensor.Vec3iZ(), tensor.Vec3iZero()}
}

// This ergonomics of this interface need to be improved, but I don't know how to
// make const structs in Go. So, in order to change between frames you need to have
// a type that implements Basis in scope
//
// Allows a vector from any frame to be converted to any other frame
func Convertf(v tensor.Vector3, from Basisf, to Basisf) tensor.Vector3 {
	fromframe := from.BasisFramef()
	toframe := to.BasisFramef()

	glob := fromframe.toGlobal(v)
	dest := toframe.fromGlobal(glob)

	return dest
}

func Converti(v tensor.Vector3i, from Basisi, to Basisi) tensor.Vector3i {
	fromframe := from.BasisFramei()
	toframe := to.BasisFramei()

	glob := fromframe.toGlobal(v)
	dest := toframe.fromGlobal(glob)

	return dest
}

// Returns the local representation of a vector on any type that implements
// Basis from the global coordinate system
func InLocalf(v tensor.Vector3, local Basisf) tensor.Vector3 {
	return local.BasisFramef().fromGlobal(v)
}

func InLocali(v tensor.Vector3i, local Basisi) tensor.Vector3i {
	return local.BasisFramei().fromGlobal(v)
}
