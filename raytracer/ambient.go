package raytracer

// Calculate light reflecting from other objects
func ambientLightCalc(scene *Scene, intersection IntersectionTriangle) (result float64) {
	sampleDirs := createSamples(intersection.IntersectionNormal, scene.Config.SamplerLimit)
	hitChannel := make(chan IntersectionTriangle, len(sampleDirs))
	for i := range sampleDirs {
		go func(scene *Scene, intersection IntersectionTriangle, dir Vector, channel chan IntersectionTriangle) {
			hit := raycastSceneIntersect(scene, intersection.Intersection, dir)
			channel <- hit
		}(scene, intersection, sampleDirs[i], hitChannel)
	}
	totalHits := 0.0
	rad := scene.ShortRadius
	if scene.Config.AmbientRadius > 0 {
		rad = scene.Config.AmbientRadius
	}
	for i := 0; i < len(sampleDirs); i++ {
		hit := <-hitChannel
		if hit.Hit && hit.Dist < rad {
			totalHits++
		}
	}
	result = 1.0 - (totalHits / float64(len(sampleDirs)))
	if intersection.Triangle.Material.Glossiness > 0 && scene.Config.RenderReflections {
		rayStart := intersection.Intersection
		rayDir := reflectVector(intersection.RayDir, intersection.IntersectionNormal)
		reflection := raycastSceneIntersect(scene, rayStart, rayDir)
		if intersection.Hit {
			ambient := ambientLightCalc(scene, reflection)
			result = (result*(1-intersection.Triangle.Material.Glossiness) + (ambient * intersection.Triangle.Material.Glossiness))
		}
	}
	return result
}

func ambientColor(scene *Scene, intersection IntersectionTriangle) (result Vector) {
	if !intersection.Hit {
		return Vector{}
	}
	sampleDirs := createSamples(intersection.IntersectionNormal, scene.Config.SamplerLimit)
	hitChannel := make(chan Vector, len(sampleDirs))
	for i := range sampleDirs {
		go func(scene *Scene, intersection IntersectionTriangle, dir Vector, channel chan Vector) {
			hit := raycastSceneIntersect(scene, intersection.Intersection, dir)
			color := calculateColor(scene, hit, 0)
			channel <- color
		}(scene, intersection, sampleDirs[i], hitChannel)
	}
	totalHits := 0.0
	totalColor := Vector{}
	for i := 0; i < len(sampleDirs); i++ {
		hit := <-hitChannel
		if vectorLength(hit) < DIFF {
			continue
		}
		totalHits++
		totalColor = addVector(totalColor, hit)
	}
	if totalHits == 0 {
		return Vector{}
	}
	return scaleVector(totalColor, 1.0/float64(totalHits))
}
