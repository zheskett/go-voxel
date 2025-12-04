package render

import (
	"github.com/chewxy/math32"
	te "github.com/zheskett/go-voxel/internal/tensor"
	vxl "github.com/zheskett/go-voxel/internal/voxel"
)

// The way the Unity HDRP prevents the brightspots is by just clamping the distance
// to prevent lights getting too bright
const (
	MinDistance float32 = 8.0
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
			brightness := math32.Max(0.0, hit.Normal.Dot(lightdir)) * lightFalloffCurve(lightdist)
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

// Performs the per-voxel lighting (attempts to at least) by caching shadow data from the voxel face center
//
// This actually now works really well, there is no flickering however there is an
// issue where walls must be < 1 voxel thick when using this or they won't actually
// be opaque, as the light ray basically jumps out to the nearest corner of the
// parent voxel
func shadeVoxel(vox *vxl.Voxels, hit vxl.RayHit, tmax float32) vxl.CachedLighting {
	intensity := te.Vec3Zero()
	direction := te.Vec3Zero()
	x, y, z := float32(hit.IntPos[0]), float32(hit.IntPos[1]), float32(hit.IntPos[2])
	voxelcenter := te.Vec3(x+0.5, y+0.5, z+0.5)
	distanceoutvoxel := math32.Sqrt(0.5 * 0.5 * 3)
	for _, light := range vox.Lights {
		lightpos := light.Position.Sub(voxelcenter)
		lightdist := lightpos.Len()
		lightdir := lightpos.Div(lightdist)
		outsidedirec := lightdir.SignVec() // Shift over to one of the corners for the shadow vector's origin
		recastpos := voxelcenter.Add(outsidedirec.Mul(distanceoutvoxel + vxl.VoxelRayDelta))
		recastray := vxl.Ray{
			Origin: recastpos,
			Dir:    lightdir,
			Tmax:   math32.Min(lightdist-distanceoutvoxel-vxl.VoxelRayDelta, tmax),
		}

		shadowcast := vox.MarchRay(recastray)

		// If we don't hit anything, the pixel has direct view of the light, as the rayline
		// has no obstruction
		if !shadowcast.Hit {
			intensity = intensity.Add(light.Color.Mul(lightFalloffCurve(lightdist)))
			direction = direction.Add(lightpos)
		}
	}
	direction = direction.Normalized()

	return vxl.CachedLighting{Light: intensity, Dir: direction}
}

func lightFalloffCurve(dist float32) float32 {
	return 1.0 / math32.Max(dist*dist, MinDistance*MinDistance)
}
