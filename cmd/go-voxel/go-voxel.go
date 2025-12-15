package main

import (
	"fmt"
	"runtime"

	"github.com/zheskett/go-voxel/internal/engine"
	ren "github.com/zheskett/go-voxel/internal/render"
	"github.com/zheskett/go-voxel/internal/scenes"
	te "github.com/zheskett/go-voxel/internal/tensor"
	"github.com/zheskett/go-voxel/internal/voxel"
	vxl "github.com/zheskett/go-voxel/internal/voxel"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// This is meantioned in the usage example on github.
	runtime.LockOSThread()
}

func main() {
	renderDist := float32(256.0)
	size := 256
	tree := voxel.OctreeInit(size)
	world := voxel.VoxelWorld{Voxels: tree, Sun: voxel.DirLight{}, Lights: make([]vxl.LightPoint, 8)}
	world.X = size
	world.Y = size
	world.Z = size
	var scene int
	fmt.Printf("Enter 1 for the big scene, 2 for room, 3 for big bunny, 4 for .vox objects, anything else for small scene\n")
	fmt.Scanln(&scene)
	switch scene {
	case 1:
		scenes.VoxelDebugSceneBig(&world)
	case 2:
		scenes.VoxelDebugEmptyScene(&world)
	case 3:
		renderDist = 560.0
		world.Voxels = voxel.OctreeInit(512)
		scenes.VoxelDebugSceneHugeBunny(&world)
	case 4:
		renderDist = 560.0
		world.Voxels = voxel.OctreeInit(1024)
		scenes.VoxelDebugSceneTrees(&world)
	default:
		scenes.VoxelDebugSceneSmall(&world)
	}

	rm, window := ren.RenderManagerInit()
	cam := ren.CameraInit()
	cam.Movespeed = 20
	cam.Lookspeed = 0.005
	cam.Fov = 90
	cam.Aspect = float32(rm.Pixels.Width) / float32(rm.Pixels.Height)
	cam.RenderDistance = renderDist
	cam.Pos = te.Vec3(10, 10, 10)

	engine := engine.Engine{}
	engine.Renderer = rm
	engine.Window = window
	engine.Camera = cam
	engine.Voxels = world
	engine.Framedata = ren.FrameDataInit()
	engine.SetCallbacks()

	LayoutCoordinateSystem(engine.Voxels)

	for {
		engine.UpdateInputs()
		engine.UpdateRender()
		engine.CheckExit()
	}
}

func LayoutCoordinateSystem(vox voxel.VoxelWorld) {
	for i := range 16 {
		vox.SetVoxel(i, 0, 0, 255, 0, 0)
		vox.SetVoxel(0, i, 0, 0, 255, 0)
		vox.SetVoxel(0, 0, i, 0, 0, 255)
	}
}
