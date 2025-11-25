// Package render provides a renderer for the voxels
package render

const (
	TextureWidth  = 320
	TextureHeight = 240
)

type Pixels []byte

func Render() Pixels {
	pixels := make(Pixels, TextureWidth*TextureHeight*4)
	for i := range pixels {
		pixels[i] = 255
		if i&0x3 == 0x1 {
			pixels[i] = 0
		}
	}
	return pixels
}
