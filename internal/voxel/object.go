package voxel

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/chewxy/math32"
	"github.com/zheskett/go-voxel/internal/parser"
	te "github.com/zheskett/go-voxel/internal/tensor"
)

type VoxelObj struct {
	X, Y, Z  int
	Presence BitArray
	Color    [3]byte
}

type ConnectivityDistance int
type plane struct {
	normVec te.Vector3
	d       float32
}

const (
	T26 ConnectivityDistance = 26
	T6  ConnectivityDistance = 6
)

const (
	setChanSize = 1000
)

var (
	cpus = runtime.NumCPU()
)

// Same as Voxelize(ParseObj(path), ...) basically
func VoxelizePath(path string, flipX, flipY, flipZ bool, cd ConnectivityDistance,
	resolution int, color [3]byte) (VoxelObj, error) {

	obj, err := parser.ParseObj(path, flipX, flipY, flipZ)
	if err != nil {
		return VoxelObj{}, err
	}

	return Voxelize(obj, cd, resolution, color)
}

// Turns an obj into voxels
//
// Algorithm from https://web.eecs.utk.edu/~huangj/papers/polygon.pdf
func Voxelize(obj parser.Obj, cd ConnectivityDistance, resolution int, color [3]byte) (VoxelObj, error) {
	if resolution < 1 {
		return VoxelObj{}, fmt.Errorf("Invalid Resolution: %v", resolution)
	}
	vLen := 2.0 / float32(resolution) // L: goes from -1 to 1
	if cd != T26 && cd != T6 {
		return VoxelObj{}, fmt.Errorf("Invalid Connectivity Distance: %v", cd)
	}

	// R_c
	boundRad := vLen / 2.0
	if cd == T26 {
		boundRad *= math32.Sqrt(3.0)
	}

	// Calculate X, Y, Z
	X := int(math32.Ceil(float32(resolution) * obj.MaxVertsPos.X))
	Y := int(math32.Ceil(float32(resolution) * obj.MaxVertsPos.Y))
	Z := int(math32.Ceil(float32(resolution) * obj.MaxVertsPos.Z))

	set := BitArrayInit(Z * Y * X)
	vObj := VoxelObj{X, Y, Z, set, color}
	setChan := make(chan int, setChanSize)

	var wg sync.WaitGroup
	wg.Go(func() { calcVertSet(setChan, obj, boundRad, vLen, X, Y, Z) }) // S_v
	wg.Go(func() { calcEdgeSet(setChan, obj, boundRad, vLen, X, Y, Z) }) // S_e
	wg.Go(func() { calcBodySet(setChan, obj, cd, vLen, X, Y, Z) })       // S_b

	go func() {
		wg.Wait()
		close(setChan)
	}()

	for i := range setChan {
		set.Set(i)
	}

	return vObj, nil
}

func (vObj *VoxelObj) Index(x, y, z int) int {
	return vObj.X*vObj.Y*z + vObj.X*y + x
}

// Flips the voxels in the x, y, and z directions.
// Flip x -> y -> z
func (vObj *VoxelObj) Flip(flipX, flipY, flipZ bool) {
	if flipX {
		for z := range vObj.Z {
			for y := range vObj.Y {
				for x := range vObj.X / 2 {
					idx1 := vObj.Index(x, y, z)
					idx2 := vObj.Index(vObj.X-x-1, y, z)
					oldSet := vObj.Presence.Get(idx1)
					vObj.Presence.Put(idx1, vObj.Presence.Get(idx2))
					vObj.Presence.Put(idx2, oldSet)
				}
			}
		}
	}
	if flipY {
		for z := range vObj.Z {
			for y := range vObj.Y / 2 {
				for x := range vObj.X {
					idx1 := vObj.Index(x, y, z)
					idx2 := vObj.Index(x, vObj.Y-y-1, z)
					oldSet := vObj.Presence.Get(idx1)
					vObj.Presence.Put(idx1, vObj.Presence.Get(idx2))
					vObj.Presence.Put(idx2, oldSet)
				}
			}
		}
	}
	if flipZ {
		for z := range vObj.Z / 2 {
			for y := range vObj.Y {
				for x := range vObj.X {
					idx1 := vObj.Index(x, y, z)
					idx2 := vObj.Index(x, y, vObj.Z-z-1)
					oldSet := vObj.Presence.Get(idx1)
					vObj.Presence.Put(idx1, vObj.Presence.Get(idx2))
					vObj.Presence.Put(idx2, oldSet)
				}
			}
		}
	}
}

func calcVertSet(setChan chan int, obj parser.Obj, boundRad float32, vLen float32, X, Y, Z int) {
	// All voxels whose voxel centers fall inside R_c are added to S_v
	for _, v := range obj.Vertices {
		cX, cY, cZ := idxPos(v, X, Y, Z, vLen)
		for i := -1; i <= 1; i++ {
			for j := -1; j <= 1; j++ {
				for k := -1; k <= 1; k++ {
					if insideSphere(cX+k, cY+j, cZ+i, boundRad, v, X, Y, Z, vLen) {
						setChan <- bitIdx(cX+k, cY+j, cZ+i, X, Y, Z)
					}
				}
			}
		}
	}
}

func calcEdgeSet(setChan chan int, obj parser.Obj, boundRad, vLen float32, X, Y, Z int) {
	// All voxels whose voxel center fall inside a cylinder with radius R_c
	// and length L, where L is the length of the edge, are added to S_e
	for _, e := range obj.Edges {
		v1, v2 := obj.Vertices[e[0]], obj.Vertices[e[1]]
		stepVec := v2.Sub(v1).Normalized().Mul(vLen * 0.5)

		// While pointing towards v2
		for pos := v1; v2.Sub(pos).Dot(stepVec) > 0; pos = pos.Add(stepVec) {
			cX, cY, cZ := idxPos(pos, X, Y, Z, vLen)
			for i := -1; i <= 1; i++ {
				for j := -1; j <= 1; j++ {
					for k := -1; k <= 1; k++ {
						if insideCylinder(cX+k, cY+j, cZ+i, boundRad, v1, v2, X, Y, Z, vLen) {
							setChan <- bitIdx(cX+k, cY+j, cZ+i, X, Y, Z)
						}
					}
				}
			}
		}
	}
}

func calcBodySet(setChan chan int, obj parser.Obj, cd ConnectivityDistance, vLen float32, X, Y, Z int) {
	// All voxels who are inside planes G and H and inside edge planes E1 - E3 are added to S_f
	invSqrt3 := 1.0 / math32.Sqrt(3.0)
	sqrt3 := math32.Sqrt(3.0)
	var wg sync.WaitGroup
	faceChan := make(chan [3]int, cpus)
	for range cpus {
		wg.Go(func() {
			for f := range faceChan {
				v1, v2, v3 := obj.Vertices[f[0]], obj.Vertices[f[1]], obj.Vertices[f[2]]
				facePlane := plane{}
				facePlane.normVec = v2.Sub(v1).Cross(v3.Sub(v1)).Normalized()
				facePlane.d = facePlane.normVec.Dot(v1) * -1

				e1, e2, e3 := plane{}, plane{}, plane{}
				e1.normVec = facePlane.normVec.Cross(v2.Sub(v1)).Normalized()
				e1.d = e1.normVec.Dot(v1) * -1
				e2.normVec = facePlane.normVec.Cross(v3.Sub(v2)).Normalized()
				e2.d = e2.normVec.Dot(v2) * -1
				e3.normVec = facePlane.normVec.Cross(v1.Sub(v3)).Normalized()
				e3.d = e3.normVec.Dot(v3) * -1

				cosBeta := max(math32.Abs(facePlane.normVec.X), math32.Abs(facePlane.normVec.Y), math32.Abs(facePlane.normVec.Z))
				t := vLen * 0.5 * cosBeta
				if cd == T26 {
					// Find closest diagonal cos
					cosAlpha := float32(0.0)
					for i := -1; i <= 1; i += 2 {
						for j := -1; j <= 1; j += 2 {
							for k := -1; k <= 1; k += 2 {
								diagVec := te.Vec3(float32(i), float32(j), float32(k)).Mul(invSqrt3)
								cosAlpha = max(cosAlpha, facePlane.normVec.Dot(diagVec))
							}
						}
					}
					t = vLen * 0.5 * sqrt3 * cosAlpha
				}

				// Need to find better way to do this
				// Create a bounding box around the face
				worldXMin, worldXMax := min(v1.X, v2.X, v3.X)-t, max(v1.X, v2.X, v3.X)+t
				worldYMin, worldYMax := min(v1.Y, v2.Y, v3.Y)-t, max(v1.Y, v2.Y, v3.Y)+t
				worldZMin, worldZMax := min(v1.Z, v2.Z, v3.Z)-t, max(v1.Z, v2.Z, v3.Z)+t
				xMin, yMin, zMin := idxPos(te.Vec3(worldXMin, worldYMin, worldZMin), X, Y, Z, vLen)
				xMax, yMax, zMax := idxPos(te.Vec3(worldXMax, worldYMax, worldZMax), X, Y, Z, vLen)

				for z := zMin; z <= zMax; z++ {
					for y := yMin; y <= yMax; y++ {
						for x := xMin; x <= xMax; x++ {
							if betweenPlanes(x, y, z, facePlane, t, X, Y, Z, vLen) &&
								insidePlaneTriangle(x, y, z, e1, e2, e3, X, Y, Z, vLen) {
								setChan <- bitIdx(x, y, z, X, Y, Z)
							}
						}
					}
				}
			}
		})
	}

	for _, f := range obj.Faces {
		faceChan <- f
	}
	close(faceChan)
	wg.Wait()
}

func bitIdx(x, y, z, X, Y, _ int) int {
	return X*Y*z + X*y + x
}

// Get closest idx of a voxel to a point
func idxPos(v te.Vector3, X, Y, Z int, vLen float32) (int, int, int) {
	vLenInv := 1.0 / vLen
	xPos := v.X*vLenInv + float32(X-1)*0.5
	yPos := v.Y*vLenInv + float32(Y-1)*0.5
	zPos := v.Z*vLenInv + float32(Z-1)*0.5
	x := int(math32.Round(xPos))
	y := int(math32.Round(yPos))
	z := int(math32.Round(zPos))

	return x, y, z
}

func toPos(x, y, z int, vLen float32, X, Y, Z int) te.Vector3 {
	xPos := (float32(x) - float32(X-1)*0.5) * vLen
	yPos := (float32(y) - float32(Y-1)*0.5) * vLen
	zPos := (float32(z) - float32(Z-1)*0.5) * vLen
	return te.Vec3(xPos, yPos, zPos)
}

func surrounds(x, y, z int, X, Y, Z int) bool {
	return x < X && y < Y && z < Z && x >= 0 && y >= 0 && z >= 0
}

func insideSphere(x, y, z int, radius float32, center te.Vector3, X, Y, Z int, vLen float32) bool {
	if !surrounds(x, y, z, X, Y, Z) {
		return false
	}

	vPos := toPos(x, y, z, vLen, X, Y, Z)
	return vPos.Sub(center).LenSqr() <= radius*radius
}

func insideCylinder(x, y, z int, radius float32, a, b te.Vector3, X, Y, Z int, vLen float32) bool {
	if !surrounds(x, y, z, X, Y, Z) {
		return false
	}

	vPos := toPos(x, y, z, vLen, X, Y, Z)
	e := b.Sub(a)
	return vPos.Sub(a).Dot(e) >= 0 &&
		vPos.Sub(b).Dot(e) <= 0 &&
		vPos.Sub(a).Cross(e).LenSqr() <= radius*radius*e.LenSqr()
}

func betweenPlanes(x, y, z int, facePlane plane, t float32, X, Y, Z int, vLen float32) bool {
	if !surrounds(x, y, z, X, Y, Z) {
		return false
	}

	vPos := toPos(x, y, z, vLen, X, Y, Z)
	distance := facePlane.normVec.Dot(vPos) + facePlane.d
	return math32.Abs(distance) <= t
}

func insidePlaneTriangle(x, y, z int, e1, e2, e3 plane, X, Y, Z int, vLen float32) bool {
	if !surrounds(x, y, z, X, Y, Z) {
		return false
	}

	vPos := toPos(x, y, z, vLen, X, Y, Z)
	distanceE1 := e1.normVec.Dot(vPos) + e1.d
	distanceE2 := e2.normVec.Dot(vPos) + e2.d
	distanceE3 := e3.normVec.Dot(vPos) + e3.d
	return distanceE1 >= 0 && distanceE2 >= 0 && distanceE3 >= 0
}
