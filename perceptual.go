package main

import (
	"image"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

// PerceptualHash represents a perceptual hash of an image
type PerceptualHash struct {
	Hash   string
	Width  int
	Height int
}

// PreprocessingOptions holds options for image preprocessing
type PreprocessingOptions struct {
	ApplyBlur            bool
	ApplyNormalization   bool // Histogram equalization
	ApplyGammaCorrection bool // Normalize gamma
	TargetSize           int  // Target size for normalization (0 = no resize)
}

// DefaultPreprocessing returns default options optimized for filtered images
func DefaultPreprocessing() PreprocessingOptions {
	return PreprocessingOptions{
		ApplyBlur:            true,
		ApplyNormalization:   true,
		ApplyGammaCorrection: true,
		TargetSize:           0, // Use algorithm-specific sizing
	}
}

// preprocessImage applies all preprocessing steps to normalize the image
// This is key for detecting filtered/edited versions of the same image
func preprocessImage(img image.Image, opts PreprocessingOptions) image.Image {
	result := img

	// Step 1: Apply gamma correction to normalize brightness
	if opts.ApplyGammaCorrection {
		result = applyGammaCorrection(result, 2.2)
	}

	// Step 2: Apply color histogram normalization
	// This helps with images that have had saturation/contrast filters applied
	if opts.ApplyNormalization {
		result = normalizeHistogram(result)
	}

	// Step 3: Apply blur on color image before converting to grayscale
	// This helps with sharpening filters and noise
	if opts.ApplyBlur {
		result = applyColorBlur(result)
	}

	return result
}

// applyGammaCorrection applies gamma correction to normalize brightness
// This helps detect images that have had brightness/contrast adjustments
func applyGammaCorrection(img image.Image, gamma float64) image.Image {
	bounds := img.Bounds()
	corrected := image.NewRGBA(bounds)

	invGamma := 1.0 / gamma
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			// Convert to 0-255 range, apply gamma, convert back
			cr := uint8(math.Pow(float64(r)/65535.0, invGamma) * 255)
			cg := uint8(math.Pow(float64(g)/65535.0, invGamma) * 255)
			cb := uint8(math.Pow(float64(b)/65535.0, invGamma) * 255)
			ca := uint8(a / 256)
			corrected.Set(x, y, color.RGBA{cr, cg, cb, ca})
		}
	}
	return corrected
}

// normalizeHistogram applies histogram equalization to normalize color distribution
// This helps with saturation filters and color grading
func normalizeHistogram(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	pixelCount := width * height

	// Calculate histograms for each channel
	var rHist, gHist, bHist [256]int
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			rHist[r/256]++
			gHist[g/256]++
			bHist[b/256]++
		}
	}

	// Calculate cumulative distribution functions (CDF)
	var rCDF, gCDF, bCDF [256]int
	rSum, gSum, bSum := 0, 0, 0
	for i := 0; i < 256; i++ {
		rSum += rHist[i]
		gSum += gHist[i]
		bSum += bHist[i]
		rCDF[i] = rSum
		gCDF[i] = gSum
		bCDF[i] = bSum
	}

	// Apply histogram equalization
	normalized := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			// Map through CDF
			nr := uint8((float64(rCDF[r/256]) / float64(pixelCount)) * 255)
			ng := uint8((float64(gCDF[g/256]) / float64(pixelCount)) * 255)
			nb := uint8((float64(bCDF[b/256]) / float64(pixelCount)) * 255)
			na := uint8(a / 256)
			normalized.Set(x, y, color.RGBA{nr, ng, nb, na})
		}
	}
	return normalized
}

// applyColorBlur applies a box blur to the color image before grayscale conversion
// This preserves more color information than blurring after grayscale
func applyColorBlur(img image.Image) image.Image {
	bounds := img.Bounds()
	blurred := image.NewRGBA(bounds)

	// 3x3 box blur on color channels
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var rSum, gSum, bSum, count int

			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					nx, ny := x+dx, y+dy
					if nx >= bounds.Min.X && nx < bounds.Max.X && ny >= bounds.Min.Y && ny < bounds.Max.Y {
						r, g, b, _ := img.At(nx, ny).RGBA()
						rSum += int(r / 256)
						gSum += int(g / 256)
						bSum += int(b / 256)
						count++
					}
				}
			}

			_, _, _, a := img.At(x, y).RGBA()
			blurred.Set(x, y, color.RGBA{
				R: uint8(rSum / count),
				G: uint8(gSum / count),
				B: uint8(bSum / count),
				A: uint8(a / 256),
			})
		}
	}
	return blurred
}

// dHash computes a difference hash (dHash) for an image
// This is fast and good for detecting near-duplicate images
func dHash(img image.Image) (string, error) {
	// Preprocess with full normalization pipeline
	opts := DefaultPreprocessing()
	processed := preprocessImage(img, opts)

	// Resize to 9x8 for dHash (we need 9 width to get 8 comparisons per row)
	resized := resizeImage(processed, 9, 8)

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
	// Preprocess with full normalization pipeline
	opts := DefaultPreprocessing()
	processed := preprocessImage(img, opts)

	// Resize to 8x8
	resized := resizeImage(processed, 8, 8)

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
	// Preprocess with full normalization pipeline
	opts := DefaultPreprocessing()
	processed := preprocessImage(img, opts)

	// Resize to 32x32 for better frequency analysis
	resized := resizeImage(processed, 32, 32)

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

// resizeImage resizes an image to the specified dimensions using high-quality interpolation
// Uses Catmull-Rom for better perceptual hashing accuracy
func resizeImage(img image.Image, width, height int) image.Image {
	srcBounds := img.Bounds()
	dst := image.NewGray(image.Rect(0, 0, width, height))

	// Use Catmull-Rom interpolation for high-quality downsampling
	// This preserves more perceptual information than nearest-neighbor
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, srcBounds, draw.Over, nil)

	return dst
}

// applyBlur applies a simple box blur to reduce noise before hashing
// NOTE: This is kept for backward compatibility but applyColorBlur is preferred
func applyBlur(img image.Image) image.Image {
	bounds := img.Bounds()
	blurred := image.NewGray(bounds)

	// Simple 3x3 box blur
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var sum int
			var count int

			// Sample 3x3 neighborhood
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					nx, ny := x+dx, y+dy
					if nx >= bounds.Min.X && nx < bounds.Max.X && ny >= bounds.Min.Y && ny < bounds.Max.Y {
						sum += grayscale(img.At(nx, ny))
						count++
					}
				}
			}

			avg := uint8(sum / count)
			blurred.SetGray(x, y, color.Gray{Y: avg})
		}
	}

	return blurred
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
	return math.Cos(x)
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

// AdaptiveThreshold returns an appropriate threshold based on hash algorithm
// and the level of variation expected between similar images
func AdaptiveThreshold(algorithm string, strictness string) int {
	// strictness: "strict" (fewer matches), "normal" (balanced), "loose" (more matches)
	baseThresholds := map[string]int{
		"dhash":  10,
		"ahash":  12,
		"phash":  8,
	}

	multipliers := map[string]float64{
		"strict": 0.6,
		"normal": 1.0,
		"loose":  1.5,
	}

	base := baseThresholds[algorithm]
	if base == 0 {
		base = 10
	}

	mult := multipliers[strictness]
	if mult == 0 {
		mult = 1.0
	}

	return int(float64(base) * mult)
}

// CompareImages returns detailed comparison info between two image files
func CompareImages(path1, path2, algorithm string) (map[string]interface{}, error) {
	hash1, err := computePerceptualHash(path1, algorithm)
	if err != nil {
		return nil, err
	}

	hash2, err := computePerceptualHash(path2, algorithm)
	if err != nil {
		return nil, err
	}

	dist := hammingDistance(hash1, hash2)
	similarity := 100.0 - (float64(dist) / 64.0 * 100.0)
	if similarity < 0 {
		similarity = 0
	}

	return map[string]interface{}{
		"hash1":      hash1,
		"hash2":      hash2,
		"distance":   dist,
		"similarity": similarity,
		"isSimilar":  dist >= 0 && dist <= 10,
	}, nil
}
