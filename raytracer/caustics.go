package raytracer

// buildPhotonMap builds a photon map :) Obvious right?
// but here is the catch, it only builds the map for caustics. (Refractions and reflections from
// each point of light) So basically direct lights and ambient occlusion do not happen here
func buildPhotonMap(scene *Scene) {
	for i := range scene.Lights {
		// This is just a single step reflection calculation.
		// so there is still a long way to improve!
		samples := sampleAllDirections(scene.Config.CausticsSamplerLimit)
		intersectChannel := make(chan IntersectionTriangle, len(samples))
		for k := range samples {
			go func(scene *Scene, position, ray Vector, c chan IntersectionTriangle) {
				result := raycastSceneIntersect(scene, position, ray)
				c <- result
			}(scene, scene.Lights[i].Position, samples[k], intersectChannel)
		}

		// Analyse hits
		caustics := make([]IntersectionTriangle, 0, len(samples))

		for k := 0; k < len(samples); k++ {
			intersect := <-intersectChannel
			if !intersect.Hit {
				continue
			}
			if intersect.Dist < DIFF {
				intersect.Hit = false
				continue
			}
			if intersect.Triangle == nil {
				continue
			}
			if intersect.Triangle.Material.Glossiness > 0 || intersect.Triangle.Material.Transmission > 0 {
				caustics = append(caustics, intersect)
			}
		}

		// Now do reverse way check if intersection point is visible from the reflection point
		for k := range caustics {
			if caustics[k].Triangle == nil {
				continue
			}
			color := addVector(
				scaleVector(scene.Lights[i].Color, caustics[k].Triangle.Material.Glossiness),
				scaleVector(caustics[k].getColor(scene, 0), caustics[k].Triangle.Material.Glossiness),
			)
			light := Light{
				Position:      caustics[k].Intersection,
				Color:         color,
				LightStrength: (scene.Lights[i].LightStrength - caustics[k].Dist) / scene.Lights[i].LightStrength,
			}
			reflect := reflectVector(caustics[k].RayDir, caustics[k].IntersectionNormal)
			hit := raycastSceneIntersect(scene, caustics[k].Intersection, reflect)
			if hit.Triangle != nil && hit.Hit {
				scale := (light.LightStrength - hit.Dist) / light.LightStrength
				if hit.Triangle.PhotonMap == nil {
					hit.Triangle.PhotonMap = make(map[Vector]Vector)
				}
				hit.Triangle.PhotonMap[hit.Intersection] = scaleVector(light.Color, scale)
			}
		}

	}
}

// func searchReflections(scene *Scene, light *Light) Vector {
// 	// This is just a single step reflection calculation.
// 	// so there is still a long way to improve!
// 	samples := sampleAllDirections(scene.Config.SamplerLimit)
// 	intersectChannel := make(chan IntersectionTriangle, len(samples))
// 	for k := range samples {
// 		go func(scene *Scene, position, ray Vector, c chan IntersectionTriangle) {
// 			result := raycastSceneIntersect(scene, position, ray)
// 			c <- result
// 		}(scene, light.Position, samples[k], intersectChannel)
// 	}

// 	// Analyse hits
// 	caustics := make([]IntersectionTriangle, 0, len(samples))

// 	for i := 0; i < len(samples); i++ {
// 		intersect := <-intersectChannel
// 		if !intersect.Hit {
// 			continue
// 		}
// 		if intersect.Dist < DIFF {
// 			intersect.Hit = false
// 			continue
// 		}
// 		if intersect.Triangle.Material.Glossiness > 0 || intersect.Triangle.Material.Transmission > 0 {
// 			caustics = append(caustics, intersect)
// 		}
// 	}

// 	// Now do reverse way check if intersection point is visible from the reflection point
// 	reflections := Vector{}
// 	refCount := 0.0
// 	for i := range caustics {
// 		color := addVector(
// 			scaleVector(light.Color, caustics[i].Triangle.Material.Glossiness),
// 			scaleVector(caustics[i].getColor(scene, 0), caustics[i].Triangle.Material.Glossiness),
// 		)
// 		light := Light{
// 			Position:      caustics[i].Intersection,
// 			Color:         color,
// 			LightStrength: (light.LightStrength - caustics[i].Dist) / light.LightStrength,
// 		}
// 		reflect := calculateLight(scene, intersection, &light, 0)
// 		if reflect[3] > 0 {
// 			reflections = addVector(reflections, reflect)
// 			refCount++
// 		}
// 	}

// 	if refCount > 0 {
// 		//		intensityScale := 1.0 / refCount
// 		// reflections = Vector{
// 		// 	reflections[0] * (reflections[3] * intensityScale),
// 		// 	reflections[1] * (reflections[3] * intensityScale),
// 		// 	reflections[2] * (reflections[3] * intensityScale),
// 		// 	(reflections[3] * intensityScale),
// 		// }
// 	}

// 	refractions := Vector{}
// 	for i := range caustics {
// 		refraction := refractVector(caustics[i].RayDir, caustics[i].IntersectionNormal, caustics[i].Triangle.Material.IndexOfRefraction)
// 		triangleTest := raycastSceneIntersect(scene, caustics[i].Intersection, refraction)
// 		if vectorDistance(triangleTest.Intersection, intersection.Intersection) < scene.Config.CausticsThreshold {
// 			color := addVector(
// 				scaleVector(light.Color, caustics[i].Triangle.Material.Transmission),
// 				scaleVector(caustics[i].getColor(scene, 0), caustics[i].Triangle.Material.Transmission),
// 			)
// 			color = scaleVector(color, (light.LightStrength-caustics[i].Dist)/light.LightStrength)
// 			refractions = addVector(refractions, color)
// 			refractions[3] = 1
// 		}
// 	}
// 	return addVector(reflections, refractions)
// }

// func causticLight(scene *Scene, intersection *IntersectionTriangle) (result Vector) {
// 	totalReflections := Vector{}
// 	refCount := 0.0
// 	for l := range scene.Lights {
// 		ref := searchReflections(scene, &scene.Lights[l], intersection)
// 		if math.IsNaN(ref[0]) || ref[3] == 0 {
// 			continue
// 		}
// 		totalReflections = addVector(totalReflections, ref)
// 		refCount++
// 	}
// 	if refCount == 0 {
// 		return Vector{}
// 	}
// 	intensityScale := 1.0 / refCount
// 	result = Vector{
// 		totalReflections[0] * (totalReflections[3] * intensityScale),
// 		totalReflections[1] * (totalReflections[3] * intensityScale),
// 		totalReflections[2] * (totalReflections[3] * intensityScale),
// 		totalReflections[3] * intensityScale,
// 	}
// 	for i := 0; i < 4; i++ {
// 		if math.IsNaN(result[i]) {
// 			return Vector{}
// 		}
// 	}
// 	return
// }

// func calculateTotalLight(scene *Scene, intersection *IntersectionTriangle, depth int) (result Vector)
