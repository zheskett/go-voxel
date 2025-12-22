package engine

import (
	"os"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/zheskett/go-voxel/internal/common"
	"github.com/zheskett/go-voxel/internal/render"
	"github.com/zheskett/go-voxel/internal/tensor"
	"github.com/zheskett/go-voxel/internal/voxel"
)

const (
	moveSpeedInc = 1.0
)

type Engine struct {
	Renderer  *render.RenderManager
	Window    *glfw.Window
	Camera    render.Camera
	Voxels    voxel.VoxelWorld
	Framedata common.FrameData
}

func (eng *Engine) UpdateInputs() {
	eng.Framedata.Update()
	eng.Framedata.ReportFps()
	eng.Camera.UpdateCamInput(&eng.Framedata)
	eng.Voxels.UpdateInputs(eng.Window, eng.Camera.Pos, eng.Camera.Fvec)
}

// TODO: This one got a lot more complicated and it is kind of slow. This should
// honestly be done in a compute shader as a lot of this is purely aesthetic
// post-processing
func (eng *Engine) UpdateRender() {
	// Clear background
	eng.Renderer.Color.FillColor(
		tensor.Vec3(
			render.BackgroundRed,
			render.BackgroundGreen,
			render.BackgroundBlue,
		),
	)
	// Render into Color-buffer
	eng.Camera.RenderVoxels(&eng.Voxels, &eng.Renderer.Color, eng.Framedata.Tick)
	// Apply gamma correction
	eng.Renderer.Color.CorrectGamma()
	// Convert to RGBA byte
	eng.Renderer.LoadPixels()
	// Dither applied to byte-buffer
	eng.Renderer.Pixels.Dither()
	// Blit to screen frame-buffer
	eng.Renderer.Render(eng.Window)
}

func (eng *Engine) SetCallbacks() {
	eng.SetScrollCallback()
	eng.SetMouseCallback()
	eng.SetKeyCallback()
}

func (eng *Engine) SetScrollCallback() {
	eng.Window.SetScrollCallback(func(_ *glfw.Window, _ float64, yoff float64) {
		eng.Camera.Movespeed = max(eng.Camera.Movespeed+float64(yoff)*moveSpeedInc, 0)
	})
}

func (eng *Engine) SetMouseCallback() {
	eng.Window.SetCursorPosCallback(func(_ *glfw.Window, xpos float64, ypos float64) {
		dx, dy := eng.Framedata.GetMouseDelta(xpos, ypos)
		eng.Camera.UpdateRotation(dy, dx)
	})
}

func (eng *Engine) SetKeyCallback() {
	eng.Window.SetKeyCallback(func(win *glfw.Window, key glfw.Key, _ int, action glfw.Action, _ glfw.ModifierKey) {
		switch action {
		case glfw.Press:
			eng.Framedata.Keys[key] = true
		case glfw.Release:
			eng.Framedata.Keys[key] = false
		}
		// Allows 'T' to be used to lock/unlock the mouse
		if key == glfw.KeyT {
			switch win.GetInputMode(glfw.CursorMode) {
			case glfw.CursorNormal:
				win.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
			default:
				win.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
			}
		}
	})
}

// Check for exit condition
func (eng *Engine) CheckExit() {
	if eng.Window.GetKey(glfw.KeyEscape) == glfw.Press || eng.Window.ShouldClose() {
		glfw.Terminate()
		os.Exit(0)
	}
}
