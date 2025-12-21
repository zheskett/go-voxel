package render

import "github.com/chewxy/math32"

// Pixels contains the data for each pixel on the screen.
// Every pixel is 4 bytes, RGBA
type Pixels struct {
	data   []byte
	Height int
	Width  int
}

func PixelsInit(width, height int) Pixels {
	data := make([]byte, width*height*4)
	for i := 0; i < width*height*4; i++ {
		data[i] = 0
	}
	return Pixels{data, height, width}
}

func (px *Pixels) FillPixels(r, g, b byte) {
	for i := 0; i < px.Width*px.Height; i++ {
		px.data[4*i+0] = r
		px.data[4*i+1] = g
		px.data[4*i+2] = b
	}
}

func (px *Pixels) SetPixel(x, y int, r, g, b byte) {
	px.data[4*(px.Width*y+x)+0] = r
	px.data[4*(px.Width*y+x)+1] = g
	px.data[4*(px.Width*y+x)+2] = b
}

func (px *Pixels) GetPixel(x, y int) [3]byte {
	return [3]byte{
		px.data[4*(px.Width*y+x)+0],
		px.data[4*(px.Width*y+x)+1],
		px.data[4*(px.Width*y+x)+2],
	}
}

func (px *Pixels) Surrounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < px.Width && y < px.Height
}

// Floyd-Steinberg dithering copied verbatim from the Wikipedia implementation
func (px *Pixels) Dither() {
	// This probably shouldn't be specific to 'Pixels' but not sure where to put this
	//
	// Actually, see the note on the function below
	for i := range px.Height - 1 {
		for j := range px.Width - 1 {
			oldColor := px.GetPixel(j, i)

			var newColor [3]byte
			for k := range 3 {
				newColor[k] = oldColor[k] / 16 * 16
			}

			px.SetPixel(j, i, newColor[0], newColor[1], newColor[2])

			qex := int(oldColor[0] - newColor[0])
			qey := int(oldColor[1] - newColor[1])
			qez := int(oldColor[2] - newColor[2])

			clamp := func(value int, rl, rh int) int {
				if value < rl {
					value = rl
				}
				if value >= rh {
					value = rh
				}
				return value
			}

			ditherPixel := func(dx, dy int, factor int) {
				pix := px.GetPixel(j+dx, i+dy)
				px.SetPixel(
					j+dx,
					i+dy,
					byte(clamp(int(pix[0])+qex*factor/16, 0, 255)),
					byte(clamp(int(pix[1])+qey*factor/16, 0, 255)),
					byte(clamp(int(pix[2])+qez*factor/16, 0, 255)),
				)
			}

			ditherPixel(1, 0, 7)
			ditherPixel(-1, 1, 3)
			ditherPixel(0, 1, 5)
			ditherPixel(1, 1, 1)
		}
	}
}

// This function and the one above seem slow, but the runtime is so overwhelmingly
// dominated by the CPU raymarching that these almost don't matter. These should
// eventually be moved into a real shader that is applied to the framebuffer, but
// for now this works as a proof of concept
func (px *Pixels) CorrectGamma() {
	for i := range px.Height - 1 {
		for j := range px.Width - 1 {
			color := px.GetPixel(j, i)
			cx, cy, cz := color[0], color[1], color[2]
			gcx := math32.Sqrt((float32(cx) / 256.0)) * 255.9999
			gcy := math32.Sqrt((float32(cy) / 256.0)) * 255.9999
			gcz := math32.Sqrt((float32(cz) / 256.0)) * 255.9999
			px.SetPixel(j, i, byte(gcx), byte(gcy), byte(gcz))
		}
	}
}

type DepthBuffer struct {
	data   []float32
	Height int
	Width  int
}

func DepthBufferInit(width, height int) DepthBuffer {
	data := make([]float32, width*height)
	return DepthBuffer{data, height, width}
}

func (db *DepthBuffer) FillDepth(value float32) {
	for i := 0; i < db.Width*db.Height; i++ {
		db.data[i] = value
	}
}

func (db *DepthBuffer) SetDepth(x, y int, depth float32) {
	db.data[db.Width*y+x] = depth
}

func (db *DepthBuffer) GetDepth(x, y int) float32 {
	return db.data[db.Width*y+x]
}

func (db *DepthBuffer) Surrounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < db.Width && y < db.Height
}
