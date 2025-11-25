package main

import (
	_ "embed"
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	WindowTitle   = "Go Voxel"
	TextureWidth  = 320
	TextureHeight = 240
	WindowWidth   = TextureWidth * 4
	WindowHeight  = TextureHeight * 4
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

	window, err := glfw.CreateWindow(WindowWidth, WindowHeight, WindowTitle, nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	gl.Init()

	var fbo, renderTexture uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	gl.GenTextures(1, &renderTexture)
	gl.BindTexture(gl.TEXTURE_2D, renderTexture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, TextureWidth, TextureHeight, 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, renderTexture, 0)

	pixels := make([]byte, TextureWidth*TextureHeight*4)
	for i := range pixels {
		pixels[i] = 255
		if i&0x3 == 0x1 {
			pixels[i] = 0
		}
	}

	for !window.ShouldClose() {
		// Do stuff here
		gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
		gl.Viewport(0, 0, TextureWidth, TextureHeight)

		gl.BindTexture(gl.TEXTURE_2D, renderTexture)
		gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, int32(TextureWidth), int32(TextureHeight), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pixels))

		gl.BindFramebuffer(gl.READ_FRAMEBUFFER, fbo)
		gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
		gl.BlitFramebuffer(0, 0, TextureWidth, TextureHeight, 0, 0, WindowWidth, WindowHeight, gl.COLOR_BUFFER_BIT, gl.NEAREST)

		window.SwapBuffers()
		glfw.PollEvents()
	}
}
