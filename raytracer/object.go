package raytracer

import "log"

var totalNodes = 0
var maxDepth = 0
var idCounter int64 = 0

// Object -
type Object struct {
	Vertices  []Vector            `json:"vertices"`
	Normals   []Vector            `json:"normals"`
	TexCoords []Vector            `json:"texcoords"`
	Matrix    Matrix              `json:"matrix"`
	Materials map[string]Material `json:"materials"`
	Children  map[string]*Object  `json:"children"`
	Triangles []Triangle
	Root      Node
	radius    float64
}

// UnifyTriangles of the object for faster processing
func (o *Object) UnifyTriangles() {
	for matName := range o.Materials {
		for indice := range o.Materials[matName].Indices {
			triangle := Triangle{}
			triangle.id = idCounter + 1
			idCounter++
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

			triangle.Smooth = face[3] == 1
			triangle.Material = o.Materials[matName]
			o.Triangles = append(o.Triangles, triangle)
		}
	}
	log.Printf("Loaded object with %d triangles", len(o.Triangles))
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
		fixObjectVectorW(obj)
		o.Children[name] = obj
	}
}

func (o *Object) calcRadius() {
	min, max := calculateBounds(o.Vertices)
	o.radius = vectorDistance(max, min)
}
