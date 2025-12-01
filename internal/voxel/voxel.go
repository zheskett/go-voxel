package voxel

import (
	"github.com/chewxy/math32"
	"github.com/zheskett/go-voxel/internal/tensor"
)

// Compact storage for an array of bools
type BitArray struct {
	bits []uint64
}

func BitArrayInit(len int) BitArray {
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

func (bits *BitArray) Reset(index int) {
	bucket := index / 64
	shift := index % 64
	mask := uint64(1) << shift
	bits.bits[bucket] ^= ^mask
}

type AABB struct {
	low  [3]int
	high [3]int
}

func AABBInit(lx, ly, lz int, hx, hy, hz int) AABB {
	return AABB{low: [3]int{lx, ly, lz}, high: [3]int{hx, hy, hz}}
}

func (bb *AABB) Subdivide() [8]AABB {
	lx, ly, lz := bb.low[0], bb.low[1], bb.low[2]
	hx, hy, hz := bb.high[0], bb.high[1], bb.high[2]
	mx, my, mz := lx+bb.high[0]-bb.low[0]/2, ly+bb.high[1]-bb.low[1]/2, lz+bb.high[2]-bb.low[2]/2

	return [8]AABB{
		AABBInit(lx, ly, lz, mx, my, mz),
		AABBInit(mx, ly, lz, hx, my, mz),
		AABBInit(lx, my, lz, mx, hy, mz),
		AABBInit(lx, ly, mz, mx, my, hz),
		AABBInit(mx, my, mz, hx, hy, hz),
		AABBInit(mx, my, lz, hx, hy, mz),
		AABBInit(lx, my, mz, mx, hy, hz),
		AABBInit(mx, ly, mz, hx, my, hz),
	}
}

type TreeNode struct {
	bb     AABB
	leaves [8]*TreeNode
}

func TreeNodeInit(bounds AABB) *TreeNode {
	return &TreeNode{bb: bounds}
}

func (node *TreeNode) IsStem() bool {
	return node.leaves[0] == nil
}

func (node *TreeNode) IsLeaf() bool {
	return node.leaves[0] != nil
}

type Octree struct {
	head *TreeNode
}

func OctreeInit(bounds AABB) Octree {
	return Octree{head: TreeNodeInit(bounds)}
}

func (tree *Octree) Insert(voxel [3]int) {

}

// Naive storage as an array
type Voxels struct {
	Z, Y, X  int
	Presence BitArray
	Color    [][3]byte
	// As of now, only support a single point-light
	Light          tensor.Vector3
	LightIntensity float32 // This isn't a good way of doing this it's just for proof of concept
}

func VoxelsInit(x, y, z int) Voxels {
	vox := Voxels{}
	presence := BitArrayInit(z * y * x)
	color := make([][3]byte, z*y*x)
	for i := 0; i < z*y*x; i++ {
		color[i] = [3]byte{0, 0, 0}
	}
	vox.Z = z
	vox.Y = y
	vox.X = x
	vox.Presence = presence
	vox.Color = color
	return vox
	// return Voxels{z, y, x, presence, color}
}

func (vox *Voxels) SetVoxel(x, y, z int, r, g, b byte) {
	idx := vox.Index(x, y, z)
	vox.Presence.Set(idx)
	vox.Color[idx] = [3]byte{r, g, b}
}

func (vox *Voxels) ResetVoxel(x, y, z int, r, g, b byte) {
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
				rayhit.Color = vox.Color[idx]
				rayhit.Position = origin.Add(direc.Mul(time))
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
