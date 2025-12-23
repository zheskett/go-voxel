// Largely adapted from https://adrianb.io/2014/08/09/perlinnoise.html and https://cs.nyu.edu/~perlin/noise/
package noise

import (
	"fmt"
	te "github.com/zheskett/go-voxel/internal/tensor"
	"math/rand/v2"
)

// Perlin2D defines perlin noise in 2D.
// Generate this with noise.GenPerlin2D()
type Perlin3D struct {
	TextureSize int
	perm        []int
}

// GenPerlin3D generates a Perlin2D. TextureSize should be a power of 2 (>2).
//
// 256 is recommended
func GenPerlin3D(TextureSize int) (Perlin3D, error) {
	if TextureSize <= 2 || (TextureSize&(TextureSize-1)) != 0 {
		return Perlin3D{}, fmt.Errorf("TextureSize should be power of 2 (>2)")
	}
	perlin := Perlin3D{TextureSize, rand.Perm(TextureSize)}
	perlin.perm = append(perlin.perm, perlin.perm...)

	return perlin, nil
}

// Sample returns the noise value at the given position
func (perlin *Perlin3D) Sample(vec te.Vector3) float32 {
	tm1 := perlin.TextureSize - 1
	xCube, yCube, zCube := int(vec.X)&tm1, int(vec.Y)&tm1, int(vec.Z)&tm1
	x, y, z := vec.X-float32(int(vec.X)), vec.Y-float32(int(vec.Y)), vec.Z-float32(int(vec.Z))
	p := perlin.perm
	aaa, aab := p[p[p[xCube]+yCube]+zCube], p[p[p[xCube]+yCube]+zCube+1]
	aba, abb := p[p[p[xCube]+yCube+1]+zCube], p[p[p[xCube]+yCube+1]+zCube+1]
	baa, bab := p[p[p[xCube+1]+yCube]+zCube], p[p[p[xCube+1]+yCube]+zCube+1]
	bba, bbb := p[p[p[xCube+1]+yCube+1]+zCube], p[p[p[xCube+1]+yCube+1]+zCube+1]
	u, v, w := fade(x), fade(y), fade(z)

	// Lerp 8 corners together
	lerp1 := lerp(grad(aaa, x, y, z), grad(baa, x-1, y, z), u)
	lerp2 := lerp(grad(aba, x, y-1, z), grad(bba, x-1, y-1, z), u)
	lerp3 := lerp(grad(aab, x, y, z-1), grad(bab, x-1, y, z-1), u)
	lerp4 := lerp(grad(abb, x, y-1, z-1), grad(bbb, x-1, y-1, z-1), u)
	return (lerp(lerp(lerp1, lerp2, v), lerp(lerp3, lerp4, v), w) + 1) * 0.5
}

// Gradient, dot product between gradient vector and relative position
func grad(hash int, x, y, z float32) float32 {
	var res float32
	switch hash & 0xF {
	case 0x0:
		res = x + y
	case 0x1:
		res = -x + y
	case 0x2:
		res = x - y
	case 0x3:
		res = -x - y
	case 0x4:
		res = x + z
	case 0x5:
		res = -x + z
	case 0x6:
		res = x - z
	case 0x7:
		res = -x - z
	case 0x8:
		res = y + z
	case 0x9:
		res = -y + z
	case 0xA:
		res = y - z
	case 0xB:
		res = -y - z
	case 0xC:
		res = y + x
	case 0xD:
		res = -y + z
	case 0xE:
		res = y - x
	case 0xF:
		res = -y - z
	}
	return res
}

func lerp(a, b, t float32) float32 {
	return a + t*(b-a)
}

// 6t^5 - 15t^4 + 10t^3
func fade(t float32) float32 {
	return t * t * t * (t*(t*6-15) + 10)
}
