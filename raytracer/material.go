package raytracer

import (
	"image"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var ImageMap map[string]image.Image
var BumpMap map[string]image.Image

type indice [4]int64

// Material -
type Material struct {
	Color             Vector   `json:"color"`
	Texture           string   `json:"texture"`
	Transmission      float64  `json:"transmission"`
	IndexOfRefraction float64  `json:"index_of_refraction"`
	Indices           []indice `json:"indices"`
	Glossiness        float64  `json:"glossiness"`
	Roughness         float64  `json:"roughness"`
	Light             bool     `json:"light"`
	LightStrength     float64  `json:"light_strength"`
}

func loadImage(scenePath, texture string) {
	textureName := texture
	_, err := os.Stat(texture)
	if os.IsNotExist(err) {
		texture = filepath.Join(scenePath, texture)
	}
	imageFile, err := os.Open(texture)
	if err != nil {
		log.Printf("Material texture [%s] can't be opened\n", texture)
		imageFile.Close() // defer has over-head
		return
	}
	src, _, err := image.Decode(imageFile)
	if err != nil {
		log.Printf("Error reading image file [%s]: [%s]\n", texture, err.Error())
		imageFile.Close()
		return
	}
	ImageMap[textureName] = src
	imageFile.Close()
	log.Printf("Image %s loaded", texture)
}

func loadBumpMap(scenePath, texture string) {
	ext := filepath.Ext(texture)
	base := strings.TrimSuffix(texture, ext)
	texturePath := filepath.Dir(texture)
	bumpTexture := filepath.Join(texturePath, base+"_bump"+ext)
	_, err := os.Stat(bumpTexture)
	if os.IsNotExist(err) {
		bumpTexture = filepath.Join(scenePath, bumpTexture)
	}
	imageFile, err := os.Open(bumpTexture)
	if err != nil {
		imageFile.Close() // defer has over-head
		return
	}
	src, _, err := image.Decode(imageFile)
	if err != nil {
		log.Printf("Error reading image file [%s]: [%s]\n", texture, err.Error())
		imageFile.Close()
		return
	}
	log.Printf("Image Bump Map %s loaded", bumpTexture)
	BumpMap[texture] = src
	imageFile.Close()
}

// Parse all material images and store them in scene object
// so we won't have to open and read for each pixel.
// TODO: Free material image if it is not being used.
// TODO: This method is complex and has more than one responsibility
// NOTE: This function assumes that objects are already flattened!
func (s *Scene) parseMaterials() {
	log.Printf("Parse material textures\n")
	scenePath := filepath.Dir(s.InputFilename)
	ImageMap = make(map[string]image.Image)
	BumpMap = make(map[string]image.Image)
	for m := range s.MasterObject.Materials {
		mat := s.MasterObject.Materials[m]
		if _, ok := ImageMap[mat.Texture]; ok {
			continue
		}
		if mat.Texture != "" {
			loadImage(scenePath, mat.Texture)
			loadBumpMap(scenePath, mat.Texture)
		}
	}

}
