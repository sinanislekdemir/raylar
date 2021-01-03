package raytracer

// DIFF floating point precision is a killing me.
const DIFF = 0.000000001

// MDIFF Imagine a CPU with no dangling float precision.
const MDIFF = -0.000000001

// ScreenToWorld conversion.
func screenToWorld(x, y, width, height int, camera Vector, proj, view Matrix) (rayDir Vector) {
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

func raycastTriangleIntersect(start, vector, p1, p2, p3 *Vector) (intersection, normal *Vector, hit bool) {
	v1 := psubVector(p2, p1)
	v2 := psubVector(p3, p1)
	pVec := pcrossProduct(vector, v2)
	det := pdot(v1, pVec)
	if (det < DIFF) && (det > -DIFF) {
		hit = false
		return
	}
	invDet := 1 / det
	tvec := psubVector(start, p1)
	u := pdot(tvec, pVec) * invDet
	if (u < 0) || (u > 1) {
		hit = false
		return
	}
	qvec := pcrossProduct(tvec, v1)
	v := pdot(vector, qvec) * invDet
	res := (v > 0) && (u+v <= 1.0)
	if !res {
		hit = false
		return
	}
	t := pdot(v2, qvec) * invDet
	if t <= 0 {
		hit = false
		return
	}
	intersection = pcombine(start, vector, 1.0, t)
	intersection[3] = 1.0
	normal = pnormalizeVector(pcrossProduct(v1, v2))
	if sameSideTest(*normal, *vector, 0) {
		iNormal := scaleVector(*normal, -1)
		normal = &iNormal
	}
	hit = true
	return
}

func raycastBoxIntersect(rayStart, rayVector *Vector, boundingBox *BoundingBox) bool {
	// IMPORTANT NOTE!
	// DO NOT FOR-LOOP HERE!!! It looks like a easy and simple idea to loop from 0 to 3 here but
	// believe me, when it does 1 million checks, it really matters!!!!!!
	// reduce the branches as much as we can!
	right := 0.0
	left := 1.0
	middle := 2.0
	inside := true
	quadrant := Vector{middle, middle, middle}
	whichPlane := 0
	candidatePlane := Vector{}

	if rayStart[0] < boundingBox[0][0] {
		quadrant[0] = left
		candidatePlane[0] = boundingBox[0][0]
		inside = false
	} else if rayStart[0] > boundingBox[1][0] {
		quadrant[0] = right
		candidatePlane[0] = boundingBox[1][0]
		inside = false
	}

	if rayStart[1] < boundingBox[0][1] {
		quadrant[1] = left
		candidatePlane[1] = boundingBox[0][1]
		inside = false
	} else if rayStart[1] > boundingBox[1][1] {
		quadrant[1] = right
		candidatePlane[1] = boundingBox[1][1]
		inside = false
	}

	if rayStart[2] < boundingBox[0][2] {
		quadrant[2] = left
		candidatePlane[2] = boundingBox[0][2]
		inside = false
	} else if rayStart[2] > boundingBox[1][2] {
		quadrant[2] = right
		candidatePlane[2] = boundingBox[1][2]
		inside = false
	}

	if inside {
		return true
	}
	maxT := Vector{-1, -1, -1}

	if quadrant[0] != middle && rayVector[0] != 0 {
		maxT[0] = (candidatePlane[0] - rayStart[0]) / rayVector[0]
	}

	if quadrant[1] != middle && rayVector[1] != 0 {
		maxT[1] = (candidatePlane[1] - rayStart[1]) / rayVector[1]
	}

	if quadrant[2] != middle && rayVector[2] != 0 {
		maxT[2] = (candidatePlane[2] - rayStart[2]) / rayVector[2]
	}

	if maxT[whichPlane] < maxT[1] {
		whichPlane = 1
	}

	if maxT[whichPlane] < maxT[2] {
		whichPlane = 2
	}

	if maxT[whichPlane] < 0 {
		return false
	}

	a := 0.0

	if whichPlane != 0 {
		a = rayStart[0] + maxT[whichPlane]*rayVector[0]
		if a < boundingBox[0][0] || a > boundingBox[1][0] {
			return false
		}
	}

	if whichPlane != 1 {
		a = rayStart[1] + maxT[whichPlane]*rayVector[1]
		if a < boundingBox[0][1] || a > boundingBox[1][1] {
			return false
		}
	}

	if whichPlane != 2 {
		a = rayStart[2] + maxT[whichPlane]*rayVector[2]
		if a < boundingBox[0][2] || a > boundingBox[1][2] {
			return false
		}
	}

	return true
}

func raycastNodeIntersect(rayStart, rayDir *Vector, node *Node, intersection *Intersection) {
	if !raycastBoxIntersect(rayStart, rayDir, node.BoundingBox) {
		return
	}

	if (node.Left != nil && node.Right != nil) && (node.Left.TriangleCount > 0 || node.Right.TriangleCount > 0) {
		raycastNodeIntersect(rayStart, rayDir, node.Left, intersection)
		raycastNodeIntersect(rayStart, rayDir, node.Right, intersection)
		return
	}

	for i := range node.Triangles {
		intersectionPoint, normal, hit := raycastTriangleIntersect(
			rayStart,
			rayDir,
			&node.Triangles[i].P1,
			&node.Triangles[i].P2,
			&node.Triangles[i].P3,
		)
		if hit {
			intersection.Hits++
			dist := pvectorDistance(intersectionPoint, rayStart)
			if node.Triangles[i].Material.Texture != "" {
				temp := Intersection{
					Hit:                true,
					IntersectionNormal: *normal,
					Intersection:       *intersectionPoint,
					Triangle:           &node.Triangles[i],
					RayDir:             *rayDir,
					RayStart:           *rayStart,
					Dist:               dist,
				}
				if temp.getColor()[3] < 1 {
					continue
				}
			}
			if dist > 0 && (intersection.Dist == -1 || dist < intersection.Dist) {
				intersection.Hit = true
				intersection.IntersectionNormal = *normal
				intersection.Intersection = *intersectionPoint
				intersection.Triangle = &node.Triangles[i]
				intersection.RayStart = *rayStart
				intersection.RayDir = *rayDir
				intersection.Dist = dist
				intersection.getNormal()
			}
		}
	}
}

func raycastObjectIntersect(object *Object, rayStart, rayDir *Vector) (intersection Intersection) {
	intersection.Dist = -1
	raycastNodeIntersect(rayStart, rayDir, &object.Root, &intersection)
	return
}

func raycastSceneIntersect(scene *Scene, position, ray Vector) Intersection {
	position = addVector(position, scaleVector(ray, GlobalConfig.RayCorrection))
	intersect := raycastObjectIntersect(scene.MasterObject, &position, &ray)
	intersect.RayDir = ray
	if !intersect.Hit {
		return intersect
	}

	if intersect.Dist < DIFF {
		intersect.Hit = false
		return intersect
	}
	intersect.RayDir = ray

	return intersect
}
