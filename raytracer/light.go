package raytracer

/*
Light related methods
*/

import (
	"math"
	"reflect"
)

// Calculate light for given light source.
// Result will be used to calculate "avarage" of the pixel color
func calculateLight(scene *Scene, intersection IntersectionTriangle, light Light, depth int) (result Vector) {
	var shortestIntersection IntersectionTriangle

	if !intersection.Hit {
		return
	}

	if intersection.Triangle.Material.Light {
		return Vector{1, 1, 1, 1}
	}

	rayDir := normalizeVector(subVector(intersection.Intersection, light.Position))
	rayStart := light.Position
	rayLength := vectorDistance(intersection.Intersection, light.Position)

	hlimit := scene.Config.LightHardLimit

	if light.LightStrength > 0 {
		hlimit = light.LightStrength
	}

	if rayLength >= hlimit {
		return
	}

	intensity := 1 - (rayLength / hlimit)

	l1 := normalizeVector(subVector(light.Position, intersection.Intersection))
	l2 := intersection.IntersectionNormal

	dotP := dot(l2, l1)
	if dotP < 0 {
		intensity = 0
	} else {
		intensity *= dotP
	}

	if intersection.Triangle.Material.LightStrength > 0 {
		intensity = 1
	}

	shortestIntersection = raycastSceneIntersect(scene, rayStart, rayDir)
	check := reflect.DeepEqual(shortestIntersection.Triangle, intersection.Triangle)
	s := math.Abs(rayLength - shortestIntersection.Dist)

	if check || s < DIFF {
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

func calculateTotalLight(scene *Scene, intersection IntersectionTriangle, depth int) (result Vector) {
	results := make([]Vector, len(scene.Lights))
	if (!intersection.Hit) || (depth >= scene.Config.MaxReflectionDepth) {
		return
	}

	totalIntensity := float64(len(scene.Lights))
	for i, light := range scene.Lights {
		results[i] = calculateLight(scene, intersection, light, depth)
	}

	if totalIntensity > 1 {
		intensityScale := 1.0 / totalIntensity
		for i := range scene.Lights {
			results[i] = Vector{
				results[i][0] * (results[i][3] * intensityScale),
				results[i][1] * (results[i][3] * intensityScale),
				results[i][2] * (results[i][3] * intensityScale),
				(results[i][3] * intensityScale),
			}
		}
	}

	result = Vector{}
	for i := range scene.Lights {
		result = addVector(result, results[i])
	}

	if intersection.Triangle.Material.Glossiness > 0 && scene.Config.RenderReflections {
		rayStart := intersection.Intersection
		rayDir := reflectVector(intersection.RayDir, intersection.IntersectionNormal)
		reflection := raycastSceneIntersect(scene, rayStart, rayDir)
		reflectionLight := calculateTotalLight(scene, reflection, depth+1)
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

	return result
}
