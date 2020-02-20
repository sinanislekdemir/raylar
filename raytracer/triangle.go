package raytracer

import "math"

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
}

// IntersectionTriangle defines the ratcast triangle intersection result
type IntersectionTriangle struct {
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

func (t *Triangle) getBoundingBox() *BoundingBox {
	result := BoundingBox{}
	result.MinExtend = t.P1
	result.MaxExtend = t.P1
	for i := 0; i < 3; i++ {
		if t.P2[i] < result.MinExtend[i] {
			result.MinExtend[i] = t.P2[i]
		}
		if t.P2[i] > result.MaxExtend[i] {
			result.MaxExtend[i] = t.P2[i]
		}
		if t.P3[i] < result.MinExtend[i] {
			result.MinExtend[i] = t.P3[i]
		}
		if t.P3[i] > result.MaxExtend[i] {
			result.MaxExtend[i] = t.P3[i]
		}
	}
	return &result
}

func (i *IntersectionTriangle) getTexCoords() Vector {
	u, v, w, _ := barycentricCoordinates(i.Triangle.P1, i.Triangle.P2, i.Triangle.P3, i.Intersection)
	tex := Vector{
		u*i.Triangle.T1[0] + v*i.Triangle.T2[0] + w*i.Triangle.T3[0],
		u*i.Triangle.T1[1] + v*i.Triangle.T2[1] + w*i.Triangle.T3[1],
	}
	return tex
}

func (i *IntersectionTriangle) getNormal() {
	if !i.Hit {
		return
	}
	if !GlobalConfig.SmoothShading {
		return
	}
	u, v, w, _ := barycentricCoordinates(i.Triangle.P1, i.Triangle.P2, i.Triangle.P3, i.Intersection)

	a := scaleVector(i.Triangle.N1, u)
	b := scaleVector(i.Triangle.N2, v)
	c := scaleVector(i.Triangle.N3, w)
	normal := normalizeVector(addVector(addVector(a, b), c))
	if !sameSideTest(normal, i.IntersectionNormal) {
		normal = scaleVector(normal, -1)
	}
	i.IntersectionNormal = normal
}

func (i *IntersectionTriangle) render(scene *Scene, depth int) Vector {
	if !i.Hit {
		return Vector{}
	}
	if depth >= GlobalConfig.MaxReflectionDepth {
		return Vector{}
	}

	samples := ambientSampling(scene, i)

	directLight := Vector{0, 0, 0, 1}
	if GlobalConfig.RenderLights {
		directLight = i.getDirectLight(scene, depth)
	}

	ambientRate := GlobalConfig.OcclusionRate
	//directLight = limitVector(directLight, ambientRate)

	light := directLight
	if directLight[0] == 0 && directLight[1] == 0 && directLight[2] == 0 {
		ambient := ambientLightCalc(scene, i, samples, GlobalConfig.SamplerLimit)
		light = Vector{
			ambient * ambientRate,
			ambient * ambientRate,
			ambient * ambientRate,
			1,
		}
	}

	color := i.getColor(scene, depth)

	if GlobalConfig.RenderAmbientColors {
		aColor := ambientColor(scene, i, samples, GlobalConfig.SamplerLimit)
		color = Vector{
			(color[0] * (1.0 - GlobalConfig.AmbientColorSharingRatio)) + (aColor[0] * GlobalConfig.AmbientColorSharingRatio),
			(color[1] * (1.0 - GlobalConfig.AmbientColorSharingRatio)) + (aColor[1] * GlobalConfig.AmbientColorSharingRatio),
			(color[2] * (1.0 - GlobalConfig.AmbientColorSharingRatio)) + (aColor[2] * GlobalConfig.AmbientColorSharingRatio),
			1,
		}
		color = limitVector(color, 1.0)
	}

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
	color = limitVector(color, 1)

	if i.Triangle.Material.Glossiness > 0 {
		dir := reflectVector(i.RayDir, i.IntersectionNormal)
		target := raycastSceneIntersect(scene, i.Intersection, dir)
		targetColor := target.render(scene, depth+1)
		color = Vector{
			color[0]*(1-i.Triangle.Material.Glossiness) + targetColor[0]*i.Triangle.Material.Glossiness,
			color[1]*(1-i.Triangle.Material.Glossiness) + targetColor[1]*i.Triangle.Material.Glossiness,
			color[2]*(1-i.Triangle.Material.Glossiness) + targetColor[2]*i.Triangle.Material.Glossiness,
			1,
		}
	}
	if i.Triangle.Material.Transmission > 0 {
		dir := refractVector(i.RayDir, i.IntersectionNormal, i.Triangle.Material.IndexOfRefraction)
		target := raycastSceneIntersect(scene, i.Intersection, dir)
		targetColor := target.render(scene, depth+1)
		color = Vector{
			color[0]*(1-i.Triangle.Material.Transmission) + targetColor[0]*i.Triangle.Material.Transmission,
			color[1]*(1-i.Triangle.Material.Transmission) + targetColor[1]*i.Triangle.Material.Transmission,
			color[2]*(1-i.Triangle.Material.Transmission) + targetColor[2]*i.Triangle.Material.Transmission,
			1,
		}
	}

	return color
}

func (i *IntersectionTriangle) getDirectLight(scene *Scene, depth int) Vector {
	return calculateTotalLight(scene, i, 0)
}

func (i *IntersectionTriangle) getColor(scene *Scene, depth int) Vector {
	if !i.Hit {
		return Vector{
			0, 0, 0, 1,
		}
	}
	if !GlobalConfig.RenderColors {
		return Vector{
			1, 1, 1, 1,
		}
	}

	material := i.Triangle.Material
	result := material.Color
	if material.Texture != "" {
		if img, ok := scene.ImageMap[material.Texture]; ok {
			// ok, we have the image. Let's calculate the pixel color;
			s := i.getTexCoords()
			// get image size
			imgBounds := img.Bounds().Max

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

			pixelX := int(float64(imgBounds.X) * s[0])
			pixelY := int(float64(imgBounds.Y) * s[1])
			r, g, b, a := img.At(pixelX, pixelY).RGBA()
			r, g, b, a = r>>8, g>>8, b>>8, a>>8

			result = Vector{
				float64(r) / 255,
				float64(g) / 255,
				float64(b) / 255,
				float64(a) / 255,
			}
		}
	}
	return result
}
