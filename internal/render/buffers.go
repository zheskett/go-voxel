package render

import (
	"runtime"
	"sync"

	"github.com/chewxy/math32"
	"github.com/zheskett/go-voxel/internal/tensor"
)

// It would be super cool if generics worked normally, considering like 3 of these
// types are basically identical

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
	// Actually, see the note on the function ColorBuffer.CorrectGamma
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

type ColorBuffer struct {
	data   []tensor.Vector3
	Height int
	Width  int
}

func ColorBufferInit(width, height int) ColorBuffer {
	data := make([]tensor.Vector3, width*height)
	return ColorBuffer{data, height, width}
}

func (cb *ColorBuffer) FillColor(value tensor.Vector3) {
	for i := 0; i < cb.Width*cb.Height; i++ {
		cb.data[i] = value
	}
}

func (cb *ColorBuffer) SetColor(x, y int, color tensor.Vector3) {
	cb.data[cb.Width*y+x] = color
}

func (cb *ColorBuffer) GetColor(x, y int) tensor.Vector3 {
	return cb.data[cb.Width*y+x]
}

func (cb *ColorBuffer) Surrounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < cb.Width && y < cb.Height
}

// This function and the one above seem slow, but the runtime is so overwhelmingly
// dominated by the CPU raymarching that these almost don't matter. These should
// eventually be moved into a real shader that is applied to the framebuffer, but
// for now this works as a proof of concept
func (px *ColorBuffer) CorrectGamma() {
	numThreads := runtime.NumCPU() - 1
	threads := sync.WaitGroup{}
	for thread := range numThreads {
		threads.Go(func() {
			startrow := thread * px.Height / numThreads
			endrow := (thread + 1) * px.Height / numThreads

			for row := startrow; row < endrow; row++ {
				for col := 0; col < px.Width; col++ {
					color := px.GetColor(col, row)
					cx, cy, cz := color.Elms()
					gcx := math32.Pow((float32(cx)/256.0), 0.4) * 255.99999
					gcy := math32.Pow((float32(cy)/256.0), 0.4) * 255.99999
					gcz := math32.Pow((float32(cz)/256.0), 0.4) * 255.99999
					px.SetColor(col, row, tensor.Vec3(gcx, gcy, gcz))
				}
			}
		})
	}
	threads.Wait()
}

// Compact storage for an array of bools
type BitArray struct {
	bits []uint64
}

func BitArrayInit(len int) BitArray {
	if len%64 != 0 {
		len += 64
	}
	len = len / 64
	bits := make([]uint64, len)
	for i := range len {
		bits[i] = 0
	}
	return BitArray{bits}
}

func (bits *BitArray) Get(index int) bool {
	bucket := index / 64
	shift := index % 64
	mask := uint64(1) << shift
	return bits.bits[bucket]&mask != 0
}

func (bits *BitArray) Set(index int) {
	bucket := index / 64
	shift := index % 64
	mask := uint64(1) << shift
	bits.bits[bucket] |= mask
}

func (bits *BitArray) Put(index int, value bool) {
	bucket := index / 64
	shift := index % 64
	mask := uint64(1) << shift
	if value {
		bits.bits[bucket] |= mask
	} else {
		bits.bits[bucket] &= ^mask
	}
}

func (bits *BitArray) Reset(index int) {
	bucket := index / 64
	shift := index % 64
	mask := uint64(1) << shift
	bits.bits[bucket] ^= mask
}

func (bits *BitArray) Clear() {
	for i := range bits.bits {
		bits.bits[i] = 0
	}
}
