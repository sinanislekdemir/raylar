package raytracer

import "math"

func calculateReflectionColor(scene *Scene, intersection IntersectionTriangle, depth int) (result Vector) {
	if !intersection.Hit {
		return Vector{0, 0, 0, -1}
	}
	bounceDir := reflectVector(intersection.RayDir, intersection.IntersectionNormal)
	bounceStart := intersection.Intersection
	reflection := raycastSceneIntersect(scene, bounceStart, bounceDir)
	if !reflection.Hit {
		return Vector{1, 1, 1, 1}
	}
	return calculateColor(scene, reflection, depth)
}

func calculateColor(scene *Scene, intersection IntersectionTriangle, depth int) (result Vector) {
	if !intersection.Hit {
		return Vector{
			0, 0, 0, 1,
		}
	}

	material := intersection.Triangle.Material
	result = material.Color
	if material.Texture != "" {
		if img, ok := scene.ImageMap[material.Texture]; ok {
			// ok, we have the image. Let's calculate the pixel color;
			s := intersection.getTexCoords()
			// get image size
			imgBounds := img.Bounds().Max

			if s[0] > 1 {
				s[0] = s[0] - math.Floor(s[0])
			}
			if s[0] < 0 {
				s[0] = math.Abs(s[0])
				s[0] = 1 - (s[0] - math.Floor(s[0]))
			}

			if s[1] > 1 {
				s[1] = s[1] - math.Floor(s[1])
			}
			if s[1] < 0 {
				s[1] = math.Abs(s[1])
				s[1] = 1 - (s[1] - math.Floor(s[1]))
			}

			s[0] = s[0] - float64(int64(s[0]))
			s[1] = s[1] - float64(int64(s[1]))

			pixelX := int(float64(imgBounds.X) * s[0])
			pixelY := int(float64(imgBounds.Y) * s[1])
			r, g, b, a := img.At(pixelX, pixelY).RGBA()
			r, g, b, a = r>>8, g>>8, b>>8, a>>8

			result = Vector{
				float64(r) / 255,
				float64(g) / 255,
				float64(b) / 255,
				float64(a) / 255,
			}
		}
	}
	if material.Glossiness > 0 && scene.Config.RenderReflections && depth < scene.Config.MaxReflectionDepth {
		reflectColor := calculateReflectionColor(scene, intersection, depth+1)
		result = Vector{
			result[0]*(1.0-material.Glossiness) + (reflectColor[0] * material.Glossiness),
			result[1]*(1.0-material.Glossiness) + (reflectColor[1] * material.Glossiness),
			result[2]*(1.0-material.Glossiness) + (reflectColor[2] * material.Glossiness),
			result[3]*(1.0-material.Glossiness) + (reflectColor[3] * material.Glossiness),
		}
	}
	return result
}
