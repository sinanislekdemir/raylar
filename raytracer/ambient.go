package raytracer

// Calculate light reflecting from other objects
func ambientLightCalc(scene *Scene, intersection *Intersection, samples []Intersection, totalDirs int) (result float64) {
	totalHits := 0.0
	rad := scene.ShortRadius
	if GlobalConfig.AmbientRadius > 0 {
		rad = GlobalConfig.AmbientRadius
	}
	for i := 0; i < len(samples); i++ {
		if samples[i].Dist < rad {
			totalHits++
		}
	}
	result = 1.0 - (totalHits / float64(totalDirs))

	return result
}

func ambientColor(scene *Scene, intersection *Intersection, samples []Intersection, totalDirs int) (result Vector) {
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

func ambientSampling(scene *Scene, intersection *Intersection) []Intersection {
	sampleDirs := createSamples(intersection.IntersectionNormal, GlobalConfig.SamplerLimit, 0)
	hitChannel := make(chan Intersection, len(sampleDirs))
	for i := range sampleDirs {
		go func(scene *Scene, intersection *Intersection, dir Vector, channel chan Intersection) {
			hit := raycastSceneIntersect(scene, intersection.Intersection, dir)
			channel <- hit
		}(scene, intersection, sampleDirs[i], hitChannel)
	}
	samples := make([]Intersection, 0, len(sampleDirs))
	for i := 0; i < len(sampleDirs); i++ {
		hit := <-hitChannel
		if hit.Hit && hit.Triangle.id != intersection.Triangle.id {
			samples = append(samples, hit)
		}
	}
	return samples
}
