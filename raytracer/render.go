package raytracer

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
)

// Render -
func Render(scene *Scene, width, height int) error {
	viewMatrix := ViewMatrix(scene.Observers[0].Position, scene.Observers[0].Target, scene.Observers[0].Up)
	projectionMatrix := PerspectiveProjection(
		scene.Observers[0].Fov,
		scene.Observers[0].AspectRatio,
		scene.Observers[0].Near,
		scene.Observers[0].Far,
	)
	scene.DepthMap = make([][]float64, width)
	for i := 0; i < width; i++ {
		scene.DepthMap[i] = make([]float64, height)
	}

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// Colors are defined by Red, Green, Blue, Alpha uint8 values.
	// cyan := color.RGBA{100, 200, 200, 0xff}

	// Set color for each pixel.
	log.Printf("Eye: %+v\n", scene.Observers[0].Position)
	rayDirx := ScreenToWorld(150, 200, 400, 300, scene.Observers[0].Position, projectionMatrix, viewMatrix)
	log.Printf("RayDir %+v\n", rayDirx)
	log.Printf("view matrix %+v\n", viewMatrix)
	log.Printf("projection %+v", projectionMatrix)
	scene.Observers[0].projection = projectionMatrix
	scene.Observers[0].view = viewMatrix
	scene.Observers[0].width = int64(width)
	scene.Observers[0].height = int64(height)

	// TODO: Multithread this part once everything seems to be better.
	// but for now, we will keep it as a single thread to make debugging easier.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, calculatePixel(scene, int64(x), int64(y)))
		}
	}

	// Encode as PNG.
	f, _ := os.Create("image.png")
	return png.Encode(f, img)
}

func calculateLight(scene *Scene, intersection IntersectionTriangle) (result Vector) {
	var shortestIntersection IntersectionTriangle
	var shortestDist float64

	if !intersection.Hit {
		return
	}
	result[3] = 1

	shortestDist = -1
	hardLimit := 40.0 // for now
	// Check if there are any triangles between intersection and the light.
	// TODO: solve this for multiple light sources. But for now, I'll just take the initial light source
	light := scene.Lights[0]
	rayDir := NormalizeVector(SubVector(light.Position, intersection.Intersection))
	rayStart := intersection.Intersection
	rayLength := VectorDistance(intersection.Intersection, light.Position)

	if rayLength >= hardLimit {
		result[3] = 1
		return
	}

	for k := range scene.Objects {
		test := raycastObjectIntersect(scene.Objects[k], rayStart, rayDir)
		if test.Hit {
			dist := VectorDistance(test.Intersection, intersection.Intersection)
			// TODO checking dist > 0 will lead into Z fighting in the future.
			// if (shortestDist == -1) || ((dist < shortestDist) && dist > 0) {
			if dist < DIFF {
				continue // Z Fighting for the very same triangle when the floating point goes punk
			}
			if (shortestDist == -1) || (dist < shortestDist) {
				shortestIntersection = test
				shortestDist = dist
			}
		}
	}
	if !shortestIntersection.Hit {
		intensity := 1 - (rayLength / hardLimit)
		l1 := NormalizeVector(SubVector(light.Position, rayStart))
		l2 := intersection.IntersectionNormal
		dotP := dot(l2, l1)
		if dotP < 0 {
			intensity = 0
		} else {
			intensity *= dotP
		}

		return Vector{
			light.Color[0] * intensity,
			light.Color[1] * intensity,
			light.Color[2] * intensity,
			1.0,
		}
	}
	// TODO: Do the global illumination
	return
}

func calculateColor(scene *Scene, intersection IntersectionTriangle) (result Vector) {
	if !intersection.Hit {
		return Vector{
			0, 0, 0, 1,
		}
	}
	material := intersection.Material
	if material.Texture != "" {
		if img, ok := scene.ImageMap[material.Texture]; ok {
			// ok, we have the image. Let's calculate the pixel color;
			s := intersection.getTexCoords()
			// get image size
			imgBounds := img.Bounds().Max

			if s[0] > 1 {
				s[0] = s[0] * 1.0
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

func calculatePixel(scene *Scene, x, y int64) (result color.RGBA) {
	var bestHit IntersectionTriangle
	var bestDist float64
	bestHit.Hit = false
	bestDist = -1
	// TODO: Antialiasing. https://en.wikipedia.org/wiki/Fast_approximate_anti-aliasing

	rayDir := ScreenToWorld(x, y, scene.Observers[0].width, scene.Observers[0].height, scene.Observers[0].Position, scene.Observers[0].projection, scene.Observers[0].view)
	for k := range scene.Objects {
		intersect := raycastObjectIntersect(scene.Objects[k], scene.Observers[0].Position, rayDir)
		if !intersect.Hit {
			continue
		}
		intersect.ObjectName = k
		dist := VectorDistance(intersect.Intersection, scene.Observers[0].Position)
		if (bestDist == -1) || (dist < bestDist) {
			bestHit = intersect
			bestDist = dist
		}
	}
	scene.DepthMap[x][y] = bestDist

	// Calculate light intensity
	if (x == 330) && (y == 790) {
		x = 330
	}
	lightIntensity := calculateLight(scene, bestHit)
	pixelColor := calculateColor(scene, bestHit)

	result = color.RGBA{
		uint8(math.Floor(lightIntensity[0] * pixelColor[0] * 255)),
		uint8(math.Floor(lightIntensity[1] * pixelColor[1] * 255)),
		uint8(math.Floor(lightIntensity[2] * pixelColor[2] * 255)),
		uint8(math.Floor(lightIntensity[3] * 255)),
	}
	return result
}
