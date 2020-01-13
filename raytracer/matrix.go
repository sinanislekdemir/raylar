package raytracer

import (
	"math"
)

// Matrix definition
type Matrix [4]Vector

var identityHmgMatrix = Matrix{
	Vector{1, 0, 0, 0},
	Vector{0, 1, 0, 0},
	Vector{0, 0, 1, 0},
	Vector{0, 0, 0, 1},
}

// PerspectiveProjection - Create perspective
func PerspectiveProjection(fovy, aspect, near, far float64) Matrix {
	ymax := near * math.Tan(fovy*math.Pi/360.0)
	xmax := ymax * aspect
	left := -xmax
	right := xmax
	bottom := -ymax
	top := ymax
	a := (right + left) / (right - left)
	b := (top + bottom) / (top - bottom)
	c := -(far + near) / (far - near)
	d := -2.0 * far * near / (far - near)
	e := 2.0 * near / (right - left)
	f := 2.0 * near / (top - bottom)
	return Matrix{
		Vector{e, 0, 0, 0},
		Vector{0, f, 0, 0},
		Vector{a, b, c, -1},
		Vector{0, 0, d, 0},
	}
}

// ViewMatrix -
func ViewMatrix(eye, target, up Vector) Matrix {
	forward := NormalizeVector(SubVector(target, eye))
	side := NormalizeVector(CrossProduct(forward, up))
	up = NormalizeVector(CrossProduct(side, forward))
	dse := dot(side, eye)
	due := dot(up, eye)
	dfe := dot(forward, eye)
	return Matrix{
		Vector{side[0], up[0], -forward[0], 0.0},
		Vector{side[1], up[1], -forward[1], 0.0},
		Vector{side[2], up[2], -forward[2], 0.0},
		Vector{dse, due, dfe, 1.0},
	}
}

// MatrixDeterminant -
func MatrixDeterminant(m Matrix) float64 {
	a := m[0][0] * MatrixDetInternal(m[1][1], m[2][1], m[3][1], m[1][2], m[2][2], m[3][2],
		m[1][3], m[2][3], m[3][3])
	b := m[0][1] * MatrixDetInternal(m[1][0], m[2][0], m[3][0], m[1][2], m[2][2], m[3][2],
		m[1][3], m[2][3], m[3][3])
	c := m[0][2] * MatrixDetInternal(m[1][0], m[2][0], m[3][0], m[1][1], m[2][1], m[3][1],
		m[1][3], m[2][3], m[3][3])
	d := m[0][3] * MatrixDetInternal(m[1][0], m[2][0], m[3][0], m[1][1], m[2][1], m[3][1],
		m[1][2], m[2][2], m[3][2])
	return a - b + c - d
}

// MatrixDetInternal -
func MatrixDetInternal(a1, a2, a3, b1, b2, b3, c1, c2, c3 float64) float64 {
	return (a1*(b2*c3-b3*c2) - b1*(a2*c3-a3*c2) + c1*(a2*b3-a3*b2))
}

// AdjointMatrix -
func AdjointMatrix(m Matrix) Matrix {
	var a1, a2, a3, a4, b1, b2, b3, b4, c1, c2, c3, c4, d1, d2, d3, d4 float64
	var result Matrix
	a1 = m[0][0]
	b1 = m[0][1]
	c1 = m[0][2]
	d1 = m[0][3]
	a2 = m[1][0]
	b2 = m[1][1]
	c2 = m[1][2]
	d2 = m[1][3]
	a3 = m[2][0]
	b3 = m[2][1]
	c3 = m[2][2]
	d3 = m[2][3]
	a4 = m[3][0]
	b4 = m[3][1]
	c4 = m[3][2]
	d4 = m[3][3]
	result[0][0] = MatrixDetInternal(b2, b3, b4, c2, c3, c4, d2, d3, d4)
	result[1][0] = -MatrixDetInternal(a2, a3, a4, c2, c3, c4, d2, d3, d4)
	result[2][0] = MatrixDetInternal(a2, a3, a4, b2, b3, b4, d2, d3, d4)
	result[3][0] = -MatrixDetInternal(a2, a3, a4, b2, b3, b4, c2, c3, c4)

	result[0][1] = -MatrixDetInternal(b1, b3, b4, c1, c3, c4, d1, d3, d4)
	result[1][1] = MatrixDetInternal(a1, a3, a4, c1, c3, c4, d1, d3, d4)
	result[2][1] = -MatrixDetInternal(a1, a3, a4, b1, b3, b4, d1, d3, d4)
	result[3][1] = MatrixDetInternal(a1, a3, a4, b1, b3, b4, c1, c3, c4)

	result[0][2] = MatrixDetInternal(b1, b2, b4, c1, c2, c4, d1, d2, d4)
	result[1][2] = -MatrixDetInternal(a1, a2, a4, c1, c2, c4, d1, d2, d4)
	result[2][2] = MatrixDetInternal(a1, a2, a4, b1, b2, b4, d1, d2, d4)
	result[3][2] = -MatrixDetInternal(a1, a2, a4, b1, b2, b4, c1, c2, c4)

	result[0][3] = -MatrixDetInternal(b1, b2, b3, c1, c2, c3, d1, d2, d3)
	result[1][3] = MatrixDetInternal(a1, a2, a3, c1, c2, c3, d1, d2, d3)
	result[2][3] = -MatrixDetInternal(a1, a2, a3, b1, b2, b3, d1, d2, d3)
	result[3][3] = MatrixDetInternal(a1, a2, a3, b1, b2, b3, c1, c2, c3)
	return result
}

// ScaleMatrix -
func ScaleMatrix(m Matrix, factor float64) Matrix {
	var result Matrix
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			result[i][j] = m[i][j] * factor
		}
	}
	return result
}

// InvertMatrix -
func InvertMatrix(m Matrix) Matrix {
	det := MatrixDeterminant(m)
	epsilon := 1e-40
	if math.Abs(det) < epsilon {
		return identityHmgMatrix
	}
	m2 := AdjointMatrix(m)
	return ScaleMatrix(m2, 1/det)
}

// MultiplyMatrix -
func MultiplyMatrix(m1, m2 Matrix) Matrix {
	var result Matrix
	result[0][0] = m1[0][0]*m2[0][0] + m1[0][1]*m2[1][0] + m1[0][2]*m2[2][0] + m1[0][3]*m2[3][0]
	result[0][1] = m1[0][0]*m2[0][1] + m1[0][1]*m2[1][1] + m1[0][2]*m2[2][1] + m1[0][3]*m2[3][1]
	result[0][2] = m1[0][0]*m2[0][2] + m1[0][1]*m2[1][2] + m1[0][2]*m2[2][2] + m1[0][3]*m2[3][2]
	result[0][3] = m1[0][0]*m2[0][3] + m1[0][1]*m2[1][3] + m1[0][2]*m2[2][3] + m1[0][3]*m2[3][3]
	result[1][0] = m1[1][0]*m2[0][0] + m1[1][1]*m2[1][0] + m1[1][2]*m2[2][0] + m1[1][3]*m2[3][0]
	result[1][1] = m1[1][0]*m2[0][1] + m1[1][1]*m2[1][1] + m1[1][2]*m2[2][1] + m1[1][3]*m2[3][1]
	result[1][2] = m1[1][0]*m2[0][2] + m1[1][1]*m2[1][2] + m1[1][2]*m2[2][2] + m1[1][3]*m2[3][2]
	result[1][3] = m1[1][0]*m2[0][3] + m1[1][1]*m2[1][3] + m1[1][2]*m2[2][3] + m1[1][3]*m2[3][3]
	result[2][0] = m1[2][0]*m2[0][0] + m1[2][1]*m2[1][0] + m1[2][2]*m2[2][0] + m1[2][3]*m2[3][0]
	result[2][1] = m1[2][0]*m2[0][1] + m1[2][1]*m2[1][1] + m1[2][2]*m2[2][1] + m1[2][3]*m2[3][1]
	result[2][2] = m1[2][0]*m2[0][2] + m1[2][1]*m2[1][2] + m1[2][2]*m2[2][2] + m1[2][3]*m2[3][2]
	result[2][3] = m1[2][0]*m2[0][3] + m1[2][1]*m2[1][3] + m1[2][2]*m2[2][3] + m1[2][3]*m2[3][3]
	result[3][0] = m1[3][0]*m2[0][0] + m1[3][1]*m2[1][0] + m1[3][2]*m2[2][0] + m1[3][3]*m2[3][0]
	result[3][1] = m1[3][0]*m2[0][1] + m1[3][1]*m2[1][1] + m1[3][2]*m2[2][1] + m1[3][3]*m2[3][1]
	result[3][2] = m1[3][0]*m2[0][2] + m1[3][1]*m2[1][2] + m1[3][2]*m2[2][2] + m1[3][3]*m2[3][2]
	result[3][3] = m1[3][0]*m2[0][3] + m1[3][1]*m2[1][3] + m1[3][2]*m2[2][3] + m1[3][3]*m2[3][3]
	return result
}
