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

func getWidthHeight(size string) (int, int, error) {
	var err error
	width := GlobalConfig.Width
	height := GlobalConfig.Height
	if strings.Contains(size, "x") {
		log.Printf("Set size to %s", size)
		split := strings.Split(size, "x")
		width, err = strconv.Atoi(split[0])
		if err != nil {
			return 0, 0, err
		}
		height, err = strconv.Atoi(split[1])
		if err != nil {
			return 0, 0, err
		}
	}
	return width, height, nil
}

func getPixelList(width, height, left, right, top, bottom, percent int) (int, []int) {
	actualWidth := width
	actualHeight := height
	if right+left > 0 {
		actualWidth = right - left
		actualHeight = bottom - top
	}

	totalPixels := actualWidth * actualHeight

	pixelList := make([]int, totalPixels)
	for i := 0; i < totalPixels; i++ {
		pixelList[i] = i
	}
	rand.Shuffle(totalPixels, func(i, j int) { pixelList[i], pixelList[j] = pixelList[j], pixelList[i] })

	to := percent * totalPixels / 100
	if percent < 100 {
		pixelList = pixelList[:to]
		totalPixels = to
	}
	return totalPixels, pixelList
}

// Render the scene, main processor.
func Render(scene *Scene, left, right, top, bottom, percent int, size *string) error {
	width, height, err := getWidthHeight(*size)
	if err != nil {
		return err
	}

	log.Printf("Start rendering scene\n")
	scene.prepare(width, height)
	start := time.Now()

	upLeft := image.Point{X: 0, Y: 0}
	log.Printf("Initial rendering: %d x %d\n", width, height)
	lowRight := image.Point{X: width, Y: height}

	img := image.NewRGBA(image.Rectangle{Min: upLeft, Max: lowRight})

	// Set color for each pixel.

	scene.Width = width
	scene.Height = height

	totalPixels, pixellist := getPixelList(width, height, left, right, top, bottom, percent)

	bar := pb.StartNew(totalPixels)

	for i := 0; i < totalPixels; i++ {
		y := int(math.Floor(float64(pixellist[i])/float64(width))) + top
		x := (pixellist[i] % width) + left
		renderPixel(scene, x, y)
		bar.Increment()
	}
	bar.Finish()

	log.Printf("Rendered scene in %f seconds\n", time.Since(start).Seconds())
	log.Printf("Second pass for antialiasing and image generation")
	renderImage(scene, img)
	// Encode as PNG.
	f, _ := os.Create(scene.OutputFilename)
	return png.Encode(f, img)
}
