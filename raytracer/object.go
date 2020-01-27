package raytracer

var totalNodes = 0
var maxDepth = 0

type indice [3]int64

// Material -
type Material struct {
	Color      Vector   `json:"color"`
	Display    int64    `json:"display"`
	Texture    string   `json:"texture"`
	Opacity    float64  `json:"opactiy"`
	Indices    []indice `json:"indices"`
	Glossiness float64  `json:"glossiness"`
	Light      bool     `json:"light"`
}

// Triangle definition
// raycasting is already expensive and trying to calculate the triangle
// in each raycast makes it harder. So we are simplifying triangle definition
type Triangle struct {
	P1       Vector
	P2       Vector
	P3       Vector
	N1       Vector
	N2       Vector
	N3       Vector
	T1       Vector
	T2       Vector
	T3       Vector
	Material Material
}

func (t *Triangle) equals(dest Triangle) bool {
	return t.P1 == dest.P1 && t.P2 == dest.P2 && t.P3 == dest.P3
}

func (t *Triangle) midPoint() Vector {
	mid := t.P1
	mid = addVector(mid, t.P2)
	mid = addVector(mid, t.P3)
	mid = scaleVector(mid, 1.0/3.0)
	return mid
}

func (t *Triangle) getBoundingBox() BoundingBox {
	result := BoundingBox{}
	result.MinExtend = t.P1
	result.MaxExtend = t.P1
	for i := 0; i < 3; i++ {
		if t.P2[i] < result.MinExtend[i] {
			result.MinExtend[i] = t.P2[i]
		}
		if t.P2[i] > result.MaxExtend[i] {
			result.MaxExtend[i] = t.P2[i]
		}
		if t.P3[i] < result.MinExtend[i] {
			result.MinExtend[i] = t.P3[i]
		}
		if t.P3[i] > result.MaxExtend[i] {
			result.MaxExtend[i] = t.P3[i]
		}
	}
	return result
}

// Object -
type Object struct {
	Vertices  []Vector            `json:"vertices"`
	Normals   []Vector            `json:"normals"`
	TexCoords []Vector            `json:"texcoords"`
	Matrix    Matrix              `json:"matrix"`
	Materials map[string]Material `json:"materials"`
	Children  map[string]Object   `json:"children"`
	Triangles []Triangle
	Root      Node
	radius    float64
}

// UnifyTriangles of the object for faster processing
func (o *Object) UnifyTriangles() {
	for matName := range o.Materials {
		for indice := range o.Materials[matName].Indices {
			triangle := Triangle{}
			face := o.Materials[matName].Indices[indice]
			triangle.P1 = o.Vertices[face[0]]
			triangle.P2 = o.Vertices[face[1]]
			triangle.P3 = o.Vertices[face[2]]
			if len(o.TexCoords) > 0 {
				triangle.T1 = o.TexCoords[face[0]]
				triangle.T2 = o.TexCoords[face[1]]
				triangle.T3 = o.TexCoords[face[2]]
			}
			triangle.N1 = o.Normals[face[0]]
			triangle.N2 = o.Normals[face[1]]
			triangle.N3 = o.Normals[face[2]]
			triangle.Material = o.Materials[matName]
			o.Triangles = append(o.Triangles, triangle)
		}
	}
	o.Vertices = nil
	o.Normals = nil
	o.TexCoords = nil
}

// KDTree Building
func (o *Object) KDTree() {
	o.Root = generateNode(&o.Triangles, 0)
}

func fixObjectVectorW(o *Object) {
	for i := range o.Vertices {
		o.Vertices[i][3] = 1.0
	}
	for i := range o.Normals {
		o.Normals[i][3] = 0.0
	}
	for i := range o.TexCoords {
		o.TexCoords[i][2] = 0.0
		o.TexCoords[i][3] = 0.0
	}
	for name, obj := range o.Children {
		fixObjectVectorW(&obj)
		o.Children[name] = obj
	}
}

func (o *Object) calcRadius() {
	min, max := calculateBounds(o.Vertices)
	o.radius = vectorDistance(max, min)
}
