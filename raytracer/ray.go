package raytracer

// DIFF -
const DIFF = 0.000000001

// IntersectionTriangle defines the ratcast triangle intersection result
type IntersectionTriangle struct {
	Hit                bool
	N1                 Vector
	N2                 Vector
	N3                 Vector
	T1                 Vector
	T2                 Vector
	T3                 Vector
	P1                 Vector
	P2                 Vector
	P3                 Vector
	Material           Material
	Intersection       Vector
	IntersectionNormal Vector
	RayStart           Vector
	ObjectName         string
}

func (i *IntersectionTriangle) getTexCoords() Vector {
	u, v, w, _ := BarycentricCoordinates(i.P1, i.P2, i.P3, i.Intersection)
	tex := Vector{
		u*i.T1[0] + v*i.T2[0] + w*i.T3[0],
		u*i.T1[1] + v*i.T2[1] + w*i.T3[1],
	}
	return tex
}

// ScreenToWorld -
func ScreenToWorld(x, y, width, height int64, camera Vector, proj, view Matrix) (rayDir Vector) {
	var xF, yF float64
	xF = (2.0*float64(x))/float64(width) - 1.0
	yF = 1.0 - (2.0*float64(y))/float64(height)

	rayStart := Vector{xF, yF, 1.0, 1.0}
	invProj := InvertMatrix(proj)
	eyeCoordsRay := VectorTransform(rayStart, invProj)
	eyeCoords := Vector{eyeCoordsRay[0], eyeCoordsRay[1], -1.0, 0.0}
	invView := InvertMatrix(view)
	rayEnd := VectorTransform(eyeCoords, invView)
	rayDir = NormalizeVector(rayEnd)
	return rayDir
}

func raycastTriangleIntersect(start, vector, p1, p2, p3 Vector) (intersection, normal Vector, hit bool) {
	v1 := SubVector(p2, p1)
	v2 := SubVector(p3, p1)
	pVec := CrossProduct(vector, v2)
	det := dot(v1, pVec)
	if (det < DIFF) && (det > -DIFF) {
		hit = false
		return
	}
	invDet := 1 / det
	tvec := SubVector(start, p1)
	u := dot(tvec, pVec) * invDet
	if (u < 0) || (u > 1) {
		hit = false
		return
	}
	qvec := CrossProduct(tvec, v1)
	v := dot(vector, qvec) * invDet
	res := (v > 0) && (u+v <= 1.0)
	if !res {
		hit = false
		return
	}
	t := dot(v2, qvec) * invDet
	if t <= 0 {
		hit = false
		return
	}
	intersection = Combine(start, vector, 1.0, t)
	intersection[3] = 1.0
	normal = CrossProduct(v1, v2)
	normal = NormalizeVector(normal)
	hit = true
	return
}

func raycastObjectIntersect(object Object, rayStart, rayDir Vector) (intersection IntersectionTriangle) {
	var p1, p2, p3, n1, n2, n3, t1, t2, t3 Vector
	var face indice
	var shortestDist float64 = -1.0
	for matName := range object.Materials {
		for indice := range object.Materials[matName].Indices {
			face = object.Materials[matName].Indices[indice]
			p1 = object.Vertices[face[0]]
			p2 = object.Vertices[face[1]]
			p3 = object.Vertices[face[2]]
			intersectionPoint, intersectionNormal, hit := raycastTriangleIntersect(rayStart, rayDir, p1, p2, p3)
			if hit {
				n1 = object.Normals[face[0]]
				n2 = object.Normals[face[1]]
				n3 = object.Normals[face[2]]
				if len(object.TexCoords) > 0 {
					t1 = object.TexCoords[face[0]]
					t2 = object.TexCoords[face[1]]
					t3 = object.TexCoords[face[2]]
				}
				dist := VectorDistance(intersectionPoint, rayStart)
				if (shortestDist == -1) || (dist < shortestDist) {
					intersection.Hit = true
					intersection.IntersectionNormal = intersectionNormal
					intersection.T3 = t3
					intersection.T2 = t2
					intersection.T1 = t1
					intersection.N3 = n3
					intersection.N2 = n2
					intersection.N1 = n1
					intersection.P3 = p3
					intersection.P2 = p2
					intersection.P1 = p1
					intersection.Intersection = intersectionPoint
					intersection.Material = object.Materials[matName]
					intersection.RayStart = rayStart
					shortestDist = dist
				}
			}
		}
	}
	return
}
