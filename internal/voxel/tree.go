package voxel

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
	mx, my, mz := lx+(bb.high[0]-bb.low[0])/2, ly+(bb.high[1]-bb.low[1])/2, lz+(bb.high[2]-bb.low[2])/2

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
	root *TreeNode
}

func OctreeInit(bounds AABB) Octree {
	return Octree{root: TreeNodeInit(bounds)}
}

func (tree *Octree) Insert(voxel [3]int) {
	// TODO:
	ErrorSilent(tree)
}

// Just used to silence the gopls errors
func ErrorSilent[T any](v T, a ...T) {

}
