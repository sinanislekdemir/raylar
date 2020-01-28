package raytracer

import (
	"image"
	"image/color"
	"math"
)

func renderPixel(scene *Scene, x, y int) {
	var bestHit IntersectionTriangle
	var pixel PixelStorage
	pixel.X = x
	pixel.Y = y

	bestHit.Hit = false
	// TODO: Antialiasing. https://en.wikipedia.org/wiki/Fast_approximate_anti-aliasing

	rayDir := ScreenToWorld(x, y, scene.Observers[0].width, scene.Observers[0].height, scene.Observers[0].Position, scene.Observers[0].projection, scene.Observers[0].view)
	bestHit = raycastSceneIntersect(scene, scene.Observers[0].Position, rayDir)
	if scene.Config.RenderLights {
		pixel.DirectLightEnergy = calculateTotalLight(scene, bestHit, 0)
	}
	if scene.Config.RenderOcclusion {
		pixel.AmbientOcclusionRate = ambientLightCalc(scene, bestHit)
	}
	if scene.Config.RenderColors {
		pixel.Color = calculateColor(scene, bestHit, 0)
	}
	if scene.Config.RenderAmbientColors {
		pixel.AmbientColor = ambientColor(scene, bestHit)
	}

	if !bestHit.Hit {
		pixel.Color = Vector{1, 1, 1, 1}
	}
	pixel.Depth = bestHit.Dist
	if scene.Config.RenderReflections && bestHit.Triangle.Material.Glossiness > 0 {
		bounceDir := reflectVector(bestHit.RayDir, bestHit.IntersectionNormal)
		bounceStart := bestHit.Intersection
		reflection := raycastSceneIntersect(scene, bounceStart, bounceDir)
		if !reflection.Hit {
			pixel.Depth += reflection.Dist
		}

	}
	scene.Pixels[x][y] = pixel
}

func renderImage(scene *Scene, image *image.RGBA) {
	maxLight := 0.0
	if scene.Config.RenderLights && scene.Config.RenderOcclusion {
		for i := 0; i < scene.Width; i++ {
			for j := 0; j < scene.Height; j++ {
				for k := 0; k < 3; k++ {
					if scene.Pixels[i][j].DirectLightEnergy[k] > maxLight {
						maxLight = scene.Pixels[i][j].DirectLightEnergy[k]
					}
				}
			}
		}
		if maxLight > 0 && scene.Config.RenderLights && scene.Config.OcclusionRate > maxLight {
			scene.Config.OcclusionRate = maxLight
		}
	}

	for i := 0; i < scene.Width; i++ {
		for j := 0; j < scene.Height; j++ {
			pixel := scene.Pixels[i][j]
			light := Vector{0, 0, 0, 1}
			if scene.Config.RenderLights {
				light = pixel.DirectLightEnergy
			}
			if scene.Config.RenderOcclusion {
				light = addVector(Vector{
					pixel.AmbientOcclusionRate * scene.Config.OcclusionRate,
					pixel.AmbientOcclusionRate * scene.Config.OcclusionRate,
					pixel.AmbientOcclusionRate * scene.Config.OcclusionRate,
					1,
				}, light)
			}
			light = limitVector(light, 1)
			pcolor := Vector{1, 1, 1, 1}

			if scene.Config.RenderColors {
				pcolor = pixel.Color
				if scene.Config.RenderAmbientColors {
					pcolor = Vector{
						(pcolor[0] * (1.0 - scene.Config.AmbientColorSharingRatio)) + (pixel.AmbientColor[0] * scene.Config.AmbientColorSharingRatio),
						(pcolor[1] * (1.0 - scene.Config.AmbientColorSharingRatio)) + (pixel.AmbientColor[1] * scene.Config.AmbientColorSharingRatio),
						(pcolor[2] * (1.0 - scene.Config.AmbientColorSharingRatio)) + (pixel.AmbientColor[2] * scene.Config.AmbientColorSharingRatio),
						1,
					}
				}
				pcolor = limitVector(pcolor, 1.0)
			}
			w := 1.0
			if pixel.Depth < 0 {
				w = 0.0
			}
			pcolor = Vector{
				pcolor[0] * light[0],
				pcolor[1] * light[1],
				pcolor[2] * light[2],
				w,
			}

			pixel.Color = pcolor
			pixel.TotalLight = light
			scene.Pixels[i][j] = pixel
		}
	}
	for i := 0; i < scene.Width; i++ {
		for j := 0; j < scene.Height; j++ {
			pcolor := scene.Pixels[i][j].Color
			pcolor = getPixelColor(scene, i, j, pcolor)
			colorRGBA := color.RGBA{
				uint8(math.Floor(pcolor[0] * 255)),
				uint8(math.Floor(pcolor[1] * 255)),
				uint8(math.Floor(pcolor[2] * 255)),
				uint8(math.Floor(pcolor[3] * 255)),
			}
			image.Set(i, j, colorRGBA)
		}
	}
}

func getPixelColor(scene *Scene, x, y int, pixelColor Vector) Vector {
	if x < 1 || x+1 == scene.Width || y < 1 || y+1 == scene.Height {
		return pixelColor
	}

	v := make([]PixelStorage, 0)
	for i := -1; i < 2; i++ {
		for j := -1; j < 2; j++ {
			p := scene.Pixels[x+i][y+j]
			if p.Color[3] > 0 {
				v = append(v, p)
			}
		}
	}
	if len(v) < 3 || pixelColor[3] < 0 {
		return pixelColor
	}

	minLight := v[0].TotalLight
	maxLight := v[0].TotalLight

	for i := range v {
		for d := 0; d < 3; d++ {
			if v[i].TotalLight[d] < minLight[d] {
				minLight[d] = v[i].TotalLight[d]
			}
			if v[i].TotalLight[d] > maxLight[d] {
				maxLight[d] = v[i].TotalLight[d]
			}
		}
	}

	check := false
	for i := 0; i < 3; i++ {
		if maxLight[i]-minLight[i] > scene.Config.EdgeDetechThreshold {
			check = true
		}
	}
	if !check {
		return pixelColor
	}

	totalColor := Vector{}
	totalHits := 0.0
	for i := range v {
		if !math.IsNaN(v[i].Color[0]) {
			totalColor = addVector(totalColor, v[i].Color)
			totalHits += 1.0
		}
	}

	totalColor = scaleVector(totalColor, 1.0/totalHits)

	return Vector{
		(pixelColor[0] + totalColor[0]) / 2.0,
		(pixelColor[1] + totalColor[1]) / 2.0,
		(pixelColor[2] + totalColor[2]) / 2.0,
		1.0,
	}

}
