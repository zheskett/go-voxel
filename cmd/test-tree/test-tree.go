package main

import (
	"fmt"

	"github.com/zheskett/go-voxel/internal/tensor"
	"github.com/zheskett/go-voxel/internal/voxel"
)

func main() {
	size := 128
	tree := voxel.BrickTreeInit(size, size, size)
	if tree.Root.IsStem() {
		panic("should have all nil leaves")
	}

	for i := range size {
		if !tree.Insert(i, i, i, 0, 0, 0) {
			panic("error inserting into tree")
		}
	}

	depth := maxDepth(&tree)
	voxels := countVoxels(&tree)
	fmt.Printf("max tree depth: %d\n", depth)
	fmt.Printf("voxels in the tree: %d\n", voxels)

	tree = voxel.BrickTreeInit(size, size, size)
	tree.Insert(10, 10, 10, 0, 255, 255)
	ray := voxel.Ray{Origin: tensor.Vec3(1, 1, 1), Dir: tensor.Vec3(1, 1, 1).Normalized(), Tmax: 1e4}
	hit := tree.MarchRay(ray)
	fmt.Printf("direct rayhit: %+v\n", hit)
	if !hit.Hit {
		panic("didn't hit tree")
	}

	aabb := voxel.BoxInit(-1, -1, -1, 1, 1, 1)
	ray = voxel.Ray{Origin: tensor.Vec3(-5, 0, 0), Dir: tensor.Vec3(1, 0, 0), Tmax: 10}
	t0, t1 := aabb.RayIntersection(ray)
	if t0 > t1 {
		panic("dirct ray didn't hit")
	}
	ray.Dir = ray.Dir.Mul(-1)
	t0, t1 = aabb.RayIntersection(ray)
	if t0 < t1 {
		panic("ray shouldn't have it")
	}

	fmt.Printf("done\n")
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
