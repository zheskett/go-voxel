package main

import (
	"os"
	"runtime/pprof"

	"github.com/zheskett/go-voxel/internal/engine"
	"github.com/zheskett/go-voxel/internal/render"
	"github.com/zheskett/go-voxel/internal/scenes"
	"github.com/zheskett/go-voxel/internal/tensor"
	"github.com/zheskett/go-voxel/internal/voxel"
)

func main() {
	file, err := os.Create("cpu.pprof")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	tree := voxel.OctreeInit(4096)
	cam := render.CameraInit()
	cam.Lookspeed = 0.005
	cam.Fov = 90
	cam.Aspect = float32(400) / float32(300)
	cam.RenderDistance = 4096
	cam.Pos = tensor.Vec3(100, 100, 100)
	engine := engine.Engine{}
	engine.Camera = cam
	engine.Voxels = voxel.VoxelWorld{Voxels: tree, Sun: voxel.DirLight{}, Lights: make([]voxel.PointLight, 0)}
	engine.Renderer = &render.RenderManager{Pixels: render.PixelsInit(400, 300)}
	scenes.VoxelDebugSceneTrees(&engine.Voxels)

	pprof.StartCPUProfile(file)
	defer pprof.StopCPUProfile()

	for range 1000 {
		engine.Camera.RenderVoxels(&engine.Voxels, &engine.Renderer.Pixels, engine.Framedata.Tick)
		engine.Camera.UpdateRotation(1, 1)
	}
}
