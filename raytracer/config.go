package raytracer

// Config keeps Raytracer Configuration
type Config struct {
	SamplerLimit             int     `json:"sampler_limit"`
	CausticsSamplerLimit     int     `json:"caustics_samples"`
	LightHardLimit           float64 `json:"light_distance"`
	Exposure                 float64 `json:"exposure"`
	MaxReflectionDepth       int     `json:"max_reflection_depth"`
	RayCorrection            float64 `json:"ray_correction"`
	OcclusionRate            float64 `json:"occlusion_rate"`
	AmbientRadius            float64 `json:"ambient_occlusion_radius"`
	RenderOcclusion          bool    `json:"render_occlusion"`
	RenderLights             bool    `json:"render_lights"`
	RenderColors             bool    `json:"render_colors"`
	RenderAmbientColors      bool    `json:"render_ambient_color"`
	AmbientColorSharingRatio float64 `json:"ambient_color_ratio"`
	RenderReflections        bool    `json:"render_reflections"`
	RenderRefractions        bool    `json:"render_refractions"`
	CausticsThreshold        float64 `json:"caustics_threshold"`
	Width                    int     `json:"width"`
	Height                   int     `json:"height"`
	EdgeDetechThreshold      float64 `json:"edge_detect_threshold"`
}

var DEFAULT = Config{
	// Default Config Settings
	SamplerLimit:             16,
	CausticsSamplerLimit:     10000,
	LightHardLimit:           100,
	Exposure:                 0.2,
	MaxReflectionDepth:       6,
	RayCorrection:            0.002,
	OcclusionRate:            0.2,
	AmbientRadius:            2.1,
	RenderOcclusion:          true,
	RenderLights:             true,
	RenderColors:             true,
	RenderAmbientColors:      true,
	AmbientColorSharingRatio: 0.5,
	RenderReflections:        true,
	RenderRefractions:        true,
	CausticsThreshold:        0,
	Width:                    1600,
	Height:                   900,
	EdgeDetechThreshold:      0.2,
}
