package engine

import (
	"os"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/zheskett/go-voxel/internal/render"
	"github.com/zheskett/go-voxel/internal/voxel"
)

const (
	moveSpeedInc = 1.0
)

type Engine struct {
	Renderer  *render.RenderManager
	Window    *glfw.Window
	Camera    render.Camera
	Voxtree   voxel.BrickTree
	Framedata render.FrameData
}

func (eng *Engine) UpdateInputs() {
	eng.Framedata.Update()
	eng.Framedata.ReportFps()
	render.UpdateCamInputGLFW(&eng.Camera, eng.Window, &eng.Framedata)
}

func (eng *Engine) UpdateRender() {
	eng.Renderer.Pixels.FillPixels(render.BackgroundRed, render.BackgroundGreen, render.BackgroundBlue)
	eng.Camera.RenderVoxels(&eng.Voxtree, &eng.Renderer.Pixels)
	eng.Renderer.Render(eng.Window)
}

// Check for exit condition
func (eng *Engine) CheckExit() {
	if eng.Window.GetKey(glfw.KeyEscape) == glfw.Press || eng.Window.ShouldClose() {
		glfw.Terminate()
		os.Exit(0)
	}
}

func (eng *Engine) SetScrollCallback() {
	eng.Window.SetScrollCallback(func(w *glfw.Window, _ float64, yoff float64) {
		eng.Camera.Movespeed = max(eng.Camera.Movespeed+float32(yoff)*moveSpeedInc, 0)
	})
}

func (eng *Engine) SetMouseCallback() {
	eng.Window.SetCursorPosCallback(func(w *glfw.Window, xpos float64, ypos float64) {
		dx, dy := eng.Framedata.GetMouseDelta(xpos, ypos)
		eng.Camera.UpdateRotationFPS(dy, dx)
	})
}
