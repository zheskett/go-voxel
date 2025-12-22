package main

import (
	"fmt"
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

	cam := render.CameraInit()
	cam.Lookspeed = 0.005
	cam.Fov = 90
	cam.Aspect = float64(400) / float64(300)
	cam.RenderDistance = 4096
	cam.Pos = tensor.Vec3Splat(256)
	engine := engine.Engine{}
	engine.Camera = cam
	engine.Voxels = voxel.VoxelWorldInit(4096)
	engine.Renderer = &render.RenderManager{
		Pixels: render.PixelsInit(400, 300),
		Color:  render.ColorBufferInit(400, 300),
	}
	scenes.VoxelDebugSceneTrees(&engine.Voxels)

	pprof.StartCPUProfile(file)
	defer pprof.StopCPUProfile()

	for range 1000 {
		engine.Camera.RenderVoxels(&engine.Voxels, &engine.Renderer.Color, engine.Framedata.Tick)
		engine.Renderer.Color.CorrectGamma()
		engine.Renderer.LoadPixels()
		engine.Renderer.Pixels.Dither()
		engine.Camera.UpdateRotation(1, 1) // Nudge the camera around so the dynamic lighting has to do some work
	}

	fmt.Println("Tree testing done.")
}
