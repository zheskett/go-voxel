package main

import (
	"runtime"

	"github.com/zheskett/go-voxel/internal/render"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// This is meantioned in the usage example on github.
	runtime.LockOSThread()
}

func main() {
	rm := render.Init()

	for {
		rm.Render()
		rm.CheckExit()
	}
}
