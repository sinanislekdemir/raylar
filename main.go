package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"

	"github.com/sinanislekdemir/raylar/raytracer"
)

func main() {
	s := raytracer.Scene{}

	sceneFile := flag.String("scene", "scene.json", "Scene File JSON")
	configFile := flag.String("config", "", "Scene Config JSON")
	outputFilename := flag.String("output", "awesome.png", "Render output image filename")
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
	if showHelp != nil && *showHelp {
		fmt.Println("--scene <scene.json>    : Scene filename")
		fmt.Println("--config <config.json>  : Render configurations")
		fmt.Println("--output <out.png>      : Output image filename")
		fmt.Println("--percent <percent>     : Render Percentage")
		fmt.Println("--profile               : Turn on profiling for golang")
		fmt.Println("--size <width>x<height> : Set width x height explicitly, overwriting config. 1600x900 eg.")
		fmt.Println("--createconfig          : Create a default config.json to modify scene parameters")
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
	if *createConfig {
		file, _ := json.MarshalIndent(raytracer.DEFAULT, "", " ")
		ferr := ioutil.WriteFile("config.json", file, 0644)
		if ferr != nil {
			log.Fatal(ferr.Error())
			os.Exit(1)
		}
		log.Printf("Created config.json")
		os.Exit(0)
	}
	err := s.Init(*sceneFile, *configFile)
	if err != nil {
		log.Fatal(err.Error())
		os.Exit(128)
	}
	log.Printf("Render %d percent of the image", *percent)
	raytracer.GlobalConfig.Percentage = *percent
	_ = raytracer.Render(&s, *left, *right, *top, *bottom, *percent, size)
}
