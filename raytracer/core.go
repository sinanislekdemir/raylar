package raytracer

import (
	"image"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cheggaaa/pb"
)

// Render -
func Render(scene *Scene, left, right, top, bottom, percent int, size *string) error {
	var err error
	width := GlobalConfig.Width
	height := GlobalConfig.Height
	if size != nil && strings.Contains(*size, "x") {
		log.Printf("Set size to %s", *size)
		split := strings.Split(*size, "x")
		width, err = strconv.Atoi(split[0])
		if err != nil {
			return err
		}
		height, err = strconv.Atoi(split[1])
		if err != nil {
			return err
		}
	}
	log.Printf("Start rendering scene\n")
	scene.prepare(width, height)
	start := time.Now()

	upLeft := image.Point{0, 0}
	log.Printf("Output image size: %d x %d\n", width, height)
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// Set color for each pixel.

	actualWidth := width
	actualHeight := height
	scene.Width = width
	scene.Height = height

	if right+left > 0 {
		actualWidth = right - left
		actualHeight = bottom - top
	}

	totalPixels := actualWidth * actualHeight

	pixellist := make([]int, totalPixels)
	for i := 0; i < totalPixels; i++ {
		pixellist[i] = i
	}

	rand.Shuffle(totalPixels, func(i, j int) { pixellist[i], pixellist[j] = pixellist[j], pixellist[i] })
	if percent < 100 {
		to := percent * totalPixels / 100
		pixellist = pixellist[:to]
		totalPixels = to
	}
	bar := pb.StartNew(totalPixels)

	for i := 0; i < totalPixels; i++ {
		y := int(math.Floor(float64(pixellist[i])/float64(actualWidth))) + top
		x := (pixellist[i] % actualWidth) + left
		renderPixel(scene, x, y)
		bar.Increment()
	}
	bar.Finish()

	log.Printf("Rendered scene in %f seconds\n", time.Since(start).Seconds())
	log.Printf("Post processing and saving file")
	renderImage(scene, img)
	// Encode as PNG.
	f, _ := os.Create(scene.OutputFilename)
	return png.Encode(f, img)
}
