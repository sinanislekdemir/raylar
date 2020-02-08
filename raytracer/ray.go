package raytracer

import "math"

// DIFF -
const DIFF = 0.000000001

// MDIFF -
const MDIFF = -0.000000001

// IntersectionTriangle defines the ratcast triangle intersection result
type IntersectionTriangle struct {
	Hit                bool
	Triangle           *Triangle
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

func (i *IntersectionTriangle) getColor(scene *Scene, depth int) Vector {
	if !i.Hit {
		return Vector{
			0, 0, 0, 1,
		}
	}

	material := i.Triangle.Material
	result := material.Color
	if material.Texture != "" {
		if img, ok := scene.ImageMap[material.Texture]; ok {
			// ok, we have the image. Let's calculate the pixel color;
			s := i.getTexCoords()
			// get image size
			imgBounds := img.Bounds().Max

			if s[0] > 1 {
				s[0] = s[0] - math.Floor(s[0])
			}
			if s[0] < 0 {
				s[0] = math.Abs(s[0])
				s[0] = 1 - (s[0] - math.Floor(s[0]))
			}

			if s[1] > 1 {
				s[1] = s[1] - math.Floor(s[1])
			}

			if s[1] < 0 {
				s[1] = math.Abs(s[1])
				s[1] = 1 - (s[1] - math.Floor(s[1]))
			}
			s[1] = 1 - s[1]

			s[0] = s[0] - float64(int64(s[0]))
			s[1] = s[1] - float64(int64(s[1]))

			pixelX := int(float64(imgBounds.X) * s[0])
			pixelY := int(float64(imgBounds.Y) * s[1])
			r, g, b, a := img.At(pixelX, pixelY).RGBA()
			r, g, b, a = r>>8, g>>8, b>>8, a>>8

			result = Vector{
				float64(r) / 255,
				float64(g) / 255,
				float64(b) / 255,
				float64(a) / 255,
			}
		}
	}
	if material.Glossiness > 0 && scene.Config.RenderReflections && depth < scene.Config.MaxReflectionDepth {
		reflectColor := calculateReflectionColor(scene, i, depth+1)
		result = Vector{
			result[0]*(1.0-material.Glossiness) + (reflectColor[0] * material.Glossiness),
			result[1]*(1.0-material.Glossiness) + (reflectColor[1] * material.Glossiness),
			result[2]*(1.0-material.Glossiness) + (reflectColor[2] * material.Glossiness),
			result[3]*(1.0-material.Glossiness) + (reflectColor[3] * material.Glossiness),
		}
	}
	if material.Transmission > 0 && scene.Config.RenderRefractions && depth < scene.Config.MaxReflectionDepth {
		refractionColor := calculateRefractionColor(scene, i, depth+1)
		result = Vector{
			result[0]*(1.0-material.Transmission) + (refractionColor[0] * material.Transmission),
			result[1]*(1.0-material.Transmission) + (refractionColor[1] * material.Transmission),
			result[2]*(1.0-material.Transmission) + (refractionColor[2] * material.Transmission),
			result[3]*(1.0-material.Transmission) + (refractionColor[3] * material.Transmission),
		}
	}
	return result
}

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
	if sameSideTest(*normal, *vector) {
		iNormal := scaleVector(*normal, -1)
		normal = &iNormal
	}
	hit = true
	return
}

func max(x, y float64) float64 {
	// I know that math.Min and math.Max exists but it has extra condition checks and I don't want them
	// at this code.
	if x > y {
		return x
	}
	return y
}

func min(x, y float64) float64 {
	// I know that math.Min and math.Max exists but it has extra condition checks and I don't want them
	// at this code.
	if x < y {
		return x
	}
	return y
}

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

func raycastBoxIntersect2(rs, rv *Vector, bb *BoundingBox) bool {
	dirFrac := Vector{
		1.0 / rv[0],
		1.0 / rv[1],
		1.0 / rv[2],
		1,
	}
	t1 := (bb.MinExtend[0] - rs[0]) * dirFrac[0]
	t2 := (bb.MaxExtend[0] - rs[0]) * dirFrac[0]
	t3 := (bb.MinExtend[1] - rs[1]) * dirFrac[1]
	t4 := (bb.MaxExtend[1] - rs[1]) * dirFrac[1]
	t5 := (bb.MinExtend[2] - rs[2]) * dirFrac[2]
	t6 := (bb.MaxExtend[2] - rs[2]) * dirFrac[2]

	tmin := max(max(min(t1, t2), min(t3, t4)), min(t5, t6))
	tmax := min(min(max(t1, t2), max(t3, t4)), max(t5, t6))
	if (tmax < 0) || (tmin > tmax) {
		return false
	}

	return true
}

func raycastNodeIntersect(rayStart, rayDir *Vector, node *Node, intersection *IntersectionTriangle) {
	if !raycastBoxIntersect(rayStart, rayDir, node.getBoundingBox()) {
		return
	}

	if node.Left.TriangleCount > 0 || node.Right.TriangleCount > 0 {
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
			}
		}
	}
}

func raycastObjectIntersect(object *Object, rayStart, rayDir *Vector) (intersection IntersectionTriangle) {
	// if !raycastSphereIntersect(rayStart, rayDir, object.Matrix[3], object.radius) {
	// 	// early exit
	// 	return
	// }
	intersection.Dist = -1
	raycastNodeIntersect(rayStart, rayDir, &object.Root, &intersection)
	return
}

func raycastSceneIntersect(scene *Scene, position, ray Vector) IntersectionTriangle {
	var bestHit IntersectionTriangle
	var bestDist float64

	bestDist = -1

	position = addVector(position, scaleVector(ray, scene.Config.RayCorrection))
	intersectChannel := make(chan IntersectionTriangle, len(scene.Objects))
	for k := range scene.Objects {
		go func(object *Object, name string, position, ray Vector, c chan IntersectionTriangle) {
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
