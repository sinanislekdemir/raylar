package raytracer

func calculateReflectionColor(scene *Scene, intersection *IntersectionTriangle, depth int) (result Vector) {
	if !intersection.Hit {
		return Vector{0, 0, 0, -1}
	}
	bounceDir := reflectVector(intersection.RayDir, intersection.IntersectionNormal)
	bounceStart := intersection.Intersection
	reflection := raycastSceneIntersect(scene, bounceStart, bounceDir)
	if !reflection.Hit {
		return Vector{1, 1, 1, 1}
	}
	return reflection.getColor(scene, depth)
}

func calculateRefractionColor(scene *Scene, intersection *IntersectionTriangle, depth int) (result Vector) {
	if !intersection.Hit || intersection.Triangle == nil {
		return Vector{0, 0, 0, -1}
	}
	refractionDir := refractVector(intersection.RayDir, intersection.IntersectionNormal, intersection.Triangle.Material.IndexOfRefraction)
	bounceStart := intersection.Intersection
	refraction := raycastSceneIntersect(scene, bounceStart, refractionDir)
	if !refraction.Hit {
		return Vector{}
	}
	return refraction.getColor(scene, depth)
}
