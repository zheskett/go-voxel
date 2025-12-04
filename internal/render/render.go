// Package render provides a renderer for the voxels
package render

import (
	"fmt"
	"runtime"
	"time"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/zheskett/go-voxel/internal/tensor"
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

// FrameData allows camera movements to be made independent of FPS for a smoother movements
type FrameData struct {
	Previous time.Time
	Deltat   float32
	Tick     uint
	mouse    tensor.Vector2
}

func FrameDataInit() FrameData {
	return FrameData{Previous: time.Now()}
}

func (data *FrameData) Update() {
	data.Deltat = float32(time.Since(data.Previous).Seconds())
	data.Previous = time.Now()
	data.Tick += 1
}

func (data *FrameData) ReportFps() {
	fmt.Printf("FPS: %.2f\n", 1.0/data.Deltat)
}

func (data *FrameData) GetMouseDelta(window *glfw.Window) (float32, float32) {
	mx_f64, my_f64 := window.GetCursorPos()
	mx, my := float32(mx_f64), float32(my_f64)
	dx, dy := data.mouse.X-mx, data.mouse.Y-my
	data.mouse = tensor.Vec2(mx, my)

	return dx, dy
}

// Pixels contains the data for each pixel on the screen.
// Every pixel is 4 bytes, RGBA
type Pixels struct {
	data   []byte
	Height int
	Width  int
}

func PixelsInit(width, height int) Pixels {
	data := make([]byte, width*height*4)
	for i := 0; i < width*height*4; i++ {
		data[i] = 0
	}
	return Pixels{data, height, width}
}

func (px *Pixels) FillPixels(r, g, b byte) {
	for i := 0; i < px.Width*px.Height; i++ {
		px.data[4*i+0] = r
		px.data[4*i+1] = g
		px.data[4*i+2] = b
	}
}

func (px *Pixels) SetPixel(x, y int, r, g, b byte) {
	px.data[4*(px.Width*y+x)+0] = r
	px.data[4*(px.Width*y+x)+1] = g
	px.data[4*(px.Width*y+x)+2] = b
}

func (px *Pixels) GetPixel(x, y int) [3]byte {
	return [3]byte{
		px.data[4*(px.Width*y+x)+0],
		px.data[4*(px.Width*y+x)+1],
		px.data[4*(px.Width*y+x)+2],
	}
}

func (px *Pixels) Surrounds(x, y int) bool {
	return x >= 0 && x < px.Width && y >= 0 && y < px.Height
}

// RenderManager contains state for the rendering
type RenderManager struct {
	renderTexture uint32
	fbo           uint32
	Pixels        Pixels
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

	return &rm, window
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
	window.SwapBuffers()
	glfw.PollEvents()
}
