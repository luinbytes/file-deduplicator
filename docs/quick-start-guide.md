# Quick Start Guide: File Deduplicator v3.0.0

Get started with File Deduplicator in 5 minutes.

## Installation

### Option 1: Install from Source (Free)

```bash
# Requires Go 1.21 or later
go install github.com/luinbytes/file-deduplicator@latest

# Verify installation
file-deduplicator --version
```

### Option 2: Download Prebuilt Binaries ($10)

Visit https://gumroad.com/l/file-deduplicator

Platforms available:
- Windows (amd64)
- macOS (Intel and Apple Silicon)
- Linux (amd64 and ARM64)

## First Run: Safe Preview

Always preview before deleting!

```bash
# Scan your Downloads folder (preview mode)
file-deduplicator -dir ~/Downloads -dry-run
```

This shows what would be found without changing anything.

## Common Use Cases

### 1. Clean Up Photo Library

```bash
# Find similar photos
file-deduplicator -dir ~/Pictures -perceptual -dry-run

# Move duplicates to review folder (safer than delete)
file-deduplicator -dir ~/Pictures -perceptual -move-to ~/Pictures/Duplicates
```

### 2. Standard Duplicate Detection

```bash
# Find exact duplicates in Documents
file-deduplicator -dir ~/Documents -dry-run
```

### 3. Specific File Types

```bash
# Only process JPEG images
file-deduplicator -dir ~/Pictures -pattern "*.jpg" -perceptual

# Only process PDFs (standard mode)
file-deduplicator -dir ~/Documents -pattern "*.pdf"
```

### 4. Safe Workflow

**Step 1: Preview**
```bash
file-deduplicator -dir ~/Pictures -perceptual -dry-run
```

**Step 2: Export Report**
```bash
file-deduplicator -dir ~/Pictures -perceptual -dry-run -export
cat .deduplicator_report.json | jq '.duplicates[] | select(.similarity < 100)'
```

**Step 3: Move, Don't Delete**
```bash
file-deduplicator -dir ~/Pictures -perceptual -move-to ~/Pictures/Similar
```

**Step 4: Review**
```bash
ls ~/Pictures/Similar
# Manually review files before deletion
```

**Step 5: Delete Only After Review**
```bash
rm -rf ~/Pictures/Similar
```

## Perceptual Mode: Finding Similar Images

### What It Does

Finds images that **look similar**, not just exact duplicates. Catches:
- Photos with filters (brightness, contrast, saturation)
- Slightly edited photos
- Screenshots saved multiple times
- Burst-mode photos

### Basic Usage

```bash
# Find similar images (default similarity: 10)
file-deduplicator -dir ~/Pictures -perceptual

# Stricter (only very similar images)
file-deduplicator -dir ~/Pictures -perceptual -similarity 5

# More lenient (catch more variations)
file-deduplicator -dir ~/Pictures -perceptual -similarity 15
```

### Algorithm Choice

```bash
# dHash (fastest, default)
file-deduplicator -dir ~/Pictures -perceptual -phash-algo dhash

# aHash (balanced)
file-deduplicator -dir ~/Pictures -perceptual -phash-algo ahash

# pHash (most accurate, slower)
file-deduplicator -dir ~/Pictures -perceptual -phash-algo phash
```

### Similarity Threshold Guide

| Threshold | Match Type | Best For |
|-----------|-----------|----------|
| `0-5` | Nearly identical | Strict cleanup |
| `10` (default) | Very similar | General use |
| `15-20` | Similar | Catching more variations |
| `25+` | Loosely related | Broad searches |

## Understanding Output

```
🔍 File Deduplicator v3.0.0 - Starting...
📁 Scanning directory: /home/user/Pictures
🖼️  Perceptual mode enabled (dhash, threshold: 10)
📊 Found 1523 files
👯 Found 8 duplicate groups

[1] Hash: 101101001011...
    Size: 2.4 MB
    Files: 3 (keeping 1, removing 2)
    Similarity: 85% (perceptual match)
    ✓ KEEP sunset_original.jpg
    ✗ DELETE sunset_edited.jpg
    ✗ DELETE sunset_blurred.jpg

======================================================================
📊 Summary: 2 duplicates, 1.6 MB of space can be freed
```

**Key Points:**
- `✓ KEEP` = Original/best version
- `✗ DELETE` = Duplicate/similar file
- `Similarity` = How similar the images are (100% = exact duplicate)
- `Size` = Total size of all files in this group

## Command Reference

### Basic Options

```
-dir string          Directory to scan (default: current directory)
-dry-run            Preview without deleting
-recursive          Scan recursively (default: true)
-verbose            Show detailed output
-move-to string     Move duplicates to folder (safer than delete)
-keep string        Keep: oldest/newest/largest/smallest/path (default: oldest)
```

### Perceptual Options

```
-perceptual          Enable perceptual image deduplication
-phash-algo string   Algorithm: dhash/ahash/phash (default: dhash)
-similarity int      Threshold 0-64 (default: 10, lower = stricter)
```

### Filtering Options

```
-pattern string      File pattern (e.g., "*.jpg", "*.pdf")
-min-size int        Minimum file size in bytes (default: 1024)
```

### Reporting Options

```
-export              Export JSON report
-undo                View undo log
```

### Other Options

```
-interactive         Ask before each deletion
-workers int         Number of worker goroutines (default: CPU cores)
-hash string         Hash: sha256/sha1/md5 (default: sha256)
```

## Examples by Scenario

### Scenario 1: Clean Up Downloads

```bash
# Preview
file-deduplicator -dir ~/Downloads -dry-run

# Move duplicates (review later)
file-deduplicator -dir ~/Downloads -move-to ~/Downloads/Duplicates
```

### Scenario 2: Burst Mode Photos

```bash
# Very strict similarity (burst shots are nearly identical)
file-deduplicator -dir ~/Photos/Burst -perceptual -similarity 3
```

### Scenario 3: Screenshot Cleanup

```bash
# Find similar screenshots
file-deduplicator -dir ~/Screenshots -perceptual -pattern "*.png"
```

### Scenario 4: Instagram Edits

```bash
# Catch photos with filters (more lenient)
file-deduplicator -dir ~/Pictures/Instagram -perceptual -similarity 15
```

### Scenario 5: Large Media Library

```bash
# Parallel processing for speed
file-deduplicator -dir ~/Media -perceptual -workers 8
```

## Troubleshooting

### "No images processed"
- Perceptual mode only supports: `.jpg`, `.jpeg`, `.png`, `.gif`, `.webp`
- Use standard mode for other file types

### "Too many/few matches"
- Adjust `-similarity` threshold
- Try different `-phash-algo` (pHash is most accurate)

### Performance Issues
- Use `-workers` to adjust parallelism
- Try dHash (faster) instead of pHash (slower)
- Use `-pattern` to limit file types

### "Unknown format" error
- Make sure you have image decoders imported if building from source
- Prebuilt binaries include all decoders

## Safety Best Practices

1. **Always dry-run first** - Preview before committing
2. **Move, don't delete** - Keep files in a separate folder
3. **Export reports** - Document everything
4. **Manual review** - Check moved files before deletion
5. **Backup first** - Have a backup of important data

## Performance Tips

- **Standard mode:** ~1000 files/sec per core
- **Perceptual mode:** ~200-500 images/sec per core
- **More workers = faster** (up to your CPU core count)
- **dHash = fastest**, **pHash = most accurate**

## Getting Help

- **Documentation:** https://github.com/luinbytes/file-deduplicator
- **Issues:** https://github.com/luinbytes/file-deduplicator/issues
- **Live demo:** https://luinbytes.github.io/file-deduplicator/

## Next Steps

1. Run a dry-run on a small test folder
2. Review the output
3. Try different similarity thresholds
4. Export a report and analyze it
5. Use `-move-to` for safe cleanup
6. Review moved files before deletion

---

**Happy cleaning! 🗑️**

*File Deduplicator v3.0.0 - Find duplicates and similar images from the command line*
