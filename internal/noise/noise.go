package noise

import (
	"github.com/chewxy/math32"
	te "github.com/zheskett/go-voxel/internal/tensor"
	"image"
	"image/color"
	"image/png"
	"os"
)

type Noise3D interface {
	Sample(vec te.Vector3) float32
}

// FBM3D samples Fractional Brownian Motion at x,y,z
//
// https://iquilezles.org/articles/fbm/
type FBM3D struct {
	Source  Noise3D
	Octaves int
	H       float32
}

// FnNoise3D is noise with smoothstep or like function applied to the output
type FnNoise3D struct {
	Source Noise3D
	Fn     func(t float32) float32
}

func (fbm *FBM3D) Sample(vec te.Vector3) float32 {
	G := math32.Exp2(-fbm.H)
	var freq, amp, total, maxVal float32 = 1.0, 1.0, 0.0, 0.0
	for range fbm.Octaves {
		total += amp * fbm.Source.Sample(vec.Mul(freq))
		maxVal += amp
		freq *= 2.0
		amp *= G
	}

	return total / maxVal
}

func (fnn3D *FnNoise3D) Sample(vec te.Vector3) float32 {
	return fnn3D.Fn(fnn3D.Fn(fnn3D.Source.Sample(vec)))
}

// Draw an image of the noise
func Draw(noise Noise3D, path string, pixels int, z float32, stepSize float32) error {
	pic := image.NewGray(image.Rect(0, 0, pixels, pixels))
	for i := range pic.Bounds().Dx() {
		for j := range pic.Bounds().Dy() {
			sample := noise.Sample(te.Vec3(float32(i)*stepSize, float32(j)*stepSize, z))
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

// CuRat is the Cubic Ration Smoothstep function for values 0-1.
func CuRat(t float32) float32 {
	return t * t * t / (3.0*t*t - 3.0*t + 1.0)
}

// QuRat is the Quadratic Rational Smoothstep function for values 0-1.
func QuRat(t float32) float32 {
	return t * t / (2.0*t*t - 2.0*t + 1.0)
}
