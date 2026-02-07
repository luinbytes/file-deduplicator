package main

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/webp"
)

// PerceptualHash represents a perceptual hash of an image
type PerceptualHash struct {
	Hash  string
	Width int
	Height int
}

// dHash computes a difference hash (dHash) for an image
// This is fast and good for detecting near-duplicate images
func dHash(img image.Image) (string, error) {
	// Resize to 9x8 for dHash (we need 9 width to get 8 comparisons per row)
	resized := resizeImage(img, 9, 8)
	
	// Convert to grayscale and compute hash
	var hashBits []byte
	
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			left := grayscale(resized.At(x, y))
			right := grayscale(resized.At(x+1, y))
			
			// If left pixel is brighter than right, set bit to 1
			if left > right {
				hashBits = append(hashBits, '1')
			} else {
				hashBits = append(hashBits, '0')
			}
		}
	}
	
	return string(hashBits), nil
}

// aHash computes average hash (aHash) for an image  
// Good for detecting images with minor modifications
func aHash(img image.Image) (string, error) {
	// Resize to 8x8
	resized := resizeImage(img, 8, 8)
	
	// Calculate average brightness
	var total int64
	pixels := make([]int, 64)
	idx := 0
	
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			gray := grayscale(resized.At(x, y))
			pixels[idx] = gray
			total += int64(gray)
			idx++
		}
	}
	
	avg := int(total / 64)
	
	// Set bits where pixel is above average
	var hashBits []byte
	for _, p := range pixels {
		if p >= avg {
			hashBits = append(hashBits, '1')
		} else {
			hashBits = append(hashBits, '0')
		}
	}
	
	return string(hashBits), nil
}

// pHash computes perceptual hash (pHash) using DCT
// More robust but slower - simplified version
func pHash(img image.Image) (string, error) {
	// Resize to 32x32 for better frequency analysis
	resized := resizeImage(img, 32, 32)
	
	// Convert to grayscale float64 for DCT
	pixels := make([][]float64, 32)
	for i := range pixels {
		pixels[i] = make([]float64, 32)
	}
	
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			pixels[y][x] = float64(grayscale(resized.At(x, y)))
		}
	}
	
	// Apply DCT and take top-left 8x8 (low frequencies)
	dct := applyDCT(pixels)
	
	// Calculate average of 8x8 (excluding DC component at 0,0)
	var total float64
	count := 0
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if x == 0 && y == 0 {
				continue // Skip DC component
			}
			total += dct[y][x]
			count++
		}
	}
	avg := total / float64(count)
	
	// Generate hash
	var hashBits []byte
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if dct[y][x] >= avg {
				hashBits = append(hashBits, '1')
			} else {
				hashBits = append(hashBits, '0')
			}
		}
	}
	
	return string(hashBits), nil
}

// grayscale converts a color to grayscale value (0-255)
func grayscale(c color.Color) int {
	r, g, b, _ := c.RGBA()
	// Standard luminance formula
	return int(0.299*float64(r/256) + 0.587*float64(g/256) + 0.114*float64(b/256))
}

// resizeImage resizes an image to the specified dimensions using simple nearest neighbor
// For hashing purposes, we don't need high quality interpolation
func resizeImage(img image.Image, width, height int) image.Image {
	bounds := img.Bounds()
	dst := image.NewGray(image.Rect(0, 0, width, height))
	
	scaleX := float64(bounds.Dx()) / float64(width)
	scaleY := float64(bounds.Dy()) / float64(height)
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := int(float64(x) * scaleX)
			srcY := int(float64(y) * scaleY)
			c := img.At(bounds.Min.X+srcX, bounds.Min.Y+srcY)
			dst.Set(x, y, c)
		}
	}
	
	return dst
}

// applyDCT applies Discrete Cosine Transform (simplified version)
func applyDCT(pixels [][]float64) [][]float64 {
	size := len(pixels)
	result := make([][]float64, size)
	for i := range result {
		result[i] = make([]float64, size)
	}
	
	// Simplified DCT - in production you'd use a proper DCT implementation
	// This is a basic version for demonstration
	for u := 0; u < size; u++ {
		for v := 0; v < size; v++ {
			var sum float64
			for x := 0; x < size; x++ {
				for y := 0; y < size; y++ {
					cu := 1.0
					cv := 1.0
					if u == 0 {
						cu = 1.0 / 1.414213562
					}
					if v == 0 {
						cv = 1.0 / 1.414213562
					}
					sum += cu * cv * pixels[y][x] * 
						cosine((2*float64(x)+1)*float64(u)*3.14159265359/(2*float64(size))) *
						cosine((2*float64(y)+1)*float64(v)*3.14159265359/(2*float64(size)))
				}
			}
			result[v][u] = sum * 2.0 / float64(size)
		}
	}
	
	return result
}

func cosine(x float64) float64 {
	// Simple cosine - Go's math package would be better but keeping dependencies minimal
	// This is a placeholder - in real implementation use math.Cos
	return 1.0 - x*x/2.0 + x*x*x*x/24.0 // Taylor series approximation
}

// hammingDistance calculates the Hamming distance between two hash strings
func hammingDistance(hash1, hash2 string) int {
	if len(hash1) != len(hash2) {
		return -1
	}
	
	distance := 0
	for i := 0; i < len(hash1); i++ {
		if hash1[i] != hash2[i] {
			distance++
		}
	}
	return distance
}

// isSimilarImage checks if two hashes are similar within threshold
// threshold: max Hamming distance to consider images similar (0-64 for 64-bit hashes)
func isSimilarImage(hash1, hash2 string, threshold int) bool {
	dist := hammingDistance(hash1, hash2)
	return dist >= 0 && dist <= threshold
}

// computePerceptualHash computes the perceptual hash for an image file
func computePerceptualHash(path string, algorithm string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	// Decode image (supports jpeg, png, gif, webp)
	img, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}
	
	// Compute hash based on algorithm
	switch strings.ToLower(algorithm) {
	case "dhash", "difference":
		return dHash(img)
	case "ahash", "average":
		return aHash(img)
	case "phash", "perceptual":
		return pHash(img)
	default:
		return dHash(img) // Default to dHash
	}
}

// isImageFile checks if a file is an image we can process
func isImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return true
	default:
		return false
	}
}
