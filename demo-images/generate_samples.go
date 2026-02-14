package main

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
)

func createSunsetImage(path string, brightness float64) error {
	img := image.NewRGBA(image.Rect(0, 0, 256, 256))

	// Create sunset gradient
	for y := 0; y < 256; y++ {
		progress := float64(y) / 256.0
		r := uint8(135 + (255-135)*progress*brightness)
		g := uint8(206 + (100-206)*progress*brightness)
		b := uint8(235 + (50-235)*progress*brightness)
		for x := 0; x < 256; x++ {
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	// Add sun
	sunX, sunY := 128, 85
	sunRadius := 32
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			dx := x - sunX
			dy := y - sunY
			if dx*dx+dy*dy < sunRadius*sunRadius {
				img.Set(x, y, color.RGBA{255, 215, 0, 255})
			}
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return jpeg.Encode(f, img, &jpeg.Options{Quality: 90})
}

func createCatImage(path string) error {
	img := image.NewRGBA(image.Rect(0, 0, 256, 256))

	// Background
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			img.Set(x, y, color.RGBA{240, 248, 255, 255})
		}
	}

	// Simple cat silhouette
	catColor := color.RGBA{100, 100, 100, 255}

	// Body
	for y := 140; y < 200; y++ {
		for x := 88; x < 168; x++ {
			cx, cy := 128, 170
			dx, dy := float64(x-cx)/40, float64(y-cy)/30
			if dx*dx+dy*dy <= 1.0 {
				img.Set(x, y, catColor)
			}
		}
	}

	// Head
	for y := 70; y < 130; y++ {
		for x := 93; x < 163; x++ {
			cx, cy := 128, 100
			dx, dy := float64(x-cx)/35, float64(y-cy)/30
			if dx*dx+dy*dy <= 1.0 {
				img.Set(x, y, catColor)
			}
		}
	}

	// Left ear
	for y := 65; y < 85; y++ {
		for x := 93; x < 113; x++ {
			if x+y > 158 && x+3*y > 300 && 3*x+y < 420 {
				img.Set(x, y, catColor)
			}
		}
	}

	// Right ear
	for y := 65; y < 85; y++ {
		for x := 143; x < 163; x++ {
			if 508-x+y > 158 && 508-x+3*y > 300 && 768-3*x+y < 420 {
				img.Set(x, y, catColor)
			}
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return jpeg.Encode(f, img, &jpeg.Options{Quality: 90})
}

func main() {
	os.MkdirAll("/home/ubuntu/file-deduplicator/demo-images/samples", 0755)

	files := []struct {
		path       string
		brightness float64
	}{
		{"samples/sunset_original.jpg", 1.0},
		{"samples/sunset_bright.jpg", 1.15},
		{"samples/sunset_dark.jpg", 0.85},
		{"samples/cat_original.jpg", 1.0},
		{"samples/cat_copy1.jpg", 1.0},
		{"samples/cat_copy2.jpg", 1.0},
	}

	for _, f := range files {
		if len(f.path) > 10 && f.path[:10] == "samples/sun" {
			err := createSunsetImage("/home/ubuntu/file-deduplicator/demo-images/"+f.path, f.brightness)
			if err != nil {
				panic(err)
			}
			println("Created:", f.path)
		} else if len(f.path) > 10 && f.path[:10] == "samples/cat" {
			err := createCatImage("/home/ubuntu/file-deduplicator/demo-images/"+f.path)
			if err != nil {
				panic(err)
			}
			println("Created:", f.path)
		}
	}
}
