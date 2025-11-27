// Package render provides a renderer for the voxels
package render

import (
	"os"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	TextureWidth  = 320
	TextureHeight = 240
	WindowTitle   = "Go Voxel"
)

type Color [3]byte

// Pixles contains the data for each pixel on the screen.
// Every pixel if 4 bytes, RGBA
type Pixels struct {
	data   []byte
	Height int
	Width  int
}

func PixelsInit(width, height int) Pixels {
	data := make([]byte, width*height*4)
	for i := range data {
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
	return x > 0 && x < px.Width && y > 0 && y < px.Height
}

// RenderManager contains state for the rendering
type RenderManager struct {
	renderTexture uint32
	fbo           uint32
	Pixels        Pixels
	Window        *glfw.Window
}

// RenderManagerInit initializes the render manager
// and initializes the opengl context
func RenderManagerInit() *RenderManager {
	// Initialize gl
	err := gl.Init()
	if err != nil {
		panic(err)
	}
	// Initialize glfw
	err = glfw.Init()
	if err != nil {
		panic(err)
	}

	rm := RenderManager{}

	// Window creation
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLAnyProfile)
	window, err := glfw.CreateWindow(TextureWidth*4, TextureHeight*4, WindowTitle, nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	rm.Window = window

	gl.GenFramebuffers(1, &rm.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, rm.fbo)
	gl.GenTextures(1, &rm.renderTexture)
	gl.BindTexture(gl.TEXTURE_2D, rm.renderTexture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, TextureWidth, TextureHeight, 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, rm.renderTexture, 0)

	rm.Pixels = PixelsInit(TextureWidth, TextureHeight)

	return &rm
}

// Render renders the current state
// It should be called each frame
func (rm *RenderManager) Render() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, rm.fbo)
	gl.Viewport(0, 0, TextureWidth, TextureHeight)

	gl.BindTexture(gl.TEXTURE_2D, rm.renderTexture)
	gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, int32(TextureWidth), int32(TextureHeight), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rm.Pixels.data))

	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, rm.fbo)
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
	gl.BlitFramebuffer(0, 0, TextureWidth, TextureHeight, 0, 0, TextureWidth*4, TextureHeight*4, gl.COLOR_BUFFER_BIT, gl.NEAREST)

	rm.Window.SwapBuffers()
	glfw.PollEvents()
}

// Check for exit condition
func (rm *RenderManager) CheckExit() {
	if rm.Window.GetKey(glfw.KeyEscape) == glfw.Press || rm.Window.ShouldClose() {
		glfw.Terminate()
		os.Exit(0)
	}
}
