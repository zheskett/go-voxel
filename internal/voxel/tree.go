package voxel

import (
	"github.com/chewxy/math32"
	"github.com/zheskett/go-voxel/internal/tensor"
)

const (
	BrickSize  int = 8
	BrickTotal int = BrickSize * BrickSize * BrickSize
)

// This is basically a slimemd down clone of the old Voxel struct.
//
// The idea for this tree works kind of the way that minecraft stores chunks,
// but using a tree for faster traversal
type Brick struct {
	Presence BitArray
	Color    [][3]byte
}

func BrickInit() Brick {
	presence := BitArrayInit(BrickTotal)
	color := make([][3]byte, BrickTotal)
	return Brick{presence, color}
}

func (brk *Brick) SetVoxel(x, y, z int, r, g, b byte) {
	idx := brk.Index(x, y, z)
	brk.Presence.Set(idx)
	brk.Color[idx] = [3]byte{r, g, b}
}

func (brk *Brick) ResetBrick(x, y, z int) {
	idx := brk.Index(x, y, z)
	brk.Presence.Reset(idx)
	brk.Color[idx] = [3]byte{0, 0, 0}
}

func (brk *Brick) Index(x, y, z int) int {
	return BrickSize*BrickSize*z + BrickSize*y + x
}

func (brk *Brick) Surrounds(x, y, z int) bool {
	return x < BrickSize && y < BrickSize && z < BrickSize && x >= 0 && y >= 0 && z >= 0
}

type AABB struct {
	Low  [3]int
	High [3]int
}

func AABBInit(lx, ly, lz int, hx, hy, hz int) AABB {
	return AABB{Low: [3]int{lx, ly, lz}, High: [3]int{hx, hy, hz}}
}

// Slab-method of AABB and ray intersection
func (bb *AABB) RayIntersection(ray Ray) (float32, float32) {
	tmin := float32(0.0)
	tmax := ray.Tmax
	dirs := ray.Dir.AsArray()
	orig := ray.Origin.AsArray()
	for i := range 3 {
		if dirs[i] != 0.0 {
			invd := 1.0 / dirs[i]
			t0 := (float32(bb.Low[i]) - orig[i]) * invd
			t1 := (float32(bb.High[i]) - orig[i]) * invd

			if invd < 0.0 {
				t0, t1 = t1, t0
			}
			if t0 > tmin {
				tmin = t0
			}
			if t1 < tmax {
				tmax = t1
			}
			if tmax < tmin {
				return 1, 0 // There isn't an intersection
			}
		}
	}

	return tmin, tmax
}

func (bb *AABB) Subdivide() [8]AABB {
	lx, ly, lz := bb.Low[0], bb.Low[1], bb.Low[2]
	hx, hy, hz := bb.High[0], bb.High[1], bb.High[2]
	mx, my, mz := lx+(bb.High[0]-bb.Low[0])/2, ly+(bb.High[1]-bb.Low[1])/2, lz+(bb.High[2]-bb.Low[2])/2

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
	Box    AABB
	Brick  *Brick
	Leaves [8]*TreeNode
}

func (node *TreeNode) IsLeaf() bool {
	return node.Brick != nil
}

type BrickTree struct {
	Root TreeNode
}

// Can repurpose this exact same function for traversing a single brick, once we go
// through the tree and determine that the leaf has voxels present
func (brk *Brick) MarchRay(ray Ray) RayHit {
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
		if brk.Surrounds(x, y, z) {
			idx := brk.Index(x, y, z)
			if brk.Presence.Get(idx) {
				rayhit.Hit = true
				rayhit.Time = time
				rayhit.IntPos = [3]int{x, y, z}
				rayhit.Position = ray.Origin.Add(ray.Dir.Mul(time))
				rayhit.Color = brk.Color[idx]
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

// Just used to silence the gopls errors
func ErrorSilent[T any](v T, a ...T) {

}
