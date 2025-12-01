package voxel

import (
	"fmt"

	"github.com/chewxy/math32"
	"github.com/zheskett/go-voxel/internal/parser"
	te "github.com/zheskett/go-voxel/internal/tensor"
)

type VoxelObj struct {
	// Position of the object in the world
	XPos, YPos, ZPos int
	Presence         BitArray
	Color            [3]byte
}

type ConnectivityDistance int

const (
	T26 ConnectivityDistance = 26
	T6  ConnectivityDistance = 6
)

// Same as Voxelize(ParseObj(path), ...) basically
func VoxelizePath(path string, t ConnectivityDistance, resolution int, color [3]byte, x, y, z int) (VoxelObj, error) {
	obj, err := parser.ParseObj(path)
	if err != nil {
		return VoxelObj{}, err
	}

	return Voxelize(obj, t, resolution, color, x, y, z)
}

// Turns an obj into voxels
//
// Algorithm from https://web.eecs.utk.edu/~huangj/papers/polygon.pdf
func Voxelize(obj parser.Obj, t ConnectivityDistance, resolution int, color [3]byte, x, y, z int) (VoxelObj, error) {
	if resolution < 1 {
		return VoxelObj{}, fmt.Errorf("Invalid Resolution: %v", resolution)
	}
	vLen := 2.0 / float32(resolution) // L: goes from -1 to 1
	if t != T26 && t != T6 {
		return VoxelObj{}, fmt.Errorf("Invalid Connectivity Distance: %v", t)
	}

	// R_c
	boundRad := vLen / 2.0
	if t == T26 {
		boundRad *= math32.Sqrt(3.0)
	}

	vertSet := calcVertSet(obj, boundRad, vLen, resolution)       // S_v
	edgeSet := BitArrayInit(resolution * resolution * resolution) // S_e
	bodySet := BitArrayInit(resolution * resolution * resolution) // S_b

	for i := range vertSet.bits {
		vertSet.bits[i] = vertSet.bits[i] | edgeSet.bits[i] | bodySet.bits[i]
	}

	return VoxelObj{x, y, z, vertSet, color}, nil
}

func calcVertSet(obj parser.Obj, boundRad, vLen float32, resolution int) BitArray {
	set := BitArrayInit(resolution * resolution * resolution)

	// All voxels whose voxel centers fall inside R_c are added to S_v
	for _, v := range obj.Vertices {
		cX, cY, cZ := idxPos(v, resolution)
		for i := -1; i <= 1; i++ {
			for j := -1; j <= 1; j++ {
				for k := -1; k <= 1; k++ {
					if insideSphere(cX+k, cY+j, cZ+i, boundRad, v, vLen, resolution) {
						set.Set(bitIdx(cX+k, cY+j, cZ+i, resolution))
					}
				}
			}
		}
	}

	return set
}

func bitIdx(x, y, z, resolution int) int {
	return resolution*resolution*z + resolution*y + x
}

// Get closest idx of a voxel to a point
func idxPos(v te.Vector3, resolution int) (int, int, int) {
	xPos := (v.X*float32(resolution) + float32(resolution)) / 2.0
	yPos := (v.Y*float32(resolution) + float32(resolution)) / 2.0
	zPos := (v.Z*float32(resolution) + float32(resolution)) / 2.0
	x := int(math32.Round(xPos))
	y := int(math32.Round(yPos))
	z := int(math32.Round(zPos))

	return x, y, z
}

func toPos(x, y, z int, vLen float32, resolution int) te.Vector3 {
	rDiv2 := float32(resolution) / 2.0
	return te.Vec3((float32(x)-rDiv2)*vLen, (float32(y)-rDiv2)*vLen, (float32(z)-rDiv2)*vLen)
}

func insideSphere(x, y, z int, radius float32, center te.Vector3, vLen float32, resolution int) bool {
	r := resolution
	if !(x < r && y < r && z < r && x >= 0 && y >= 0 && z >= 0) {
		return false
	}

	vPos := toPos(x, y, z, vLen, resolution)
	return vPos.Sub(center).LenSqr() < radius*radius
}
