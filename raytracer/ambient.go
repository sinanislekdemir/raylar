package raytracer

// Calculate light reflecting from other objects
func ambientLightCalc(scene *Scene, intersection *IntersectionTriangle, samples []IntersectionTriangle, totalDirs int) (result float64) {
	totalHits := 0.0
	rad := scene.ShortRadius
	if scene.Config.AmbientRadius > 0 {
		rad = scene.Config.AmbientRadius
	}
	for i := 0; i < len(samples); i++ {
		if samples[i].Hit && samples[i].Dist < rad {
			totalHits++
		}
	}
	result = 1.0 - (totalHits / float64(totalDirs))

	if intersection.Triangle != nil && intersection.Triangle.Material.Glossiness > 0 && scene.Config.RenderReflections {
		rayStart := intersection.Intersection
		rayDir := reflectVector(intersection.RayDir, intersection.IntersectionNormal)
		reflection := raycastSceneIntersect(scene, rayStart, rayDir)
		if intersection.Hit {
			ambient := ambientLightCalc(scene, &reflection, samples, totalDirs)
			result = (result*(1-intersection.Triangle.Material.Glossiness) + (ambient * intersection.Triangle.Material.Glossiness))
		}
	}
	return result
}

func ambientColor(scene *Scene, intersection *IntersectionTriangle, samples []IntersectionTriangle, totalDirs int) (result Vector) {
	if !intersection.Hit {
		return Vector{}
	}

	totalHits := 0.0
	totalColor := Vector{}
	for i := 0; i < len(samples); i++ {
		color := samples[i].getColor(scene, 0)
		if vectorLength(color) < DIFF {
			continue
		}
		totalHits++
		totalColor = addVector(totalColor, color)
	}
	if totalHits == 0 {
		return Vector{}
	}

	return scaleVector(totalColor, 1.0/totalHits)
}

func ambientSampling(scene *Scene, intersection *IntersectionTriangle) []IntersectionTriangle {
	sampleDirs := createSamples(intersection.IntersectionNormal, scene.Config.SamplerLimit)
	hitChannel := make(chan IntersectionTriangle, len(sampleDirs))
	for i := range sampleDirs {
		go func(scene *Scene, intersection *IntersectionTriangle, dir Vector, channel chan IntersectionTriangle) {
			hit := raycastSceneIntersect(scene, intersection.Intersection, dir)
			channel <- hit
		}(scene, intersection, sampleDirs[i], hitChannel)
	}
	samples := make([]IntersectionTriangle, 0, len(sampleDirs))
	for i := 0; i < len(sampleDirs); i++ {
		hit := <-hitChannel
		if hit.Hit && hit.Triangle.id != intersection.Triangle.id {
			samples = append(samples, hit)
		}
	}
	return samples
}
