package voxel

import (
	"github.com/chewxy/math32"
	te "github.com/zheskett/go-voxel/internal/tensor"
)

const (
	// All raymarched position data is ambiguous as it gives the face lies exactly on
	// the shared face of two neighbor voxels. This distance offset is used in:
	// vox = hit_position - hit_normal * VoxelRayDelta
	// to find the actual voxel the ray hit
	VoxelRayDelta = 0.05
)

type Ray struct {
	Origin te.Vector3
	Dir    te.Vector3
	Tmax   float32
}

type RayHit struct {
	Hit      bool
	Time     float32
	Color    [3]byte
	IntPos   [3]int
	Position te.Vector3
	Normal   te.Vector3
}

type Marchable interface {
	MarchRay(ray Ray) RayHit
}

type MarchData struct {
	Pos   Vec3i
	Step  Vec3i
	Inv   te.Vector3
	Timev te.Vector3
	Time  float32
	Side  axis
}

func MarchDataInit(ray Ray) MarchData {
	ox, oy, oz := ray.Origin.Elms()
	dx, dy, dz := ray.Dir.Elms()
	x, y, z := int(math32.Floor(ox)), int(math32.Floor(oy)), int(math32.Floor(oz))
	adx, ady, adz := math32.Abs(dx), math32.Abs(dy), math32.Abs(dz)
	fractx, fracty, fractz := ox-float32(x), oy-float32(y), oz-float32(z)

	var stepx, stepy, stepz int
	var invx, invy, invz float32
	var timex, timey, timez float32

	inf := math32.Inf(1)
	if adx < 1e-9 {
		stepx = 0
		invx = inf
		timex = inf
	} else {
		invx = 1.0 / adx
		if dx > 0 {
			stepx = 1
			timex = invx * (1.0 - fractx)
		} else {
			stepx = -1
			timex = invx * fractx
		}
	}
	if ady < 1e-9 {
		stepy = 0
		invy = inf
		timey = inf
	} else {
		invy = 1.0 / ady
		if dy > 0 {
			stepy = 1
			timey = invy * (1.0 - fracty)
		} else {
			stepy = -1
			timey = invy * fracty
		}
	}
	if adz < 1e-9 {
		stepz = 0
		invz = inf
		timez = inf
	} else {
		invz = 1.0 / adz
		if dz > 0 {
			stepz = 1
			timez = invz * (1.0 - fractz)
		} else {
			stepz = -1
			timez = invz * fractz
		}
	}

	return MarchData{
		Pos:   Vec3(x, y, z),
		Step:  Vec3(stepx, stepy, stepz),
		Inv:   te.Vec3(invx, invy, invz),
		Timev: te.Vec3(timex, timey, timez),
		Time:  0.0,
		Side:  none,
	}
}

func (march *MarchData) step() {
	if march.Timev.X < march.Timev.Y {
		if march.Timev.X < march.Timev.Z {
			march.Pos.X += march.Step.X
			march.Time = march.Timev.X
			march.Timev.X += march.Inv.X
			march.Side = axisX
		} else {
			march.Pos.Z += march.Step.Z
			march.Time = march.Timev.Z
			march.Timev.Z += march.Inv.Z
			march.Side = axisZ
		}
	} else {
		if march.Timev.Y < march.Timev.Z {
			march.Pos.Y += march.Step.Y
			march.Time = march.Timev.Y
			march.Timev.Y += march.Inv.Y
			march.Side = axisY
		} else {
			march.Pos.Z += march.Step.Z
			march.Time = march.Timev.Z
			march.Timev.Z += march.Inv.Z
			march.Side = axisZ
		}
	}
}

func (march *MarchData) ScaleToBox(box Box, ray Ray) {
	size := box.sizeScalar()
	march.Step = march.Step.Mul(size)
	pos := ray.Origin.Add(ray.Dir.Mul(march.Time))
	high := box.high.AsVec3f()
	low := box.low.AsVec3f()
	if march.Step.X > 0 {
		march.Timev.X = march.Time + (high.X-pos.X)*march.Inv.X
	} else {
		march.Timev.X = march.Time + (pos.X-low.X)*march.Inv.X
	}
	if march.Step.Y > 0 {
		march.Timev.Y = march.Time + (high.Y-pos.Y)*march.Inv.Y
	} else {
		march.Timev.Y = march.Time + (pos.Y-low.Y)*march.Inv.Y
	}
	if march.Step.Z > 0 {
		march.Timev.Z = march.Time + (high.Z-pos.Z)*march.Inv.Z
	} else {
		march.Timev.Z = march.Time + (pos.Z-low.Z)*march.Inv.Z
	}
}
