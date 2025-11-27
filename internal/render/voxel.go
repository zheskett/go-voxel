package render

// Naive storage as an array
type Voxels struct {
	z, y, x  int
	presense []bool
	color    []Color
}

func VoxelsInit(x, y, z int) Voxels {
	presense := make([]bool, y*x*z)
	color := make([]Color, y*x*z)
	for i := range presense {
		presense[i] = false
		color[i] = Color{0, 0, 0}
	}
	return Voxels{z, y, x, presense, color}
}

func (vox *Voxels) Index(x, y, z int) int {
	return vox.x*vox.y*z + vox.x*y + x
}

func (vox *Voxels) SetVoxel(x, y, z int, r, g, b byte) {
	idx := vox.Index(x, y, z)
	vox.presense[idx] = true
	vox.color[idx] = [3]byte{r, g, b}
}

func (vox*Voxels) MarchRay(ray Ray) RayHit {
	// TODO
	return RayHit{}
}
