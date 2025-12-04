package voxel

import (
	"fmt"
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

// Same as Voxelize(ParseObj(path), ...) basically
func VoxelizePath(path string, cd ConnectivityDistance, resolution int, color [3]byte) (VoxelObj, error) {
	obj, err := parser.ParseObj(path)
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

	set := BitArrayInit(resolution * resolution * resolution)
	calcVertSet(&set, obj, boundRad, vLen, resolution) // S_v
	calcEdgeSet(&set, obj, boundRad, vLen, resolution) // S_e
	calcBodySet(&set, obj, cd, vLen, resolution)       // S_b
	vObj := VoxelObj{resolution, resolution, resolution, set, color}
	squash(&vObj, resolution)

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
		for z := range vObj.X / 2 {
			for y := range vObj.Y {
				for x := range vObj.Z {
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

func calcVertSet(set *BitArray, obj parser.Obj, boundRad, vLen float32, resolution int) {
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
}

func calcEdgeSet(set *BitArray, obj parser.Obj, boundRad, vLen float32, resolution int) {
	// All voxels whose voxel center fall inside a cylinder with radius R_c
	// and length L, where L is the length of the edge, are added to S_e
	for _, e := range obj.Edges {
		v1, v2 := obj.Vertices[e[0]], obj.Vertices[e[1]]
		stepVec := v2.Sub(v1).Normalized().Mul(vLen * 0.5)

		// While pointing towards v2
		for pos := v1; v2.Sub(pos).Dot(stepVec) > 0; pos = pos.Add(stepVec) {
			cX, cY, cZ := idxPos(pos, resolution)
			for i := -1; i <= 1; i++ {
				for j := -1; j <= 1; j++ {
					for k := -1; k <= 1; k++ {
						if insideCylinder(cX+k, cY+j, cZ+i, boundRad, v1, v2, vLen, resolution) {
							set.Set(bitIdx(cX+k, cY+j, cZ+i, resolution))
						}
					}
				}
			}
		}
	}
}

func calcBodySet(set *BitArray, obj parser.Obj, cd ConnectivityDistance, vLen float32, resolution int) {
	// All voxels who are inside planes G and H and inside edge planes E1 - E3 are added to S_f
	invSqrt3 := 1.0 / math32.Sqrt(3.0)
	sqrt3 := math32.Sqrt(3.0)
	for _, f := range obj.Faces {
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
		xMin, yMin, zMin := idxPos(te.Vec3(worldXMin, worldYMin, worldZMin), resolution)
		xMax, yMax, zMax := idxPos(te.Vec3(worldXMax, worldYMax, worldZMax), resolution)

		for z := zMin; z <= zMax; z++ {
			for y := yMin; y <= yMax; y++ {
				for x := xMin; x <= xMax; x++ {
					if betweenPlanes(x, y, z, facePlane, t, vLen, resolution) &&
						insidePlaneTriangle(x, y, z, e1, e2, e3, vLen, resolution) {
						set.Set(bitIdx(x, y, z, resolution))
					}
				}
			}
		}
	}
}

func bitIdx(x, y, z, resolution int) int {
	return resolution*resolution*z + resolution*y + x
}

// Get closest idx of a voxel to a point
func idxPos(v te.Vector3, resolution int) (int, int, int) {
	xPos := (v.X*float32(resolution) + float32(resolution)) * 0.5
	yPos := (v.Y*float32(resolution) + float32(resolution)) * 0.5
	zPos := (v.Z*float32(resolution) + float32(resolution)) * 0.5
	x := int(math32.Round(xPos))
	y := int(math32.Round(yPos))
	z := int(math32.Round(zPos))

	return x, y, z
}

func toPos(x, y, z int, vLen float32, resolution int) te.Vector3 {
	rDiv2 := float32(resolution) * 0.5
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

func insideCylinder(x, y, z int, radius float32, a, b te.Vector3, vLen float32, resolution int) bool {
	r := resolution
	if !(x < r && y < r && z < r && x >= 0 && y >= 0 && z >= 0) {
		return false
	}

	vPos := toPos(x, y, z, vLen, resolution)
	e := b.Sub(a)
	return vPos.Sub(a).Dot(e) > 0 &&
		vPos.Sub(b).Dot(e) < 0 &&
		vPos.Sub(a).Cross(e).LenSqr() < radius*radius*e.LenSqr()
}

func betweenPlanes(x, y, z int, facePlane plane, t float32, vLen float32, resolution int) bool {
	r := resolution
	if !(x < r && y < r && z < r && x >= 0 && y >= 0 && z >= 0) {
		return false
	}

	vPos := toPos(x, y, z, vLen, resolution)
	distance := facePlane.normVec.Dot(vPos) + facePlane.d
	return math32.Abs(distance) <= t
}

func insidePlaneTriangle(x, y, z int, e1, e2, e3 plane, vLen float32, resolution int) bool {
	r := resolution
	if !(x < r && y < r && z < r && x >= 0 && y >= 0 && z >= 0) {
		return false
	}

	vPos := toPos(x, y, z, vLen, resolution)
	distanceE1 := e1.normVec.Dot(vPos) + e1.d
	distanceE2 := e2.normVec.Dot(vPos) + e2.d
	distanceE3 := e3.normVec.Dot(vPos) + e3.d
	return distanceE1 > 0 && distanceE2 > 0 && distanceE3 > 0
}

// Squashes the empty space on each axis to get accurate X, Y, Z
func squash(vObj *VoxelObj, resolution int) {
	minX, minY, minZ, maxX, maxY, maxZ := findBounds(vObj, resolution)
	newX, newY, newZ := maxX-minX+1, maxY-minY+1, maxZ-minZ+1
	newPresence := BitArrayInit(newX * newY * newZ)
	currX, currY, currZ := 0, 0, 0
	for z := minZ; z <= maxZ; z++ {
		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				isFilled := vObj.Presence.Get(vObj.Index(x, y, z))
				if isFilled {
					newPresence.Set(currZ*newY*newX + currY*newX + currX)
				}
				currX++
			}
			currX = 0
			currY++
		}
		currY = 0
		currZ++
	}
	vObj.Presence = newPresence
	vObj.X, vObj.Y, vObj.Z = newX, newY, newZ
}

func findBounds(vObj *VoxelObj, resolution int) (int, int, int, int, int, int) {
	minXC, minYC, minZC := 0, 0, 0
	maxXC, maxYC, maxZC := resolution-1, resolution-1, resolution-1

	var wg sync.WaitGroup
	wg.Go(func() {
		foundmin, foundmax := false, false
		for x := range resolution {
			for z := range resolution {
				for y := range resolution {
					if !foundmin && vObj.Presence.Get(vObj.Index(x, y, z)) {
						minXC = x
						foundmin = true
					}
					if !foundmax && vObj.Presence.Get(vObj.Index(resolution-x-1, y, z)) {
						maxXC = resolution - x - 1
						foundmax = true
					}
					if foundmin && foundmax {
						return
					}
				}
			}
		}
	})

	wg.Go(func() {
		foundmin, foundmax := false, false
		for y := range resolution {
			for z := range resolution {
				for x := range resolution {
					if !foundmin && vObj.Presence.Get(vObj.Index(x, y, z)) {
						minYC = y
						foundmin = true
					}
					if !foundmax && vObj.Presence.Get(vObj.Index(x, resolution-y-1, z)) {
						maxYC = resolution - y - 1
						foundmax = true
					}
					if foundmin && foundmax {
						return
					}
				}
			}
		}
	})

	wg.Go(func() {
		foundmin, foundmax := false, false
		for z := range resolution {
			for y := range resolution {
				for x := range resolution {
					if !foundmin && vObj.Presence.Get(vObj.Index(x, y, z)) {
						minZC = z
						foundmin = true
					}
					if !foundmax && vObj.Presence.Get(vObj.Index(x, y, resolution-z-1)) {
						maxZC = resolution - z - 1
						foundmax = true
					}
					if foundmin && foundmax {
						return
					}
				}
			}
		}
	})

	wg.Wait()

	return minXC, minYC, minZC, maxXC, maxYC, maxZC
}
