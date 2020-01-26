package raytracer

import (
	"math"
)

// DIFF -
const DIFF = 0.000000001

// MDIFF -
const MDIFF = -0.000000001

// IntersectionTriangle defines the ratcast triangle intersection result
type IntersectionTriangle struct {
	Hit                bool
	Triangle           Triangle
	Intersection       Vector
	IntersectionNormal Vector
	RayStart           Vector
	RayDir             Vector
	ObjectName         string
	Dist               float64
	Hits               int
}

func (i *IntersectionTriangle) getTexCoords() Vector {
	u, v, w, _ := barycentricCoordinates(i.Triangle.P1, i.Triangle.P2, i.Triangle.P3, i.Intersection)
	tex := Vector{
		u*i.Triangle.T1[0] + v*i.Triangle.T2[0] + w*i.Triangle.T3[0],
		u*i.Triangle.T1[1] + v*i.Triangle.T2[1] + w*i.Triangle.T3[1],
	}
	return tex
}

// ScreenToWorld -
func ScreenToWorld(x, y, width, height int, camera Vector, proj, view Matrix) (rayDir Vector) {
	var xF, yF float64
	xF = (2.0*float64(x))/float64(width) - 1.0
	yF = 1.0 - (2.0*float64(y))/float64(height)

	rayStart := Vector{xF, yF, 1.0, 1.0}
	invProj := invertMatrix(proj)
	eyeCoordsRay := vectorTransform(rayStart, invProj)
	eyeCoords := Vector{eyeCoordsRay[0], eyeCoordsRay[1], -1.0, 0.0}
	invView := invertMatrix(view)
	rayEnd := vectorTransform(eyeCoords, invView)
	rayDir = normalizeVector(rayEnd)
	return rayDir
}

func raycastTriangleIntersect(start, vector, p1, p2, p3 Vector) (intersection Vector, hit bool) {
	v1 := subVector(p2, p1)
	v2 := subVector(p3, p1)
	pVec := crossProduct(vector, v2)
	det := dot(v1, pVec)
	if (det < DIFF) && (det > -DIFF) {
		hit = false
		return
	}
	invDet := 1 / det
	tvec := subVector(start, p1)
	u := dot(tvec, pVec) * invDet
	if (u < 0) || (u > 1) {
		hit = false
		return
	}
	qvec := crossProduct(tvec, v1)
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
	intersection = combine(start, vector, 1.0, t)
	intersection[3] = 1.0
	hit = true
	return
}

func raycastSphereIntersect(start, vector, center Vector, radius float64) bool {
	proj := projectPoint(center, start, vector)
	if proj < 0 {
		proj = 0.0
	}
	vc := combine(start, vector, 1.0, proj)
	dist := math.Pow(vectorDistance(center, vc), 2)
	return dist < math.Pow(radius, 2)
}

func raycastBoxIntersect(rayStart, rayVector Vector, boundingBox BoundingBox) bool {
	right := 0.0
	left := 1.0
	middle := 2.0
	inside := true
	quadrant := Vector{}
	whichPlane := 0
	maxT := Vector{}
	candidatePlane := Vector{}
	for i := 0; i < 3; i++ {
		if rayStart[i] < boundingBox.MinExtend[i] {
			quadrant[i] = left
			candidatePlane[i] = boundingBox.MinExtend[i]
			inside = false
		} else if rayStart[i] > boundingBox.MaxExtend[i] {
			quadrant[i] = right
			candidatePlane[i] = boundingBox.MaxExtend[i]
			inside = false
		} else {
			quadrant[i] = middle
		}
	}
	if inside {
		return true
	}

	for i := 0; i < 3; i++ {
		if quadrant[i] != middle && rayVector[i] != 0 {
			maxT[i] = (candidatePlane[i] - rayStart[i]) / rayVector[i]
		} else {
			maxT[i] = -1
		}
	}
	whichPlane = 0
	for i := 1; i < 3; i++ {
		if maxT[whichPlane] < maxT[i] {
			whichPlane = i
		}
	}
	if maxT[whichPlane] < 0 {
		return false
	}
	coord := Vector{}
	for i := 0; i < 3; i++ {
		if whichPlane != i {
			coord[i] = rayStart[i] + maxT[whichPlane]*rayVector[i]
			if coord[i] < boundingBox.MinExtend[i] || coord[i] > boundingBox.MaxExtend[i] {
				return false
			}
			coord[i] = candidatePlane[i]
		}
	}
	return true
}

func raycastNodeIntersect(rayStart, rayDir Vector, node *Node, intersection *IntersectionTriangle) {
	if !raycastBoxIntersect(rayStart, rayDir, node.getBoundingBox()) {
		return
	}

	if node.Left.TriangleCount > 0 || node.Right.TriangleCount > 0 {
		raycastNodeIntersect(rayStart, rayDir, node.Left, intersection)
		raycastNodeIntersect(rayStart, rayDir, node.Right, intersection)
		return
	}

	for i := range node.Triangles {
		// if dot(rayDir, node.Triangles[i].N1) > 0 {
		// 	continue
		// }
		intersectionPoint, hit := raycastTriangleIntersect(
			rayStart,
			rayDir,
			node.Triangles[i].P1,
			node.Triangles[i].P2,
			node.Triangles[i].P3,
		)
		if hit {
			intersection.Hits++
			dist := vectorDistance(intersectionPoint, rayStart)
			if dist > 0 && (intersection.Dist == -1 || dist < intersection.Dist) {
				intersection.Hit = true
				intersection.IntersectionNormal = node.Triangles[i].N1
				intersection.Intersection = intersectionPoint
				intersection.Triangle = node.Triangles[i]
				intersection.RayStart = rayStart
				intersection.RayDir = rayDir
				intersection.Dist = dist
			}
		}
	}
}

func raycastObjectIntersect(object Object, rayStart, rayDir Vector) (intersection IntersectionTriangle) {
	if !raycastSphereIntersect(rayStart, rayDir, object.Matrix[3], object.radius) {
		// early exit
		return
	}
	intersection.Dist = -1
	raycastNodeIntersect(rayStart, rayDir, &object.Root, &intersection)
	return
}

func raycastSceneIntersect(scene *Scene, position, ray Vector) IntersectionTriangle {
	var bestHit IntersectionTriangle
	var bestDist float64
	bestDist = -1
	// slightly change position to avoid self collision
	position = addVector(position, scaleVector(ray, scene.Config.RayCorrection))
	intersectChannel := make(chan IntersectionTriangle, len(scene.Objects))
	for k := range scene.Objects {
		go func(object Object, name string, position, ray Vector, c chan IntersectionTriangle) {
			result := raycastObjectIntersect(object, position, ray)
			result.ObjectName = name
			c <- result
		}(scene.Objects[k], k, position, ray, intersectChannel)
	}
	totalHits := 0
	for i := 0; i < len(scene.Objects); i++ {
		intersect := <-intersectChannel
		if !intersect.Hit {
			continue
		}

		if intersect.Dist < DIFF {
			intersect.Hit = false
			continue
		}
		totalHits += intersect.Hits
		if (bestDist == -1) || (intersect.Dist < bestDist) {
			bestHit = intersect
			bestDist = intersect.Dist
		}
	}
	bestHit.RayDir = ray
	bestHit.Dist = bestDist
	bestHit.Hits = totalHits

	return bestHit
}
