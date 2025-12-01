package main

import (
	"fmt"
	"math/rand"
	"runtime"

	ren "github.com/zheskett/go-voxel/internal/render"
	te "github.com/zheskett/go-voxel/internal/tensor"
	vxl "github.com/zheskett/go-voxel/internal/voxel"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// This is meantioned in the usage example on github.
	runtime.LockOSThread()
}

func main() {
	rm := ren.RenderManagerInit()
	cam := ren.CameraInit()
	cam.Movespeed = 15 // 15 voxels/s second walking
	cam.Lookspeed = 2  // 2 rad/s rotation
	cam.Fov = 90
	cam.Aspect = float32(rm.Pixels.Width) / float32(rm.Pixels.Height)

	cam.Pos = te.Vec3(16, 4, 16)
	cam.RenderDistance = 256.0
	vox := vxl.VoxelsInit(256, 256, 256)
	vox.Light = te.Vec3(50, 5, 30)
	vox.LightIntensity = 30.0
	fdata := ren.FrameDataInit()
	voxelDebugSceneSmall(&vox)

	// cam.Pos = te.Vec3(16, 4, 16)
	// cam.RenderDistance = 256.0
	// vox := vxl.VoxelsInit(256, 256, 256)
	// vox.Light = te.Vec3(64, 32, 96)
	// vox.LightIntensity = 70.0
	// fdata := ren.FrameDataInit()
	// voxelDebugSceneBig(&vox)

	fmt.Printf("total voxels: %d\n", vox.X*vox.Y*vox.Z)

	for {
		rm.Pixels.FillPixels(15, 25, 40)
		cam.RenderVoxels(&vox, &rm.Pixels)
		ren.UpdateCamInputGLFW(&cam, rm.Window, &fdata)
		fdata.Update()
		fdata.ReportFps()
		rm.Render()
		rm.CheckExit()
	}
}

func voxelDebugSceneSmall(vox *vxl.Voxels) {
	// Make a floor and ceiling
	for i := 0; i < vox.X; i++ {
		for j := 0; j < vox.Z; j++ {
			vox.SetVoxel(i, 0, j, 220, 180, 180)
			vox.SetVoxel(i, 30, j, 180, 180, 180)
		}
	}
	// Make walls
	for i := range 100 {
		for j := range 100 {
			vox.SetVoxel(100, i, j, 200, 180, 180)
			vox.SetVoxel(0, i, j, 180, 220, 180)
			vox.SetVoxel(j, i, 0, 200, 180, 180)
			vox.SetVoxel(j, i, 100, 180, 180, 220)
		}
	}
	// Make a small wall for shadows
	for i := range 55 {
		for j := range 100 {
			vox.SetVoxel(35, j, i, 200, 180, 180)
		}
	}

	for i := 64; i < 70; i++ {
		for j := 16; j < 22; j++ {
			for k := 64; k < 70; k++ {
				vox.SetVoxel(i, j, k, 200, 200, 200)
			}
		}
	}
}

func voxelDebugSceneBig(vox *vxl.Voxels) {
	// Make a teal and purple checkerboard "ground"
	for i := 0; i < vox.Z; i++ {
		for j := 0; j < vox.X; j++ {
			for k := 0; k < 1+(i+j)/16; k++ {
				if (i+j)%2 == 0 {
					vox.SetVoxel(i, k, j, 70, 200, 200)
				} else {
					vox.SetVoxel(i, k, j, 200, 30, 200)
				}
			}
		}
	}
	// Same thing as above but on the roof with a more extreme slope
	for i := 0; i < vox.Z; i++ {
		for j := 0; j < vox.X; j++ {
			for k := vox.Y - 1; k > vox.Y-(i+j)/4; k-- {
				if (i/4+j/4)%2 == 0 {
					vox.SetVoxel(i, k, j, 200, 3, 180)
				} else {
					vox.SetVoxel(i, k, j, 150, 200, 20)
				}
			}
		}
	}
	// Floating red cube
	for i := 10; i < 15; i++ {
		for j := 10; j < 15; j++ {
			for k := 10; k < 15; k++ {
				vox.SetVoxel(i, j, k, 180, 50, 50)
			}
		}
	}
	// Floating orange cube
	for i := 32; i < 40; i++ {
		for j := 32; j < 40; j++ {
			for k := 32; k < 40; k++ {
				vox.SetVoxel(i, j, k, 200, 100, 100)
			}
		}
	}
	// A larger checkerboard wall one one side
	for i := 0; i < vox.X; i++ {
		for j := 0; j < vox.Y; j++ {
			if (i/15+j/15)%2 == 0 {
				vox.SetVoxel(0, i, j, 30, 30, 30)
			} else {
				vox.SetVoxel(0, i, j, 200, 200, 200)
			}
		}
	}
	// Some green pillar
	for i := 1; i < 16; i++ {
		vox.SetVoxel(5, i, 5, 30, 255, 30)
		vox.SetVoxel(vox.X-1, i, 0, 30, 255, 30)
		vox.SetVoxel(vox.X-1, i, vox.Z-1, 30, 255, 30)
		vox.SetVoxel(50, i, 50, 30, 255, 30)
		vox.SetVoxel(25, i, 9, 30, 255, 30)
	}
	// A big ominous ball
	// Also a bunch of random colored voxels
	center, radius := te.Vec3(64, 64, 64), 24
	for i := 0; i < vox.Z; i++ {
		for j := 0; j < vox.Y; j++ {
			for k := 0; k < vox.X; k++ {
				point := te.Vec3(float32(k), float32(j), float32(i))
				if center.Sub(point).LenSqr() < float32(radius*radius) {
					vox.SetVoxel(k, j, i, 20, 20, 20)
				}
				if rand.Intn(2500) == 0 {
					vox.SetVoxel(k, j, i, byte(rand.Intn(255)), byte(rand.Intn(255)), byte(rand.Intn(255)))
				}
			}
		}
	}
}
