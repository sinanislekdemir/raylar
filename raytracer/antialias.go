package raytracer

import (
	"math"
	"math/rand"
)

func getPixel(scene *Scene, x, y int) Vector {
	if GlobalConfig.AntialiasSamples > 64 {
		GlobalConfig.AntialiasSamples = 64
	}
	if GlobalConfig.AntialiasSamples == 0 {
		return scene.Pixels[x][y].Color
	}
	sw := scene.Width * 8
	sh := scene.Height * 8
	totalColor := Vector{}
	totalHits := 0.0
	p := rand.Perm(64)
	for _, n := range p[:GlobalConfig.AntialiasSamples] {
		yi := int(math.Floor(float64(n)/float64(8))) + (y * 8) - 4
		xi := (n % 8) + (x * 8) - 4
		rayDir := screenToWorld(xi, yi, sw, sh, scene.Observers[0].Position, *scene.Observers[0].Projection, scene.Observers[0].view)
		hit := raycastSceneIntersect(scene, scene.Observers[0].Position, rayDir)
		render := hit.render(scene, 0)
		totalColor = addVector(totalColor, render)
		totalHits += 1.0
	}
	totalColor = scaleVector(totalColor, 1.0/totalHits)
	totalColor[3] = 1
	return totalColor
}
