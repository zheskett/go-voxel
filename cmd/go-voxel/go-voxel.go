package main

import (
	"fmt"
	"runtime"

	"github.com/zheskett/go-voxel/internal/engine"
	ren "github.com/zheskett/go-voxel/internal/render"
	"github.com/zheskett/go-voxel/internal/scenes"
	te "github.com/zheskett/go-voxel/internal/tensor"
	vxl "github.com/zheskett/go-voxel/internal/voxel"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// This is meantioned in the usage example on github.
	runtime.LockOSThread()
}

func main() {
	vox := vxl.VoxelsInit(256, 256, 256)
	renderDist := float32(256.0)
	var scene int
	fmt.Printf("Enter 1 for the big scene, 2 for room, 3 for big bunny, 4 for trees, anything else for small scene\n")
	fmt.Scanln(&scene)
	switch scene {
	case 1:
		scenes.VoxelDebugSceneBig(&vox)
	case 2:
		scenes.VoxelDebugEmptyScene(&vox)
	case 3:
		scenes.VoxelDebugSceneHugeBunny(&vox)
		renderDist = 560.0
	case 4:
		scenes.VoxelDebugSceneTrees(&vox)
		renderDist = 560.0
	default:
		scenes.VoxelDebugSceneSmall(&vox)
	}
	rm, window := ren.RenderManagerInit()
	cam := ren.CameraInit()
	cam.Movespeed = 20
	cam.Lookspeed = 0.005
	cam.Fov = 90
	cam.Aspect = float32(rm.Pixels.Width) / float32(rm.Pixels.Height)
	cam.Pos = te.Vec3(16, 4, 16)
	cam.RenderDistance = renderDist

	engine := engine.Engine{}
	engine.Renderer = rm
	engine.Window = window
	engine.Camera = cam
	engine.Voxels = vox
	engine.Framedata = ren.FrameDataInit()
	engine.SetScrollCallback()

	for {
		engine.UpdateInputs()
		engine.UpdateRender()
		engine.CheckExit()
	}
}
