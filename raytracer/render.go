package raytracer

import (
	"image"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/cheggaaa/pb"
)

// Render -
func Render(scene *Scene, left, right, top, bottom int) error {
	width := scene.Config.Width
	height := scene.Config.Height
	log.Printf("Start rendering scene\n")
	start := time.Now()
	view := viewMatrix(scene.Observers[0].Position, scene.Observers[0].Target, scene.Observers[0].Up)
	projectionMatrix := perspectiveProjection(
		scene.Observers[0].Fov,
		scene.Observers[0].AspectRatio,
		scene.Observers[0].Near,
		scene.Observers[0].Far,
	)

	upLeft := image.Point{0, 0}
	log.Printf("Output image size: %d x %d\n", width, height)
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// Set color for each pixel.
	if scene.Observers[0].Projection == nil {
		scene.Observers[0].Projection = &projectionMatrix
	}

	scene.Observers[0].view = view
	scene.Observers[0].width = width
	scene.Observers[0].height = height

	actualWidth := width
	actualHeight := height
	scene.Width = width
	scene.Height = height

	if right+left > 0 {
		actualWidth = right - left
		actualHeight = bottom - top
	}

	// TODO: Multithread this part once everything seems to be better.
	// but for now, we will keep it as a single thread to make debugging easier.
	totalPixels := actualWidth * actualHeight
	scene.Pixels = make([][]PixelStorage, width)
	for i := 0; i < width; i++ {
		scene.Pixels[i] = make([]PixelStorage, height)
	}

	pixellist := make([]int, totalPixels)
	for i := 0; i < totalPixels; i++ {
		pixellist[i] = i
	}
	bar := pb.StartNew(totalPixels)

	rand.Shuffle(totalPixels, func(i, j int) { pixellist[i], pixellist[j] = pixellist[j], pixellist[i] })

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
