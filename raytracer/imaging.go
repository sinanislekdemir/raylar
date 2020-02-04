package raytracer

import (
	"image"
	"image/color"
	"math"
)

func renderPixel(scene *Scene, x, y int) {
	var bestHit IntersectionTriangle
	var pixel PixelStorage
	var samples []IntersectionTriangle
	pixel.X = x
	pixel.Y = y

	bestHit.Hit = false

	rayDir := screenToWorld(x, y, scene.Observers[0].width, scene.Observers[0].height, scene.Observers[0].Position, *scene.Observers[0].Projection, scene.Observers[0].view)
	bestHit = raycastSceneIntersect(scene, scene.Observers[0].Position, rayDir)

	// We need some sampling from around
	if scene.Config.RenderAmbientColors || scene.Config.RenderOcclusion {
		samples = ambientSampling(scene, &bestHit)
	} else {
		samples = make([]IntersectionTriangle, 0)
	}

	if scene.Config.RenderLights {
		pixel.DirectLightEnergy = calculateTotalLight(scene, &bestHit, 0)
	}
	if scene.Config.RenderOcclusion {
		pixel.AmbientOcclusionRate = ambientLightCalc(scene, &bestHit, samples, scene.Config.SamplerLimit)
	}
	if scene.Config.RenderColors {
		pixel.Color = bestHit.getColor(scene, 0)
	}
	if scene.Config.RenderAmbientColors {
		pixel.AmbientColor = ambientColor(scene, &bestHit, samples, scene.Config.SamplerLimit)
	}

	if !bestHit.Hit {
		pixel.Color = Vector{1, 1, 1, 1}
	}

	pixel.Depth = bestHit.Dist
	if bestHit.Triangle != nil {
		if scene.Config.RenderReflections && bestHit.Triangle.Material.Glossiness > 0 {
			bounceDir := reflectVector(bestHit.RayDir, bestHit.IntersectionNormal)
			bounceStart := bestHit.Intersection
			reflection := raycastSceneIntersect(scene, bounceStart, bounceDir)
			if !reflection.Hit {
				pixel.Depth += reflection.Dist
			}
		}
		if scene.Config.RenderRefractions && bestHit.Triangle.Material.Transmission > 0 {
			bounceDir := refractVector(bestHit.RayDir, bestHit.IntersectionNormal, bestHit.Triangle.Material.IndexOfRefraction)
			bounceStart := bestHit.Intersection
			refraction := raycastSceneIntersect(scene, bounceStart, bounceDir)
			if !refraction.Hit {
				pixel.Depth += refraction.Dist
			}
		}
	}

	scene.Pixels[x][y] = pixel
}

func renderImage(scene *Scene, image *image.RGBA) {
	maxLight := 0.0
	// Calculate the occlusion rate to apply to the scene
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
			// Calcualte illumination for the pixel fragment
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
			// Some kind of weird thing might exceed illumination factor over 1
			light = limitVector(light, 1)

			// Imagine that the pixel is actually white.
			pcolor := Vector{1, 1, 1, 1}

			// Do we have any colors?
			if scene.Config.RenderColors {
				// Yep, use the pixel color (from material)
				pcolor = pixel.Color
				// Does the colors effect each-other?
				if scene.Config.RenderAmbientColors {
					pcolor = Vector{
						(pcolor[0] * (1.0 - scene.Config.AmbientColorSharingRatio)) + (pixel.AmbientColor[0] * scene.Config.AmbientColorSharingRatio),
						(pcolor[1] * (1.0 - scene.Config.AmbientColorSharingRatio)) + (pixel.AmbientColor[1] * scene.Config.AmbientColorSharingRatio),
						(pcolor[2] * (1.0 - scene.Config.AmbientColorSharingRatio)) + (pixel.AmbientColor[2] * scene.Config.AmbientColorSharingRatio),
						1,
					}
				}
				// Just in case if we exceed 1. Ideally won't happen. But a correction would be nice.
				pcolor = limitVector(pcolor, 1.0)
			}

			pixelAlpha := 1.0
			if pixel.Depth < 0 {
				pixelAlpha = 0.0
			}
			pcolor = Vector{
				pcolor[0] * light[0],
				pcolor[1] * light[1],
				pcolor[2] * light[2],
				pixelAlpha,
			}

			pixel.Color = pcolor
			pixel.TotalLight = light
			scene.Pixels[i][j] = pixel
		}
	}
	// Unify pixel depths between 0-1.
	maxDepth := scene.Pixels[0][0].Depth

	for i := 0; i < scene.Width; i++ {
		for j := 0; j < scene.Height; j++ {
			if maxDepth < scene.Pixels[i][j].Depth {
				maxDepth = scene.Pixels[i][j].Depth
			}
		}
	}
	for i := 0; i < scene.Width; i++ {
		for j := 0; j < scene.Height; j++ {
			scene.Pixels[i][j].Depth /= maxDepth
			if math.IsNaN(scene.Pixels[i][j].Depth) {
				scene.Pixels[i][j].Depth = 0
			}
		}
	}

	// Print pixels onto image
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
	aaRadius := 1
	if x < aaRadius || x+aaRadius >= scene.Width || y < aaRadius || y+aaRadius >= scene.Height {
		return pixelColor
	}

	v := make([]PixelStorage, 0)
	for i := -aaRadius; i <= aaRadius; i++ {
		for j := -aaRadius; j <= aaRadius; j++ {
			p := scene.Pixels[x+i][y+j]
			v = append(v, p)
		}
	}

	if len(v) < 3 || pixelColor[3] < 0 {
		return pixelColor
	}

	minDepth := v[0].Depth
	maxDepth := v[0].Depth
	minColor := vectorSum(v[0].Color)
	maxColor := vectorSum(v[0].Color)

	for i := range v {
		if v[i].Depth <= 0 {
			continue
		}
		if v[i].Depth < minDepth {
			minDepth = v[i].Depth
		}
		if v[i].Depth > maxDepth {
			maxDepth = v[i].Depth
		}
		vsum := vectorSum(v[i].Color)
		if vsum < minColor {
			minColor = vsum
		}
		if vsum > maxColor {
			maxColor = vsum
		}
	}

	check := (maxDepth-minDepth > scene.Config.EdgeDetechThreshold) || (maxColor-minColor > scene.Config.EdgeDetechThreshold)

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
	totalColor[3] = 1
	return totalColor
}
