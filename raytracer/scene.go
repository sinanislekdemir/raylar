package raytracer

import (
	"encoding/json"
	"image"
	_ "image/jpeg" // fuck you go-linter
	_ "image/png"  // fuck you go-linter
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/cheggaaa/pb"
)

// EnvironmentMap cache.
var EnvironmentMap [][]Vector
var hasEnvironmentMap bool

// Light structure.
type Light struct {
	Position      Vector  `json:"position"`
	Color         Vector  `json:"color"`
	Active        bool    `json:"active"`
	LightStrength float64 `json:"light_strength"`
	Directional   bool    `json:"directional_light"`
	Direction     Vector  `json:"direction"`
	Samples       []Vector
}

// Camera structure.
type Camera struct {
	Position    Vector  `json:"position"`
	Target      Vector  `json:"target"`
	Up          Vector  `json:"up"`
	Fov         float64 `json:"fov"`
	AspectRatio float64 `json:"aspect_ratio"`
	Zoom        float64 `json:"zoom"`
	Near        float64 `json:"near"`
	Far         float64 `json:"far"`
	Perspective bool    `json:"perspective"`
	Projection  *Matrix `json:"projection"`
	view        Matrix
	width       int
	height      int
}

// PixelStorage to Store pixel information before turning it into a png
// we need to do this for post-processing.
type PixelStorage struct {
	WorldLocation     Intersection
	DirectLightEnergy Vector
	Color             Vector
	AmbientColor      Vector
	Depth             float64
	X                 int
	Y                 int
}

// Scene main structure.
type Scene struct {
	Objects        map[string]*Object `json:"objects"`
	MasterObject   *Object
	Lights         []Light  `json:"lights"`
	Cameras        []Camera `json:"observers"`
	Pixels         [][]PixelStorage
	Width          int
	Height         int
	ShortRadius    float64
	InputFilename  string
	OutputFilename string
}

// Init scene.
func (s *Scene) Init(sceneFile, configFile, environmentMap string) error {
	log.Print("Initializing the scene")
	if configFile == "" {
		log.Print("No config set, setting defaults")
		GlobalConfig = DEFAULT
	} else {
		err := loadConfig(configFile)
		if err != nil {
			return err
		}
	}
	if GlobalConfig.EnvironmentMap != "" && environmentMap == "" {
		environmentMap = GlobalConfig.EnvironmentMap
	}
	if environmentMap != "" {
		s.loadEnvironmentMap(environmentMap)
	}
	return s.loadJSON(sceneFile)
}

func (s *Scene) loadEnvironmentMap(mapFilename string) {
	imageFile, err := os.Open(mapFilename)
	if err != nil {
		log.Printf("Environment Map [%s] can't be opened\n", mapFilename)
		imageFile.Close() // defer has over-head
		return
	}
	src, _, err := image.Decode(imageFile)
	if err != nil {
		log.Printf("Error reading image file [%s]: [%s]\n", mapFilename, err.Error())
		imageFile.Close()
		return
	}

	imgBounds := src.Bounds().Max
	EnvironmentMap = make([][]Vector, imgBounds.X)
	for i := 0; i < imgBounds.X; i++ {
		EnvironmentMap[i] = make([]Vector, imgBounds.Y)
		for j := 0; j < imgBounds.Y; j++ {
			r, g, b, a := src.At(i, j).RGBA()
			r, g, b, a = r>>8, g>>8, b>>8, a>>8

			result := Vector{
				float64(r) / 255,
				float64(g) / 255,
				float64(b) / 255,
				float64(a) / 255,
			}
			EnvironmentMap[i][j] = result
		}
	}
	imageFile.Close()
	hasEnvironmentMap = true
}

func (s *Scene) loadJSON(jsonFile string) error {
	start := time.Now()
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
	s.InputFilename = jsonFile
	log.Printf("Fixing object Ws\n")
	for name := range s.Objects {
		s.Objects[name].fixW()
		s.Objects[name].calcRadius()
	}

	log.Printf("Loaded scene in %f seconds\n", time.Since(start).Seconds())
	return nil
}

func (s *Scene) mergeAll() {
	gigaMesh := Object{
		Matrix: identityHmgMatrix,
	}
	gigaMesh.Materials = make(map[string]Material)
	gigaMesh.Triangles = make([]Triangle, 0)
	for obj := range s.Objects {
		for k, m := range s.Objects[obj].Materials {
			gigaMesh.Materials[k] = m
		}
		gigaMesh.Triangles = append(gigaMesh.Triangles, s.Objects[obj].Triangles...)
		s.Objects[obj] = nil
	}
	gigaMesh.calcRadius()
	log.Printf("Build KDTree")
	gigaMesh.KDTree()
	log.Printf("Built %d nodes with %d max depth, object ready", totalNodes, maxDepth)
	s.Objects = nil
	s.MasterObject = &gigaMesh
}

func (s *Scene) prepare(width, height int) {
	s.Width = width
	s.Height = height
	// Order of below calls is important!
	log.Printf("Init scene")
	s.flatten()
	// log.Printf("After flatten")
	// PrintMemUsage()
	s.processObjects()
	// log.Printf("After objects processing")
	// PrintMemUsage()
	s.mergeAll()
	// log.Printf("After mergeall")
	// PrintMemUsage()
	s.parseMaterials()
	s.fixLightPos()
	s.loadLights()
	s.prepareMatrices()
	log.Printf("After parse materials")
	PrintMemUsage()
	s.scanPixels()
	log.Printf("When we prep scene")
	PrintMemUsage()
	if GlobalConfig.RenderCaustics {
		s.buildPhotonMap()
	}
	log.Printf("Done init scene")
}

func (s *Scene) prepareMatrices() {
	view := viewMatrix(s.Cameras[0].Position, s.Cameras[0].Target, s.Cameras[0].Up)
	projectionMatrix := perspectiveProjection(
		s.Cameras[0].Fov,
		float64(s.Width)/float64(s.Height),
		s.Cameras[0].Near,
		s.Cameras[0].Far,
	)
	if s.Cameras[0].Projection == nil {
		s.Cameras[0].Projection = &projectionMatrix
	}

	s.Cameras[0].view = view
	s.Cameras[0].width = s.Width
	s.Cameras[0].height = s.Height
}

func (s *Scene) scanPixels() {
	log.Printf("Scanning pixels on view")
	bar := pb.StartNew(s.Width * s.Height)
	s.Pixels = make([][]PixelStorage, s.Width)
	for i := 0; i < s.Width; i++ {
		s.Pixels[i] = make([]PixelStorage, s.Height)
		for j := 0; j < s.Height; j++ {
			s.Pixels[i][j].Color = GlobalConfig.TransparentColor
		}
	}
	log.Println("After pixel storage")
	PrintMemUsage()

	for i := 0; i < s.Width; i++ {
		for j := 0; j < s.Height; j++ {
			rayDir := screenToWorld(i, j, s.Width, s.Height, s.Cameras[0].Position, *s.Cameras[0].Projection, s.Cameras[0].view)
			bestHit := raycastSceneIntersect(s, s.Cameras[0].Position, rayDir)
			s.Pixels[i][j].WorldLocation = bestHit
			bar.Increment()
		}
	}
	log.Printf("After pixel raycasts")
	PrintMemUsage()
	bar.Finish()
	log.Printf("Done scanning pixels")
}

func (s *Scene) buildPhotonMap() {
	log.Print("Building photon map")
	buildPhotonMap(s)
}

func (s *Scene) loadLights() {
	for i := range s.MasterObject.Triangles {
		if !s.MasterObject.Triangles[i].Material.Light {
			continue
		}
		mat := s.MasterObject.Triangles[i].Material
		lights := sampleTriangle(s.MasterObject.Triangles[i], GlobalConfig.LightSampleCount)
		strength := s.MasterObject.Triangles[i].Material.LightStrength
		for li := range lights {
			light := Light{
				Position:      lights[li],
				Color:         mat.Color,
				Active:        true,
				LightStrength: strength,
				// HitExceptions: make(map[int64]bool),
			}
			s.Lights = append(s.Lights, light)
		}
	}
}

// Lights have 0 as w but they are not vectors, they are positions;
// so we need to set them to 1.0.
func (s *Scene) fixLightPos() {
	for i := range s.Lights {
		s.Lights[i].Position[3] = 1.0
	}
}

func (s *Scene) flatten() {
	log.Printf("Flatten Scene Objects\n")
	s.Objects = flattenSceneObjects(s.Objects)
}

// Flatten Scene Objects and move them to root
// So, we won't have to multiply matrices each time.
func flattenSceneObjects(objects map[string]*Object) map[string]*Object {
	result := make(map[string]*Object)
	for k := range objects {
		result[k] = objects[k]
		if len(objects[k].Children) > 0 {
			flatList := flattenSceneObjects(objects[k].Children)
			for subKey := range flatList {
				subObj := flatList[subKey]
				subObj.Matrix = multiplyMatrix(subObj.Matrix, objects[k].Matrix)
				result[k+subKey] = subObj
			}
		}
	}
	return result
}

// TODO: This is a bit heavy, refactor.
func (s *Scene) processObjects() {
	log.Printf("Transform object vertices to absolute and build KDTrees")

	for k := range s.Objects {
		log.Printf("Prepare object %s", k)
		obj := s.Objects[k]
		log.Printf("Local to absolute")
		absoluteVertices := localToAbsoluteList(obj.Vertices, obj.Matrix)
		for i := 0; i < len(absoluteVertices); i++ {
			obj.Vertices[i] = absoluteVertices[i]
		}
		log.Printf("Unify triangles")
		obj.UnifyTriangles()
		totalNodes = 0
		maxDepth = 0
		s.Objects[k] = obj
	}
}

// Parse all material images and store them in scene object
// so we won't have to open and read for each pixel.
// TODO: Free material image if it is not being used.
// TODO: This method is complex and has more than one responsibility
// NOTE: This function assumes that objects are already flattened!
func (s *Scene) parseMaterials() {
	log.Printf("Parse material textures\n")
	scenePath := filepath.Dir(s.InputFilename)
	BumpMapNormals = make(map[string][][]Vector)
	Images = make(map[string][][]Vector)
	for m := range s.MasterObject.Materials {
		mat := s.MasterObject.Materials[m]
		if _, ok := Images[mat.Texture]; ok {
			continue
		}
		if mat.Texture != "" {
			loadImage(scenePath, mat.Texture)
			loadBumpMap(scenePath, mat.Texture)
		}
	}
}
