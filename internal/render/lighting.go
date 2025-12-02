package render

import (
	"github.com/chewxy/math32"
	te "github.com/zheskett/go-voxel/internal/tensor"
	vxl "github.com/zheskett/go-voxel/internal/voxel"
)

func ShadePixel(vox *vxl.Voxels, hit vxl.RayHit) te.Vector3 {
	intensity := te.Vec3Zero()
	for _, light := range vox.Lights {
		lightpos := light.Position.Sub(hit.Position)
		lightdist := lightpos.Len()
		lightdir := lightpos.Div(lightdist)
		recastpos := hit.Position.Add(hit.Normal.Mul(vxl.VoxelRayDelta))
		recastray := vxl.Ray{
			Origin: recastpos,
			Dir:    lightdir,
			Tmax:   lightdist,
		}

		shadowcast := vox.MarchRay(recastray)

		if !shadowcast.Hit {
			brightness := math32.Max(0.0, hit.Normal.Dot(lightdir))
			intensity = intensity.Add(light.Color.Mul(brightness))
		}
	}
	intensity = intensity.Div(float32(len(vox.Lights)))

	return intensity
}
