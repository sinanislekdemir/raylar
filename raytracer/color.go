package raytracer

import "math"

func calculateColor(scene *Scene, intersection IntersectionTriangle) (result Vector) {
	if !intersection.Hit {
		return Vector{
			0, 0, 0, 1,
		}
	}
	material := intersection.Triangle.Material
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

			return Vector{
				float64(r) / 255,
				float64(g) / 255,
				float64(b) / 255,
				float64(a) / 255,
			}
		}
	}
	return material.Color
}
