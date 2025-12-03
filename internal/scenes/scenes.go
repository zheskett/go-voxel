// This def needs to change I just didn't know where to put these
package scenes

import (
	"math/rand"

	te "github.com/zheskett/go-voxel/internal/tensor"
	vxl "github.com/zheskett/go-voxel/internal/voxel"
)

func VoxelDebugEmptyScene(vox *vxl.Voxels) {
	brightness := 100.0
	vox.Lights = append(vox.Lights, vxl.Light{
		Position: te.Vec3(float32(vox.X/2), 90, float32(vox.Z/2)),
		Color:    te.Vec3(1.0, 0.95, 0.95).Mul(float32(brightness)),
	})
	vox.Lights = append(vox.Lights, vxl.Light{
		Position: te.Vec3(10, 10, 10),
		Color:    te.Vec3(1.0, 0.3, 0.3).Mul(float32(brightness / 4.0)),
	})
	vox.Lights = append(vox.Lights, vxl.Light{
		Position: te.Vec3(10, 10, float32(vox.Y-10)),
		Color:    te.Vec3(0.3, 1.0, 0.3).Mul(float32(brightness / 4.0)),
	})
	vox.Lights = append(vox.Lights, vxl.Light{
		Position: te.Vec3(float32(vox.X-10), 10, 10),
		Color:    te.Vec3(0.3, 0.3, 1.0).Mul(float32(brightness / 4.0)),
	})
	for i := 0; i < vox.X; i++ {
		for j := 0; j < vox.Z; j++ {
			// Floor and ceiling
			vox.SetVoxel(i, 0, j, 200, 200, 200)
			vox.SetVoxel(i, 100, j, 200, 200, 200)
			// Walls
			vox.SetVoxel(0, i, j, 200, 200, 200)
			vox.SetVoxel(vox.X-1, i, j, 200, 200, 200)
			vox.SetVoxel(j, i, 0, 200, 200, 200)
			vox.SetVoxel(j, i, vox.Z-1, 200, 200, 200)
		}
	}
	obj, err := vxl.VoxelizePath("assets/bunny.obj", vxl.T6, 200, [3]byte{220, 220, 220})
	if err != nil {
		panic(err)
	}
	obj.Flip(false, true, false)
	vox.AddVoxelObj(obj, vox.X/2-35, 0, vox.Z/2-35)
	cow, err := vxl.VoxelizePath("assets/cow.obj", vxl.T6, 165, [3]byte{160, 82, 45})
	if err != nil {
		panic(err)
	}
	cow.Flip(false, true, false)
	vox.AddVoxelObj(cow, vox.X/8-35, -40, vox.Z/8-35)
}

// A small room with 3 colored lights and boxes everywhere
func VoxelDebugSceneSmall(vox *vxl.Voxels) {
	brightness := 22
	light := vxl.Light{
		Position: te.Vec3(50, 15, 30),
		Color:    te.Vec3(0.5, 0.5, 1.0).Mul(float32(brightness)),
	}
	vox.Lights = append(vox.Lights, light)
	light = vxl.Light{
		Position: te.Vec3(20, 7, 22),
		Color:    te.Vec3(1.0, 0.5, 0.5).Mul(float32(brightness)),
	}
	vox.Lights = append(vox.Lights, light)
	light = vxl.Light{
		Position: te.Vec3(88, 30, 88),
		Color:    te.Vec3(0.5, 1.0, 0.5).Mul(float32(brightness)),
	}
	vox.Lights = append(vox.Lights, light)

	// Make a floor and ceiling
	for i := 0; i < vox.X; i++ {
		for j := 0; j < vox.Z; j++ {
			vox.SetVoxel(i, 0, j, 220, 180, 180)
			vox.SetVoxel(i, 40, j, 180, 180, 180)
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
			vox.SetVoxel(36, j, i, 200, 180, 180)
		}
	}

	for i := 60; i < 70; i++ {
		for j := 28; j < 40; j++ {
			for k := 60; k < 70; k++ {
				vox.SetVoxel(i, j, k, 200, 200, 200)
			}
		}
	}

	for i := range 100 {
		for j := range 100 {
			for k := range 100 {
				if i%10 == 0 && j%10 == 0 && k%10 == 0 {
					vox.SetVoxel(i, j, k, 200, 200, 200)
				}
			}
		}
	}
}

// A massive open scene with a bunch of random stuff
func VoxelDebugSceneBig(vox *vxl.Voxels) {
	brightness := 120
	light := vxl.Light{
		Position: te.Vec3(64, 32, 96),
		Color:    te.Vec3(1.0, 1.0, 1.0).Mul(float32(brightness)),
	}
	vox.Lights = append(vox.Lights, light)
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
