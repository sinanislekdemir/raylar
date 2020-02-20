package raytracer

import (
	"image"
	"image/color"
	"math"
)

func renderPixel(scene *Scene, x, y int) {
	var bestHit IntersectionTriangle
	var pixel PixelStorage
	// var samples []IntersectionTriangle
	pixel.X = x
	pixel.Y = y

	bestHit.Hit = false

	bestHit = scene.Pixels[x][y].WorldLocation
	if !bestHit.Hit {
		pixel.Color = Vector{1, 1, 1, 1}
	}

	pixel.Color = bestHit.render(scene, 0)
	pixel.Depth = bestHit.Dist

	if bestHit.Triangle != nil {
		if GlobalConfig.RenderReflections && bestHit.Triangle.Material.Glossiness > 0 {
			bounceDir := reflectVector(bestHit.RayDir, bestHit.IntersectionNormal)
			bounceStart := bestHit.Intersection
			reflection := raycastSceneIntersect(scene, bounceStart, bounceDir)
			if !reflection.Hit {
				pixel.Depth += reflection.Dist
			}
		}
		if GlobalConfig.RenderRefractions && bestHit.Triangle.Material.Transmission > 0 {
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
	transparent := false

	for i := range v {
		if v[i].Depth <= 0 {
			transparent = true
			break
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

	check := (maxDepth-minDepth > GlobalConfig.EdgeDetechThreshold) || (maxColor-minColor > GlobalConfig.EdgeDetechThreshold)

	if transparent {
		check = true
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
	totalColor[3] = 1
	return totalColor
}
