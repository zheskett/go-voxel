package main

import (
	_ "embed"
	"runtime"
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	WindowTitle   = "Go Voxel"
	TextureWidth  = 320
	TextureHeight = 240
	WaindowWidth  = TextureWidth * 4
	WindowHeight  = TextureHeight * 4
)

//go:embed shaders/fragment.glsl
var fragmentShaderSource string

//go:embed shaders/vertex.glsl
var vertexShaderSource string

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

	window, err := glfw.CreateWindow(WaindowWidth, WindowHeight, WindowTitle, nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	gl.Init()

	program := createProgram()
	gl.UseProgram(program)

	// Fullscreen quad vertices (position + texcoord)
	vertices := []float32{
		// positions   // texCoords
		-1.0, -1.0, 0.0, 1.0,
		1.0, -1.0, 1.0, 1.0,
		1.0, 1.0, 1.0, 0.0,

		-1.0, -1.0, 0.0, 1.0,
		1.0, 1.0, 1.0, 0.0,
		-1.0, 1.0, 0.0, 0.0,
	}

	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	//nolint:gosec
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, glOffset(0))
	gl.EnableVertexAttribArray(0)
	//nolint:gosec
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, glOffset(2*4))
	gl.EnableVertexAttribArray(1)

	pixels := make([]byte, TextureWidth*TextureHeight*4)
	for i := range pixels {
		pixels[i] = 255
		if i%2 == 0 {
			pixels[i] = 0
		}
	}

	var renderTexture uint32
	gl.GenTextures(1, &renderTexture)
	gl.BindTexture(gl.TEXTURE_2D, renderTexture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	for !window.ShouldClose() {
		// Do stuff here
		gl.ClearColor(0.0, 0.0, 0.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		gl.BindTexture(gl.TEXTURE_2D, renderTexture)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(TextureWidth), int32(TextureHeight), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pixels))

		gl.UseProgram(program)
		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.TRIANGLES, 0, 6)

		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func createProgram() uint32 {
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	csources, free := gl.Strs(vertexShaderSource + "\x00")
	gl.ShaderSource(vertexShader, 1, csources, nil)
	free()
	gl.CompileShader(vertexShader)

	// Check vertex shader compilation
	var status int32
	gl.GetShaderiv(vertexShader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(vertexShader, gl.INFO_LOG_LENGTH, &logLength)
		log := make([]byte, logLength)
		gl.GetShaderInfoLog(vertexShader, logLength, nil, &log[0])
		panic("Vertex shader error: " + string(log))
	}

	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	csources, free = gl.Strs(fragmentShaderSource + "\x00")
	gl.ShaderSource(fragmentShader, 1, csources, nil)
	free()
	gl.CompileShader(fragmentShader)

	// Check fragment shader compilation
	gl.GetShaderiv(fragmentShader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(fragmentShader, gl.INFO_LOG_LENGTH, &logLength)
		log := make([]byte, logLength)
		gl.GetShaderInfoLog(fragmentShader, logLength, nil, &log[0])
		panic("Fragment shader error: " + string(log))
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	// Check program linking
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		log := make([]byte, logLength)
		gl.GetProgramInfoLog(program, logLength, nil, &log[0])
		panic("Program link error: " + string(log))
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program
}

func glOffset(offset int) unsafe.Pointer {
	//nolint:gosec
	return gl.PtrOffset(offset)
}
