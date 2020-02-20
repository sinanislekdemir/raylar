package raytracer

/*
Light related methods
*/

import (
	"math"
)

// Calculate light for given light source.
// Result will be used to calculate "avarage" of the pixel color
func calculateLight(scene *Scene, intersection *IntersectionTriangle, light *Light, depth int) (result Vector) {
	var shortestIntersection IntersectionTriangle

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
	rayStart := light.Position
	rayLength := vectorDistance(intersection.Intersection, light.Position)

	intensity := (1 / (rayLength * rayLength)) * GlobalConfig.Exposure
	intensity *= dotP * light.LightStrength

	if intersection.Triangle.Material.LightStrength > 0 {
		intensity = intersection.Triangle.Material.LightStrength * GlobalConfig.Exposure
	}

	shortestIntersection = raycastSceneIntersect(scene, rayStart, rayDir)
	s := math.Abs(rayLength - shortestIntersection.Dist)

	if (shortestIntersection.Triangle != nil && shortestIntersection.Triangle.id == intersection.Triangle.id) || s < DIFF {
		if !sameSideTest(intersection.IntersectionNormal, shortestIntersection.IntersectionNormal) {
			return
		}
		return Vector{
			light.Color[0] * intensity,
			light.Color[1] * intensity,
			light.Color[2] * intensity,
			intensity,
		}
	}
	return
}

func calculateTotalLight(scene *Scene, intersection *IntersectionTriangle, depth int) (result Vector) {
	if (!intersection.Hit) || (depth >= GlobalConfig.MaxReflectionDepth) {
		return
	}

	if intersection.Triangle.Material.Light {
		c := scaleVector(intersection.Triangle.Material.Color, intersection.Triangle.Material.LightStrength)
		return c
	}

	lightChan := make(chan Vector, len(scene.Lights))

	for i := range scene.Lights {
		go func(scene *Scene, intersection *IntersectionTriangle, light *Light, depth int, lightChan chan Vector) {
			lightChan <- calculateLight(scene, intersection, light, depth)
		}(scene, intersection, &scene.Lights[i], depth, lightChan)
	}

	result = Vector{}
	for i := 0; i < len(scene.Lights); i++ {
		light := <-lightChan
		if light[3] > 0 {
			result = addVector(result, light)
		}
	}

	if GlobalConfig.PhotonSpacing > 0 {
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
