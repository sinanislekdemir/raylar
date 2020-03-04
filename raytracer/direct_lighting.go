package raytracer

/*
Light related methods
*/

import (
	"math"
)

func calculateDirectionalLight(scene *Scene, intersection *Intersection, light *Light, depth int) (result Vector) {
	var shortestIntersection Intersection
	if !intersection.Hit {
		return
	}

	lightD := scaleVector(light.Direction, -1)

	dotP := dot(intersection.IntersectionNormal, lightD)
	if dotP < 0 {
		return
	}

	samples := sampleSphere(0.5, GlobalConfig.LightSampleCount)
	totalHits := 0.0
	totalLight := Vector{}

	for i := range samples {
		rayStart := scaleVector(lightD, 999999999999)
		rayStart = addVector(rayStart, intersection.Intersection)
		rayStart = addVector(rayStart, samples[i])
		dir := normalizeVector(subVector(intersection.Intersection, rayStart))

		shortestIntersection = raycastSceneIntersect(scene, rayStart, dir)
		if (shortestIntersection.Triangle != nil && shortestIntersection.Triangle.id == intersection.Triangle.id) || shortestIntersection.Dist < DIFF {
			if !sameSideTest(intersection.IntersectionNormal, shortestIntersection.IntersectionNormal, 0) {
				return
			}

			intensity := dotP * light.LightStrength
			intensity *= GlobalConfig.Exposure

			totalLight = addVector(totalLight, Vector{
				light.Color[0] * intensity,
				light.Color[1] * intensity,
				light.Color[2] * intensity,
				intensity,
			})
			totalHits += 1.0
		}

		// Let things pass if this is a regular glass
		if (shortestIntersection.Hit && shortestIntersection.Triangle != nil) && (shortestIntersection.Triangle.id != intersection.Triangle.id) && (shortestIntersection.Triangle.Material.Transmission > 0) && (!shortestIntersection.Triangle.Smooth) {
			col := shortestIntersection.getColor(scene, depth)
			lColor := Vector{
				light.Color[0] * col[0],
				light.Color[1] * col[1],
				light.Color[2] * col[2],
				1,
			}

			intensity := (1 / (shortestIntersection.Dist * shortestIntersection.Dist)) * GlobalConfig.Exposure
			intensity *= dotP * light.LightStrength * shortestIntersection.Triangle.Material.Transmission
			if intensity > DIFF {
				subLight := Light{
					Position:      shortestIntersection.Intersection,
					Color:         lColor,
					Active:        true,
					LightStrength: intensity,
				}
				return calculateDirectionalLight(scene, intersection, &subLight, depth)
			}
		}
	}
	if totalHits > 0 {
		return scaleVector(totalLight, totalHits/float64(GlobalConfig.LightSampleCount))
	}

	return
}

// Calculate light for given light source.
// Result will be used to calculate "avarage" of the pixel color
func calculateLight(scene *Scene, intersection *Intersection, light *Light, depth int) (result Vector) {
	if !intersection.Hit {
		return
	}

	l1 := normalizeVector(subVector(light.Position, intersection.Intersection))
	l2 := intersection.IntersectionNormal

	dotP := dot(l2, l1)
	if dotP < 0 {
		return
	}

	if intersection.Triangle.Material.Light {
		if intersection.Triangle.Material.LightStrength == 0 {
			intersection.Triangle.Material.LightStrength = light.LightStrength
		}
		return Vector{
			GlobalConfig.Exposure * light.Color[0] * intersection.Triangle.Material.LightStrength,
			GlobalConfig.Exposure * light.Color[1] * intersection.Triangle.Material.LightStrength,
			GlobalConfig.Exposure * light.Color[2] * intersection.Triangle.Material.LightStrength,
			1,
		}
	}

	rayDir := normalizeVector(subVector(intersection.Intersection, light.Position))
	rayLength := vectorDistance(intersection.Intersection, light.Position)

	shortestIntersection := raycastSceneIntersect(scene, light.Position, rayDir)
	s := math.Abs(rayLength - shortestIntersection.Dist)

	if (shortestIntersection.Triangle != nil && shortestIntersection.Triangle.id == intersection.Triangle.id) || s < DIFF {
		if !sameSideTest(intersection.IntersectionNormal, shortestIntersection.IntersectionNormal, 0) {
			return
		}

		intensity := (1 / (rayLength * rayLength)) * GlobalConfig.Exposure
		intensity *= dotP * light.LightStrength

		if intersection.Triangle.Material.LightStrength > 0 {
			intensity = intersection.Triangle.Material.LightStrength * GlobalConfig.Exposure
		}

		return Vector{
			light.Color[0] * intensity,
			light.Color[1] * intensity,
			light.Color[2] * intensity,
			intensity,
		}
	}

	// Let things pass if this is a regular glass
	if (shortestIntersection.Hit && shortestIntersection.Triangle != nil) && (shortestIntersection.Triangle.id != intersection.Triangle.id) && (shortestIntersection.Triangle.Material.Transmission > 0) && (!shortestIntersection.Triangle.Smooth) {
		col := shortestIntersection.getColor(scene, depth)
		lColor := Vector{
			light.Color[0] * col[0],
			light.Color[1] * col[1],
			light.Color[2] * col[2],
			1,
		}

		intensity := (1 / (shortestIntersection.Dist * shortestIntersection.Dist)) * GlobalConfig.Exposure
		intensity *= dotP * light.LightStrength * shortestIntersection.Triangle.Material.Transmission
		if intensity > DIFF {
			subLight := Light{
				Position:      shortestIntersection.Intersection,
				Color:         lColor,
				Active:        true,
				LightStrength: intensity,
			}
			return calculateLight(scene, intersection, &subLight, depth)
		}
	}
	return
}

func calculateTotalLight(scene *Scene, intersection *Intersection, depth int) (result Vector) {
	if (!intersection.Hit) || (depth >= GlobalConfig.MaxReflectionDepth) {
		return
	}

	if intersection.Triangle.Material.Light {
		c := scaleVector(intersection.Triangle.Material.Color, intersection.Triangle.Material.LightStrength)
		return c
	}

	lightChan := make(chan Vector, len(scene.Lights))

	for i := range scene.Lights {
		go func(scene *Scene, intersection *Intersection, light *Light, depth int, lightChan chan Vector) {
			if light.Directional {
				lightChan <- calculateDirectionalLight(scene, intersection, light, depth)
			} else {
				lightChan <- calculateLight(scene, intersection, light, depth)
			}
		}(scene, intersection, &scene.Lights[i], depth, lightChan)
	}

	result = Vector{}
	for i := 0; i < len(scene.Lights); i++ {
		light := <-lightChan
		if light[3] > 0 {
			result = addVector(result, light)
		}
	}

	if GlobalConfig.PhotonSpacing > 0 && GlobalConfig.RenderCaustics {
		if intersection.Triangle.Photons != nil && len(intersection.Triangle.Photons) > 0 {
			for i := range intersection.Triangle.Photons {
				if vectorDistance(intersection.Triangle.Photons[i].Location, intersection.Intersection) < GlobalConfig.PhotonSpacing {
					c := scaleVector(intersection.Triangle.Photons[i].Color, GlobalConfig.Exposure)
					result = addVector(result, c)
				}
			}
		}
	}

	return result
}
