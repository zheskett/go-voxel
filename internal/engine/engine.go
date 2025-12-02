package engine

import (
	"os"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/zheskett/go-voxel/internal/render"
	"github.com/zheskett/go-voxel/internal/voxel"
)

type Engine struct {
	Renderer  *render.RenderManager
	Window    *glfw.Window
	Camera    render.Camera
	Voxels    voxel.Voxels
	Framedata render.FrameData
}

func (eng *Engine) UpdateInputs() {
	eng.Framedata.Update()
	eng.Framedata.ReportFps()
	// render.UpdateCamInputGLFW(&eng.Camera, eng.Window, &eng.Framedata)
	render.UpdateCamInputGLFWFPS(&eng.Camera, eng.Window, &eng.Framedata)
}

func (eng *Engine) UpdateRender() {
	eng.Renderer.Pixels.FillPixels(render.BackgroundRed, render.BackgroundGreen, render.BackgroundBlue)
	eng.Camera.RenderVoxels(&eng.Voxels, &eng.Renderer.Pixels)
	eng.Renderer.Render(eng.Window)
}

// Check for exit condition
func (eng *Engine) CheckExit() {
	if eng.Window.GetKey(glfw.KeyEscape) == glfw.Press || eng.Window.ShouldClose() {
		glfw.Terminate()
		os.Exit(0)
	}
}
