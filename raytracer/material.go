package raytracer

import (
	"image"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Images map to hold image data in memory for repeating images.
var Images map[string][][]Vector

// BumpMapNormals cache to hold bump map information in memory.
var BumpMapNormals map[string][][]Vector

type indice [4]int64

// Material definition.
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

func loadImage(scenePath, texture string) (imageHasAlpha bool) {
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

	imgBounds := src.Bounds().Max
	Images[textureName] = make([][]Vector, imgBounds.X)
	for i := 0; i < imgBounds.X; i++ {
		Images[textureName][i] = make([]Vector, imgBounds.Y)
		for j := 0; j < imgBounds.Y; j++ {
			r, g, b, a := src.At(i, j).RGBA()
			r, g, b, a = r>>8, g>>8, b>>8, a>>8

			result := Vector{
				float64(r) / 255,
				float64(g) / 255,
				float64(b) / 255,
				float64(a) / 255,
			}
			if result[3] < 1 {
				imageHasAlpha = true
			}

			Images[textureName][i][j] = result
		}
	}
	imageFile.Close()

	log.Printf("Image %s loaded: Alpha %t", texture, imageHasAlpha)
	return imageHasAlpha
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
	imgBounds := src.Bounds().Max
	BumpMapNormals[texture] = make([][]Vector, imgBounds.X)
	for i := 0; i < imgBounds.X; i++ {
		BumpMapNormals[texture][i] = make([]Vector, imgBounds.Y)
		for j := 0; j < imgBounds.Y; j++ {
			r, g, b, _ := src.At(i, j).RGBA()
			r, g, b = r>>8, g>>8, b>>8

			bump := normalizeVector(Vector{
				(float64(r) / 255),
				(float64(g) / 255),
				(float64(b) / 255),
				1,
			})

			BumpMapNormals[texture][i][j] = normalizeVector(subVector(scaleVector(bump, 2), Vector{1, 1, 1, 0}))
		}
	}
	imageFile.Close()
}
