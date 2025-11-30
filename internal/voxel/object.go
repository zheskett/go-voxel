package voxel

type VoxelObj struct {
	// Position of the object in the world
	XPos, YPos, ZPos int
	Vox              Voxels
}
