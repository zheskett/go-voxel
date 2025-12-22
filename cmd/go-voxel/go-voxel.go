package main

import (
	"fmt"
	"runtime"

	"github.com/zheskett/go-voxel/internal/common"
	"github.com/zheskett/go-voxel/internal/engine"
	ren "github.com/zheskett/go-voxel/internal/render"
	"github.com/zheskett/go-voxel/internal/scenes"
	"github.com/zheskett/go-voxel/internal/tensor"
	"github.com/zheskett/go-voxel/internal/voxel"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// This is meantioned in the usage example on github.
	runtime.LockOSThread()
}

func main() {
	renderDist := float32(256.0)
	var world voxel.VoxelWorld
	var scene int
	fmt.Println("1 for the big scene\n" +
		"2 for room\n" +
		"3 for big bunny\n" +
		"4 for sponza\n" +
		"5 for nuke\n" +
		"Or anything else for small scene")
	fmt.Scanln(&scene)
	switch scene {
	case 1:
		renderDist = 512.0
		scenes.VoxelDebugSceneBig(&world)
	case 2:
		renderDist = 512.0
		scenes.VoxelDebugEmptyScene(&world)
	case 3:
		renderDist = 1024.0
		scenes.VoxelDebugSceneHugeBunny(&world)
	case 4:
		renderDist = 1024.0
		scenes.VoxelDebugSceneTrees(&world)
	case 5:
		renderDist = 4096.0
		scenes.VoxelDebugSceneNuke(&world)
	default:
		scenes.VoxelDebugSceneSmall(&world)
	}
	scenes.LayoutCoordinateSystem(&world)

	rm, window := ren.RenderManagerInit()
	cam := ren.CameraInit()
	cam.Pos = tensor.Vec3Splat(16)
	cam.Movespeed = 20
	cam.Lookspeed = 0.005
	cam.Fov = 90
	cam.Aspect = float32(rm.Pixels.Width) / float32(rm.Pixels.Height)
	cam.RenderDistance = renderDist

	engine := engine.Engine{}
	engine.Renderer = rm
	engine.Window = window
	engine.Camera = cam
	engine.Voxels = world
	engine.Framedata = common.FrameDataInit()
	engine.SetCallbacks()

	for {
		engine.UpdateInputs()
		engine.UpdateRender()
		engine.CheckExit()
	}
}
