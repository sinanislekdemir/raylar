package raytracer

import (
	"math/rand"
	"sort"
)

var sampleCache [][]Vector

func createSamples(normal Vector, limit int) []Vector {
	if sampleCache == nil {
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
