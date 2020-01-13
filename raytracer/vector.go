package raytracer

import "math"

// Vector definition
type Vector [4]float64

// SubVector -
// substract vector v2 from v1
func SubVector(v1, v2 Vector) Vector {
	return Vector{
		v1[0] - v2[0],
		v1[1] - v2[1],
		v1[2] - v2[2],
		1,
	}
}

// CrossProduct - v1 X v2
func CrossProduct(v1, v2 Vector) Vector {
	return Vector{
		v1[1]*v2[2] - v1[2]*v2[1],
		v1[2]*v2[0] - v1[0]*v2[2],
		v1[0]*v2[1] - v1[1]*v2[0],
	}
}

func vectorNorm(v Vector) float64 {
	return math.Sqrt((v[0] * v[0]) + (v[1] * v[1]) + (v[2] * v[2]))
}

func dot(v1, v2 Vector) float64 {
	return v1[0]*v2[0] + v1[1]*v2[1] + v1[2]*v2[2]
}

// Combine -
func Combine(v1, v2 Vector, f1, f2 float64) Vector {
	x := (f1 * v1[0]) + (f2 * v2[0])
	y := (f1 * v1[1]) + (f2 * v2[1])
	z := (f1 * v1[2]) + (f2 * v2[2])
	w := (f1 * v1[3]) + (f2 * v2[3])
	return Vector{x, y, z, w}
}

// NormalizeVector -
func NormalizeVector(v Vector) Vector {
	vn := vectorNorm(v)
	if vn == 0 {
		return v
	}
	invlen := 1.0 / vn
	return Vector{
		v[0] * invlen,
		v[1] * invlen,
		v[2] * invlen,
		v[3],
	}
}

// VectorTransform -
func VectorTransform(v Vector, m Matrix) Vector {
	var result Vector
	result[0] = v[0]*m[0][0] + v[1]*m[1][0] + v[2]*m[2][0] + v[3]*m[3][0]
	result[1] = v[0]*m[0][1] + v[1]*m[1][1] + v[2]*m[2][1] + v[3]*m[3][1]
	result[2] = v[0]*m[0][2] + v[1]*m[1][2] + v[2]*m[2][2] + v[3]*m[3][2]
	result[3] = v[0]*m[0][3] + v[1]*m[1][3] + v[2]*m[2][3] + v[3]*m[3][3]
	return result
}

// VectorDistance -
func VectorDistance(v1, v2 Vector) float64 {
	diff := SubVector(v2, v1)
	return vectorNorm(diff)
}

func absVector(v Vector) Vector {
	return Vector{
		math.Abs(v[0]),
		math.Abs(v[1]),
		math.Abs(v[2]),
		math.Abs(v[3]),
	}
}

// BarycentricCoordinates is a hell of a thing
func BarycentricCoordinates(v1, v2, v3, p Vector) (u, v, w float64, success bool) {
	var a1, a2 int64
	var n, e1, e2, pt Vector

	e1 = SubVector(v1, v3)
	e2 = SubVector(v2, v3)
	pt = SubVector(p, v3)

	n = CrossProduct(e1, e2)
	n = absVector(n)
	a1 = 0
	if n[1] > n[a1] {
		a1 = 1
	}
	if n[2] > n[a1] {
		a1 = 2
	}
	switch a1 {
	case 0:
		a1 = 1
		a2 = 2
	case 1:
		a1 = 0
		a2 = 2
	default:
		a1 = 0
		a2 = 1
	}
	u = (pt[a2]*e2[a1] - pt[a1]*e2[a2]) / (e1[a2]*e2[a1] - e1[a1]*e2[a2])
	v = (pt[a2]*e1[a1] - pt[a1]*e1[a2]) / (e2[a2]*e1[a1] - e2[a1]*e1[a2])
	w = 1 - u - v

	success = (u >= DIFF) && (v >= DIFF) && (u+v <= 1.0+DIFF)
	return
}

func calculateBounds(vlist []Vector) (min, max Vector) {
	if len(vlist) == 0 {
		return
	}
	min = vlist[0]
	max = vlist[0]
	for i := range vlist {
		for j := 0; j < 4; j++ {
			if vlist[i][j] < min[j] {
				min[j] = vlist[i][j]
			}
			if vlist[i][j] > max[j] {
				max[j] = vlist[i][j]
			}
		}
	}
	return
}
