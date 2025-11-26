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

// Pixles contains the data for each pixel on the screen.
// Every pixel if 4 bytes, RGBA
type Pixels []byte

// RenderManager contains state for the rendering
type RenderManager struct {
	renderTexture uint32
	fbo           uint32
	pixels        Pixels
	window        *glfw.Window
}

// Init initializes the render manager
// and initializes the opengl context
func Init() *RenderManager {
	err := gl.Init()
	if err != nil {
		panic(err)
	}
	// Initialize glfw
	err = glfw.Init()
	if err != nil {
		panic(err)
	}

	var rm RenderManager = RenderManager{}

	// Window creation
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLAnyProfile)
	window, err := glfw.CreateWindow(TextureWidth*4, TextureHeight*4, WindowTitle, nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	rm.window = window

	gl.GenFramebuffers(1, &rm.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, rm.fbo)

	gl.GenTextures(1, &rm.renderTexture)
	gl.BindTexture(gl.TEXTURE_2D, rm.renderTexture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, TextureWidth, TextureHeight, 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, rm.renderTexture, 0)

	rm.pixels = make(Pixels, TextureWidth*TextureHeight*4)

	return &rm
}

// Render renders the current state
// It should be called each frame
func (rm *RenderManager) Render() {
	for i := range rm.pixels {
		rm.pixels[i] = 255
		if i&0x3 == 0x1 {
			rm.pixels[i] = 0
		}
	}

	rm.pixels[4*100] = 0
	rm.pixels[4*101] = 0
	rm.pixels[4*102] = 0

	gl.BindFramebuffer(gl.FRAMEBUFFER, rm.fbo)
	gl.Viewport(0, 0, TextureWidth, TextureHeight)

	gl.BindTexture(gl.TEXTURE_2D, rm.renderTexture)
	gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, int32(TextureWidth), int32(TextureHeight), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rm.pixels))

	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, rm.fbo)
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
	gl.BlitFramebuffer(0, 0, TextureWidth, TextureHeight, 0, 0, TextureWidth*4, TextureHeight*4, gl.COLOR_BUFFER_BIT, gl.NEAREST)

	rm.window.SwapBuffers()
	glfw.PollEvents()
}

func (rm *RenderManager) CheckExit() {
	if rm.window.GetKey(glfw.KeyEscape) == glfw.Press || rm.window.ShouldClose() {
		glfw.Terminate()
		os.Exit(0)
	}
}
