package main

import (
	"github.com/sinanislekdemir/raylar/raytracer"
)

func main() {
	s := raytracer.Scene{}
	s.LoadConfig("config.json")
	s.LoadJSON("scene.json")

	// fx, _ := os.Create("test.prof")
	// pprof.StartCPUProfile(fx)
	// defer pprof.StopCPUProfile()

	// raytracer.Render(&s, 1600, 900, 0, 0, 0, 0)
	// raytracer.Render(&s, 800, 450, 0, 0, 0, 0)
	// raytracer.Render(&s, 800, 450, 130, 132, 194, 200)
	raytracer.Render(&s, 0, 0, 0, 0)
	// raytracer.Render(&s, 200, 112, 0, 0, 0, 0)
	// raytracer.Render(&s, 800, 450, 327, 332, 215, 220)
	// raytracer.Render(&s, 160, 90, 0, 0, 0, 0)
}
