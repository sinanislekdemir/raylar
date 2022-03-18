package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/sinanislekdemir/raylar/raytracer"
)

var buildTime string

func main() {
	s := raytracer.Scene{}

	sceneFile := ""

	configFile := flag.String("config", "", "Scene Config JSON")
	outputFilename := flag.String("output", "awesome.png", "Render output image filename")
	environmentMap := flag.String("environment", "", "Environment map image file for infinite reflections")
	percent := flag.Int("percent", 100, "Render completion percentage")
	size := flag.String("size", "", "width x height: Eg: 1600x900")
	left := flag.Int("left", 0, "Left X")
	right := flag.Int("right", 0, "Right X")
	top := flag.Int("top", 0, "Top")
	bottom := flag.Int("bottom", 0, "Bottom")
	profiling := flag.Bool("profile", false, "Set 1 for debugging")
	showHelp := flag.Bool("help", false, "Show help!")
	createConfig := flag.Bool("createconfig", false, "Create config")

	flag.Parse()

	if flag.NArg() == 0 {
		sceneFile = "scene.json"
	} else {
		sceneFile = flag.Arg(0)
		if sceneFile == "" {
			sceneFile = "scene.json"
		}
	}

	if showHelp != nil && *showHelp {
		fmt.Println("--config <config.json>  : Render configurations")
		fmt.Println("--output <out.png>      : Output image filename")
		fmt.Println("--percent <percent>     : Render Percentage")
		fmt.Println("--profile               : Turn on profiling for golang")
		fmt.Println("--size <width>x<height> : Set width x height explicitly, overwriting config. 1600x900 eg.")
		fmt.Println("--createconfig          : Create a default config.json to modify scene parameters")
		fmt.Println("--environment           : Environment map image file for infinite reflections")
		os.Exit(0)
	}

	if outputFilename == nil {
		s.OutputFilename = "out.png"
	} else {
		s.OutputFilename = *outputFilename
	}

	if *profiling {
		fx, _ := os.Create("profiling.prof")
		_ = pprof.StartCPUProfile(fx)
		defer pprof.StopCPUProfile()
	}

	fmt.Printf("Raylar - Build %s", buildTime)
	if *createConfig {
		err := raytracer.CreateConfig("config.json")
		if err != nil {
			fmt.Println(err.Error())
		}
		return
	}

	if _, err := os.Stat("config.json"); os.IsNotExist(err) {
		err := raytracer.CreateConfig("config.json")
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		cf := "config.json"
		configFile = &cf
	}

	err := s.Init(sceneFile, *configFile, *environmentMap)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Printf("Render %d percent of the image", *percent)
	raytracer.GlobalConfig.Percentage = *percent
	_ = raytracer.Render(&s, *left, *right, *top, *bottom, *percent, size)
}
