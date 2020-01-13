package main

import (
	"github.com/sinanislekdemir/raylar/raytracer"
)

func main() {
	f := "scene.json"
	s := raytracer.Scene{}
	s.LoadJSON(f)
	raytracer.Render(&s, 1600, 1200)
}
