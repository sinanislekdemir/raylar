package raytracer

import (
	"math"
)

// Matrix definition.
type Matrix [4]Vector

var identityHmgMatrix = Matrix{
	Vector{1, 0, 0, 0},
	Vector{0, 1, 0, 0},
	Vector{0, 0, 1, 0},
	Vector{0, 0, 0, 1},
}

// perspectiveProjection - Create perspective.
func perspectiveProjection(fovy, aspect, near, far float64) Matrix {
	ymax := near * math.Tan(fovy*math.Pi/360.0)
	xmax := ymax * aspect
	left := -xmax
	right := xmax
	bottom := -ymax
	top := ymax
	temp := 2.0 * near
	temp2 := right - left
	temp3 := top - bottom
	temp4 := far - near
	return Matrix{
		Vector{temp / temp2, 0, 0, 0},
		Vector{0, temp / temp3, 0, 0},
		Vector{(right + left) / temp2, (top + bottom) / temp3, (-far - near) / temp4, -1},
		Vector{0, 0, (-temp * far) / temp4, 0},
	}
}

// viewMatrix calculation.
func viewMatrix(eye, target, up Vector) Matrix {
	forward := normalizeVector(subVector(target, eye))
	side := normalizeVector(crossProduct(forward, up))
	up = normalizeVector(crossProduct(side, forward))
	dse := -dot(side, eye)
	due := -dot(up, eye)
	dfe := dot(forward, eye)
	return Matrix{
		Vector{side[0], up[0], -forward[0], 0.0},
		Vector{side[1], up[1], -forward[1], 0.0},
		Vector{side[2], up[2], -forward[2], 0.0},
		Vector{dse, due, dfe, 1.0},
	}
}

// matrixDeterminant calculation.
func matrixDeterminant(m Matrix) float64 {
	a := m[0][0] * matrixDetInternal(m[1][1], m[2][1], m[3][1], m[1][2], m[2][2], m[3][2],
		m[1][3], m[2][3], m[3][3])
	b := m[0][1] * matrixDetInternal(m[1][0], m[2][0], m[3][0], m[1][2], m[2][2], m[3][2],
		m[1][3], m[2][3], m[3][3])
	c := m[0][2] * matrixDetInternal(m[1][0], m[2][0], m[3][0], m[1][1], m[2][1], m[3][1],
		m[1][3], m[2][3], m[3][3])
	d := m[0][3] * matrixDetInternal(m[1][0], m[2][0], m[3][0], m[1][1], m[2][1], m[3][1],
		m[1][2], m[2][2], m[3][2])
	return a - b + c - d
}

// matrixDetInternal calculation.
func matrixDetInternal(a1, a2, a3, b1, b2, b3, c1, c2, c3 float64) float64 {
	return (a1*(b2*c3-b3*c2) - b1*(a2*c3-a3*c2) + c1*(a2*b3-a3*b2))
}

// adjointMatrix calculation.
func adjointMatrix(m Matrix) Matrix {
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
	result[0][0] = matrixDetInternal(b2, b3, b4, c2, c3, c4, d2, d3, d4)
	result[1][0] = -matrixDetInternal(a2, a3, a4, c2, c3, c4, d2, d3, d4)
	result[2][0] = matrixDetInternal(a2, a3, a4, b2, b3, b4, d2, d3, d4)
	result[3][0] = -matrixDetInternal(a2, a3, a4, b2, b3, b4, c2, c3, c4)

	result[0][1] = -matrixDetInternal(b1, b3, b4, c1, c3, c4, d1, d3, d4)
	result[1][1] = matrixDetInternal(a1, a3, a4, c1, c3, c4, d1, d3, d4)
	result[2][1] = -matrixDetInternal(a1, a3, a4, b1, b3, b4, d1, d3, d4)
	result[3][1] = matrixDetInternal(a1, a3, a4, b1, b3, b4, c1, c3, c4)

	result[0][2] = matrixDetInternal(b1, b2, b4, c1, c2, c4, d1, d2, d4)
	result[1][2] = -matrixDetInternal(a1, a2, a4, c1, c2, c4, d1, d2, d4)
	result[2][2] = matrixDetInternal(a1, a2, a4, b1, b2, b4, d1, d2, d4)
	result[3][2] = -matrixDetInternal(a1, a2, a4, b1, b2, b4, c1, c2, c4)

	result[0][3] = -matrixDetInternal(b1, b2, b3, c1, c2, c3, d1, d2, d3)
	result[1][3] = matrixDetInternal(a1, a2, a3, c1, c2, c3, d1, d2, d3)
	result[2][3] = -matrixDetInternal(a1, a2, a3, b1, b2, b3, d1, d2, d3)
	result[3][3] = matrixDetInternal(a1, a2, a3, b1, b2, b3, c1, c2, c3)
	return result
}

// scaleMatrix calculation.
func scaleMatrix(m Matrix, factor float64) Matrix {
	var result Matrix
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			result[i][j] = m[i][j] * factor
		}
	}
	return result
}

// invertMatrix calculation.
func invertMatrix(m Matrix) Matrix {
	det := matrixDeterminant(m)
	epsilon := 1e-40
	if math.Abs(det) < epsilon {
		return identityHmgMatrix
	}
	m2 := adjointMatrix(m)
	return scaleMatrix(m2, 1/det)
}

// multiplyMatrix operation.
func multiplyMatrix(m1, m2 Matrix) Matrix {
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
