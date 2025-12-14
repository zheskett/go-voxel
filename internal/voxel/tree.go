package voxel

import (
	"github.com/chewxy/math32"
	te "github.com/zheskett/go-voxel/internal/tensor"
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
	tmin := float32(0.0)
	tmax := math32.Inf(1)
	dirs := ray.Dir.AsArray()
	orig := ray.Origin.AsArray()
	low := box.Low.AsArray()
	high := box.high().AsArray()

	for i := range 3 {
		if dirs[i] == 0.0 {
			if orig[i] < float32(low[i]) || orig[i] >= float32(high[i]) {
				return math32.Inf(1), math32.Inf(-1)
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
			return math32.Inf(1), math32.Inf(-1) // There isn't an intersection
		}
	}

	if tmin < 0.0 {
		panic("this shouldn't happen")
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
func (box *Box) index(x, y, z int) int {
	if x > 1 || y > 1 || z > 1 || x < 0 || y < 0 || z < 0 {
		panic("using relative indexing")
	}
	return (x << 2) | (y << 1) | z
}

type Voxel struct {
	Present bool
	Color   [3]byte
}

func VoxelInit() Voxel {
	return Voxel{Present: false, Color: [3]byte{0, 0, 0}}
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

// Has leaves that need to be searched in order
func (node *TreeNode) IsStem() bool {
	return node.Leaves[0] != nil
}

// Has a voxel
func (node *TreeNode) IsLeaf() bool {
	return node.Leaves[0] == nil
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

	octantcoords := GetOctantCoords(pos, node.Box)
	linear := node.Box.index(octantcoords.X, octantcoords.Y, octantcoords.Z)
	return node.Leaves[linear].RecursiveInsert(x, y, z, r, g, b)
}

func (node *TreeNode) subdivide() {
	parts := node.Box.subdivide()
	for i := range 8 {
		node.Leaves[i] = TreeNodeInit(parts[i], node)
	}
}

type Octree struct {
	Root    TreeNode
	Z, Y, X int // For compatibility with the old ones
}

func OctreeInit(size int) Octree {
	// Currently, the whole tree is 'lopsided' to one side and not centered around zero
	// to allow for direct translation from the array storage without coordinate system
	// transformations
	return Octree{*TreeNodeInit(BoxInit(0, 0, 0, size), nil), size, size, size} // Root has no stem
}

func (oc *Octree) Insert(x, y, z int, r, g, b byte) bool {
	return oc.Root.RecursiveInsert(x, y, z, r, g, b)
}

// Entry point for sending a ray into the tree
func (oc *Octree) MarchRay(ray Ray) RayHit {
	rayhit := RayHit{Hit: false}

	tmin, tmax := oc.Root.Box.RayIntersection(ray)
	if tmin > tmax {
		return rayhit // The ray misses the the whole tree
	}

	if tmin < 0.0 {
		panic("should have been clamped by the slab AABB")
	}

	data := MarchDataInit(tmin, tmax, ray)
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

// Climbs the walker to the closest upward stem
func (tw *TreeWalker) Ascend() {
	tw.Node = tw.Node.Stem
	tw.level -= 1

	if tw.level < 0 {
		panic("error ascending tree")
	}
}

// Drops down a level into the relative indexed node
func (tw *TreeWalker) Descend(x, y, z int) {
	idx := tw.Node.Box.index(x, y, z)
	tw.Node = tw.Node.Leaves[idx]
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
		oct := GetOctantCoords(pos, tw.Node.Box)
		if !tw.Node.Box.surrounds(te.Vec3i(x, y, z)) {
			panic("why is this happening to me")
		}
		tw.Descend(oct.X, oct.Y, oct.Z)
	}
}

func GetOctantCoords(pos te.Vector3i, box Box) te.Vector3i {
	if !box.surrounds(pos) {
		panic("The box doesn't contain the voxel even")
	}
	center := box.center()
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
	return te.Vec3i(x, y, z)
}

func (tw *TreeWalker) StateMarchRay(ray Ray, data MarchData) RayHit {
	rayhit := RayHit{Hit: false}
	for {
		if data.Time > data.Tmax {
			break
		}

		pos := ray.Origin.Add(ray.Dir.Mul(data.Time))
		x, y, z := int(math32.Floor(pos.X)), int(math32.Floor(pos.Y)), int(math32.Floor(pos.Z))
		tw.GotoAbsolute(x, y, z)

		if tw.Node.IsEmpty() {
			_, nodeexit := tw.Node.Box.RayIntersection(ray)
			data.Time = math32.Max(nodeexit, data.Time+VoxelRayDelta)
			continue
		}

		if tw.Node.IsLeaf() {
			return RayHit{
				Hit:   true,
				Color: tw.Node.Voxel.Color,
			}
		}

		data.Time += VoxelRayDelta
	}

	return rayhit
}
