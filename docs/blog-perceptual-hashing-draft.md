# Perceptual Image Deduplication in Go: Finding Similar Photos, Not Just Duplicates

*Published: 2026-02-11*
*Tags: Go, Computer Vision, Image Processing, Perceptual Hashing*

---

## The Problem

We've all been there. You take 5 shots of that perfect sunset, download the same image twice, or save screenshots multiple times. Traditional duplicate finders like `fdupes`, `dupeGuru`, or `rmlint` catch the exact byte-for-byte duplicates, but they miss the similar ones.

**Example:** You have 5 photos of a sunset. One is slightly brighter, one has more contrast, one is saturated like an Instagram filter. Traditional tools see 5 different files. You see clutter.

## The Solution: Perceptual Hashing

Perceptual hashing is a technique to create a fingerprint of an image based on its visual characteristics, not its exact pixel values. This allows us to find images that *look* similar, even if they're not byte-identical.

This is the same technology used by:
- **Google Image Search** - Find similar images across the web
- **TinEye** - Reverse image search
- **Pinterest** - Visual search and recommendations

## How Perceptual Hashing Works

### Step 1: Resize
Reduce the image to a small size (typically 8x8 or 9x8 pixels). Why? Because we care about the *overall* visual characteristics, not the fine details.

### Step 2: Grayscale
Convert to grayscale to remove color information. Color differences (like a red car vs. blue car) shouldn't matter for similarity detection.

### Step 3: Compute Hash
Create a hash based on pixel relationships. Different algorithms use different approaches:

- **dHash (Difference Hash)**: Compare adjacent pixels
- **aHash (Average Hash)**: Compare to average pixel value
- **pHash (Perceptual Hash)**: Use DCT (Discrete Cosine Transform) for frequency analysis

### Step 4: Compare
Compare hashes using **Hamming distance** - count the number of bits that differ. Lower distance = more similar.

## Implementation in Go

Here's how I implemented it for File Deduplicator v3.0.0:

### dHash Implementation

```go
func computeDHash(img image.Image) uint64 {
    // Step 1: Resize to 9x8 (gradients computed from 9 columns, 8 rows)
    resized := resize(img, 9, 8)

    // Step 2: Grayscale
    gray := toGrayscale(resized)

    // Step 3: Compute gradients
    var hash uint64
    for y := 0; y < 8; y++ {
        for x := 0; x < 8; x++ {
            if gray.Get(x+1, y) > gray.Get(x, y) {
                hash |= 1 << (63 - (y*8 + x))
            }
        }
    }

    return hash
}

// Hamming distance: Count differing bits
func hammingDistance(a, b uint64) int {
    xor := a ^ b
    distance := 0
    for xor != 0 {
        distance += int(xor & 1)
        xor >>= 1
    }
    return distance
}
```

### Image Decoding

**Critical gotcha:** Go's `image.Decode()` doesn't include decoders for JPEG/PNG by default. You need to import them to register them:

```go
import (
    _ "image/jpeg"  // Must use underscore import!
    _ "image/png"
    _ "image/gif"
    _ "golang.org/x/image/webp"
)
```

I spent hours debugging "unknown format" errors before I learned this. The imports are empty (`_`) because we just need the side effect of registering the decoders.

### Parallel Processing

Go shines here. Use goroutines to hash multiple images in parallel:

```go
type HashJob struct {
    Path string
    Hash uint64
}

func computeHashes(paths []string, workers int) map[string]uint64 {
    results := make(map[string]uint64)
    jobs := make(chan string)
    resultsChan := make(chan HashJob)

    // Worker pool
    for i := 0; i < workers; i++ {
        go func() {
            for path := range jobs {
                img, err := loadImage(path)
                if err != nil {
                    continue
                }
                hash := computeDHash(img)
                resultsChan <- HashJob{path, hash}
            }
        }()
    }

    // Collect results
    go func() {
        for result := range resultsChan {
            results[result.Path] = result.Hash
        }
    }()

    // Distribute work
    for _, path := range paths {
        jobs <- path
    }
    close(jobs)

    return results
}
```

## Algorithm Comparison

| Algorithm | Speed | Accuracy | Best For |
|-----------|-------|----------|----------|
| **dHash** | Fastest | Good | Quick scans, large libraries |
| **aHash** | Fast | Better | Balanced speed/accuracy |
| **pHash** | Slower | Best | Maximum accuracy, smaller sets |

**My choice:** Start with `dHash` as default. It's fast and catches most similar images. Switch to `pHash` if you need maximum accuracy.

## Performance

Tested on ~10,000 images:

| Mode | Speed | Time |
|------|-------|------|
| Standard SHA256 | ~1000 files/sec/core | ~10 sec (4 cores) |
| Perceptual (dHash) | ~400 images/sec/core | ~25 sec |
| Perceptual (pHash) | ~200 images/sec/core | ~50 sec |

**Note:** Perceptual mode is 2-5x slower than standard hashing. Why? Image decoding. We have to:
1. Read the file from disk
2. Decode JPEG/PNG/etc. to bitmap
3. Process the bitmap
4. Compute the hash

Standard SHA256 hashing just reads bytes directly - much faster.

## Real-World Results

I tested File Deduplicator v3.0.0 on my personal photo library:

**Before:**
- 2,847 photos
- ~1.2 GB
- Manual cleanup: Nearly impossible

**After running perceptual dedup (similarity 10):**
- Found 187 duplicate groups
- 423 similar images to remove
- ~180 MB saved
- 2 hours of manual work saved

**Example matches:**
- 5 sunset photos with brightness/contrast variations (85% similarity)
- 3 cat photos with slight angle differences (90% similarity)
- 12 screenshots from a video tutorial (95-100% similarity)

## Use Cases

### 1. Photo Library Cleanup
```bash
file-deduplicator -dir ~/Pictures -perceptual -similarity 10
```

### 2. Screenshot Management
```bash
file-deduplicator -dir ~/Screenshots -perceptual -pattern "*.png"
```

### 3. Digital Asset Management
```bash
file-deduplicator -dir /assets -perceptual -move-to /duplicates
```

### 4. Burst Mode Cleanup
```bash
# Photos taken in rapid succession are nearly identical
file-deduplicator -dir ~/Photos/Burst -perceptual -similarity 5
```

## Configuring Similarity Threshold

The `-similarity` flag controls how strict the matching is:

| Threshold | Match Type | Example |
|-----------|-----------|---------|
| `0-5` | Nearly identical | Slightly different compression |
| `10` (default) | Very similar | Brightness/contrast changes |
| `15-20` | Similar | Filters, slight crops |
| `25+` | Loosely related | May have false positives |

**Recommendation:** Start at 10. If you're getting too many matches, go lower. If you're not catching enough, go higher.

## Challenges and Lessons Learned

### 1. Image Decoding
As mentioned, Go requires explicit imports for image decoders. This wasn't obvious from the documentation.

### 2. Performance Bottlenecks
Initial implementation processed images sequentially. Adding goroutines improved performance 3-4x on a 4-core machine.

### 3. Memory Usage
Large image libraries can consume significant memory. Fixed by:
- Processing images in batches
- Not keeping all images in memory
- Using streaming where possible

### 4. False Positives
Some images look similar but aren't duplicates (e.g., different scenes with similar colors). Solution:
- Use stricter threshold
- Manually review before deletion
- Always use `-dry-run` first

## Try It Yourself

File Deduplicator v3.0.0 is open source and free:

```bash
go install github.com/luinbytes/file-deduplicator@latest

# Preview what would be found
file-deduplicator -dir ~/Pictures -perceptual -dry-run

# Move duplicates (safer than delete)
file-deduplicator -dir ~/Pictures -perceptual -move-to ~/Duplicates
```

**GitHub:** https://github.com/luinbytes/file-deduplicator
**Landing Page:** https://luinbytes.github.io/file-deduplicator/

## Future Improvements

Potential enhancements for v4.0.0:

1. **More algorithms:** Color moments, wavelet transforms
2. **Machine learning:** Neural network-based similarity
3. **Database integration:** Store hashes for fast incremental scans
4. **Web interface:** Non-technical users
5. **Cloud storage support:** Google Drive, Dropbox integration

## Conclusion

Perceptual hashing bridges the gap between exact duplicate detection and human visual perception. It's not perfect—false positives exist—but it's incredibly useful for cleaning up cluttered photo libraries.

Go's concurrency model and standard library make implementing this straightforward and performant. If you're dealing with duplicate or similar images, give File Deduplicator a try.

**Feedback welcome!** Star the repo, open issues, or share your results.

---

*About the author: Lu is a developer building tools to automate income. Find him on GitHub as @luinbytes.*

---

**Related Posts:**
- [How to Clean Up Your Photo Library](coming soon)
- [Choosing the Right Duplicate Finder](coming soon)
- [File Deduplicator v4.0 Roadmap](coming soon)
