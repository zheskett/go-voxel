package main

import (
	"fmt"

	"github.com/zheskett/go-voxel/internal/voxel"
)

func main() {
	size := 128
	tree := voxel.OctreeInit(size)
	for i := range size {
		if !tree.Insert(i, i, i, 0, 0, 0) {
			panic("error inserting into tree")
		}
	}

	depth := maxDepth(&tree)
	voxels := countVoxels(&tree)
	fmt.Printf("max tree depth: %d\n", depth)
	fmt.Printf("voxels in the tree: %d\n", voxels)

	walker := voxel.TreeWalkerInit(&tree)
	walker.GotoAbsolute(64, 0, 64)
	assert(walker.Node.Box.Size == 64)
	walker.GotoAbsolute(64, 64, 64)
	assert(walker.Node.Box.Size == 1)
	walker.GotoAbsolute(64, 70, 64)
	fmt.Printf("node: %v", walker.Node.Box)

	fmt.Printf("done\n")
}

func assert(arg bool) {
	if !arg {
		panic("assert failed")
	}
}

func countVoxels(br *voxel.Octree) int {
	return recurCountVoxels(&br.Root)
}

func recurCountVoxels(node *voxel.TreeNode) int {
	if node == nil {
		return 0
	}
	if node.IsLeaf() {
		count := 0
		if node.Voxel.Present {
			count++
		}
		return count
	}
	if node.IsStem() {
		count := 0
		for i := range 8 {
			count += recurCountVoxels(node.Leaves[i])
		}
		return count
	}
	return 0
}

func maxDepth(br *voxel.Octree) int {
	return recurMaxDepth(&br.Root)
}

func recurMaxDepth(node *voxel.TreeNode) int {
	if node.IsLeaf() {
		return 0
	}
	if node.IsStem() {
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
