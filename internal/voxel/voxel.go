package voxel

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	te "github.com/zheskett/go-voxel/internal/tensor"
)

// Just a point light
type PointLight struct {
	Position te.Vector3
	Color    te.Vector3 // Can have mag > 1 for a bright light
}

// A directional light source which emulates a point light at infinity with no intensity falloff
type DirLight struct {
	Dir te.Vector3
	Col te.Vector3 // Can NOT have mag > 1 for directional light
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

// Manager type for the state of the voxels
type VoxelWorld struct {
	X, Y, Z int
	Voxels  Octree
	Sun     DirLight
	Lights  []PointLight
}

func VoxelWorldInit(size int) VoxelWorld {
	return VoxelWorld{Voxels: OctreeInit(size), Lights: make([]PointLight, 0), X: size, Y: size, Z: size}
}

func (vox *VoxelWorld) SetVoxel(x, y, z int, r, g, b byte) {
	vox.Voxels.Insert(x, y, z, r, g, b)
}

func (vox *VoxelWorld) ResetVoxel(x, y, z int) {
	vox.Voxels.Remove(x, y, z)
}

// Adds a voxel object to the world
func (vox *VoxelWorld) AddVoxelObj(vObj VoxelObj, x, y, z int) {
	for xyz, cIdx := range vObj.Voxels {
		vx, vy, vz := int(xyz[0]), int(xyz[1]), int(xyz[2])
		if vox.Voxels.Root.Box.surrounds(te.Vec3i(x+vx, y+vy, z+vz)) {
			clr := vObj.ColorPalette[cIdx]
			vox.SetVoxel(x+vx, y+vy, z+vz, clr.R, clr.G, clr.B)
		}
	}
}

// This is super temporary and just a proof of concept
func (vox *VoxelWorld) UpdateInputs(window *glfw.Window, pos te.Vector3, dir te.Vector3) {
	ray := Ray{Origin: pos, Dir: dir, Tmax: 100.0}
	hit := vox.Voxels.MarchRay(ray)
	if !hit.Hit {
		return
	}
	if window.GetMouseButton(glfw.MouseButtonLeft) == glfw.Press {
		x, y, z := hit.IntPos[0], hit.IntPos[1], hit.IntPos[2]
		vox.ResetVoxel(x, y, z)
	}
	if window.GetMouseButton(glfw.MouseButtonRight) == glfw.Press {
		voxel := hit.Position.Add(hit.Normal.Mul(VoxelRayDelta))
		x, y, z := int(voxel.X), int(voxel.Y), int(voxel.Z)
		vox.SetVoxel(x, y, z, 255, 255, 255)
		for i := -1; i <= 1; i++ {
			for j := -1; j <= 1; j++ {
				for k := -1; k <= 1; k++ {
					// Only can add a pure white voxel right now
					vox.SetVoxel(x+i, y+j, z+k, 255, 255, 255)
				}
			}
		}
	}
}
