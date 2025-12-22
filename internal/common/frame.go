package common

import (
	"fmt"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/zheskett/go-voxel/internal/tensor"
)

// FrameData allows camera movements to be made independent of FPS for a smoother movements
type FrameData struct {
	Tick     uint
	Deltat   float64
	Previous time.Time

	Keys  map[glfw.Key]bool
	mouse tensor.Vector2
}

func FrameDataInit() FrameData {
	return FrameData{Previous: time.Now(), Keys: make(map[glfw.Key]bool)}
}

func (data *FrameData) Update() {
	data.Deltat = float64(time.Since(data.Previous).Seconds())
	data.Previous = time.Now()
	data.Tick += 1
}

func (data *FrameData) GetMouseDelta(mxc, myc float64) (float64, float64) {
	mx, my := float64(mxc), float64(myc)
	dx, dy := data.mouse.X-mx, data.mouse.Y-my
	data.mouse = tensor.Vec2(mx, my)
	return dx, dy
}

func (data *FrameData) ReportFps() {
	fmt.Printf("FPS: %.2f\n", 1.0/data.Deltat)
}
