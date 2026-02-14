# Demo Images - File Deduplicator

This directory contains materials for creating demo videos and screenshots showcasing the perceptual image deduplication feature.

## Files

### Documentation
- `DEMO_VIDEO_GUIDE.md` - Comprehensive guide for recording demo videos
- `demo_script.sh` - Automated demo script for terminal demonstrations

### Samples (to be added)
- `samples/` - Directory for sample images (create these manually)

## Quick Start

### 1. Create Sample Images

Add sample images to `samples/` directory with these characteristics:

**Sunset Photos (recommended):**
- 1 original photo
- 4 variations with edits:
  - Brightness +15%
  - Contrast +20%
  - Slight blur
  - Saturation +30% (Instagram-style)

**Cat/Other Photos (recommended):**
- 1 original photo
- 2 identical copies with different filenames

### 2. Run Demo Script

```bash
cd /home/ubuntu/file-deduplicator/demo-images
./demo_script.sh
```

This script will:
1. Create a demo directory
2. Check for sample images
3. Show current state
4. Run file-deduplicator with perceptual mode
5. Display safe cleanup workflow

### 3. Record Demo Video

Follow `DEMO_VIDEO_GUIDE.md` for:
- Video script (30-45 seconds)
- Recording setup
- Post-processing tips
- Distribution channels

## Why Perceptual Deduplication?

Traditional duplicate finders (fdupes, dupeGuru) only catch **exact duplicates**.

File Deduplicator catches **similar images**:
- Photos with brightness/contrast adjustments
- Instagram-style filters
- Slightly blurred copies
- Screenshots saved multiple times
- Burst-mode photos

## Demo Script Output Example

```
==========================================
File Deduplicator v3.0.0 Demo Script
==========================================

[STEP 1] Setting up demo directory...
✓ Created demo directory: /home/user/file-dedup-demo

[STEP 2] Checking for sample images...
✓ Found 10 sample images in demo directory

-rw-r--r-- 1 user user 256K sunset_original.jpg
-rw-r--r-- 1 user user 251K sunset_bright.jpg
-rw-r--r-- 1 user user 248K sunset_contrast.jpg
-rw-r--r-- 1 user user 249K sunset_blurred.jpg
-rw-r--r-- 1 user user 252K sunset_saturated.jpg
-rw-r--r-- 1 user user 125K cat_original.jpg
-rw-r--r-- 1 user user 125K cat_copy1.jpg
-rw-r--r-- 1 user user 125K cat_copy2.jpg

[STEP 3] Current state: Cluttered photo library...
Total size: 1.8M
Number of files: 10

[STEP 4] Traditional approach: Using fdupes...
Note: Traditional tools only find exact duplicates...
Exact duplicates found: 1

[STEP 5] File Deduplicator: Perceptual image deduplication...

Command: ./file-deduplicator -dir /home/user/file-dedup-demo -perceptual -similarity 10 -dry-run

🔍 File Deduplicator v3.0.0 - Starting...
📁 Scanning directory: /home/user/file-dedup-demo
🖼️  Perceptual mode enabled (dhash, threshold: 10)
📊 Found 10 files
🔐 Computed 10 hashes
👯 Found 2 duplicate groups

🖼️  Similar Images Found:
======================================================================

[1] Hash: 101101001011...
    Size: 1.3 MB
    Files: 5 (keeping 1, removing 4)
    Similarity: 85% (perceptual match)
    ✓ KEEP /home/user/file-dedup-demo/sunset_original.jpg
    ✗ DELETE /home/user/file-dedup-demo/sunset_bright.jpg
    ✗ DELETE /home/user/file-dedup-demo/sunset_contrast.jpg
    ✗ DELETE /home/user/file-dedup-demo/sunset_blurred.jpg
    ✗ DELETE /home/user/file-dedup-demo/sunset_saturated.jpg

[2] Hash: 010011101001...
    Size: 375 KB
    Files: 3 (keeping 1, removing 2)
    Similarity: 100% (exact duplicate)
    ✓ KEEP /home/user/file-dedup-demo/cat_original.jpg
    ✗ DELETE /home/user/file-dedup-demo/cat_copy1.jpg
    ✗ DELETE /home/user/file-dedup-demo/cat_copy2.jpg

======================================================================
📊 Summary: 7 duplicates/similar files, 1.4 MB of space can be freed

[STEP 6] Summary of findings...

Perceptual mode finds:
  - Exact duplicates (100% similarity)
  - Similar images with edits (brightness, contrast, filters)
  - Near-identical screenshots and burst photos

[STEP 7] Safe cleanup workflow...

To safely clean up, run:

  # Step 1: Preview (what you just saw above)
  ./file-deduplicator -dir /home/user/file-dedup-demo -perceptual -similarity 10 -dry-run

  # Step 2: Move to safe folder (not delete)
  ./file-deduplicator -dir /home/user/file-dedup-demo -perceptual -similarity 10 -move-to /home/user/file-dedup-demo_duplicates

  # Step 3: Review moved files
  ls /home/user/file-dedup-demo_duplicates

  # Step 4: Delete only after review
  rm -rf /home/user/file-dedup-demo_duplicates

==========================================
Demo complete!
==========================================
```

## Tips for Great Demos

1. **Use real photos** - Personal photos look more authentic
2. **Show variety** - Different types of edits demonstrate capabilities
3. **Keep it short** - Under 60 seconds for social media
4. **Large terminal** - Make text readable (12-14pt min)
5. **Before/after** - Show the dramatic space savings

## Next Steps

Once you have sample images:
1. Run `demo_script.sh` to test the workflow
2. Record screen with OBS, Loom, or QuickTime
3. Edit and polish following `DEMO_VIDEO_GUIDE.md`
4. Upload to GitHub, Product Hunt, and social media

## Resources

- Main README: `../README.md`
- Issue tracker: https://github.com/luinbytes/file-deduplicator/issues
- Landing page: https://luinbytes.github.io/file-deduplicator/
