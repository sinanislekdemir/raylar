package raytracer

import (
	"encoding/json"
	_ "image/jpeg" // fuck you go-linter
	_ "image/png"  // fuck you go-linter
	"io/ioutil"
	"log"
	"time"

	"github.com/cheggaaa/pb"
)

// Light -
type Light struct {
	Position      Vector  `json:"position"`
	Color         Vector  `json:"color"`
	Active        bool    `json:"active"`
	LightStrength float64 `json:"light_strength"`
	Directional   bool    `json:"directional_light"`
	Direction     Vector  `json:"direction"`
	Samples       []Vector
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

// Scene -
type Scene struct {
	Objects        map[string]*Object `json:"objects"`
	MasterObject   *Object
	Lights         []Light    `json:"lights"`
	Observers      []Observer `json:"observers"`
	Pixels         [][]PixelStorage
	Width          int
	Height         int
	ShortRadius    float64
	InputFilename  string
	OutputFilename string
}

// Init scene
func (s *Scene) Init(sceneFile string, configFile string) error {
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
	return s.loadJSON(sceneFile)
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
	for name, obj := range s.Objects {
		fixObjectVectorW(obj)
		obj.calcRadius()
		s.Objects[name] = obj
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
	s.processObjects()
	s.mergeAll()
	s.parseMaterials()
	s.fixLightPos()
	s.loadLights()
	s.prepareMatrices()
	s.scanPixels()
	if GlobalConfig.RenderCaustics {
		s.buildPhotonMap()
	}
	log.Printf("Done init scene")
}

func (scene *Scene) prepareMatrices() {
	view := viewMatrix(scene.Observers[0].Position, scene.Observers[0].Target, scene.Observers[0].Up)
	projectionMatrix := perspectiveProjection(
		scene.Observers[0].Fov,
		float64(scene.Width)/float64(scene.Height),
		scene.Observers[0].Near,
		scene.Observers[0].Far,
	)
	if scene.Observers[0].Projection == nil {
		scene.Observers[0].Projection = &projectionMatrix
	}

	scene.Observers[0].view = view
	scene.Observers[0].width = scene.Width
	scene.Observers[0].height = scene.Height
}

func (scene *Scene) scanPixels() {
	log.Printf("Scanning pixels on view")
	bar := pb.StartNew(scene.Width * scene.Height)
	scene.Pixels = make([][]PixelStorage, scene.Width)
	for i := 0; i < scene.Width; i++ {
		scene.Pixels[i] = make([]PixelStorage, scene.Height)
		for j := 0; j < scene.Height; j++ {
			scene.Pixels[i][j].Color = GlobalConfig.TransparentColor
		}
	}

	for i := 0; i < scene.Width; i++ {
		for j := 0; j < scene.Height; j++ {
			rayDir := screenToWorld(i, j, scene.Width, scene.Height, scene.Observers[0].Position, *scene.Observers[0].Projection, scene.Observers[0].view)
			bestHit := raycastSceneIntersect(scene, scene.Observers[0].Position, rayDir)
			scene.Pixels[i][j].WorldLocation = bestHit
			bar.Increment()
		}
	}
	bar.Finish()
	log.Printf("Done scanning pixels")
}

func (s *Scene) buildPhotonMap() {
	log.Print("Building photon map")
	buildPhotonMap(s)
}

func (s *Scene) loadLights() {
	// for i := range s.Lights {
	// 	s.Lights[i].HitExceptions =
	// }
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
// so we need to set them to 1.0
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
// So, we won't have to multiply matrices each time
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

// TODO: This is a bit heavy, refactor
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
