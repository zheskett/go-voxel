package voxel

import (
	"github.com/chewxy/math32"
)

// Compact storage for an array of bools
type BitArray struct {
	bits []uint64
}

func BitFlagsInit(len int) BitArray {
	len = len / 64
	if len%64 != 0 {
		len += 1
	}
	bits := make([]uint64, len)
	for i := range len {
		bits[i] = 0
	}
	return BitArray{bits}
}

func (bits *BitArray) Get(index int) bool {
	bucket := index / 64
	shift := index % 64
	mask := uint64(1) << shift
	return bits.bits[bucket]&mask != 0
}

func (bits *BitArray) Set(index int) {
	bucket := index / 64
	shift := index % 64
	mask := uint64(1) << shift
	bits.bits[bucket] |= mask
}

// Naive storage as an array
type Voxels struct {
	Z, Y, X  int
	Presence BitArray
	Color    [][3]byte
}

func VoxelsInit(x, y, z int) Voxels {
	presence := BitFlagsInit(z * y * z)
	color := make([][3]byte, z*y*x)
	for i := 0; i < z*y*x; i++ {
		color[i] = [3]byte{0, 0, 0}
	}
	return Voxels{x, y, z, presence, color}
}

func (vox *Voxels) SetVoxel(x, y, z int, r, g, b byte) {
	idx := vox.Index(x, y, z)
	vox.Presence.Set(idx)
	vox.Color[idx] = [3]byte{r, g, b}
}

func (vox *Voxels) Index(x, y, z int) int {
	return vox.X*vox.Y*z + vox.X*y + x
}

func (vox *Voxels) Surrounds(x, y, z int) bool {
	return x < vox.X && y < vox.Y && z < vox.Z && x >= 0 && y >= 0 && z >= 0
}

func (vox *Voxels) MarchRay(ray Ray) RayHit {
	rayhit := RayHit{Hit: false}
	origin, direc, tmax := ray.Origin, ray.Dir, ray.Tmax

	ox, oy, oz := origin.Elms()
	dx, dy, dz := direc.Elms()

	// Ok, this is a huge mess and needs to be cleaned up
	x, y, z := int(math32.Floor(ox)), int(math32.Floor(oy)), int(math32.Floor(oz))
	adx, ady, adz := math32.Abs(dx), math32.Abs(dy), math32.Abs(dz)
	invx, invy, invz := 1.0/adx, 1.0/ady, 1.0/adz
	fractx, fracty, fractz := ox-float32(x), oy-float32(y), oz-float32(z)

	var stepx, stepy, stepz int
	timex, timey, timez := invx, invy, invz
	if dx > 0 {
		stepx = 1
		timex *= 1.0 - fractx
	} else {
		stepx = -1
		timex *= fractx
	}
	if dy > 0 {
		stepy = 1
		timey *= 1.0 - fracty
	} else {
		stepy = -1
		timey *= fracty
	}
	if dz > 0 {
		stepz = 1
		timez *= 1.0 - fractz
	} else {
		stepz = -1
		timez *= fractz
	}

	time := float32(0.0)
	for {
		if time > tmax {
			break
		}
		if vox.Surrounds(x, y, z) {
			idx := vox.Index(x, y, z)
			if vox.Presence.Get(idx) {
				rayhit.Color = vox.Color[idx]
				rayhit.Hit = true
				break
			}
		}

		if timex < timey {
			if timex < timez {
				x += stepx
				time = timex
				timex += invx
			} else {
				z += stepz
				time = timez
				timez += invz
			}
		} else {
			if timey < timez {
				y += stepy
				time = timey
				timey += invy
			} else {
				z += stepz
				time = timez
				timez += invz
			}
		}
	}

	return rayhit
}
