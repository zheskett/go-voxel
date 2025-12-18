package render

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
