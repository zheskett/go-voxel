package render

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

	// I really don't like how there isn't methods to get x, y, components of vectors
	x, y, z := int(origin[0]), int(origin[1]), int(origin[2])
	var stepx, stepy, stepz int
	// There has to be a better way to write this
	if direc[0] > 0 {
		stepx = 1
	} else {
		stepx = -1
	}
	if direc[1] > 0 {
		stepy = 1
	} else {
		stepy = -1
	}
	if direc[2] > 0 {
		stepz = 1
	} else {
		stepz = -1
	}

	// Extracts the distance to closest boundary
	// Is either fract(x) or 1 - fract(x)
	var timex, timey, timez float32
	if stepx > 0 {
		timex = direc[0] - float32(int32(direc[0]))
	} else {
		timex = 1.0 - (direc[0] - float32(int32(direc[0])))
	}
	if stepy > 0 {
		timey = direc[1] - float32(int32(direc[1]))
	} else {
		timey = 1.0 - (direc[1] - float32(int32(direc[1])))
	}
	if stepz > 0 {
		timez = direc[2] - float32(int32(direc[2]))
	} else {
		timez = 1.0 - (direc[2] - float32(int32(direc[2])))
	}
	timex /= direc[0]
	timey /= direc[1]
	timez /= direc[2]

	tstepx := float32(stepx) / direc[0]
	tstepy := float32(stepy) / direc[1]
	tstepz := float32(stepz) / direc[2]

	time := float32(0.0)
	for {
		if time > tmax {
			break
		}
		if vox.Surrounds(x, y, z) {
			idx := vox.Index(x, y, z)
			if vox.Presence[idx] {
				rayhit.Color = vox.Color[idx]
				break
			}
		}
		if timex < timey {
			if timex < timez {
				x += stepx
				time = timex
				timex += tstepx
			} else {
				z += stepz
				time = timez
				timez += tstepz
			}
		} else {
			if timey < timez {
				y += stepy
				time = timey
				timey += tstepy
			} else {
				z += stepz
				time = timez
				timez += tstepz
			}
		}
	}

	return rayhit
}
