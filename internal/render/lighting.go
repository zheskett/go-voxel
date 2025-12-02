package render

import (
	"github.com/chewxy/math32"
	te "github.com/zheskett/go-voxel/internal/tensor"
	vxl "github.com/zheskett/go-voxel/internal/voxel"
)

// Performs the per-pixel lighting by sending secondary rays back towards all of the lights in the scene
// Much slower than below funcs, but looks very nice
func GetPixelShading(vox *vxl.Voxels, hit vxl.RayHit, tmax float32) te.Vector3 {
	intensity := te.Vec3Zero()
	for _, light := range vox.Lights {
		lightpos := light.Position.Sub(hit.Position)
		lightdist := lightpos.Len()
		lightdir := lightpos.Div(lightdist)
		recastpos := hit.Position.Add(hit.Normal.Mul(vxl.VoxelRayDelta))
		recastray := vxl.Ray{
			Origin: recastpos,
			Dir:    lightdir,
			Tmax:   math32.Min(lightdist, tmax),
		}

		shadowcast := vox.MarchRay(recastray)

		// If we don't hit anything, the pixel has direct view of the light, as the rayline
		// has no obstruction
		if !shadowcast.Hit {
			brightness := math32.Max(0.0, hit.Normal.Dot(lightdir)) / lightdist
			intensity = intensity.Add(light.Color.Mul(brightness))
		}
	}

	return intensity
}

// Gets the per-voxel lighting from cache or calculating it
func GetVoxelShading(vox *vxl.Voxels, hit vxl.RayHit, tmax float32) te.Vector3 {
	x, y, z := hit.IntPos[0], hit.IntPos[1], hit.IntPos[2]
	idx := vox.Index(x, y, z)

	var light vxl.CachedLighting
	if vox.LightCached.Get(idx) {
		light = vox.Lighting[idx]
	} else {
		light = shadeVoxel(vox, hit, tmax)
		vox.Lighting[idx] = light
		vox.LightCached.Set(idx)
	}

	brightness := math32.Max(0.0, hit.Normal.Dot(light.Dir))
	return light.Light.Mul(brightness)
}

// Performs the per-voxel lighting (attemps to atleast) by caching shadow data from the voxel face center
func shadeVoxel(vox *vxl.Voxels, hit vxl.RayHit, tmax float32) vxl.CachedLighting {
	intensity := te.Vec3Zero()
	direction := te.Vec3Zero()
	x, y, z := float32(hit.IntPos[0]), float32(hit.IntPos[1]), float32(hit.IntPos[2])
	// The flickering is coming from this. The race condition technically shouldn't
	// matter (because the lights don't move), but what is happening is depdending on
	// how the camera is oriented a different face of the voxel is being struck first,
	// leading to a different exposure angle to the lights and it gets biased towards
	// one color.
	voxelcenter := te.Vec3(x+0.5, y+0.5, z+0.5).Add(hit.Normal.Mul(0.5))
	for _, light := range vox.Lights {
		lightpos := light.Position.Sub(voxelcenter)
		lightdist := lightpos.Len()
		lightdir := lightpos.Div(lightdist)
		recastpos := voxelcenter.Add(hit.Normal.Mul(vxl.VoxelRayDelta + 0.5))
		recastray := vxl.Ray{
			Origin: recastpos,
			Dir:    lightdir,
			Tmax:   math32.Min(lightdist, tmax),
		}

		shadowcast := vox.MarchRay(recastray)

		// If we don't hit anything, the pixel has direct view of the light, as the rayline
		// has no obstruction
		if !shadowcast.Hit {
			intensity = intensity.Add(light.Color.Div(lightdist))
			direction = direction.Add(lightpos)
		}
	}
	direction = direction.Normalized()

	return vxl.CachedLighting{Light: intensity, Dir: direction}
}
