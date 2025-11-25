package main

import (
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/zheskett/go-voxel/internal/render"
)

const (
	WindowTitle  = "Go Voxel"
	WindowWidth  = render.TextureWidth * 4
	WindowHeight = render.TextureHeight * 4
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
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, render.TextureWidth, render.TextureHeight, 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, renderTexture, 0)

	for !window.ShouldClose() {
		// Do stuff here

		pixels := render.Render()

		gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
		gl.Viewport(0, 0, render.TextureWidth, render.TextureHeight)

		gl.BindTexture(gl.TEXTURE_2D, renderTexture)
		gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, int32(render.TextureWidth), int32(render.TextureHeight), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pixels))

		gl.BindFramebuffer(gl.READ_FRAMEBUFFER, fbo)
		gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
		gl.BlitFramebuffer(0, 0, render.TextureWidth, render.TextureHeight, 0, 0, WindowWidth, WindowHeight, gl.COLOR_BUFFER_BIT, gl.NEAREST)

		window.SwapBuffers()
		glfw.PollEvents()
	}
}
