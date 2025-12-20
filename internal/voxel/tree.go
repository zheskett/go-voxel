package voxel

import (
	"github.com/chewxy/math32"
	te "github.com/zheskett/go-voxel/internal/tensor"
)

const (
	axisXBit = 1 << 2
	axisYBit = 1 << 1
	axisZBit = 1 << 0
)

// Axis Aligned Bounding Box
type Box struct {
	Low  te.Vector3i // Stores the low position of a cubic bounding box
	Size int         // The full side lenghts of the cubic bounding box
}

func BoxInit(lx, ly, lz int, size int) Box {
	return Box{Low: te.Vec3i(lx, ly, lz), Size: size}
}

func (box *Box) isUnit() bool {
	return box.Size == 1
}

func (box *Box) center() te.Vector3i {
	if box.Size%2 != 0 {
		panic("There is a large problem if this is ever called on a unit AABB")
	}
	return box.Low.Add(te.Vec3iSplat(box.Size / 2))
}

func (box *Box) high() te.Vector3i {
	return box.Low.Add(te.Vec3iSplat(box.Size))
}

// Returns if a point is fully encased by the box. The convention we are using is [min, max)
func (box *Box) surrounds(v te.Vector3i) bool {
	high := box.high()
	return v.X >= box.Low.X && v.Y >= box.Low.Y && v.Z >= box.Low.Z &&
		v.X < high.X && v.Y < high.Y && v.Z < high.Z
}

// Slab-method of AABB and ray intersection
// Returns (min_t, max_t) corresponding to the rays' entrance and exit time
func (box *Box) RayIntersection(ray Ray) (float32, float32) {
	// TODO: This needs to not be recalcualted each time -- either take MarchData or store that on the ray
	invs := ray.Dir.Inv()
	low := box.Low.AsVec3f()
	high := box.high().AsVec3f()

	tx1 := (low.X - ray.Origin.X) * invs.X
	tx2 := (high.X - ray.Origin.X) * invs.X
	tmin := min(tx1, tx2)
	tmax := max(tx1, tx2)

	ty1 := (low.Y - ray.Origin.Y) * invs.Y
	ty2 := (high.Y - ray.Origin.Y) * invs.Y
	tmin = max(tmin, min(ty1, ty2))
	tmax = min(tmax, max(ty1, ty2))

	tz1 := (low.Z - ray.Origin.Z) * invs.Z
	tz2 := (high.Z - ray.Origin.Z) * invs.Z
	tmin = max(tmin, min(tz1, tz2))
	tmax = min(tmax, max(tz1, tz2))

	if tmax < tmin || tmax < 0 {
		return math32.Inf(1), math32.Inf(-1)
	}

	return tmin, tmax
}

func (box *Box) subdivide() [8]Box {
	low, mid := box.Low.AsArray(), box.center().AsArray()
	lx, ly, lz := low[0], low[1], low[2]
	mx, my, mz := mid[0], mid[1], mid[2]
	halfsize := box.Size / 2

	return [8]Box{
		BoxInit(lx, ly, lz, halfsize), // 000
		BoxInit(lx, ly, mz, halfsize), // 001
		BoxInit(lx, my, lz, halfsize), // 010
		BoxInit(lx, my, mz, halfsize), // 011
		BoxInit(mx, ly, lz, halfsize), // 100
		BoxInit(mx, ly, mz, halfsize), // 101
		BoxInit(mx, my, lz, halfsize), // 110
		BoxInit(mx, my, mz, halfsize), // 111
	}
}

// Returns the linear index into the Box assuming relative coordinates [0, 2)
func (box *Box) Index(x, y, z int) int {
	center := box.center()
	index := 0
	if !box.surrounds(te.Vec3i(x, y, z)) {
		panic("This index function is only valid for a point containted within the box")
	}
	if x >= center.X {
		index |= axisXBit
	}
	if y >= center.Y {
		index |= axisYBit
	}
	if z >= center.Z {
		index |= axisZBit
	}
	return index
}

// This function is pretty slow and I am almost certain that there is a more
// efficient way to get this but I don't know a method yet
func (box *Box) GetHitNormal(ray Ray) te.Vector3 {
	invs := ray.Dir.Inv()
	low := box.Low.AsVec3f()
	high := box.high().AsVec3f()

	tx1 := (low.X - ray.Origin.X) * invs.X
	tx2 := (high.X - ray.Origin.X) * invs.X
	ty1 := (low.Y - ray.Origin.Y) * invs.Y
	ty2 := (high.Y - ray.Origin.Y) * invs.Y
	tz1 := (low.Z - ray.Origin.Z) * invs.Z
	tz2 := (high.Z - ray.Origin.Z) * invs.Z

	tex := min(tx1, tx2)
	tey := min(ty1, ty2)
	tez := min(tz1, tz2)

	// Goodness gracious
	normal := te.Vec3Zero()
	if tex > tey && tex > tez {
		if ray.Dir.X > 0.0 {
			normal = te.Vec3X().Neg()
		} else {
			normal = te.Vec3X()
		}
	} else if tey > tez {
		if ray.Dir.Y > 0.0 {
			normal = te.Vec3Y().Neg()
		} else {
			normal = te.Vec3Y()
		}
	} else {
		if ray.Dir.Z > 0.0 {
			normal = te.Vec3Z().Neg()
		} else {
			normal = te.Vec3Z()
		}
	}

	return normal
}

// The same as 'CachedLighting' but with a tick that it was set
type VoxelLighting struct {
	Light te.Vector3
	Dir   te.Vector3
	Tick  uint
}

func VoxelLightingInit() VoxelLighting {
	return VoxelLighting{Light: te.Vec3Zero(), Dir: te.Vec3Zero(), Tick: 0}
}

type Voxel struct {
	Present bool
	Color   [3]byte
	Light   VoxelLighting
}

func VoxelInit() Voxel {
	return Voxel{Present: false, Light: VoxelLightingInit()}
}

// Doubly linked octant node
type TreeNode struct {
	Box    Box
	Voxel  Voxel
	Stem   *TreeNode
	Leaves [8]*TreeNode
}

func TreeNodeInit(box Box, stem *TreeNode) *TreeNode {
	return &TreeNode{box, VoxelInit(), stem, [8]*TreeNode{}}
}

// If we are at the top of the tree
func (node *TreeNode) IsRoot() bool {
	return node.Stem == nil
}

// Basically returns if we can jump that entire octant
func (node *TreeNode) IsEmpty() bool {
	return !node.Voxel.Present
}

// Has a voxel
func (node *TreeNode) IsLeaf() bool {
	return node.Voxel.Present
}

// Has leaves that need to be searched in order
func (node *TreeNode) IsStem() bool {
	return node.Leaves[0] != nil
}

func (node *TreeNode) query(pos te.Vector3i) *Voxel {
	if node.IsLeaf() {
		return &node.Voxel
	}
	if node.IsStem() {
		index := node.Box.Index(pos.X, pos.Y, pos.Z)
		return node.Leaves[index].query(pos)
	}
	return nil // This shouldn't ever happen, this shouldn't be called blinding without knowing a voxel is there
}

func (node *TreeNode) RecursiveInsert(x, y, z int, r, g, b byte) bool {
	pos := te.Vec3i(x, y, z)

	// Point isn't in the tree
	if !node.Box.surrounds(pos) {
		return false
	}

	// There is no brick, but one can be directly created
	if node.Box.isUnit() {
		node.Voxel.Present = true
		node.Voxel.Color = [3]byte{r, g, b}
		return true
	}

	// Otherwise, we need to split apart into bricks
	if !node.IsStem() {
		node.subdivide()
	}

	index := node.Box.Index(pos.X, pos.Y, pos.Z)
	return node.Leaves[index].RecursiveInsert(x, y, z, r, g, b)
}

func (node *TreeNode) RecursiveRemove(x, y, z int) {
	pos := te.Vec3i(x, y, z)
	if !node.Box.surrounds(pos) {
		return
	}

	if node.IsLeaf() {
		node.Voxel = Voxel{}
		return
	}

	index := node.Box.Index(pos.X, pos.Y, pos.Z)
	node.Leaves[index].RecursiveRemove(x, y, z)
}

func (node *TreeNode) subdivide() {
	parts := node.Box.subdivide()
	for i := range 8 {
		node.Leaves[i] = TreeNodeInit(parts[i], node)
	}
}

type Octree struct {
	Root TreeNode
}

func OctreeInit(size int) Octree {
	// Currently, the whole tree is 'lopsided' to one side and not centered around zero
	// to allow for direct translation from the array storage without coordinate system
	// transformations
	return Octree{*TreeNodeInit(BoxInit(0, 0, 0, size), nil)} // Root has no stem
}

func (oc *Octree) Insert(x, y, z int, r, g, b byte) bool {
	return oc.Root.RecursiveInsert(x, y, z, r, g, b)
}

func (oc *Octree) Remove(x, y, z int) {
	oc.Root.RecursiveRemove(x, y, z)
}

func (oc *Octree) GetVoxel(x, y, z int) *Voxel {
	return oc.Root.query(te.Vec3i(x, y, z))
}

// Entry point for sending a ray into the tree
func (oc *Octree) MarchRay(ray Ray) RayHit {
	rayhit := RayHit{Hit: false}

	tmin, tmax := oc.Root.Box.RayIntersection(ray)
	if tmin > tmax {
		return rayhit // The ray misses the the whole tree
	}

	data := MarchDataInit(max(0.0, tmin), min(tmax, ray.Tmax), ray)
	walker := TreeWalkerInit(oc)
	return walker.StateMarchRay(ray, data)
}

type TreeWalker struct {
	Node  *TreeNode
	level int
}

func TreeWalkerInit(tree *Octree) TreeWalker {
	return TreeWalker{&tree.Root, 0}
}

// Climbs the tree to the closest upward stem
func (tw *TreeWalker) Ascend() {
	tw.Node = tw.Node.Stem
	tw.level -= 1

	if tw.level < 0 {
		panic("error ascending tree")
	}
}

// Drops down a level into the relative indexed node
func (tw *TreeWalker) Descend(index int) {
	tw.Node = tw.Node.Leaves[index]
	tw.level += 1

	if tw.level > 32 {
		panic("error descending tree")
	}
}

func (tw *TreeWalker) GotoAbsolute(x, y, z int) {
	pos := te.Vec3i(x, y, z)

	// Climb up until the node surrounds the desired position
	for !tw.Node.Box.surrounds(pos) {
		if tw.Node.IsRoot() {
			return
		}
		tw.Ascend()
	}

	// Continue descending down into the smallest node that surrounds the position
	for tw.Node.IsStem() {
		if !tw.Node.Box.surrounds(pos) {
			panic("Why is this happening to me")
		}
		index := tw.Node.Box.Index(pos.X, pos.Y, pos.Z)
		tw.Descend(index)
	}
}

func (tw *TreeWalker) StateMarchRay(ray Ray, data MarchData) RayHit {
	rayhit := RayHit{Hit: false}

	// This keeps endlessly looping becuase of null rays? This just
	// keeps the program from freezing until I can figure out how that is happening
	//
	// It has something to do with bounce-rays from lighting being cast on null-hits.
	// The error is likely coming from floating point errors in in grazing rays
	for range 100 {
		if data.Time > data.Tmax {
			break
		}

		pos := ray.Origin.Add(ray.Dir.Mul(data.Time + VoxelRayDelta))
		x, y, z := int(math32.Floor(pos.X)), int(math32.Floor(pos.Y)), int(math32.Floor(pos.Z))
		tw.GotoAbsolute(x, y, z)

		if tw.Node.IsEmpty() {
			_, nodeexit := tw.Node.Box.RayIntersection(ray)
			data.Time = max(nodeexit, data.Time+VoxelRayDelta)
		} else if tw.Node.IsLeaf() {
			return RayHit{
				Hit:      true,
				Time:     data.Time,
				Color:    tw.Node.Voxel.Color,
				IntPos:   [3]int{x, y, z},
				Position: ray.Origin.Add(ray.Dir.Mul(data.Time)),
				Normal:   tw.Node.Box.GetHitNormal(ray),
				Voxel:    &tw.Node.Voxel,
			}
		}
	}

	return rayhit
}
