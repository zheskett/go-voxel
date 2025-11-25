package main

import (
	"runtime"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/zheskett/go-voxel/internal/render"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// This is meantioned in the usage example on github.
	runtime.LockOSThread()
}

func main() {
	// Initialize glfw
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	// Window creation
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(render.WindowWidth, render.WindowHeight, render.WindowTitle, nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	rm := render.Init()

	for !window.ShouldClose() {
		// Do stuff here
		rm.Render()

		window.SwapBuffers()
		glfw.PollEvents()
	}
}
