package voxel

import (
	"github.com/chewxy/math32"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/zheskett/go-voxel/internal/tensor"
)

// Compact storage for an array of bools
type BitArray struct {
	bits []uint64
}

func BitArrayInit(len int) BitArray {
	if len%64 != 0 {
		len += 64
	}
	len = len / 64
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

func (bits *BitArray) Put(index int, value bool) {
	bucket := index / 64
	shift := index % 64
	mask := uint64(1) << shift
	if value {
		bits.bits[bucket] |= mask
	} else {
		bits.bits[bucket] &= ^mask
	}
}

func (bits *BitArray) Reset(index int) {
	bucket := index / 64
	shift := index % 64
	mask := uint64(1) << shift
	bits.bits[bucket] ^= mask
}

func (bits *BitArray) Clear() {
	for i := range bits.bits {
		bits.bits[i] = 0
	}
}

// Just a point light
type Light struct {
	Position tensor.Vector3
	Color    tensor.Vector3 // Can have mag > 1 for a bright light
}

// Lighting info for a single voxel
type CachedLighting struct {
	Light tensor.Vector3 // The cumulative lighting it gets
	Dir   tensor.Vector3 // The weighted direction of all lights in the scene w.r.t. that voxel
}

// Naive storage as an array
type Voxels struct {
	Z, Y, X  int
	Presence BitArray
	Color    [][3]byte

	// Actually, this would be really easy to bake lighting as long as we aren't moving the lights at runtime
	// Doing realtime lighting just seems more interesting tho
	LightCached BitArray // Whether or not we already having lighting data for that frame
	Lighting    []CachedLighting

	Lights []Light // Shouldn't be in here probably, maybe in another larger structure holding all worlds stuff
}

func VoxelsInit(x, y, z int) Voxels {
	presence := BitArrayInit(z * y * x)
	color := make([][3]byte, z*y*x)
	lighting := make([]CachedLighting, z*y*x)
	lightcache := BitArrayInit(z * y * x)
	lights := make([]Light, 0)
	for i := 0; i < z*y*x; i++ {
		color[i] = [3]byte{0, 0, 0}
	}
	return Voxels{z, y, x, presence, color, lightcache, lighting, lights}
}

func (vox *Voxels) SetVoxel(x, y, z int, r, g, b byte) {
	idx := vox.Index(x, y, z)
	vox.Presence.Set(idx)
	vox.Color[idx] = [3]byte{r, g, b}
}

func (vox *Voxels) ResetVoxel(x, y, z int) {
	idx := vox.Index(x, y, z)
	vox.Presence.Reset(idx)
	vox.Color[idx] = [3]byte{0, 0, 0}
}

func (vox *Voxels) Index(x, y, z int) int {
	return vox.X*vox.Y*z + vox.X*y + x
}

func (vox *Voxels) Surrounds(x, y, z int) bool {
	return x < vox.X && y < vox.Y && z < vox.Z && x >= 0 && y >= 0 && z >= 0
}

// Enum for axis
// Probably unnecessary for this use
type axis uint8

const (
	axisX axis = iota
	axisY
	axisZ
	none
)

func (vox *Voxels) MarchRay(ray Ray) RayHit {
	rayhit := RayHit{Hit: false}
	origin, direc, tmax := ray.Origin, ray.Dir, ray.Tmax

	ox, oy, oz := origin.Elms()
	dx, dy, dz := direc.Elms()

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

	side := none
	time := float32(0.0)
	for {
		if time > tmax {
			break
		}
		if vox.Surrounds(x, y, z) {
			idx := vox.Index(x, y, z)
			if vox.Presence.Get(idx) {
				rayhit.Hit = true
				rayhit.Time = time
				rayhit.IntPos = [3]int{x, y, z}
				rayhit.Position = ray.Origin.Add(ray.Dir.Mul(time))
				rayhit.Color = vox.Color[idx]
				switch side {
				case axisX:
					rayhit.Normal = tensor.Vec3(1, 0, 0).Mul(-float32(stepx))
				case axisY:
					rayhit.Normal = tensor.Vec3(0, 1, 0).Mul(-float32(stepy))
				case axisZ:
					rayhit.Normal = tensor.Vec3(0, 0, 1).Mul(-float32(stepz))
				default:
					rayhit.Normal = tensor.Vec3(0, 0, 0)
				}
				break
			}
		}

		if timex < timey {
			if timex < timez {
				x += stepx
				time = timex
				timex += invx
				side = axisX
			} else {
				z += stepz
				time = timez
				timez += invz
				side = axisZ
			}
		} else {
			if timey < timez {
				y += stepy
				time = timey
				timey += invy
				side = axisY
			} else {
				z += stepz
				time = timez
				timez += invz
				side = axisZ
			}
		}
	}

	return rayhit
}

// Adds a voxel object to the world
func (vox *Voxels) AddVoxelObj(vObj VoxelObj, x, y, z int) {
	for xyz, cIdx := range vObj.Voxels {
		vx, vy, vz := int(xyz[0]), int(xyz[1]), int(xyz[2])
		if vox.Surrounds(x+vx, y+vy, z+vz) {
			clr := vObj.ColorPalete[cIdx]
			vox.SetVoxel(x+vx, y+vy, z+vz, clr.R, clr.G, clr.B)
		}
	}
}

// This is super temporary and just a proof of concept
func (vox *Voxels) UpdateInputs(window *glfw.Window, pos tensor.Vector3, dir tensor.Vector3) {
	ray := Ray{Origin: pos, Dir: dir, Tmax: 100.0}
	hit := vox.MarchRay(ray)
	if !hit.Hit {
		return
	}
	if window.GetMouseButton(glfw.MouseButtonLeft) == glfw.Press {
		x, y, z := hit.IntPos[0], hit.IntPos[1], hit.IntPos[2]
		vox.ResetVoxel(x, y, z)
	}
	if window.GetMouseButton(glfw.MouseButtonRight) == glfw.Press {
		voxel := hit.Position.Add(hit.Normal.Mul(VoxelRayDelta))
		x, y, z := int(voxel.X), int(voxel.Y), int(voxel.Z)
		vox.SetVoxel(x, y, z, 255, 255, 255)
	}
}
