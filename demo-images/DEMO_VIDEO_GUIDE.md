# Demo Video Guide: Perceptual Image Deduplication

## Overview

This guide walks through creating a compelling demo video showing file-deduplicator's perceptual image deduplication feature in action.

## Demo Concept

**Before:** A photo library cluttered with duplicate and similar photos
**After:** Clean, organized library with duplicates removed

## Required Assets

### Sample Images

Create or gather sample images showing these scenarios:

1. **Sunset Photos** (5 variations)
   - `sunset_original.jpg` - Original photo
   - `sunset_bright.jpg` - Brightened version (+15%)
   - `sunset_contrast.jpg` - Increased contrast (+20%)
   - `sunset_blurred.jpg` - Slight blur
   - `sunset_saturated.jpg` - Instagram-style saturation (+30%)

2. **Cat Photos** (3 variations)
   - `cat_original.jpg` - Original
   - `cat_copy1.jpg` - Identical copy (different filename)
   - `cat_copy2.jpg` - Identical copy (different filename)

### How to Create Sample Images

**Option 1: Use Real Photos**
1. Take a photo (e.g., sunset)
2. Apply simple edits:
   - Brightness: +15% (Photoshop, GIMP, or Preview)
   - Contrast: +20%
   - Blur: Gaussian blur 0.5px
   - Saturation: +30%

**Option 2: Use Existing Photo Library**
- Grab 8-10 photos from your personal collection
- Make copies with slight edits
- Rename copies with descriptive names

**Option 3: Use Online Tools**
- Photopea.com (online Photoshop)
- Apply filters to same base image
- Download variations

## Video Script (30-45 seconds)

### Scene 1: Problem (5-8 seconds)
**Visual:** Show folder with ~10 photos, many look similar
**Voiceover:** "Ever found your photo library cluttered with duplicate and similar photos? You take five shots of that perfect sunset, but only need to keep one."

### Scene 2: Traditional Tools Don't Help (5-8 seconds)
**Visual:** Show command `fdupes ~/Pictures` finding only 1 exact duplicate
**Voiceover:** "Traditional duplicate finders only catch exact copies. They miss those similar shots with filters, crops, or compression."

### Scene 3: File Deduplicator Perceptual Mode (15-20 seconds)
**Visual:**
```
$ file-deduplicator -dir ~/Photos -perceptual -similarity 10
```
Show output:
```
🔍 File Deduplicator v3.0.0 - Starting...
📁 Scanning directory: ~/Photos
🖼️  Perceptual mode enabled (dhash, threshold: 10)
👯 Found 3 duplicate groups

[1] Similarity: 85%
  ✓ KEEP sunset_original.jpg
  ✗ DELETE sunset_bright.jpg
  ✗ DELETE sunset_contrast.jpg
  ✗ DELETE sunset_blurred.jpg
  ✗ DELETE sunset_saturated.jpg

[2] Similarity: 100%
  ✓ KEEP cat_original.jpg
  ✗ DELETE cat_copy1.jpg
  ✗ DELETE cat_copy2.jpg
```
**Voiceover:** "File Deduplicator uses perceptual hashing to find similar images, not just exact duplicates. It detects photos with filters, brightness changes, and other edits."

### Scene 4: Results (5-10 seconds)
**Visual:** Show clean folder with 2 photos instead of 10
**Voiceover:** "In seconds, your library is clean. Free up space and find what you actually want."

### Scene 5: CTA (3-5 seconds)
**Visual:** Show QR code or link to GitHub/Gumroad
**Voiceover:** "Get it free or grab the binaries for just $10. Clean up your photos today."

## Recording Setup

### Required Tools
- Screen recording: OBS Studio, Loom, QuickTime, or built-in tools
- Terminal: Large font, light background (for visibility)
- Image viewer: Preview (Mac) or similar for side-by-side comparison

### Recommended Settings
- Resolution: 1920x1080 (1080p)
- Frame rate: 30fps
- Audio: Clear voiceover, optional background music (quiet)

## Step-by-Step Recording

1. **Prepare demo folder**
   ```bash
   mkdir ~/file-dedup-demo
   cd ~/file-dedup-demo
   # Add your 8-10 sample images here
   ```

2. **Show problem (Scene 1)**
   - Open Finder/Explorer showing demo folder
   - Show 8-10 images in grid view
   - Point out similar ones with mouse

3. **Show traditional limitation (Scene 2)**
   - Open terminal
   - Run `fdupes ~/file-dedup-demo`
   - Show only 1 exact duplicate found

4. **Show file-deduplicator (Scene 3)**
   - Clear terminal
   - Run: `./file-deduplicator -dir ~/file-dedup-demo -perceptual -similarity 10 -dry-run`
   - Let output scroll
   - Point out the similarity percentages

5. **Show actual cleanup**
   - Run without dry-run: `./file-deduplicator -dir ~/file-dedup-demo -perceptual -similarity 10 -move-to ~/duplicates`
   - Show folder with fewer images

6. **Show cleanup (Scene 4)**
   - Show empty ~/duplicates folder
   - Show clean demo folder

7. **CTA (Scene 5)**
   - Show GitHub page: https://github.com/luinbytes/file-deduplicator
   - Show Gumroad link (once set up)

## Alternative: Animated GIF

If video is too complex, create an animated GIF:
1. Take screenshots of terminal output at each step
2. Use tools like `gifsicle` or online gif makers
3. Combine into single animated GIF (~5-10 seconds)
4. Add text overlays explaining each step

## Post-Processing

1. **Add text overlays**
   - "Before: 10 photos"
   - "After: 2 photos"
   - "Space saved: X MB"

2. **Highlight key features**
   - Perceptual similarity detection
   - Configurable threshold
   - Safe move (not delete)

3. **Polish**
   - Trim silences
   - Add background music (optional, keep it quiet)
   - Normalize audio levels

## Distribution

1. **GitHub** - Add to README.md
   ```markdown
   [![Demo Video](screenshot-thumb.png)](demo-video.mp4)
   ```

2. **Product Hunt** - Upload demo video
3. **Reddit / r/Python, r/golang** - Embed in post
4. **Hacker News** - Mention in Show HN post

## Notes

- Keep video under 60 seconds
- Show, don't just tell - visual demonstration is key
- Make terminal text large enough to read (12-14pt font minimum)
- Test on different screens to ensure readability
- Prepare script but speak naturally
- Consider adding subtitles for accessibility

## Example Commands for Demo

```bash
# Show problem
ls ~/file-dedup-demo
# Output: 10 photos

# Traditional approach (fails)
fdupes ~/file-dedup-demo
# Output: 1 exact duplicate only

# File deduplicator with perceptual mode
./file-deduplicator -dir ~/file-dedup-demo -perceptual -similarity 10 -dry-run
# Output: Found 3 groups with varying similarity (85%, 90%, 100%)

# Safe cleanup
./file-deduplicator -dir ~/file-dedup-demo -perceptual -similarity 10 -move-to ~/duplicates

# Show results
ls ~/file-dedup-demo  # 2 photos
ls ~/duplicates      # 8 similar photos moved
```

## Task Status

✅ Demo guide created
⏳ Sample images needed (create from personal photos or use online tools)
⏳ Video recording needed (requires Lu's input)
⏳ Post-processing and distribution
