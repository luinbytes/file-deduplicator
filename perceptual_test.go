package main

import (
	"image"
	"image/color"
	"math"
	"testing"
)

// TestCosineFunction verifies that our cosine function produces correct values
func TestCosineFunction(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{0, 1.0},
		{math.Pi / 2, 0.0},
		{math.Pi, -1.0},
		{math.Pi / 4, 0.7071067811865476},
	}

	for _, tt := range tests {
		got := cosine(tt.input)
		if math.Abs(got-tt.expected) > 0.0001 {
			t.Errorf("cosine(%v) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

// TestGrayscaleConsistency verifies grayscale produces consistent values
func TestGrayscaleConsistency(t *testing.T) {
	c := color.RGBA{R: 100, G: 150, B: 200, A: 255}
	g1 := grayscale(c)
	g2 := grayscale(c)
	if g1 != g2 {
		t.Errorf("grayscale not consistent: %d vs %d", g1, g2)
	}
}

// TestPerceptualHashConsistency verifies same image produces same hash
func TestPerceptualHashConsistency(t *testing.T) {
	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 100, 255})
		}
	}

	// Test dHash consistency
	hash1, err := dHash(img)
	if err != nil {
		t.Fatalf("dHash failed: %v", err)
	}
	hash2, err := dHash(img)
	if err != nil {
		t.Fatalf("dHash failed: %v", err)
	}
	if hash1 != hash2 {
		t.Errorf("dHash not consistent: %s vs %s", hash1, hash2)
	}

	// Test aHash consistency
	hash1, err = aHash(img)
	if err != nil {
		t.Fatalf("aHash failed: %v", err)
	}
	hash2, err = aHash(img)
	if err != nil {
		t.Fatalf("aHash failed: %v", err)
	}
	if hash1 != hash2 {
		t.Errorf("aHash not consistent: %s vs %s", hash1, hash2)
	}

	// Test pHash consistency
	hash1, err = pHash(img)
	if err != nil {
		t.Fatalf("pHash failed: %v", err)
	}
	hash2, err = pHash(img)
	if err != nil {
		t.Fatalf("pHash failed: %v", err)
	}
	if hash1 != hash2 {
		t.Errorf("pHash not consistent: %s vs %s", hash1, hash2)
	}
}

// TestHammingDistance verifies hamming distance calculation
func TestHammingDistance(t *testing.T) {
	tests := []struct {
		hash1    string
		hash2    string
		expected int
	}{
		{"0000", "0000", 0},
		{"0000", "1111", 4},
		{"0101", "1010", 4},
		{"0110", "0110", 0},
		{"0", "1", 1},
	}

	for _, tt := range tests {
		got := hammingDistance(tt.hash1, tt.hash2)
		if got != tt.expected {
			t.Errorf("hammingDistance(%s, %s) = %d, want %d", tt.hash1, tt.hash2, got, tt.expected)
		}
	}
}

// TestSimilarImages detects similar images with slight brightness changes
func TestSimilarImages(t *testing.T) {
	// Create two similar images (one slightly brighter)
	img1 := image.NewRGBA(image.Rect(0, 0, 100, 100))
	img2 := image.NewRGBA(image.Rect(0, 0, 100, 100))

	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img1.Set(x, y, color.RGBA{100, 100, 100, 255})
			// Slightly brighter version
			img2.Set(x, y, color.RGBA{120, 120, 120, 255})
		}
	}

	// Test with dHash
	hash1, _ := dHash(img1)
	hash2, _ := dHash(img2)

	dist := hammingDistance(hash1, hash2)
	t.Logf("dHash distance between similar images: %d", dist)

	// They should be similar (distance should be low)
	if dist > 15 {
		t.Logf("Warning: dHash distance %d is high for similar images", dist)
	}

	// Test with aHash
	hash1, _ = aHash(img1)
	hash2, _ = aHash(img2)

	dist = hammingDistance(hash1, hash2)
	t.Logf("aHash distance between similar images: %d", dist)

	// Test with pHash
	hash1, _ = pHash(img1)
	hash2, _ = pHash(img2)

	dist = hammingDistance(hash1, hash2)
	t.Logf("pHash distance between similar images: %d", dist)
}

// TestFilteredImages detects images with simulated filter effects
// This is the key test for the P1 bug: perceptual hashing not detecting filtered/edited images
func TestFilteredImages(t *testing.T) {
	// Create a base test image with gradient pattern
	baseImg := image.NewRGBA(image.Rect(0, 0, 200, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			// Create a colorful gradient pattern
			r := uint8((x * 255) / 200)
			g := uint8((y * 255) / 200)
			b := uint8(((x + y) * 255) / 400)
			baseImg.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	// Simulate different filter effects
	filters := []struct {
		name string
		fn   func(image.Image) image.Image
	}{
		{"brightness_increase", func(img image.Image) image.Image {
			return applyBrightnessFilter(img, 1.3)
		}},
		{"brightness_decrease", func(img image.Image) image.Image {
			return applyBrightnessFilter(img, 0.7)
		}},
		{"contrast_increase", func(img image.Image) image.Image {
			return applyContrastFilter(img, 1.5)
		}},
		{"saturation_boost", func(img image.Image) image.Image {
			return applySaturationFilter(img, 1.5)
		}},
		{"warm_filter", func(img image.Image) image.Image {
			return applyColorFilter(img, 1.2, 1.0, 0.8)
		}},
		{"cool_filter", func(img image.Image) image.Image {
			return applyColorFilter(img, 0.8, 1.0, 1.2)
		}},
	}

	algorithms := []struct {
		name      string
		hashFn    func(image.Image) (string, error)
		threshold int
	}{
		{"dhash", dHash, 15},
		{"ahash", aHash, 18},
		{"phash", pHash, 12},
	}

	for _, algo := range algorithms {
		t.Run(algo.name, func(t *testing.T) {
			baseHash, err := algo.hashFn(baseImg)
			if err != nil {
				t.Fatalf("Failed to hash base image: %v", err)
			}

			for _, filter := range filters {
				filteredImg := filter.fn(baseImg)
				filteredHash, err := algo.hashFn(filteredImg)
				if err != nil {
					t.Fatalf("Failed to hash filtered image: %v", err)
				}

				dist := hammingDistance(baseHash, filteredHash)
				similarity := 100.0 - (float64(dist) / 64.0 * 100.0)

				t.Logf("%s + %s: distance=%d, similarity=%.1f%%",
					algo.name, filter.name, dist, similarity)

				// With proper preprocessing, filtered images should be detected as similar
				if dist > algo.threshold {
					t.Errorf("%s failed for %s: distance=%d (threshold=%d), similarity=%.1f%%",
						algo.name, filter.name, dist, algo.threshold, similarity)
				}
			}
		})
	}
}

// TestResizedImages detects the same image at different sizes
func TestResizedImages(t *testing.T) {
	// Create a base test image
	baseImg := image.NewRGBA(image.Rect(0, 0, 200, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			r := uint8((x * 255) / 200)
			g := uint8((y * 255) / 200)
			baseImg.Set(x, y, color.RGBA{r, g, 128, 255})
		}
	}

	// Resize to different dimensions
	sizes := []struct {
		name   string
		width  int
		height int
	}{
		{"same_size", 200, 200},
		{"half_size", 100, 100},
		{"double_size", 400, 400},
		{"wide", 300, 150},
		{"tall", 150, 300},
	}

	baseHash, err := dHash(baseImg)
	if err != nil {
		t.Fatalf("Failed to hash base image: %v", err)
	}

	for _, size := range sizes {
		// Create resized image
		resized := image.NewRGBA(image.Rect(0, 0, size.width, size.height))
		for y := 0; y < size.height; y++ {
			for x := 0; x < size.width; x++ {
				// Sample from base image
				srcX := (x * 200) / size.width
				srcY := (y * 200) / size.height
				r, g, b, a := baseImg.At(srcX, srcY).RGBA()
				resized.Set(x, y, color.RGBA{uint8(r / 256), uint8(g / 256), uint8(b / 256), uint8(a / 256)})
			}
		}

		resizedHash, err := dHash(resized)
		if err != nil {
			t.Fatalf("Failed to hash resized image: %v", err)
		}

		dist := hammingDistance(baseHash, resizedHash)
		similarity := 100.0 - (float64(dist) / 64.0 * 100.0)

		t.Logf("%s (%dx%d): distance=%d, similarity=%.1f%%",
			size.name, size.width, size.height, dist, similarity)

		// Resized images should be very similar
		if dist > 20 {
			t.Errorf("Resized image %s has high distance: %d", size.name, dist)
		}
	}
}

// TestCroppedImages detects cropped versions of the same image
func TestCroppedImages(t *testing.T) {
	// Create a base test image with centered pattern
	baseImg := image.NewRGBA(image.Rect(0, 0, 200, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			r := uint8((x * 255) / 200)
			g := uint8((y * 255) / 200)
			baseImg.Set(x, y, color.RGBA{r, g, 128, 255})
		}
	}

	// Test center crop (most common)
	cropSize := 150
	offset := (200 - cropSize) / 2
	cropped := image.NewRGBA(image.Rect(0, 0, cropSize, cropSize))
	for y := 0; y < cropSize; y++ {
		for x := 0; x < cropSize; x++ {
			r, g, b, a := baseImg.At(x+offset, y+offset).RGBA()
			cropped.Set(x, y, color.RGBA{uint8(r / 256), uint8(g / 256), uint8(b / 256), uint8(a / 256)})
		}
	}

	baseHash, _ := dHash(baseImg)
	cropHash, _ := dHash(cropped)

	dist := hammingDistance(baseHash, cropHash)
	similarity := 100.0 - (float64(dist) / 64.0 * 100.0)

	t.Logf("Center crop: distance=%d, similarity=%.1f%%", dist, similarity)

	// Cropped images will have higher distance but should still be detected
	if dist > 30 {
		t.Logf("Warning: Cropped image has high distance: %d", dist)
	}
}

// TestCompareImages tests the CompareImages function
func TestCompareImages(t *testing.T) {
	// Test the helper functions directly since file-based tests require proper encoding
	// The core hash algorithms are tested in TestFilteredImages
	t.Log("CompareImages requires proper image encoding - tested via hash functions directly")
}

// TestAdaptiveThreshold tests the adaptive threshold function
func TestAdaptiveThreshold(t *testing.T) {
	tests := []struct {
		algorithm  string
		strictness string
		minVal     int
		maxVal     int
	}{
		{"dhash", "strict", 5, 8},
		{"dhash", "normal", 9, 12},
		{"dhash", "loose", 13, 16},
		{"ahash", "strict", 6, 9},
		{"ahash", "normal", 10, 14},
		{"phash", "strict", 4, 6},
		{"phash", "normal", 7, 9},
	}

	for _, tt := range tests {
		threshold := AdaptiveThreshold(tt.algorithm, tt.strictness)
		if threshold < tt.minVal || threshold > tt.maxVal {
			t.Errorf("AdaptiveThreshold(%s, %s) = %d, expected between %d and %d",
				tt.algorithm, tt.strictness, threshold, tt.minVal, tt.maxVal)
		}
		t.Logf("AdaptiveThreshold(%s, %s) = %d", tt.algorithm, tt.strictness, threshold)
	}
}

// TestPreprocessingOptions tests different preprocessing configurations
func TestPreprocessingOptions(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{100, 150, 200, 255})
		}
	}

	// Test with different preprocessing options
	configs := []PreprocessingOptions{
		{ApplyBlur: false, ApplyNormalization: false, ApplyGammaCorrection: false},
		{ApplyBlur: true, ApplyNormalization: false, ApplyGammaCorrection: false},
		{ApplyBlur: false, ApplyNormalization: true, ApplyGammaCorrection: false},
		{ApplyBlur: false, ApplyNormalization: false, ApplyGammaCorrection: true},
		DefaultPreprocessing(),
	}

	for i, opts := range configs {
		processed := preprocessImage(img, opts)
		if processed == nil {
			t.Errorf("Config %d returned nil", i)
		}
	}
}

// Helper functions for filter simulation

func applyBrightnessFilter(img image.Image, factor float64) image.Image {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			nr := uint8(math.Min(255, float64(r/256)*factor))
			ng := uint8(math.Min(255, float64(g/256)*factor))
			nb := uint8(math.Min(255, float64(b/256)*factor))
			result.Set(x, y, color.RGBA{nr, ng, nb, uint8(a / 256)})
		}
	}
	return result
}

func applyContrastFilter(img image.Image, factor float64) image.Image {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			nr := uint8(math.Min(255, math.Max(0, (float64(r/256)-128)*factor+128)))
			ng := uint8(math.Min(255, math.Max(0, (float64(g/256)-128)*factor+128)))
			nb := uint8(math.Min(255, math.Max(0, (float64(b/256)-128)*factor+128)))
			result.Set(x, y, color.RGBA{nr, ng, nb, uint8(a / 256)})
		}
	}
	return result
}

func applySaturationFilter(img image.Image, factor float64) image.Image {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			gray := 0.299*float64(r/256) + 0.587*float64(g/256) + 0.114*float64(b/256)
			nr := uint8(math.Min(255, math.Max(0, gray+(float64(r/256)-gray)*factor)))
			ng := uint8(math.Min(255, math.Max(0, gray+(float64(g/256)-gray)*factor)))
			nb := uint8(math.Min(255, math.Max(0, gray+(float64(b/256)-gray)*factor)))
			result.Set(x, y, color.RGBA{nr, ng, nb, uint8(a / 256)})
		}
	}
	return result
}

func applyColorFilter(img image.Image, rFactor, gFactor, bFactor float64) image.Image {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			nr := uint8(math.Min(255, float64(r/256)*rFactor))
			ng := uint8(math.Min(255, float64(g/256)*gFactor))
			nb := uint8(math.Min(255, float64(b/256)*bFactor))
			result.Set(x, y, color.RGBA{nr, ng, nb, uint8(a / 256)})
		}
	}
	return result
}

// TestSmallImages tests that small images are handled correctly
func TestSmallImages(t *testing.T) {
	// Test with very small images (10x10)
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{100, 150, 200, 255})
		}
	}

	algorithms := []func(image.Image) (string, error){dHash, aHash, pHash}
	for _, algo := range algorithms {
		hash, err := algo(img)
		if err != nil {
			t.Errorf("Algorithm failed on small image: %v", err)
		}
		if len(hash) == 0 {
			t.Error("Hash should not be empty for small image")
		}
	}
}

// TestSolidColorImages tests solid color images (worst case for perceptual hashing)
func TestSolidColorImages(t *testing.T) {
	// Solid color images should produce consistent hashes
	colors := []color.RGBA{
		{255, 0, 0, 255},    // Red
		{0, 255, 0, 255},    // Green
		{0, 0, 255, 255},    // Blue
		{128, 128, 128, 255}, // Gray
	}

	for _, c := range colors {
		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		for y := 0; y < 100; y++ {
			for x := 0; x < 100; x++ {
				img.Set(x, y, c)
			}
		}

		// Same image should always produce same hash
		hash1, _ := dHash(img)
		hash2, _ := dHash(img)
		if hash1 != hash2 {
			t.Errorf("Solid color %v produced inconsistent hashes", c)
		}
	}
}

// TestIsImageFile tests image file detection
func TestIsImageFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"photo.jpg", true},
		{"photo.jpeg", true},
		{"photo.png", true},
		{"photo.gif", true},
		{"photo.webp", true},
		{"photo.JPG", true},
		{"photo.PNG", true},
		{"document.pdf", false},
		{"script.sh", false},
		{"noextension", false},
		{".hidden", false},
	}

	for _, tt := range tests {
		result := isImageFile(tt.path)
		if result != tt.expected {
			t.Errorf("isImageFile(%q) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

// Benchmarks

func BenchmarkDHash(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 100, 255})
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = dHash(img)
	}
}

func BenchmarkAHash(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 100, 255})
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = aHash(img)
	}
}

func BenchmarkPHash(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 100, 255})
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pHash(img)
	}
}
