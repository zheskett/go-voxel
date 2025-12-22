// Package render provides a renderer for the voxels
package render

import (
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// Window info
const (
	TextureWidth  = 400
	TextureHeight = 300
	WindowUpscale = 4
	WindowTitle   = "Go Voxel"
)

// Window clear color
const (
	BackgroundRed   = 15
	BackgroundGreen = 25
	BackgroundBlue  = 40
)

// Number of goroutines that are dispatched to render the frame
const (
	RenderThreads = 16
)

// Set the minimum luminance even for complete shadow
const (
	MinLuminosity = 0.01
)

// RenderManager contains state for the rendering
type RenderManager struct {
	// Handles for glfw draw call
	renderTexture uint32
	fbo           uint32

	// Buffers that are used for the software renderer to write into
	Pixels Pixels      // RGBA byte buffer (displayed to screen)
	Color  ColorBuffer // Vec3 color buffer (for internal rendering)
	Depth  DepthBuffer // Depth buffer
}

// RenderManagerInit initializes the render manager
// and initializes the opengl context
func RenderManagerInit() (*RenderManager, *glfw.Window) {
	rm := RenderManager{}

	// Initialize glfw
	err := glfw.Init()
	if err != nil {
		panic(err)
	}

	// Window creation
	switch runtime.GOOS {
	case "darwin": // MacOS
		glfw.WindowHint(glfw.ContextVersionMajor, 3)
		glfw.WindowHint(glfw.ContextVersionMinor, 3)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	case "windows": // Windows
		glfw.WindowHint(glfw.ContextVersionMajor, 3)
		glfw.WindowHint(glfw.ContextVersionMinor, 3)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCompatProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.False)
	default:
		glfw.WindowHint(glfw.ContextVersionMajor, 3)
		glfw.WindowHint(glfw.ContextVersionMinor, 1)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLAnyProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.False)
	}
	window, err := glfw.CreateWindow(TextureWidth*WindowUpscale, TextureHeight*WindowUpscale, WindowTitle, nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)

	// Initialize gl
	err = gl.Init()
	if err != nil {
		panic(err)
	}

	gl.GenFramebuffers(1, &rm.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, rm.fbo)
	gl.GenTextures(1, &rm.renderTexture)
	gl.BindTexture(gl.TEXTURE_2D, rm.renderTexture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, TextureWidth, TextureHeight, 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, rm.renderTexture, 0)

	rm.Pixels = PixelsInit(TextureWidth, TextureHeight)
	rm.Color = ColorBufferInit(TextureWidth, TextureHeight)

	return &rm, window
}

func (rm *RenderManager) LoadPixels() {
	for row := range rm.Pixels.Height {
		for col := range rm.Pixels.Width {
			color := rm.Color.GetColor(col, row)
			color = color.ComponentMin(255.99999)
			rm.Pixels.SetPixel(col, row, byte(color.X), byte(color.Y), byte(color.Z))
		}
	}
}

// Render renders the current state
// It should be called each frame
func (rm *RenderManager) Render(window *glfw.Window) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, rm.fbo)
	gl.Viewport(0, 0, TextureWidth, TextureHeight)

	gl.BindTexture(gl.TEXTURE_2D, rm.renderTexture)
	gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, int32(TextureWidth), int32(TextureHeight), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rm.Pixels.data))

	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, rm.fbo)
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)

	fbWidth, fbHeight := window.GetFramebufferSize()
	gl.BlitFramebuffer(0, 0, TextureWidth, TextureHeight, 0, 0, int32(fbWidth), int32(fbHeight), gl.COLOR_BUFFER_BIT, gl.NEAREST)
	window.SetAspectRatio(fbWidth, fbHeight)
	window.SwapBuffers()
	glfw.PollEvents()
}
