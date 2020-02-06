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
		return Vector{
			scene.Config.Exposure * light.Color[0] * intersection.Triangle.Material.LightStrength,
			scene.Config.Exposure * light.Color[1] * intersection.Triangle.Material.LightStrength,
			scene.Config.Exposure * light.Color[2] * intersection.Triangle.Material.LightStrength,
			1,
		}
	}

	rayDir := normalizeVector(subVector(intersection.Intersection, light.Position))
	rayStart := light.Position
	rayLength := vectorDistance(intersection.Intersection, light.Position)

	intensity := (1 / (rayLength * rayLength)) * scene.Config.Exposure
	intensity *= dotP * light.LightStrength

	if intersection.Triangle.Material.LightStrength > 0 {
		intensity = intersection.Triangle.Material.LightStrength * scene.Config.Exposure
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
	// results := make([]Vector, len(scene.Lights))
	if (!intersection.Hit) || (depth >= scene.Config.MaxReflectionDepth) {
		return
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

	result[3] = 1

	if intersection.Triangle.Material.Glossiness > 0 && scene.Config.RenderReflections {
		rayStart := intersection.Intersection
		rayDir := reflectVector(intersection.RayDir, intersection.IntersectionNormal)
		reflection := raycastSceneIntersect(scene, rayStart, rayDir)
		reflectionLight := calculateTotalLight(scene, &reflection, depth+1)
		if !intersection.Hit {
			reflectionLight = Vector{}
		}
		result = Vector{
			result[0]*(1.0-intersection.Triangle.Material.Glossiness) + (reflectionLight[0] * intersection.Triangle.Material.Glossiness),
			result[1]*(1.0-intersection.Triangle.Material.Glossiness) + (reflectionLight[1] * intersection.Triangle.Material.Glossiness),
			result[2]*(1.0-intersection.Triangle.Material.Glossiness) + (reflectionLight[2] * intersection.Triangle.Material.Glossiness),
			result[3]*(1.0-intersection.Triangle.Material.Glossiness) + (reflectionLight[3] * intersection.Triangle.Material.Glossiness),
		}
	}

	if intersection.Triangle.Material.Transmission > 0 && scene.Config.RenderRefractions {
		rayStart := intersection.Intersection
		rayDir := refractVector(intersection.RayDir, intersection.IntersectionNormal, intersection.Triangle.Material.IndexOfRefraction)
		refraction := raycastSceneIntersect(scene, rayStart, rayDir)
		refractionLight := calculateTotalLight(scene, &refraction, depth+1)
		if !intersection.Hit {
			refractionLight = Vector{}
		}
		result = Vector{
			result[0]*(1.0-intersection.Triangle.Material.Transmission) + (refractionLight[0] * intersection.Triangle.Material.Transmission),
			result[1]*(1.0-intersection.Triangle.Material.Transmission) + (refractionLight[1] * intersection.Triangle.Material.Transmission),
			result[2]*(1.0-intersection.Triangle.Material.Transmission) + (refractionLight[2] * intersection.Triangle.Material.Transmission),
			result[3]*(1.0-intersection.Triangle.Material.Transmission) + (refractionLight[3] * intersection.Triangle.Material.Transmission),
		}
	}

	if scene.Config.CausticsThreshold > 0 && intersection.Triangle.PhotonMap != nil {
		for c := range intersection.Triangle.PhotonMap {
			if vectorDistance(intersection.Intersection, c) < scene.Config.CausticsThreshold {
				result = addVector(result, intersection.Triangle.PhotonMap[c])
			}
		}
	}

	return result
}
