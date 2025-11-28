package render

import "math"

// Naive storage as an array
type Voxels struct {
	Z, Y, X  int
	Presence []bool
	Color    [][3]byte
}

func VoxelsInit(x, y, z int) Voxels {
	presense := make([]bool, z*y*x)
	color := make([][3]byte, z*y*x)
	for i := 0; i < z*y*x; i++ {
		presense[i] = false
		color[i] = [3]byte{0, 0, 0}
	}
	return Voxels{z, y, x, presense, color}
}

func (vox *Voxels) SetVoxel(x, y, z int, r, g, b byte) {
	idx := vox.Index(x, y, z)
	vox.Presence[idx] = true
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
	origin, direc, tmax := ray.Origin, ray.Direc, ray.Tmax

	// Ok, this is a huge mess and needs to be cleaned up
	ox, oy, oz := origin.Elem()
	x, y, z := int(math.Floor(float64(ox))), int(math.Floor(float64(oy))), int(math.Floor(float64(oz)))
	dx, dy, dz := direc.Elem()
	adx, ady, adz := float32(math.Abs(float64(dx))), float32(math.Abs(float64(dy))), float32(math.Abs(float64(dz)))
	invx, invy, invz := 1.0/adx, 1.0/ady, 1.0/adz
	fractx, fracty, fractz := ox-float32(x), oy-float32(y), oz-float32(z)

	var stepx, stepy, stepz int
	var timex, timey, timez float32
	if dx > 0 {
		stepx = 1
		timex = 1.0 - fractx
	} else {
		stepx = -1
		timex = fractx
	}
	if dy > 0 {
		stepy = 1
		timey = 1.0 - fracty
	} else {
		stepy = -1
		timey = fracty
	}
	if dz > 0 {
		stepz = 1
		timez = 1.0 - fractz
	} else {
		stepz = -1
		timez = fractz
	}
	timex *= invx
	timey *= invy
	timez *= invz

	time := float32(0.0)
	for {
		if time > tmax {
			break
		}
		if vox.Surrounds(x, y, z) {
			idx := vox.Index(x, y, z)
			if vox.Presence[idx] {
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
