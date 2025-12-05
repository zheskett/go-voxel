package main

import (
	"fmt"

	"github.com/zheskett/go-voxel/internal/tensor"
	"github.com/zheskett/go-voxel/internal/voxel"
)

func main() {
	size := 16
	tree := voxel.BrickTreeInit(size, size, size)
	if tree.Root.Brick != nil {
		panic("what??")
	}
	if tree.Root.IsBranch() {
		panic("don't understand")
	}
	if tree.Root.IsLeaf() {
		panic("there shouldn't be data there yet")
	}
	if !tree.Root.IsEmtpy() {
		panic("why is there no assert")
	}

	for i := range size {
		tree.Insert(i, i, i, 0, 0, 0)
	}

	depth := maxDepth(&tree)
	fmt.Printf("max tree depth: %d\n", depth)
	voxels := countVoxels(&tree)
	fmt.Printf("voxels in the tree: %d\n", voxels)
	bricks := countBricks(&tree)
	fmt.Printf("bricks in the tree: %d\n", bricks)

	fmt.Printf("done\n")

	tree = voxel.BrickTreeInit(size, size, size)
	tree.Insert(10, 10, 10, 0, 255, 255)
	ray := voxel.Ray{Origin: tensor.Vec3(1, 1, 1), Dir: tensor.Vec3(1, 1, 1).Normalized(), Tmax: 1e4}
	hit := tree.MarchRay(ray)
	fmt.Printf("direct rayhit: %+v\n", hit)
}

func countBricks(br *voxel.BrickTree) int {
	return recurCountBricks(&br.Root)
}

func recurCountBricks(node *voxel.TreeNode) int {
	if node == nil {
		return 0
	}
	if node.IsLeaf() {
		return 1
	}
	bricks := 0
	for i := range 8 {
		bricks += recurCountBricks(node.Leaves[i])
	}
	return bricks
}

func countVoxels(br *voxel.BrickTree) int {
	return recurCountVoxels(&br.Root)
}

func recurCountVoxels(node *voxel.TreeNode) int {
	if node == nil {
		return 0
	}
	if node.IsLeaf() {
		count := 0
		for i := range voxel.BrickTotal {
			if node.Brick.Presence.Get(i) {
				count++
			}
		}
		return count
	}
	if node.IsBranch() {
		count := 0
		for i := range 8 {
			count += recurCountVoxels(node.Leaves[i])
		}
		return count
	}
	return 0
}

func maxDepth(br *voxel.BrickTree) int {
	return recurMaxDepth(&br.Root)
}

func recurMaxDepth(node *voxel.TreeNode) int {
	if node.IsLeaf() {
		return 0
	}
	if node.IsBranch() {
		maxdepth := 0
		for i := range 8 {
			if node.Leaves[i] != nil {
				depth := recurMaxDepth(node.Leaves[i])
				maxdepth = max(maxdepth, depth)
			}
		}
		return maxdepth + 1
	}
	return 0
}
