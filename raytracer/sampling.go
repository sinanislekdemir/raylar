package raytracer

import (
	"math/rand"
	"sort"
)

var sampleCache [][]Vector

func createCache() {
	sampleCache = make([][]Vector, 10)
	// Create 10 different cache variations
	for index := 0; index < 10; index++ {
		sampleCache[index] = make([]Vector, 100000)
		for i := 0; i < 100000; i++ {
			x := rand.Float64() - 0.5
			y := rand.Float64() - 0.5
			z := rand.Float64() - 0.5
			v := normalizeVector(Vector{x, y, z, 0})
			sampleCache[index][i] = v
		}
	}
}

func createSamples(normal Vector, limit int, shifting float64) []Vector {
	if sampleCache == nil {
		createCache()
	}
	index := rand.Int() % 10

	result := make([]Vector, 0, limit)
	result = append(result, normal)

	for i := range sampleCache[index] {
		if sameSideTest(sampleCache[index][i], normal, shifting) {
			result = append(result, sampleCache[index][i])
			if len(result) == limit {
				break
			}
		}
	}
	return result
}

func sampleSphere(radius float64, limit int) []Vector {
	result := make([]Vector, limit)
	for i := 0; i < limit; i++ {
		result[i] = Vector{
			(rand.Float64() - 0.5) * radius,
			(rand.Float64() - 0.5) * radius,
			(rand.Float64() - 0.5) * radius,
			1,
		}
	}
	return result
}

func sampleTriangle(triangle Triangle, count int) []Vector {
	result := make([]Vector, count)
	for i := 0; i < count; i++ {
		vl := []float64{
			rand.Float64(), rand.Float64(),
		}
		sort.Float64s(vl)
		s := vl[0]
		t := vl[1]

		result[i] = Vector{
			s*triangle.P1[0] + (t-s)*triangle.P2[0] + (1-t)*triangle.P3[0],
			s*triangle.P1[1] + (t-s)*triangle.P2[1] + (1-t)*triangle.P3[1],
			s*triangle.P1[2] + (t-s)*triangle.P2[2] + (1-t)*triangle.P3[2],
			1,
		}
	}
	return result
}
