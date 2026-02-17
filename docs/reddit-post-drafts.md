# Reddit Post Drafts for File Deduplicator v3.0.0

---

## r/DataHoarder

### Title
```
I built a CLI tool that finds similar images (not just exact duplicates) to clean up photo libraries
```

### Body
```
Hi fellow data hoarders,

I've been struggling with cluttered photo libraries for years. Duplicate finders like fdupes, dupeGuru, rmlint all miss the similar photos - those 5 sunset shots, the screenshots saved multiple times, the burst-mode photos.

I just launched v3.0.0 of [File Deduplicator](https://github.com/luinbytes/file-deduplicator) with perceptual image deduplication.

**What's different:**
- Uses perceptual hashing to find similar images, not just exact duplicates
- Catches photos with filters, brightness changes, compression artifacts
- Multiple algorithms (dHash, aHash, pHash) with configurable similarity threshold
- Fast, cross-platform CLI tool written in Go

**Example usage:**
```bash
# Find similar photos in your library
file-deduplicator -dir ~/Pictures -perceptual

# Preview first (dry-run)
file-deduplicator -dir ~/Pictures -perceptual -dry-run

# Move duplicates instead of deleting
file-deduplicator -dir ~/Pictures -perceptual -move-to ~/Duplicates
```

**Try it free:**
```bash
go install github.com/luinbytes/file-deduplicator@latest
```

Or buy prebuilt binaries ($10 one-time): https://gumroad.com/l/file-deduplicator

This uses the same perceptual hashing technology as Google Image Search and TinEye. I'd love feedback from the data hoarding community on how it works with your libraries.

GitHub: https://github.com/luinbytes/file-deduplicator
Live demo: https://luinbytes.github.io/file-deduplicator/
```

**Tags:** None (let community vote)

**Best time to post:** 8-11am ET or 6-9pm ET

---

## r/Go

### Title
```
[RFC] v3.0.0 - Perceptual image deduplication in Go (find similar photos, not just exact duplicates)
```

### Body
```
Hi r/golang,

I just released v3.0.0 of File Deduplicator, adding perceptual image deduplication. It finds similar images using computer vision techniques, not just exact byte-level duplicates.

**The problem:**
Most duplicate finders only catch exact duplicates. They miss photos with edits, compression, or different filenames.

**The solution:**
Perceptual hashing - resize, grayscale, compute hash based on pixel relationships, compare using Hamming distance.

**Implementation details:**
- Three algorithms: dHash (gradient-based), aHash (average), pHash (DCT-based)
- Configurable similarity threshold (0-64, lower = stricter)
- Image decoding for PNG, JPEG, GIF, WebP (required explicit imports, learned this the hard way 😅)
- Parallel processing with goroutines
- Hybrid mode: standard dedup for non-images, perceptual for images

**Code example (perceptual hash):**
```go
func computeDHash(img image.Image) uint64 {
    // Resize to 8x8, grayscale, compute gradient
    // Returns 64-bit hash
}
```

**Performance:**
- Standard SHA256: ~1000 files/sec per core
- Perceptual mode: ~200-500 images/sec per core (image decoding bottleneck)

**Try it:**
```bash
go install github.com/luinbytes/file-deduplicator@latest
file-deduplicator -dir ~/Pictures -perceptual
```

Open source, MIT licensed. Would love feedback on:
1. Algorithm choices (dHash vs aHash vs pHash)
2. Performance optimizations
3. Go code quality (still learning!)

GitHub: https://github.com/luinbytes/file-deduplicator
```

**Flair:** Project

**Best time to post:** 8-11am ET

---

## r/golang

### Title
```
Show: File Deduplicator v3.0.0 - CLI tool with perceptual image deduplication in Go
```

### Body
```
(Same as r/Go post, but shorter)
```

**Flair:** Show

---

## r/photography

### Title
```
Free CLI tool to find similar photos (not just duplicates) - helps clean up cluttered libraries
```

### Body
```
Hi r/photography,

I built a CLI tool called File Deduplicator that helps clean up cluttered photo libraries by finding similar images, not just exact duplicates.

**Why it's useful for photographers:**
- Catches burst-mode photos that are nearly identical
- Finds photos with slight edits (brightness, contrast, filters)
- Helps identify which versions to keep
- Saves storage space on your drives
- Works on RAW, JPEG, PNG, and more

**How it works:**
It uses perceptual hashing - the same tech Google Image Search uses to find similar images. It detects:
- Photos with Instagram-style filters
- Brightness/contrast adjustments
- Slightly blurred copies
- Screenshots saved multiple times

**Example:**
```bash
# Find similar photos in your library
file-deduplicator -dir ~/Pictures -perceptual

# Preview first (safe mode)
file-deduplicator -dir ~/Pictures -perceptual -dry-run

# Move duplicates to review later
file-deduplicator -dir ~/Pictures -perceptual -move-to ~/Duplicates
```

**Try it free:**
```bash
go install github.com/luinbytes/file-deduplicator@latest
```

It's open source, cross-platform, and free to build. Prebuilt binaries available for $10 if you don't want to compile.

Would love feedback from photographers on how it works with real photo libraries!

GitHub: https://github.com/luinbytes/file-deduplicator
Demo: https://luinbytes.github.io/file-deduplicator/
```

**Flair:** Help (or leave as default)

**Best time to post:** 8-11am ET or 6-9pm ET

---

## r/photographs

### Title
```
Tool I built to help find similar photos for cleanup (perceptual hashing, not just exact duplicates)
```

### Body
```
(Similar to r/photography but more casual)
```

---

## r/SideProject

### Title
```
Just launched v3.0.0 of my side project: CLI tool that finds similar images using perceptual hashing
```

### Body
```
Hi r/SideProject,

After 6 months of work, I just launched v3.0.0 of File Deduplicator. It's a CLI tool for finding duplicate files, with a killer feature: perceptual image deduplication.

**The journey:**
- Started as a weekend project to clean up my photo library
- Realized existing tools only found exact duplicates
- Implemented perceptual hashing (dHash, aHash, pHash)
- Now launching v3.0.0 with full perceptual support

**Tech stack:**
- Go for cross-platform support
- Goroutines for parallel processing
- Three hash algorithms with configurable thresholds

**Monetization:**
- Free to build from source (MIT licensed)
- $10 for prebuilt binaries (convenience pricing)
- Goal: $10k/year (~83 sales/month)

**Launch channels:**
- Hacker News (Show HN) - submitting today
- Reddit (this post!)
- GitHub Trending (optimizing repo)

**What I learned:**
1. Go's image.Decode requires explicit imports for PNG/JPEG decoders (spent hours debugging)
2. Perceptual hashing is CPU-intensive but worth it
3. Building for multiple platforms is tedious (automation needed)
4. Community feedback is invaluable

**Try it:**
```bash
go install github.com/luinbytes/file-deduplicator@latest
file-deduplicator -dir ~/Pictures -perceptual
```

GitHub: https://github.com/luinbytes/file-deduplicator
Landing page: https://luinbytes.github.io/file-deduplicator/

Would love feedback on:
1. Product itself
2. Marketing approach
3. Monetization strategy
4. What to do next

Thanks for reading! 🚀
```

**Flair:** Launch

---

## Submission Tips

### General Rules
1. **Read subreddit rules** before posting
2. **Use appropriate flairs** when available
3. **Engage with comments** - reply quickly, be helpful
4. **Don't spam** - one post per subreddit, space them out
5. **Follow up** - share updates if you make improvements

### Best Times to Post
- **r/DataHoarder:** 8-11am ET or 6-9pm ET
- **r/Go:** 8-11am ET (business hours)
- **r/golang:** Same as r/Go
- **r/photography:** 8-11am ET or 6-9pm ET (photographers active mornings/evenings)
- **r/SideProject:** Anytime, community is international

### What to Track
- Upvotes and engagement
- Comments and questions
- GitHub stars increase
- Landing page traffic
- Any Gumroad sales

### Engagement Strategy
1. **Reply to every comment** (first 6 hours critical)
2. **Be humble and open** to feedback
3. **Share learnings** from building the project
4. **Ask for help** if stuck on something
5. **Follow up** with improvements based on feedback

### Things to Avoid
- ❌ Posting the same thing multiple times
- ❌ Being defensive about criticism
- ❌ Over-promoting or sounding salesy
- ❌ Ignoring negative feedback
- ❌ Deleting and reposting

---

## Post Schedule (Recommended)

**Day 1:**
- Morning: Hacker News (Show HN) submit at 7am PT
- Afternoon: r/DataHoarder (if HN post does well)
- Evening: r/SideProject (casual launch)

**Day 2-3:**
- r/Go (morning, business hours)
- r/golang (same or next day)

**Day 4-5:**
- r/photography (morning)
- r/photographs (evening, different audience)

**Ongoing:**
- Share demos on Twitter/X
- Respond to GitHub issues
- Write blog posts (technical deep dives)

---

## Template for Engagement

**Positive comment:**
```
Thanks [user]! Really appreciate the kind words. Let me know if you run into any issues or have feature requests!
```

**Feature request:**
```
Great idea [user]! I've added that to the issue tracker: https://github.com/luinbytes/file-deduplicator/issues/[number]. Prioritizing for v3.1.0 if there's enough interest.
```

**Criticism:**
```
Fair point [user]. That's definitely a limitation. Currently working on [solution], expect it in v3.1.0. Thanks for the feedback!
```

**Bug report:**
```
Thanks for reporting [user]! I'll investigate this. Could you share your OS and version? File an issue here and I'll track it: https://github.com/luinbytes/file-deduplicator/issues/new
```

---

*Created: 2026-02-11*
*Ready to post: Yes (once Lu reviews and approves)*
