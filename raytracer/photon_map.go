package raytracer

import (
	"log"
	"math"
	"runtime"
	"sync"
)

func lightSampleToScreen(scene *Scene, light *Light) []Vector {
	result := make([]Vector, 0, scene.Width*scene.Height)
	for i := range scene.Pixels {
		for j := range scene.Pixels[i] {
			if scene.Pixels[i][j].WorldLocation.Hit {
				vec := subVector(scene.Pixels[i][j].WorldLocation.Intersection, light.Position)
				vec = normalizeVector(vec)
				result = append(result, vec)
			}
		}
	}
	return result
}

func buildPhotonMap(scene *Scene) {
	log.Printf("Analysing scene for caustic surfaces")
	causticSampleLocations := make([]Vector, 0)
	for k := range scene.Objects {
		for tri := range scene.Objects[k].Triangles {
			if scene.Objects[k].Triangles[tri].Material.Glossiness > 0 || scene.Objects[k].Triangles[tri].Material.Transmission > 0 {
				locations := sampleTriangle(scene.Objects[k].Triangles[tri], scene.Config.CausticsSamplerLimit)
				causticSampleLocations = append(causticSampleLocations, locations...)
			}
		}
	}
	log.Printf("Found %d sample photons", len(causticSampleLocations))
	for i := range scene.Lights {
		var wg sync.WaitGroup
		workCount := runtime.NumCPU() * 8
		batchSize := int(math.Floor(float64(len(causticSampleLocations)) / float64(workCount)))
		for k := 0; k < workCount-1; k++ {
			from := batchSize * k
			to := batchSize * (k + 1)
			if to > len(causticSampleLocations) {
				to = len(causticSampleLocations)
			}
			sample := causticSampleLocations[from:to]
			wg.Add(1)
			go func(scene *Scene, samples []Vector, light *Light, wg *sync.WaitGroup) {
				for sampleIndex := range samples {
					dir := normalizeVector(subVector(samples[sampleIndex], light.Position))
					photon := Photon{
						Location:  light.Position,
						Color:     light.Color,
						Direction: dir,
						Intensity: light.LightStrength,
					}
					tracePhoton(scene, &photon, 0)
				}
				wg.Done()
			}(scene, sample, &scene.Lights[i], &wg)
			wg.Wait()
		}
	}
	log.Printf("Done building photon map")
}
