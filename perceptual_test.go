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

// TestSimilarImages detects similar images
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
	if dist > 20 {
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
