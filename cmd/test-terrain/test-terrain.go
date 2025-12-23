package main

import (
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/zheskett/go-voxel/internal/engine"
	"github.com/zheskett/go-voxel/internal/scenes"
)

func main() {
	file, err := os.Create("cpu.pprof")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	engine := engine.Engine{}
	pprof.StartCPUProfile(file)
	scenes.VoxelDebugSceneTerrain(&engine.Voxels)
	pprof.StopCPUProfile()
	fmt.Println("Testing done.")
}
