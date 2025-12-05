package voxel

import (
	"github.com/chewxy/math32"
	"github.com/zheskett/go-voxel/internal/tensor"
)

const (
	BrickSize  int = 8
	BrickTotal int = BrickSize * BrickSize * BrickSize
)

// An integer 3D vector
type IVec3 struct {
	x, y, z int
}

func IV3(x, y, z int) IVec3 {
	return IVec3{x, y, z}
}

func (v IVec3) AsArray() [3]int {
	return [3]int{v.x, v.y, v.z}
}

func (v1 IVec3) Add(v2 IVec3) IVec3 {
	return IV3(v1.x+v2.x, v1.y+v2.y, v1.z+v2.z)
}

func (v1 IVec3) Sub(v2 IVec3) IVec3 {
	return IV3(v1.x-v2.x, v1.y-v2.y, v1.z-v2.z)
}

// Axis Aligned Bounding Box
type AABB struct {
	Low  IVec3 // This storage method can be simplified this is just the easiest
	High IVec3
}

func AABBInit(lx, ly, lz int, hx, hy, hz int) AABB {
	return AABB{Low: IV3(lx, ly, lz), High: IV3(hx, hy, hz)}
}

func (box *AABB) Size() IVec3 {
	return box.High.Sub(box.Low)
}

// Returns if a point is fully encased by the box. The convention we are using is [min, max)
func (box *AABB) Surrounds(v IVec3) bool {
	return v.x >= box.Low.x && v.y >= box.Low.y && v.z >= box.Low.z &&
		v.x < box.High.x && v.y < box.High.y && v.z < box.High.z
}

// Slab-method of AABB and ray intersection
func (box *AABB) RayIntersection(ray Ray) (float32, float32) {
	tmin := float32(0.0)
	tmax := ray.Tmax
	dirs := ray.Dir.AsArray()
	orig := ray.Origin.AsArray()
	low := box.Low.AsArray()
	high := box.High.AsArray()
	for i := range 3 {
		if dirs[i] != 0.0 {
			invd := 1.0 / dirs[i]
			t0 := (float32(low[i]) - orig[i]) * invd
			t1 := (float32(high[i]) - orig[i]) * invd

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

func (box *AABB) Subdivide() [8]AABB {
	low, high := box.Low.AsArray(), box.High.AsArray()
	lx, ly, lz := low[0], low[1], low[2]
	hx, hy, hz := high[0], high[1], high[2]
	mx, my, mz := lx+(high[0]-low[0])/2, ly+(high[1]-low[1])/2, lz+(high[2]-low[2])/2

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

func TreeNodeInit(box AABB) TreeNode {
	return TreeNode{box, nil, [8]*TreeNode{}}
}

func (node *TreeNode) IsBranch() bool {
	return node.Brick == nil && node.Leaves[0] != nil
}

func (node *TreeNode) IsLeaf() bool {
	return node.Brick != nil
}

func (node *TreeNode) recursiveInsert(x, y, z int, r, g, b byte) bool {
	pos := IV3(x, y, z)

	// Point isn't in the tree
	if !node.Box.Surrounds(pos) {
		return false
	}
	// We already have a brick, so just put the voxel in it
	if node.IsLeaf() {
		localpos := pos.Sub(node.Box.Low)
		node.Brick.SetVoxel(localpos.x, localpos.y, localpos.z, r, g, b)
		return true
	}

	boxsize := node.Box.Size()
	// There is no brick, but one can be directly created
	if boxsize.x == BrickSize && boxsize.y == BrickSize && boxsize.z == BrickSize {
		brick := BrickInit()
		node.Brick = &brick
		localpos := pos.Sub(node.Box.Low)
		node.Brick.SetVoxel(localpos.x, localpos.y, localpos.z, r, g, b)
		return true
	}

	// Otherwise, we need to split apart into bricks
	node.Subdivide()
	for i := range 8 {
		if node.Leaves[i].recursiveInsert(x, y, z, r, g, b) {
			return true
		}
	}

	return false
}

func (node *TreeNode) Subdivide() {
	parts := node.Box.Subdivide()
	for i := range 8 {
		child := TreeNodeInit(parts[i])
		node.Leaves[i] = &child
	}
}

func (node *TreeNode) MarchRay(ray Ray) RayHit {
	rayhit := RayHit{Hit: false}

	tmin, tmax := node.Box.RayIntersection(ray)
	if tmax < tmin || tmin > ray.Tmax {
		return rayhit // Never hits the bounding box
	}

	if node.IsLeaf() {
		// If we do hit the bounding box, shift the ray origin right next to it and
		// traverse the brick just like we would with the dense voxel storage
		localorigin := ray.Origin.Sub(tensor.Vec3(float32(node.Box.Low.x), float32(node.Box.Low.y), float32(node.Box.Low.z)))

		localray := Ray{
			Origin: localorigin,
			Dir:    ray.Dir,
			Tmax:   ray.Tmax - tmin,
		}

		hit := node.Brick.MarchRay(localray)
		if hit.Hit {
			hit.Time += tmin
			hit.Position = ray.Origin.Add(ray.Dir.Mul(hit.Time))

			hit.IntPos[0] += node.Box.Low.x
			hit.IntPos[1] += node.Box.Low.y
			hit.IntPos[2] += node.Box.Low.z
		}
		return hit
	}

	// If it is a branch, recursively dive into each leaf
	// This can be heavily optimized with some bitmasking trick instead of checking
	// all 8 leaves, but I don't understand how to do that optimization yet
	//
	// Basically, this always has to check all 8 leaves while one average it should
	// only take 4 checks to find a hit
	if node.IsBranch() {
		closesthit := RayHit{Hit: false}
		closesttime := ray.Tmax

		for i := range 8 {
			if node.Leaves[i] != nil {
				hit := node.Leaves[i].MarchRay(ray)
				if hit.Hit && hit.Time < closesttime {
					closesthit = hit
					closesttime = hit.Time
				}
			}
		}

		return closesthit
	}

	return rayhit
}

type BrickTree struct {
	Root TreeNode
}

func BrickTreeInit(x, y, z int) BrickTree {
	// This is just for now, becuase I can't even get it to work with a self-similar one
	if x%BrickSize != 0 || y%BrickSize != 0 || z%BrickSize != 0 {
		panic("Current tree must be multiples of 64 until it is working properly")
	}

	// 	Currently, the whole tree is 'lopsided' to one side and not centered around zero
	// to allow for direct translation from the array storage without coordinate system
	// transformations
	return BrickTree{TreeNodeInit(AABBInit(0, 0, 0, x, y, z))}
}

func (bt *BrickTree) Insert(x, y, z int, r, g, b byte) {
	bt.Root.recursiveInsert(x, y, z, r, g, b)
}

func (bt *BrickTree) MarchRay(ray Ray) RayHit {
	return bt.Root.MarchRay(ray)
}

// This is basically a slimmed down clone of the old Voxel struct.
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
