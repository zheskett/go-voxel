package voxel

import (
	"github.com/zheskett/go-voxel/internal/tensor"
)

// This should later be changed to 4^3 so that we can fit in a single u64
const (
	BrickSize  int = 8
	BrickTotal int = BrickSize * BrickSize * BrickSize
)

// An integer 3D vector
type Vec3i struct {
	X, Y, Z int
}

func Vec3(x, y, z int) Vec3i {
	return Vec3i{x, y, z}
}

func (v1 Vec3i) AsArray() [3]int {
	return [3]int{v1.X, v1.Y, v1.Z}
}

func (v1 Vec3i) AsVec3f() tensor.Vector3 {
	return tensor.Vec3(float32(v1.X), float32(v1.Y), float32(v1.Z))
}

func (v1 Vec3i) Add(v2 Vec3i) Vec3i {
	return Vec3(v1.X+v2.X, v1.Y+v2.Y, v1.Z+v2.Z)
}

func (v1 Vec3i) Sub(v2 Vec3i) Vec3i {
	return Vec3(v1.X-v2.X, v1.Y-v2.Y, v1.Z-v2.Z)
}

func (v1 Vec3i) Mul(s int) Vec3i {
	return Vec3(v1.X*s, v1.Y*s, v1.Z*s)
}

// Need to be careful becuase this is integer division
func (v1 Vec3i) Div(s int) Vec3i {
	return Vec3(v1.X/s, v1.Y/s, v1.Z/s)
}

// Axis Aligned Bounding Box
type Box struct {
	low  Vec3i // This storage method can be simplified this is just the easiest
	high Vec3i
}

func BoxInit(lx, ly, lz int, hx, hy, hz int) Box {
	return Box{low: Vec3(lx, ly, lz), high: Vec3(hx, hy, hz)}
}

func (box *Box) size() Vec3i {
	return box.high.Sub(box.low)
}

func (box *Box) center() Vec3i {
	return box.low.Add(box.high).Div(2)
}

// Returns if a point is fully encased by the box. The convention we are using is [min, max)
func (box *Box) surrounds(v Vec3i) bool {
	return v.X >= box.low.X && v.Y >= box.low.Y && v.Z >= box.low.Z &&
		v.X < box.high.X && v.Y < box.high.Y && v.Z < box.high.Z
}

// Slab-method of AABB and ray intersection
func (box *Box) rayIntersection(ray Ray) (float32, float32) {
	tmin := float32(0.0)
	tmax := ray.Tmax
	dirs := ray.Dir.AsArray()
	orig := ray.Origin.AsArray()
	low := box.low.AsArray()
	high := box.high.AsArray()

	for i := range 3 {
		if dirs[i] == 0.0 {
			if orig[i] < float32(low[i]) || orig[i] >= float32(high[i]) {
				return 1, 0
			}
			continue
		}
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

	return tmin, tmax
}

func (box *Box) subdivide() [8]Box {
	low, high := box.low.AsArray(), box.high.AsArray()
	lx, ly, lz := low[0], low[1], low[2]
	hx, hy, hz := high[0], high[1], high[2]
	mx, my, mz := (high[0]+low[0])/2, (high[1]+low[1])/2, (high[2]+low[2])/2

	return [8]Box{
		BoxInit(lx, ly, lz, mx, my, mz), // 000
		BoxInit(lx, ly, mz, mx, my, hz), // 001
		BoxInit(lx, my, lz, mx, hy, mz), // 010
		BoxInit(lx, my, mz, mx, hy, hz), // 011
		BoxInit(mx, ly, lz, hx, my, mz), // 100
		BoxInit(mx, ly, mz, hx, my, hz), // 101
		BoxInit(mx, my, lz, hx, hy, mz), // 110
		BoxInit(mx, my, mz, hx, hy, hz), // 111
	}
}

// Returns the linear index into the Box assuming relative coordinates [0, 2)
func (box *Box) index(x, y, z int) int {
	if x > 1 || y > 1 || z > 1 {
		panic("using relative indexing")
	}
	return (x << 2) | (y << 1) | z
}

func canBrick(box *Box) bool {
	size := box.size()
	return size.X == BrickSize && size.Y == BrickSize && size.Z == BrickSize
}

type TreeWalker struct {
	node  *TreeNode
	level int
}

func TreeWalkerInit(tree *BrickTree) TreeWalker {
	return TreeWalker{&tree.Root, 0}
}

// Climbs the walker to the closest upward stem
func (tw *TreeWalker) Ascend() {
	tw.node = tw.node.Stem
	tw.level -= 1

	if tw.level < 0 {
		panic("error descending tree")
	}
}

// Drops down a level into the relative indexed node
func (tw *TreeWalker) Descend(x, y, z int) {
	idx := tw.node.Box.index(x, y, z)
	tw.node = tw.node.Leaves[idx]
	tw.level += 1

	if tw.level > 32 {
		panic("error descending tree")
	}
}

func (tw *TreeWalker) GotoAbsolute(x, y, z int) {
	pos := Vec3(x, y, z)
	// If the current node doesn't surround our point, ascend and try again
	if !tw.node.Box.surrounds(pos) {
		tw.Ascend()
		tw.GotoAbsolute(x, y, z)
	}

	// Need someway to keep descending until we hit a leaf/empty cell that is the lowest container
	if tw.node.IsStem() {
		center := tw.node.Box.center()
		node := tw.GetOctantCoords(pos, center)
		tw.Descend(node.X, node.Y, node.Z)
		tw.GotoAbsolute(x, y, z)
	}

	// Otherwise, it is empty or a brick and is the smallest possible container
}

func (tw *TreeWalker) GetOctantCoords(pos, center Vec3i) Vec3i {
	var x, y, z int
	if pos.X < center.X {
		x = 0
	} else {
		x = 1
	}
	if pos.Y < center.Y {
		y = 0
	} else {
		y = 1
	}
	if pos.Z < center.Z {
		z = 0
	} else {
		z = 1
	}
	return Vec3(x, y, z)
}

// Doubly linked octant node
type TreeNode struct {
	Box    Box
	Brick  *Brick
	Stem   *TreeNode
	Leaves [8]*TreeNode
}

func TreeNodeInit(box Box, stem *TreeNode) TreeNode {
	return TreeNode{box, nil, stem, [8]*TreeNode{}}
}

// If we are at the top of the tree
func (node *TreeNode) IsRoot() bool {
	return node.Stem == nil
}

// Basically returns if we can jump that entire octant
func (node *TreeNode) IsEmtpy() bool {
	return node.Brick == nil && node.Leaves[0] == nil
}

// Has leaves that need to be searched in order
func (node *TreeNode) IsStem() bool {
	return node.Brick == nil && node.Leaves[0] != nil
}

// Has an active brick that we need to search
func (node *TreeNode) IsLeaf() bool {
	return node.Brick != nil && node.Leaves[0] == nil
}

func (node *TreeNode) RecursiveInsert(x, y, z int, r, g, b byte) bool {
	pos := Vec3(x, y, z)

	// Point isn't in the tree
	if !node.Box.surrounds(pos) {
		return false
	}

	// We already have a brick, so just put the voxel in it
	if node.IsLeaf() {
		node.insertLocalBrick(pos, r, g, b)
		return true
	}

	// There is no brick, but one can be directly created
	if node.IsEmtpy() && canBrick(&node.Box) {
		brick := BrickInit()
		node.Brick = &brick
		node.insertLocalBrick(pos, r, g, b)
		return true
	}

	// Otherwise, we need to split apart into bricks
	if !node.IsStem() {
		node.subdivide()
	}

	for i := range 8 {
		if node.Leaves[i].RecursiveInsert(x, y, z, r, g, b) {
			return true
		}
	}

	return false
}

func (node *TreeNode) insertLocalBrick(pos Vec3i, r byte, g byte, b byte) {
	localpos := pos.Sub(node.Box.low)
	node.Brick.Set(localpos.X, localpos.Y, localpos.Z, r, g, b)
}

func (node *TreeNode) subdivide() {
	parts := node.Box.subdivide()
	for i := range 8 {
		child := TreeNodeInit(parts[i], node)
		node.Leaves[i] = &child
	}
}

func (node *TreeNode) MarchRay(ray Ray) RayHit {
	rayhit := RayHit{Hit: false}

	tmin, tmax := node.Box.rayIntersection(ray)
	if tmax < tmin || tmin > ray.Tmax {
		return rayhit // Never hits the bounding box
	}

	if node.IsLeaf() {
		originatentry := ray.Origin.Add(ray.Dir.Mul(tmin))
		localorigin := originatentry.Sub(tensor.Vec3(
			float32(node.Box.low.X),
			float32(node.Box.low.Y),
			float32(node.Box.low.Z),
		))

		brickray := Ray{
			Origin: localorigin,
			Dir:    ray.Dir,
			Tmax:   tmax - tmin,
		}

		hit := node.Brick.MarchRay(brickray)
		if hit.Hit {
			hit.Time += tmin
			hit.Position = ray.Origin.Add(ray.Dir.Mul(hit.Time))
			hit.IntPos[0] += node.Box.low.X
			hit.IntPos[1] += node.Box.low.Y
			hit.IntPos[2] += node.Box.low.Z
			return hit
		}
	}

	// If it is a branch, recursively dive into each leaf
	// This can be heavily optimized with some bitmasking trick instead of checking
	// all 8 leaves, but I don't understand how to do that optimization yet
	//
	// Basically, this always has to check all 8 leaves while one average it should
	// only take 4 checks to find a hit
	if node.IsStem() {
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
		panic("Current tree must be multiples of 8 until it is working properly")
	}

	// Currently, the whole tree is 'lopsided' to one side and not centered around zero
	// to allow for direct translation from the array storage without coordinate system
	// transformations
	return BrickTree{TreeNodeInit(BoxInit(0, 0, 0, x, y, z), nil)}
}

func (bt *BrickTree) Insert(x, y, z int, r, g, b byte) bool {
	return bt.Root.RecursiveInsert(x, y, z, r, g, b)
}

// Entry point for sending a ray into the tree
func (bt *BrickTree) MarchRay(ray Ray) RayHit {
	return bt.Root.MarchRay(ray)
}

type SubMarch interface {
	March(ray Ray, march MarchData) RayHit
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

func (brk *Brick) Set(x, y, z int, r, g, b byte) {
	idx := brk.Index(x, y, z)
	brk.Presence.Set(idx)
	brk.Color[idx] = [3]byte{r, g, b}
}

func (brk *Brick) Reset(x, y, z int) {
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

	march := MarchDataInit(ray)

	for {
		if march.Time > ray.Tmax {
			break
		}
		if brk.Surrounds(march.Pos.X, march.Pos.Y, march.Pos.Z) {
			idx := brk.Index(march.Pos.X, march.Pos.Y, march.Pos.Z)
			if brk.Presence.Get(idx) {
				rayhit.Hit = true
				rayhit.Time = march.Time
				rayhit.IntPos = [3]int{march.Pos.X, march.Pos.Y, march.Pos.Z}
				rayhit.Position = ray.Origin.Add(ray.Dir.Mul(march.Time))
				rayhit.Color = brk.Color[idx]
				switch march.Side {
				case axisX:
					rayhit.Normal = tensor.Vec3(1, 0, 0).Mul(-float32(march.Step.X))
				case axisY:
					rayhit.Normal = tensor.Vec3(0, 1, 0).Mul(-float32(march.Step.Y))
				case axisZ:
					rayhit.Normal = tensor.Vec3(0, 0, 1).Mul(-float32(march.Step.Z))
				default:
					rayhit.Normal = tensor.Vec3(0, 0, 0)
				}
				break
			}
		}

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

	return rayhit
}
