package raytracer

import (
	"math"
)

// Triangle definition
// raycasting is already expensive and trying to calculate the triangle
// in each raycast makes it harder. So we are simplifying triangle definition
type Triangle struct {
	id       int64
	P1       Vector
	P2       Vector
	P3       Vector
	N1       Vector
	N2       Vector
	N3       Vector
	T1       Vector
	T2       Vector
	T3       Vector
	Material Material
	Photons  []Photon
	Smooth   bool
}

// Intersection defines the ratcast triangle intersection result
type Intersection struct {
	Hit                bool
	Triangle           *Triangle
	Intersection       Vector
	IntersectionNormal Vector
	RayStart           Vector
	RayDir             Vector
	ObjectName         string
	Dist               float64
	Hits               int
}

func (t *Triangle) equals(dest Triangle) bool {
	return t.P1 == dest.P1 && t.P2 == dest.P2 && t.P3 == dest.P3
}

func (t *Triangle) midPoint() Vector {
	mid := t.P1
	mid = addVector(mid, t.P2)
	mid = addVector(mid, t.P3)
	mid = scaleVector(mid, 1.0/3.0)
	return mid
}

func (t *Triangle) getBoundingBox() BoundingBox {
	result := BoundingBox{}
	result[0] = t.P1
	result[1] = t.P1
	result.extendVector(t.P2)
	result.extendVector(t.P3)
	return result
}

func (i *Intersection) getTexCoords() Vector {
	u, v, w, _ := barycentricCoordinates(i.Triangle.P1, i.Triangle.P2, i.Triangle.P3, i.Intersection)
	tex := Vector{
		u*i.Triangle.T1[0] + v*i.Triangle.T2[0] + w*i.Triangle.T3[0],
		u*i.Triangle.T1[1] + v*i.Triangle.T2[1] + w*i.Triangle.T3[1],
	}
	return tex
}

func (i *Intersection) hasBumpMap() bool {
	material := i.Triangle.Material
	if material.Texture != "" {
		if _, ok := BumpMapNormals[material.Texture]; ok {
			return true
		}
	}
	return false
}

func (i *Intersection) getBumpNormal() Vector {
	material := i.Triangle.Material
	if material.Texture != "" {
		// ok, we have the image. Let's calculate the pixel color;
		s := i.getTexCoords()
		// get image size
		if s[0] > 1 {
			s[0] -= math.Floor(s[0])
		}
		if s[0] < 0 {
			s[0] = math.Abs(s[0])
			s[0] = 1 - (s[0] - math.Floor(s[0]))
		}

		if s[1] > 1 {
			s[1] -= math.Floor(s[1])
		}

		if s[1] < 0 {
			s[1] = math.Abs(s[1])
			s[1] = 1 - (s[1] - math.Floor(s[1]))
		}
		s[1] = 1 - s[1]

		s[0] -= float64(int64(s[0]))
		s[1] -= float64(int64(s[1]))

		pixelX := int(float64(len(BumpMapNormals[material.Texture])) * s[0])
		pixelY := int(float64(len(BumpMapNormals[material.Texture][0])) * s[1])

		bump := BumpMapNormals[material.Texture][pixelX][pixelY]
		t := crossProduct(i.IntersectionNormal, Vector{0, -1, 0, 0})
		if vectorLength(t) < DIFF {
			t = crossProduct(i.IntersectionNormal, Vector{0, 0, 1, 0})
		}
		t = normalizeVector(t)
		bV := normalizeVector(crossProduct(i.IntersectionNormal, t))
		tbnMatrix := Matrix{t, bV, i.IntersectionNormal, Vector{0, 0, 0, 1}}
		wNormal := normalizeVector(vectorTransform(bump, tbnMatrix))
		wNormal[3] = 0
		return wNormal
	}
	return i.IntersectionNormal
}

func (i *Intersection) getNormal() {
	if !i.Hit {
		return
	}

	if i.Triangle.Smooth {
		u, v, w, _ := barycentricCoordinates(i.Triangle.P1, i.Triangle.P2, i.Triangle.P3, i.Intersection)

		N1 := i.Triangle.N1
		N2 := i.Triangle.N2
		N3 := i.Triangle.N3
		if !sameSideTest(N1, i.IntersectionNormal, 0) {
			N1 = scaleVector(N1, -1)
			N2 = scaleVector(N2, -1)
			N3 = scaleVector(N3, -1)
		}

		a := scaleVector(N1, u)
		b := scaleVector(N2, v)
		c := scaleVector(N3, w)
		normal := normalizeVector(addVector(addVector(a, b), c))

		i.IntersectionNormal = normal
	}
	if i.hasBumpMap() && GlobalConfig.RenderBumpMap {
		i.IntersectionNormal = i.getBumpNormal()
	}
}

func (i *Intersection) render(scene *Scene, depth int) Vector {
	if !i.Hit {
		if !hasEnvironmentMap {
			return GlobalConfig.TransparentColor
		}
		u := math.Atan2(i.RayDir[0], i.RayDir[1])/(2*math.Pi) + 0.5
		v := i.RayDir[2]*0.5 + 0.5
		w := float64(len(EnvironmentMap)) - 1
		h := float64(len(EnvironmentMap[0])) - 1
		pixelX := int(w * u)
		pixelY := int(h - h*v)
		return EnvironmentMap[pixelX][pixelY]
	}
	if depth >= GlobalConfig.MaxReflectionDepth {
		return i.getColor()
	}

	// We use same samples for both color sampling as well as
	// global illumination calculation.
	samples := ambientSampling(scene, i)

	// Initial light to render
	light := Vector{}

	// Light that reaches intersection point without any obstacles
	if GlobalConfig.RenderLights {
		light = i.getDirectLight(scene, depth)
	}

	// Do we have occlusion? If so, keep in mind that, we are not actually doing a real
	// global illumination sampling as it is way too expensive _for now_
	// Instead, we are taking a short-cut that modern games also do, an idea by CryTek I suppose?
	// We are doing an ambient occlusion
	if GlobalConfig.RenderOcclusion {
		aRate := ambientLightCalc(scene, i, samples, GlobalConfig.SamplerLimit)
		aRate *= GlobalConfig.OcclusionRate

		// Add ambient light to direct light.
		// In a perfect world, we should first calculate the lights then do the occlusion
		// but that is time consuming, so we again, cheat by assumptions.
		light = Vector{
			light[0] + aRate,
			light[1] + aRate,
			light[2] + aRate,
			1,
		}
	}

	// Get color
	color := i.getColor()

	if GlobalConfig.RenderAmbientColors {
		// Get ambient colors and apply to existing color
		aColor := ambientColor(scene, i, samples, GlobalConfig.SamplerLimit)
		color = Vector{
			(color[0] * (1.0 - GlobalConfig.AmbientColorSharingRatio)) + (aColor[0] * GlobalConfig.AmbientColorSharingRatio),
			(color[1] * (1.0 - GlobalConfig.AmbientColorSharingRatio)) + (aColor[1] * GlobalConfig.AmbientColorSharingRatio),
			(color[2] * (1.0 - GlobalConfig.AmbientColorSharingRatio)) + (aColor[2] * GlobalConfig.AmbientColorSharingRatio),
			1,
		}
		color = limitVector(color, 1.0)
	}

	// Do we have any transparency?
	pAlpha := 1.0
	if i.Dist < 0 {
		pAlpha = 0
	}

	color = Vector{
		color[0] * light[0],
		color[1] * light[1],
		color[2] * light[2],
		pAlpha,
	}
	dirs := make([]Vector, 0, int(math.Floor(i.Triangle.Material.Roughness*10)))

	// When light is too shiny, we have to limit color to white as it can't exceed white.
	color = limitVector(color, 1)

	// END OF MAIN RENDERING OF THE INTERSECTION
	// NOW WE DO THE TRACING PART

	// Do we have a glossy (metalic) material or a glass / transmissive material?
	if i.Triangle.Material.Glossiness > 0 || i.Triangle.Material.Transmission > 0 {
		if i.Triangle.Material.Roughness == 0 {
			// If we have a roughness, it means we need to sample intersection color from multiple directions to give
			// it the roughness it needs.
			dirs = append(dirs, i.IntersectionNormal)
		} else {
			numNormals := int(math.Floor(i.Triangle.Material.Roughness * 10))
			if numNormals > 0 {
				dirSamples := createSamples(i.IntersectionNormal, numNormals, 1-i.Triangle.Material.Roughness)
				dirs = append(dirs, dirSamples...)
			}
		}
	}

	if i.Triangle.Material.Glossiness > 0 && GlobalConfig.RenderReflections {
		// Do the reflection!
		collColor := Vector{}
		colChan := make(chan Vector, len(dirs))
		// Sample from reflected directions
		for m := range dirs {
			go func(scene *Scene, intersection *Intersection, dir Vector, depth int, colChan chan Vector) {
				dir = reflectVector(intersection.RayDir, dir)
				target := raycastSceneIntersect(scene, intersection.Intersection, dir)
				colChan <- target.render(scene, depth)
			}(scene, i, dirs[m], depth+1, colChan)
		}
		for m := 0; m < len(dirs); m++ {
			targetColor := <-colChan
			collColor = addVector(collColor, targetColor)
		}
		collColor = scaleVector(collColor, 1.0/float64(len(dirs)))

		color = Vector{
			color[0]*(1-i.Triangle.Material.Glossiness) + collColor[0]*i.Triangle.Material.Glossiness,
			color[1]*(1-i.Triangle.Material.Glossiness) + collColor[1]*i.Triangle.Material.Glossiness,
			color[2]*(1-i.Triangle.Material.Glossiness) + collColor[2]*i.Triangle.Material.Glossiness,
			1,
		}
	}
	if i.Triangle.Material.Transmission > 0 && GlobalConfig.RenderRefractions {
		// Do the refraction!
		collColor := Vector{}
		colChan := make(chan Vector, len(dirs))
		for m := range dirs {
			go func(scene *Scene, intersection *Intersection, dir Vector, depth int, colChan chan Vector) {
				dir = refractVector(intersection.RayDir, intersection.IntersectionNormal, intersection.Triangle.Material.IndexOfRefraction)
				target := raycastSceneIntersect(scene, intersection.Intersection, dir)
				colChan <- target.render(scene, depth)
			}(scene, i, dirs[m], depth+1, colChan)
		}
		for m := 0; m < len(dirs); m++ {
			targetColor := <-colChan
			collColor = addVector(collColor, targetColor)
		}
		collColor = scaleVector(collColor, 1.0/float64(len(dirs)))
		trans := i.Triangle.Material.Transmission * (1 - i.Triangle.Material.Roughness)

		color = Vector{
			color[0]*(1-trans) + collColor[0]*trans,
			color[1]*(1-trans) + collColor[1]*trans,
			color[2]*(1-trans) + collColor[2]*trans,
			1,
		}
	}

	return color
}

func (i *Intersection) getDirectLight(scene *Scene, depth int) Vector {
	return calculateTotalLight(scene, i, 0)
}

func (i *Intersection) getColor() Vector {
	if !GlobalConfig.RenderColors {
		return Vector{
			1, 1, 1, 1,
		}
	}

	material := i.Triangle.Material
	result := material.Color
	if material.Texture != "" {
		if _, ok := Images[material.Texture]; ok {
			// ok, we have the image. Let's calculate the pixel color;
			s := i.getTexCoords()
			// get image size

			if s[0] > 1 {
				s[0] -= math.Floor(s[0])
			}
			if s[0] < 0 {
				s[0] = math.Abs(s[0])
				s[0] = 1 - (s[0] - math.Floor(s[0]))
			}

			if s[1] > 1 {
				s[1] -= math.Floor(s[1])
			}

			if s[1] < 0 {
				s[1] = math.Abs(s[1])
				s[1] = 1 - (s[1] - math.Floor(s[1]))
			}
			s[1] = 1 - s[1]

			s[0] -= float64(int64(s[0]))
			s[1] -= float64(int64(s[1]))

			pixelX := int(float64(len(Images[material.Texture])) * s[0])
			pixelY := int(float64(len(Images[material.Texture][0])) * s[1])
			result = Images[material.Texture][pixelX][pixelY]
		}
	}
	return result
}
