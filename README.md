# File Deduplicator v3.0.0 🖼️

A fast, parallel CLI tool to find and remove duplicate files using SHA256 hashing — **now with perceptual image deduplication!** Find similar photos, not just exact duplicates.

## What's New in v3.0

### 🖼️ Perceptual Image Deduplication
The killer feature that sets this apart from every other duplicate finder:

- **Find similar images**, not just exact duplicates — catch those 5 sunset shots you took
- **Multiple algorithms**: dHash (fast), aHash (balanced), pHash (most robust)
- **Configurable similarity**: Adjust threshold to match your needs
- **Smart grouping**: Groups similar photos together with similarity percentage
- **Supports**: JPG, PNG, GIF, WebP

### Why This Matters
Other tools only catch exact duplicates. This catches:
- Photos with slight edits (filters, crops, compression)
- Screenshots saved multiple times
- Downloaded images with different filenames
- Burst-mode photos that are nearly identical

## Features

### Standard Features
- 🚀 **Fast Parallel Processing** - Uses multiple goroutines for hashing
- 🔐 **Multiple Hash Algorithms** - SHA256, SHA1, MD5
- 📊 **Smart File Selection** - Keep oldest, newest, largest, smallest, or specific path
- ❓ **Interactive Mode** - Ask before deleting each duplicate
- 📦 **Move Instead of Delete** - Move duplicates to a safe folder
- 📁 **Pattern Filtering** - Only process files matching a pattern
- 📄 **Export Reports** - Generate JSON reports
- 🌳 **Recursive Scanning** - Scan directories recursively

### Perceptual Features (NEW)
- 🖼️ **Perceptual Hashing** - Find similar images using computer vision
- 🎨 **Algorithm Choice** - dHash, aHash, or pHash
- 📏 **Similarity Threshold** - Fine-tune what counts as "similar"
- 🧠 **Hybrid Mode** - Standard dedup for non-images, perceptual for images

## Installation

### Build from Source

```bash
git clone https://github.com/luinbytes/file-deduplicator.git
cd file-deduplicator
go mod tidy
go build -o file-deduplicator
```

Or install directly:

```bash
go install github.com/luinbytes/file-deduplicator@latest
```

## Usage

### Basic Duplicate Detection

```bash
# Find and delete exact duplicates
file-deduplicator -dir /path/to/scan

# Preview only (dry run)
file-deduplicator -dir /path/to/scan -dry-run

# Interactive mode
file-deduplicator -dir /path/to/scan -interactive
```

### Perceptual Image Deduplication (NEW)

```bash
# Find similar images in your Photos folder
file-deduplicator -dir ~/Pictures -perceptual

# Use faster algorithm (dHash)
file-deduplicator -dir ~/Pictures -perceptual -phash-algo dhash

# Stricter similarity (lower = more similar required)
file-deduplicator -dir ~/Pictures -perceptual -similarity 5

# More lenient similarity (catches more matches)
file-deduplicator -dir ~/Pictures -perceptual -similarity 15

# Most robust algorithm (slower but better)
file-deduplicator -dir ~/Pictures -perceptual -phash-algo phash
```

### Real-World Examples

**Clean up Downloads folder:**
```bash
file-deduplicator -dir ~/Downloads -move-to ~/Duplicates -dry-run
```

**Organize photo library:**
```bash
# First pass - see what would be found
file-deduplicator -dir ~/Pictures -perceptual -similarity 10 -dry-run -export

# Review the report
cat .deduplicator_report.json | jq '.duplicates[] | select(.similarity < 100)'

# Run for real with move (safer than delete)
file-deduplicator -dir ~/Pictures -perceptual -similarity 10 -move-to ~/Pictures/Similar
```

**Keep only largest versions:**
```bash
file-deduplicator -dir ~/Photos -perceptual -keep largest
```

**Focus on specific file types:**
```bash
# Only JPEGs
file-deduplicator -dir ~/Pictures -pattern "*.jpg" -perceptual

# Only PNG screenshots
file-deduplicator -dir ~/Screenshots -pattern "*.png" -perceptual
```

## Options

### Standard Options

| Option | Default | Description |
|--------|---------|-------------|
| `-dir string` | `.` | Directory to scan |
| `-recursive` | `true` | Scan recursively |
| `-dry-run` | `false` | Preview without deleting |
| `-verbose` | `false` | Detailed output |
| `-workers int` | NumCPU | Worker goroutines |
| `-min-size int` | `1024` | Minimum file size (bytes) |
| `-interactive` | `false` | Ask before each delete |
| `-move-to string` | `""` | Move duplicates here |
| `-keep string` | `oldest` | Keep: oldest/newest/largest/smallest/first/path |
| `-hash string` | `sha256` | Hash: sha256/sha1/md5 |
| `-pattern string` | `""` | File pattern (e.g., `*.jpg`) |
| `-export` | `false` | Export JSON report |
| `-undo` | `false` | View undo log |

### Perceptual Options (NEW)

| Option | Default | Description |
|--------|---------|-------------|
| `-perceptual` | `false` | Enable perceptual image deduplication |
| `-phash-algo` | `dhash` | Algorithm: dhash/ahash/phash |
| `-similarity` | `10` | Threshold 0-64 (lower = stricter) |

### Algorithm Comparison

| Algorithm | Speed | Accuracy | Best For |
|-----------|-------|----------|----------|
| **dHash** (default) | Fastest | Good | Quick scans, large libraries |
| **aHash** | Fast | Better | Balanced speed/accuracy |
| **pHash** | Slower | Best | Maximum accuracy, smaller sets |

### Similarity Threshold Guide

| Threshold | Match Type | Use Case |
|-----------|-----------|----------|
| `0-5` | Nearly identical | Strict dedup, minor edits only |
| `10` (default) | Very similar | Good balance |
| `15-20` | Similar | Catches more variations |
| `25+` | Loosely related | Broad matches |

## How Perceptual Hashing Works

1. **Resize** image to small size (8x8 or 9x8)
2. **Grayscale** to remove color info
3. **Compute hash** based on pixel relationships
4. **Compare** hashes using Hamming distance
5. **Group** images with similar hashes

This is the same technology used by:
- Google Image Search
- TinEye reverse image search
- Pinterest visual search

## Output Example

```
🔍 File Deduplicator v3.0.0 - Starting...
📁 Scanning directory: /home/user/Pictures
🖼️  Perceptual mode enabled (dhash, threshold: 10)
📊 Found 1523 files
🔐 Computed 1523 hashes
👯 Found 8 duplicate groups

🖼️  Similar Images Found:
======================================================================

[1] Hash: 101101001011...
    Size: 2.4 MB
    Files: 3 (keeping 1, removing 2)
    Similarity: 85% (perceptual match)
    ✓ KEEP /home/user/Pictures/Vacation/sunset.jpg (modified: 2026-01-15 18:30:00)
    ✗ DELETE /home/user/Pictures/Vacation/sunset_edited.jpg (modified: 2026-01-15 19:15:00)
    ✗ DELETE /home/user/Downloads/sunset_final.png (modified: 2026-01-16 09:20:00)

[2] Hash: 010011101001...
    Size: 1.8 MB
    Files: 2 (keeping 1, removing 1)
    Similarity: 90% (perceptual match)
    ✓ KEEP /home/user/Pictures/Cats/fluffy_original.jpg
    ✗ DELETE /home/user/Pictures/Cats/fluffy_copy(1).jpg

======================================================================
📊 Summary: 3 duplicates/similar files, 6.6 MB of space can be freed (2 perceptual groups)
```

## Safety Features

- **Dry run first** - Always preview with `-dry-run`
- **Move, don't delete** - Use `-move-to` to keep files safe
- **Export reports** - Document everything with `-export`
- **Undo log** - Track operations (informational)
- **Skip hidden files** - `.hidden` files ignored by default

## Best Practices

### For Photo Libraries
1. Start with `-dry-run` to see what would be found
2. Use `-similarity 10` as a balanced starting point
3. Export report: `-export` and review `.deduplicator_report.json`
4. Move, don't delete: `-move-to ~/Pictures/Similar`
5. Review moved files before permanent deletion

### For General Files
1. Use standard mode (no `-perceptual`) for non-image files
2. Pattern filtering for specific types: `-pattern "*.pdf"`
3. Keep criteria: `-keep newest` for backup scenarios

## Troubleshooting

### "No images processed"
- Perceptual mode only processes: `.jpg`, `.jpeg`, `.png`, `.gif`, `.webp`
- Use standard mode for other file types

### "Too many/few matches"
- Adjust `-similarity` threshold
- Try different `-phash-algo`

### Performance
- Standard mode: ~1000 files/sec per core
- Perceptual mode: ~200-500 images/sec per core (image decoding takes time)
- Use `-workers` to adjust parallelism

## Changelog

### v3.0.0
- ✨ **Perceptual image deduplication** - Find similar images, not just exact duplicates
- ✨ Multiple perceptual hash algorithms (dHash, aHash, pHash)
- ✨ Similarity threshold configuration
- ✨ Hybrid mode (standard + perceptual)
- ✨ Enhanced reporting with similarity percentages

### v2.0.0
- Fast parallel SHA256 hashing
- Smart file selection (oldest/newest/largest/smallest/path)
- Interactive mode
- Move instead of delete
- Pattern filtering
- JSON export

## License

MIT License

## Author

Created by Lu ([@luinbytes](https://github.com/luinbytes))

---

**Clean up your files, reclaim your space! 🔍🗑️🖼️**
