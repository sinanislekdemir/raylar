package raytracer

import (
	"math"
)

// Vector definition
type Vector [4]float64

// func degToRad(degrees float64) float64 {
// 	return degrees * (math.Pi / 180.0)
// }

// func radToDeg(radians float64) float64 {
// 	return radians * (180.0 / math.Pi)
// }

// func vecToSpherical(v Vector) (r, phi, theta float64) {
// 	r = vectorLength(v)
// 	phi = math.Atan(v[1] / v[0])
// 	theta = math.Acos(v[2] / r)
// 	return
// }

// func sphericalToVec(r, phi, theta float64) (v Vector) {
// 	v[0] = r * math.Sin(theta) * math.Cos(phi)
// 	v[1] = r * math.Sin(theta) * math.Sin(phi)
// 	v[2] = r * math.Cos(theta)
// 	return
// }

func sameSideTest(v1, v2 Vector, shifting float64) bool {
	return dot(v1, v2)-shifting > -DIFF
}

// subVector -
// substract vector v2 from v1
func subVector(v1, v2 Vector) Vector {
	return Vector{
		v1[0] - v2[0],
		v1[1] - v2[1],
		v1[2] - v2[2],
		1,
	}
}

func psubVector(v1, v2 *Vector) *Vector {
	return &Vector{
		v1[0] - v2[0],
		v1[1] - v2[1],
		v1[2] - v2[2],
		1,
	}
}

func addVector(v1, v2 Vector) Vector {
	return Vector{
		v1[0] + v2[0],
		v1[1] + v2[1],
		v1[2] + v2[2],
		v1[0] + v2[0],
	}
}

func limitVector(v Vector, factor float64) Vector {
	result := Vector{v[0], v[1], v[2], v[3]}
	for i := 0; i < 3; i++ {
		if result[i] > factor {
			result[i] = factor
		}
	}
	return result
}

func upscaleVector(v Vector, factor float64) Vector {
	result := Vector{v[0], v[1], v[2], v[3]}
	for i := 0; i < 3; i++ {
		if result[i] < factor {
			result[i] = factor
		}
	}
	return result
}

func limitVectorByVector(v, factor Vector) Vector {
	result := Vector{v[0], v[1], v[2], v[3]}
	for i := 0; i < 3; i++ {
		if result[i] > factor[i] {
			result[i] = factor[i]
		}
	}
	return result
}

func scaleVector(v Vector, factor float64) Vector {
	if factor == 0 {
		return Vector{}
	}
	return Vector{
		v[0] * factor,
		v[1] * factor,
		v[2] * factor,
		v[3],
	}
}

// crossProduct - v1 X v2
func crossProduct(v1, v2 Vector) Vector {
	return Vector{
		v1[1]*v2[2] - v1[2]*v2[1],
		v1[2]*v2[0] - v1[0]*v2[2],
		v1[0]*v2[1] - v1[1]*v2[0],
	}
}

func pcrossProduct(v1, v2 *Vector) *Vector {
	return &Vector{
		v1[1]*v2[2] - v1[2]*v2[1],
		v1[2]*v2[0] - v1[0]*v2[2],
		v1[0]*v2[1] - v1[1]*v2[0],
	}
}

func vectorNorm(v Vector) float64 {
	return (v[0] * v[0]) + (v[1] * v[1]) + (v[2] * v[2])
}

func pvectorNorm(v *Vector) float64 {
	return (v[0] * v[0]) + (v[1] * v[1]) + (v[2] * v[2])
}

func dot(v1, v2 Vector) float64 {
	return v1[0]*v2[0] + v1[1]*v2[1] + v1[2]*v2[2]
}

func pdot(v1, v2 *Vector) float64 {
	return v1[0]*v2[0] + v1[1]*v2[1] + v1[2]*v2[2]
}

// combine -
func combine(v1, v2 Vector, f1, f2 float64) Vector {
	x := (f1 * v1[0]) + (f2 * v2[0])
	y := (f1 * v1[1]) + (f2 * v2[1])
	z := (f1 * v1[2]) + (f2 * v2[2])
	w := (f1 * v1[3]) + (f2 * v2[3])
	return Vector{x, y, z, w}
}

func pcombine(v1, v2 *Vector, f1, f2 float64) *Vector {
	x := (f1 * v1[0]) + (f2 * v2[0])
	y := (f1 * v1[1]) + (f2 * v2[1])
	z := (f1 * v1[2]) + (f2 * v2[2])
	w := (f1 * v1[3]) + (f2 * v2[3])
	return &Vector{x, y, z, w}
}

// normalizeVector -
func pnormalizeVector(v *Vector) *Vector {
	vn := pvectorNorm(v)
	if vn == 0 {
		return v
	}
	invlen := 1.0 / math.Sqrt(vn)
	return &Vector{
		v[0] * invlen,
		v[1] * invlen,
		v[2] * invlen,
		v[3],
	}
}

func normalizeVector(v Vector) Vector {
	vn := vectorNorm(v)
	if vn == 0 {
		return v
	}
	invlen := 1.0 / math.Sqrt(vn)
	return Vector{
		v[0] * invlen,
		v[1] * invlen,
		v[2] * invlen,
		v[3],
	}
}

// vectorTransform -
func vectorTransform(v Vector, m Matrix) Vector {
	var result Vector
	result[0] = v[0]*m[0][0] + v[1]*m[1][0] + v[2]*m[2][0] + v[3]*m[3][0]
	result[1] = v[0]*m[0][1] + v[1]*m[1][1] + v[2]*m[2][1] + v[3]*m[3][1]
	result[2] = v[0]*m[0][2] + v[1]*m[1][2] + v[2]*m[2][2] + v[3]*m[3][2]
	result[3] = v[0]*m[0][3] + v[1]*m[1][3] + v[2]*m[2][3] + v[3]*m[3][3]
	return result
}

// vectorDistance -
func vectorDistance(v1, v2 Vector) float64 {
	diff := subVector(v2, v1)
	return vectorLength(diff)
}

func pvectorDistance(v1, v2 *Vector) float64 {
	diff := psubVector(v2, v1)
	return vectorLength(*diff)
}

func absVector(v Vector) Vector {
	return Vector{
		math.Abs(v[0]),
		math.Abs(v[1]),
		math.Abs(v[2]),
		math.Abs(v[3]),
	}
}

// barycentricCoordinates is a hell of a thing
func barycentricCoordinates(v1, v2, v3, p Vector) (u, v, w float64, success bool) {
	var a1, a2 int64
	var n, e1, e2, pt Vector

	e1 = subVector(v1, v3)
	e2 = subVector(v2, v3)
	pt = subVector(p, v3)

	n = crossProduct(e1, e2)
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

// TODO: Refactor
func localToAbsoluteList(vertices []Vector, matrix Matrix) []Vector {
	result := make([]Vector, len(vertices))
	for i := 0; i < len(vertices); i++ {
		result[i] = vectorTransform(vertices[i], matrix)
	}
	return result
}

func vectorLength(v Vector) float64 {
	return math.Sqrt(vectorNorm(v))
}

func reflectVector(v, n Vector) Vector {
	return combine(v, n, 1.0, -2*dot(v, n))
}

func vectorSum(v Vector) float64 {
	return v[0] + v[1] + v[2]
}

func refractVector(v, n Vector, ior float64) Vector {
	if ior < DIFF {
		return v
	}
	ior = 1.0 / ior
	nDotI := dot(n, v)
	k := 1.0 - ior*ior*(1.0-nDotI*nDotI)
	if k < 0 {
		return Vector{}
	} else {
		return subVector(scaleVector(v, ior), scaleVector(n, (ior*nDotI+math.Sqrt(k))))
	}
}
