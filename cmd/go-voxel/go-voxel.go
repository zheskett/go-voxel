package main

import (
	"fmt"
	"runtime"

	"github.com/zheskett/go-voxel/cmd/scenes"
	"github.com/zheskett/go-voxel/internal/engine"
	ren "github.com/zheskett/go-voxel/internal/render"
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
	var scene int
	fmt.Printf("enter 1 for the big scene and anything else for the small room with lots of lights\n")
	fmt.Scanln(&scene)
	switch scene {
	case 1:
		scenes.VoxelDebugSceneBig(&vox)
	default:
		scenes.VoxelDebugSceneSmall(&vox)
	}
	rm, window := ren.RenderManagerInit()
	cam := ren.CameraInit()
	cam.Movespeed = 15
	cam.Lookspeed = 0.005
	cam.Fov = 90
	cam.Aspect = float32(rm.Pixels.Width) / float32(rm.Pixels.Height)
	cam.Pos = te.Vec3(16, 4, 16)
	cam.RenderDistance = 256.0

	engine := engine.Engine{}
	engine.Renderer = rm
	engine.Window = window
	engine.Camera = cam
	engine.Voxels = vox
	engine.Framedata = ren.FrameDataInit()

	for {
		engine.UpdateInputs()
		engine.UpdateRender()
		engine.CheckExit()
	}
}
