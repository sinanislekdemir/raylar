package raytracer

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg" // fuck you go-linter
	"io/ioutil"
	"log"
	"os"
)

type indice [3]int64

// Material -
type Material struct {
	Color   Vector   `json:"color"`
	Display int64    `json:"display"`
	Texture string   `json:"texture"`
	Opacity float64  `json:"opactiy"`
	Indices []indice `json:"indices"`
}

// Object -
type Object struct {
	Vertices  []Vector            `json:"vertices"`
	Normals   []Vector            `json:"normals"`
	TexCoords []Vector            `json:"texcoords"`
	Matrix    Matrix              `json:"matrix"`
	Materials map[string]Material `json:"materials"`
	Children  map[string]Object   `json:"children"`
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

// Light -
type Light struct {
	Position Vector `json:"position"`
	Color    Vector `json:"color"`
	Active   bool   `json:"active"`
}

// Observer -
type Observer struct {
	Position    Vector  `json:"position"`
	Target      Vector  `json:"target"`
	Up          Vector  `json:"up"`
	Fov         float64 `json:"fov"`
	AspectRatio float64 `json:"aspect_ratio"`
	Zoom        float64 `json:"zoom"`
	Near        float64 `json:"near"`
	Far         float64 `json:"far"`
	Perspective bool    `json:"perspective"`
	projection  Matrix
	view        Matrix
	width       int64
	height      int64
}

// Scene -
type Scene struct {
	Objects   map[string]Object `json:"objects"`
	Lights    []Light           `json:"lights"`
	Observers []Observer        `json:"observers"`
	DepthMap  [][]float64
	ImageMap  map[string]image.Image
}

// LoadJSON -
func (s *Scene) LoadJSON(jsonFile string) error {
	log.Printf("Loading file: %s\n", jsonFile)
	file, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		log.Fatalf("Error while reading file: %s", err.Error())
		return err
	}
	log.Printf("Unmarshal JSON\n")
	err = json.Unmarshal(file, &s)
	if err != nil {
		log.Fatalf("Error unmarshalling %s", err.Error())
		return err
	}
	log.Printf("Fixing object Ws\n")
	for name, obj := range s.Objects {
		fixObjectVectorW(&obj)
		s.Objects[name] = obj
	}
	log.Printf("Flatten Scene Objects\n")
	s.Objects = flattenSceneObjects(s.Objects)
	log.Printf("Transform object vertices to absolute")
	s.Objects = transformObjectsToAbsolute(s.Objects)
	s.ImageMap = parseMaterials(s.Objects)
	fmt.Print(s.Objects["cube2"].Materials["default"].Color)
	return nil
}

// Parse all material images and store them in scene object
// so we won't have to open and read for each pixel.
// TODO: Free material image if it is not being used.
// NOTE: This function assumes that objects are already flattened!
func parseMaterials(objects map[string]Object) map[string]image.Image {
	result := make(map[string]image.Image)
	for k := range objects {
		for m := range objects[k].Materials {
			mat := objects[k].Materials[m]
			if _, ok := result[mat.Texture]; ok {
				continue
			}
			if mat.Texture != "" {
				inFile, err := os.Open(mat.Texture)
				if err != nil {
					log.Printf("Material texture [%s] can't be opened for material [%s]\n", mat.Texture, m)
					inFile.Close()
					continue
				}
				src, _, err := image.Decode(inFile)
				if err != nil {
					log.Printf("Error reading image file [%s]: [%s]\n", mat.Texture, err.Error())
					inFile.Close()
					continue
				}
				result[mat.Texture] = src
				inFile.Close()
			}
		}
	}
	return result
}

// Flatten Scene Objects and move them to root
// So, we won't have to multiply matrices each time
func flattenSceneObjects(objects map[string]Object) map[string]Object {
	result := make(map[string]Object)
	for k := range objects {
		result[k] = objects[k]
		if len(objects[k].Children) > 0 {
			flatList := flattenSceneObjects(objects[k].Children)
			for subKey := range flatList {
				subObj := flatList[subKey]
				subObj.Matrix = MultiplyMatrix(subObj.Matrix, objects[k].Matrix)
				result[k+subKey] = subObj
			}
		}
	}
	return result
}

// TODO: Refactor
func transformObjectToAbsolute(vertices []Vector, matrix Matrix) []Vector {
	result := make([]Vector, len(vertices))
	for i := 0; i < len(vertices); i++ {
		result[i] = VectorTransform(vertices[i], matrix)
	}
	return result
}

// TODO: This is a bit heavy, refactor
func transformObjectsToAbsolute(objects map[string]Object) map[string]Object {
	for k := range objects {
		absoluteVertices := transformObjectToAbsolute(objects[k].Vertices, objects[k].Matrix)
		for i := 0; i < len(absoluteVertices); i++ {
			objects[k].Vertices[i] = absoluteVertices[i]
		}
	}
	return objects
}
