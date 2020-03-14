package raytracer

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// Config keeps Raytracer Configuration
type Config struct {
	AmbientColorSharingRatio float64 `json:"ambient_color_ratio"`
	AmbientRadius            float64 `json:"ambient_occlusion_radius"`
	AntialiasSamples         int     `json:"antialias_samples"`
	CausticsSamplerLimit     int     `json:"caustics_samples"`
	EdgeDetechThreshold      float64 `json:"edge_detect_threshold"`
	EnvironmentMap           string  `json:"environment_map"`
	Exposure                 float64 `json:"exposure"`
	Height                   int     `json:"height"`
	LightSampleCount         int     `json:"light_sample_count"`
	MaxReflectionDepth       int     `json:"max_reflection_depth"`
	OcclusionRate            float64 `json:"occlusion_rate"`
	PhotonSpacing            float64 `json:"photon_spacing"`
	RayCorrection            float64 `json:"ray_correction"`
	RenderAmbientColors      bool    `json:"render_ambient_color"`
	RenderBumpMap            bool    `json:"render_bump_map"`
	RenderCaustics           bool    `json:"render_caustics"`
	RenderColors             bool    `json:"render_colors"`
	RenderLights             bool    `json:"render_lights"`
	RenderOcclusion          bool    `json:"render_occlusion"`
	RenderReflections        bool    `json:"render_reflections"`
	RenderRefractions        bool    `json:"render_refractions"`
	SamplerLimit             int     `json:"sampler_limit"`
	TransparentColor         Vector  `json:"transparent_color"`
	Width                    int     `json:"width"`
	Percentage               int
}

var DEFAULT = Config{
	// Default Config Settings
	AmbientColorSharingRatio: 0.5,
	AmbientRadius:            2.1,
	AntialiasSamples:         8,
	CausticsSamplerLimit:     10000,
	EdgeDetechThreshold:      0.7,
	Exposure:                 0.2,
	Height:                   900,
	LightSampleCount:         16,
	MaxReflectionDepth:       3,
	OcclusionRate:            0.2,
	Percentage:               100,
	PhotonSpacing:            0.005,
	RayCorrection:            0.002,
	RenderAmbientColors:      true,
	RenderBumpMap:            true,
	RenderCaustics:           false,
	RenderColors:             true,
	RenderLights:             true,
	RenderOcclusion:          true,
	RenderReflections:        true,
	RenderRefractions:        true,
	SamplerLimit:             16,
	TransparentColor:         Vector{0, 0, 0, 0},
	Width:                    1600,
}

var GlobalConfig = Config{}

// LoadConfig file for the render
func loadConfig(jsonFile string) error {
	var config Config
	log.Printf("Loading configuration from %s", jsonFile)
	file, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		log.Printf("Error while reading file: %s", err.Error())
		return nil
	}
	log.Printf("Unmarshal JSON\n")
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatalf("Error unmarshalling %s", err.Error())
		return err
	}
	GlobalConfig = config
	return nil
}
