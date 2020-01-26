package raytracer

import (
	"log"
	"math/rand"
)

var sampleCache [][]Vector

func createSamples(normal Vector, limit int) []Vector {
	if sampleCache == nil {
		log.Printf("Creating sampling cache for 10.000 Vectors")
		sampleCache = make([][]Vector, 10)
		// Create 10 different cache variations
		for index := 0; index < 10; index++ {
			sampleCache[index] = make([]Vector, 1000)
			for i := 0; i < 1000; i++ {
				x := rand.Float64() - 0.5
				y := rand.Float64() - 0.5
				z := rand.Float64() - 0.5
				v := normalizeVector(Vector{x, y, z, 0})
				sampleCache[index][i] = v
			}
		}
	}
	index := rand.Int() % 10

	result := make([]Vector, 0)
	result = append(result, normal)

	for i := range sampleCache[index] {
		if sameSideTest(sampleCache[index][i], normal) {
			result = append(result, sampleCache[index][i])
			if len(result) == limit {
				break
			}
		}
	}
	return result
}
