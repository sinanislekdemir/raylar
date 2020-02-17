package raytracer

import "math"

// Photon information to follow
type Photon struct {
	Location  Vector // You can never be sure!
	Direction Vector
	Color     Vector
	Intensity float64
}

/// trace a photon's path
func tracePhoton(scene *Scene, photon *Photon, depth int) {
	if photon.Intensity < DIFF {
		return
	}
	if depth > scene.Config.MaxReflectionDepth {
		return
	}
	hit := raycastSceneIntersect(scene, photon.Location, photon.Direction)
	if !hit.Hit {
		return
	}

	rayLength := vectorDistance(hit.Intersection, photon.Location)
	dotP := dot(hit.IntersectionNormal, scaleVector(photon.Direction, -1))
	if dotP < 0 {
		return
	}
	inv_dist_sqr := 1.0 / (rayLength * rayLength)
	photon.Intensity *= inv_dist_sqr
	photon.Intensity *= dotP

	trace := scaleVector(photon.Color, photon.Intensity)
	if math.IsNaN(trace[0]) {
		return
	}

	if hit.Triangle.Material.Glossiness == 0 && hit.Triangle.Material.Transmission == 0 {
		if hit.Triangle.Photons == nil {
			hit.Triangle.Photons = make([]Photon, 0)
		}
		hit.Triangle.Photons = append(hit.Triangle.Photons, Photon{
			Location:  hit.Intersection,
			Color:     trace,
			Direction: photon.Direction,
		})
	}

	if hit.Triangle.Material.Glossiness > 0 {
		reflect := reflectVector(photon.Direction, hit.IntersectionNormal)
		reflectedPhoton := Photon{
			Location:  hit.Intersection,
			Direction: reflect,
			Color:     photon.Color,
			Intensity: photon.Intensity * hit.Triangle.Material.Glossiness,
		}
		tracePhoton(scene, &reflectedPhoton, depth+1)
	}
	if hit.Triangle.Material.Transmission > 0 {
		refract := refractVector(photon.Direction, hit.IntersectionNormal, hit.Triangle.Material.IndexOfRefraction)
		refractedPhoton := Photon{
			Location:  hit.Intersection,
			Direction: refract,
			Color:     photon.Color,
			Intensity: photon.Intensity * hit.Triangle.Material.Transmission,
		}
		tracePhoton(scene, &refractedPhoton, depth+1)
	}
}
