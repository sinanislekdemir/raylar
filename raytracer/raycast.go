package raytracer

// DIFF -
const DIFF = 0.000000001

// MDIFF -
const MDIFF = -0.000000001

// ScreenToWorld -
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

// This is the branchless check but it doesn't help much with the performance.
// func raycastBoxIntersect(rayStart, rayVector *Vector, boundingBox *BoundingBox) bool {
// 	inv := Vector{
// 		1.0 / rayVector[0],
// 		1.0 / rayVector[1],
// 		1.0 / rayVector[2],
// 	}
// 	t1 := (boundingBox.MinExtend[0] - rayStart[0]) * inv[0]
// 	t2 := (boundingBox.MaxExtend[0] - rayStart[0]) * inv[0]

// 	tmin := math.Min(t1, t2)
// 	tmax := math.Max(t1, t2)
// 	for i := 1; i < 3; i++ {
// 		t1 = (boundingBox.MinExtend[i] - rayStart[i]) * inv[i]
// 		t2 = (boundingBox.MaxExtend[i] - rayStart[i]) * inv[i]

// 		tmin = math.Max(tmin, math.Min(t1, t2))
// 		tmax = math.Min(tmax, math.Max(t1, t2))
// 	}
// 	return tmax > math.Max(tmin, 0)
// }

func raycastBoxIntersect(rayStart, rayVector *Vector, boundingBox *BoundingBox) bool {
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

func raycastNodeIntersect(rayStart, rayDir *Vector, node *Node, intersection *Intersection) {
	if !raycastBoxIntersect(rayStart, rayDir, node.getBoundingBox()) {
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
	var bestHit Intersection
	var bestDist float64

	bestDist = -1

	position = addVector(position, scaleVector(ray, GlobalConfig.RayCorrection))
	intersectChannel := make(chan Intersection, len(scene.Objects))
	for k := range scene.Objects {
		go func(object *Object, name string, position, ray Vector, c chan Intersection) {
			result := raycastObjectIntersect(object, &position, &ray)
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
