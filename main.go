package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/sinanislekdemir/raylar/raytracer"
)

func main() {
	s := raytracer.Scene{}

	sceneFile := flag.String("scene", "scene.json", "Scene File JSON")
	configFile := flag.String("config", "config.json", "Scene Config JSON")
	outputFilename := flag.String("output", "out.png", "Render output image filename")
	percent := flag.Int("percent", 100, "Render completion percentage")
	showHelp := flag.Bool("help", false, "Show help!")
	flag.Parse()
	if showHelp != nil && *showHelp {
		fmt.Println("--scene <scene.json>   : Scene filename")
		fmt.Println("--config <config.json> : Render configurations")
		fmt.Println("--output <out.png>     : Output image filename")
		fmt.Println("--percent <percent>    : Render Percentage")
		os.Exit(0)
	}
	if configFile != nil {
		_ = s.LoadConfig(*configFile)
	}
	if outputFilename == nil {
		s.OutputFilename = "out.png"
	} else {
		s.OutputFilename = *outputFilename
	}
	if sceneFile != nil {
		_ = s.LoadJSON(*sceneFile)
		if s.Config.Profiling {
			fx, _ := os.Create("profiling.prof")
			_ = pprof.StartCPUProfile(fx)
			defer pprof.StopCPUProfile()
		}
	}
	log.Printf("Render %d percent of the image", *percent)
	_ = raytracer.Render(&s, 0, 0, 0, 0, *percent)
}
