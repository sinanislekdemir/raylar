package raytracer

import "math"

// BoundingBox -
type BoundingBox [2]Vector

// Node for KDTree
type Node struct {
	Triangles     []Triangle
	TriangleCount int
	BoundingBox   *BoundingBox
	Left          *Node
	Right         *Node
	depth         int
}

func (b *BoundingBox) extend(o BoundingBox) {
	b[0][0] = math.Min(b[0][0], o[0][0])
	b[0][1] = math.Min(b[0][1], o[0][1])
	b[0][2] = math.Min(b[0][2], o[0][2])

	b[1][0] = math.Max(b[1][0], o[1][0])
	b[1][1] = math.Max(b[1][1], o[1][1])
	b[1][2] = math.Max(b[1][2], o[1][2])
}

func (b *BoundingBox) extendVector(v Vector) {
	b[0][0] = math.Min(b[0][0], v[0])
	b[0][1] = math.Min(b[0][1], v[1])
	b[0][2] = math.Min(b[0][2], v[2])

	b[1][0] = math.Max(b[1][0], v[0])
	b[1][1] = math.Max(b[1][1], v[1])
	b[1][2] = math.Max(b[1][2], v[2])
}

func (b *BoundingBox) longestAxis() int {
	result := 0
	fdiff := 0.0
	dist := b[1][0] - b[0][0]
	for i := 1; i < 3; i++ {
		fdiff = b[1][i] - b[0][i]
		if fdiff > dist {
			result = i
		}
	}
	return result
}

func (b *BoundingBox) center() Vector {
	return Vector{
		(b[1][0] - b[0][0]) / 2.0,
		(b[1][1] - b[0][1]) / 2.0,
		(b[1][2] - b[0][2]) / 2.0,
		1,
	}
}

func (b *BoundingBox) inside(v Vector) bool {
	return (v[0]+DIFF >= b[0][0] && v[0]-DIFF <= b[1][0] &&
		v[1]+DIFF >= b[0][1] && v[1]-DIFF <= b[1][1] &&
		v[2]+DIFF >= b[0][2] && v[2]-DIFF <= b[1][2])
}

func (n *Node) getBoundingBox() *BoundingBox {
	if n.Triangles == nil || len(n.Triangles) == 0 {
		n.BoundingBox = &BoundingBox{}
		return n.BoundingBox
	}
	bb := n.Triangles[0].getBoundingBox()
	n.BoundingBox = &bb
	if len(n.Triangles) == 1 {
		return n.BoundingBox
	}
	for i := 1; i < len(n.Triangles); i++ {
		bb.extend(n.Triangles[i].getBoundingBox())
	}
	n.BoundingBox = &bb
	return n.BoundingBox
}

func (n *Node) midPoint() Vector {
	mid := Vector{}
	if n.Triangles == nil {
		return mid
	}
	for i := 0; i < len(n.Triangles); i++ {
		mid = addVector(n.Triangles[i].midPoint(), mid)
	}
	return scaleVector(mid, 1.0/float64(len(n.Triangles)))
}

func generateNode(tris *[]Triangle, depth int) (result Node) {
	totalNodes++
	if depth > maxDepth {
		maxDepth = depth
	}
	result.Triangles = *tris
	result.TriangleCount = len(result.Triangles)
	result.Left = nil
	result.Right = nil
	result.depth = depth
	result.getBoundingBox()

	if result.Triangles == nil || len(result.Triangles) == 0 {
		return
	}

	if len(result.Triangles) == 1 {
		result.Left = &Node{}
		result.Right = &Node{}
		return
	}

	midP := result.midPoint()
	leftTris := make([]Triangle, 0)
	rightTris := make([]Triangle, 0)

	axis := result.BoundingBox.longestAxis()
	for i := range result.Triangles {
		mp := result.Triangles[i].midPoint()
		if midP[axis] >= mp[axis] {
			rightTris = append(rightTris, result.Triangles[i])
		} else {
			leftTris = append(leftTris, result.Triangles[i])
		}
	}

	if len(leftTris) == 0 && len(rightTris) > 0 {
		leftTris = rightTris
	}
	if len(rightTris) == 0 && len(leftTris) > 0 {
		rightTris = leftTris
	}

	matches := 0
	ratio := true
	if len(result.Triangles) < 10000 {
		for i := range leftTris {
			for j := range rightTris {
				if leftTris[i].equals(rightTris[j]) {
					matches++
					ratio = (float64(matches)/float64(len(leftTris)) < 0.5 && float64(matches)/float64(len(rightTris)) < 0.5)
					if !ratio {
						break
					}
				}
			}
			if !ratio {
				break
			}
		}
	}

	if ratio && depth < 50 {
		leftNode := generateNode(&leftTris, depth+1)
		rightNode := generateNode(&rightTris, depth+1)
		result.Left = &leftNode
		result.Right = &rightNode
		result.Triangles = nil
	} else {
		result.Left = &Node{}
		result.Right = &Node{}
	}
	return result
}
