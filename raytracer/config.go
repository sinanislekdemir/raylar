package raytracer

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// Config keeps Raytracer Configuration
type Config struct {
	SamplerLimit             int     `json:"sampler_limit"`
	LightSampleCount         int     `json:"light_sample_count"`
	CausticsSamplerLimit     int     `json:"caustics_samples"`
	Exposure                 float64 `json:"exposure"`
	MaxReflectionDepth       int     `json:"max_reflection_depth"`
	RayCorrection            float64 `json:"ray_correction"`
	OcclusionRate            float64 `json:"occlusion_rate"`
	AmbientRadius            float64 `json:"ambient_occlusion_radius"`
	RenderOcclusion          bool    `json:"render_occlusion"`
	RenderLights             bool    `json:"render_lights"`
	RenderColors             bool    `json:"render_colors"`
	RenderAmbientColors      bool    `json:"render_ambient_color"`
	RenderCaustics           bool    `json:"render_caustics"`
	AmbientColorSharingRatio float64 `json:"ambient_color_ratio"`
	RenderReflections        bool    `json:"render_reflections"`
	RenderRefractions        bool    `json:"render_refractions"`
	PhotonSpacing            float64 `json:"photon_spacing"`
	Width                    int     `json:"width"`
	Height                   int     `json:"height"`
	EdgeDetechThreshold      float64 `json:"edge_detect_threshold"`
	MergeAll                 bool    `json:"merge_all"`
	AntialiasSamples         int     `json:"antialias_samples"`
}

var DEFAULT = Config{
	// Default Config Settings
	SamplerLimit:             16,
	LightSampleCount:         16,
	CausticsSamplerLimit:     10000,
	PhotonSpacing:            0.005,
	Exposure:                 0.2,
	MaxReflectionDepth:       3,
	RayCorrection:            0.002,
	OcclusionRate:            0.2,
	AmbientRadius:            2.1,
	RenderOcclusion:          true,
	RenderLights:             true,
	RenderColors:             true,
	RenderAmbientColors:      true,
	RenderCaustics:           false,
	AmbientColorSharingRatio: 0.5,
	RenderReflections:        true,
	RenderRefractions:        true,
	Width:                    1600,
	Height:                   900,
	EdgeDetechThreshold:      0.2,
	MergeAll:                 false,
	AntialiasSamples:         8,
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
