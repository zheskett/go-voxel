// Largely adapted from https://adrianb.io/2014/08/09/perlinnoise.html and https://cs.nyu.edu/~perlin/noise/
package noise

import (
	"fmt"
	"image/color"
	"image/png"
	"math/rand"
	"os"

	"image"

	"github.com/chewxy/math32"
	te "github.com/zheskett/go-voxel/internal/tensor"
)

// Perlin2D defines perlin noise in 2D.
// Generate this with noise.GenPerlin2D()
type Perlin3D struct {
	TextureSize int
	perm        []int
}

// GenPerlin2D generates a Perlin2D. TextureSize should be a power of 2 (>2).
//
// 256 is recommended
func GenPerlin2D(TextureSize int) (Perlin3D, error) {
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

// FBM samples Fractional Brownian Motion at x,y,z
//
// https://iquilezles.org/articles/fbm/
func (perlin *Perlin3D) FBM(vec te.Vector3, octaves int, H float32) float32 {
	G := math32.Exp2(-H)
	var freq, amp, total, maxVal float32 = 1.0, 1.0, 0.0, 0.0
	for range octaves {
		total += amp * perlin.Sample(vec.Mul(freq))
		maxVal += amp
		freq *= 2.0
		amp *= G
	}

	return total / maxVal
}

// Draw an image of the noise
func (perlin *Perlin3D) Draw(path string, pixels int, z float32, stepSize float32) error {
	pic := image.NewGray(image.Rect(0, 0, pixels, pixels))
	for i := range pic.Bounds().Dx() {
		for j := range pic.Bounds().Dy() {
			sample := perlin.Sample(te.Vec3(float32(i)*stepSize, float32(j)*stepSize, z))
			pic.SetGray(i, j, color.Gray{uint8(math32.Round(sample * 255))})
		}
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	defer file.Close()
	err = png.Encode(file, pic)
	if err != nil {
		return err
	}

	return nil
}

// Draw an image of the FBM noise
func (perlin *Perlin3D) DrawFBM(path string, pixels int, z float32, octaves int, H, stepSize float32) error {
	pic := image.NewGray(image.Rect(0, 0, pixels, pixels))
	for i := range pic.Bounds().Dx() {
		for j := range pic.Bounds().Dy() {
			sample := perlin.FBM(te.Vec3(float32(i)*stepSize, float32(j)*stepSize, z), octaves, H)
			pic.SetGray(i, j, color.Gray{uint8(math32.Round(sample * 255))})
		}
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	defer file.Close()
	err = png.Encode(file, pic)
	if err != nil {
		return err
	}

	return nil
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

// CuRat is the Cubic Ration Smoothstep function for values 0-1.
func CuRat(t float32) float32 {
	return t * t * t / (3.0*t*t - 3.0*t + 1.0)
}
