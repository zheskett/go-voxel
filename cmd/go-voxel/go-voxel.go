package main

import (
	"runtime"

	"github.com/zheskett/go-voxel/internal/engine"
	ren "github.com/zheskett/go-voxel/internal/render"
	"github.com/zheskett/go-voxel/internal/voxel"
	vxl "github.com/zheskett/go-voxel/internal/voxel"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// This is meantioned in the usage example on github.
	runtime.LockOSThread()
}

func main() {
	// vox := vxl.VoxelsInit(256, 256, 256)
	// brk := vxl.BrickTreeInit(256, 256, 256)
	// renderDist := float32(256.0)
	// var scene int
	// fmt.Printf("Enter 1 for the big scene, 2 for room, 3 for big bunny, anything else for small scene\n")
	// fmt.Scanln(&scene)
	// switch scene {
	// case 1:
	// 	scenes.VoxelDebugSceneBig(&vox)
	// case 2:
	// 	scenes.VoxelDebugEmptyScene(&vox)
	// case 3:
	// 	scenes.VoxelDebugSceneHugeBunny(&vox)
	// 	renderDist = 560.0
	// default:
	// 	scenes.VoxelDebugSceneSmall(&vox)
	// }

	renderDist := float32(512.0)
	size := 512
	tree := voxel.BrickTreeInit(size, size, size)
	tree.Insert(10, 10, 10, 0, 255, 255)
	rm, window := ren.RenderManagerInit()
	cam := ren.CameraInit()
	cam.Movespeed = 20
	cam.Lookspeed = 0.005
	cam.Fov = 90
	cam.Aspect = float32(rm.Pixels.Width) / float32(rm.Pixels.Height)
	cam.RenderDistance = renderDist

	engine := engine.Engine{}
	engine.Renderer = rm
	engine.Window = window
	engine.Camera = cam
	// engine.Voxels = vox
	engine.Voxtree = tree
	engine.Framedata = ren.FrameDataInit()
	engine.SetScrollCallback()

	VoxelDebugSceneSmall(&tree)
	// brk.Insert(0, 0, 10, 255, 0, 0)

	for {
		engine.UpdateInputs()
		engine.UpdateRender()
		engine.CheckExit()
	}
}

func VoxelDebugSceneSmall(vox *vxl.BrickTree) {
	// Make a floor and ceiling
	for i := 1; i < 256; i++ {
		for j := 1; j < 256; j++ {
			vox.Insert(i, 0, j, 220, 180, 180)
			vox.Insert(i, 40, j, 180, 180, 180)
		}
	}
	// Make walls
	for i := range 100 {
		for j := range 100 {
			vox.Insert(100, i, j, 200, 180, 180)
			vox.Insert(0, i, j, 180, 220, 180)
			vox.Insert(j, i, 0, 200, 180, 180)
			vox.Insert(j, i, 100, 180, 180, 220)
		}
	}
	// Make a small wall for shadows
	for i := range 55 {
		for j := range 100 {
			vox.Insert(35, j, i, 200, 180, 180)
			vox.Insert(36, j, i, 200, 180, 180)
		}
	}

	for i := 60; i < 70; i++ {
		for j := 28; j < 40; j++ {
			for k := 60; k < 70; k++ {
				vox.Insert(i, j, k, 200, 200, 200)
			}
		}
	}

	for i := range 100 {
		for j := range 100 {
			for k := range 100 {
				if i%10 == 0 && j%10 == 0 && k%10 == 0 {
					vox.Insert(i, j, k, 200, 200, 200)
				}
			}
		}
	}
}
