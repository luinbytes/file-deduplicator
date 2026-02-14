# Show HN: File Deduplicator v3.0.0 - CLI tool that finds similar images, not just exact duplicates

## Title (Copy this exactly)
```
Show HN: File Deduplicator v3.0.0 - CLI tool that finds similar images using perceptual hashing
```

## URL
https://github.com/luinbytes/file-deduplicator

## Post Body (Copy this)

Hi HN,

I just launched v3.0.0 of File Deduplicator, a fast CLI tool for finding and removing duplicate files. The big news: **perceptual image deduplication**.

Most duplicate finders only catch exact duplicates. They miss those 5 sunset shots you took, or screenshots saved with different names.

File Deduplicator uses perceptual hashing to find **similar images** - photos with filters, brightness changes, crops, compression artifacts. It catches:
- Instagram-style edits (brightness, contrast, saturation)
- Screenshots saved multiple times
- Downloaded images with different filenames
- Burst-mode photos that are nearly identical

## How it works

1. **Resize** images to 8x8 or 9x8 pixels
2. **Grayscale** to remove color info
3. **Compute hash** based on pixel relationships
4. **Compare** hashes using Hamming distance
5. **Group** images below similarity threshold

This is the same technology used by Google Image Search and TinEye.

## Example

```bash
# Find similar photos in your library
file-deduplicator -dir ~/Pictures -perceptual

# Stricter similarity (only very similar images)
file-deduplicator -dir ~/Pictures -perceptual -similarity 5

# More lenient (catch more variations)
file-deduplicator -dir ~/Pictures -perceptual -similarity 15
```

## Why build this?

I was tired of my photo library being cluttered with nearly identical photos. fdupes, dupeGuru, and rmlint all missed the similar ones. Only exact byte-for-byte matches.

Perceptual hashing solves this. It's fast, cross-platform, and scriptable.

## Try it out

**Free:**
```bash
go install github.com/luinbytes/file-deduplicator@latest
file-deduplicator -dir ~/Pictures -perceptual -dry-run
```

**Prebuilt binaries** ($10 one-time, convenience):
https://gumroad.com/l/file-deduplicator

**Live demo:**
https://luinbytes.github.io/file-deduplicator/

## Tech stack

- Go for speed and cross-platform support
- Goroutines for parallel processing
- Three algorithms: dHash (fast), aHash (balanced), pHash (most robust)

## What's new in v3.0.0

- ✨ Perceptual image deduplication
- ✨ Multiple hash algorithms (dHash/aHash/pHash)
- ✨ Configurable similarity threshold
- ✨ Hybrid mode (standard + perceptual)
- ✨ Enhanced reporting with similarity percentages

Feedback welcome!

---

## Additional Notes (Not part of post)

### Timing Tips
- **Best times:** 7-10am Pacific (US West Coast), or 3-5pm Pacific (Europe waking up)
- **Worst times:** 10pm-5am Pacific (low activity)
- **Weekends:** Generally lower engagement, but can still work

### Submission Checklist
- [ ] Title starts with "Show HN"
- [ ] URL is correct (GitHub repo or landing page)
- [ ] No ask for upvotes in the post
- [ ] Easy to try (free build from source)
- [ ] Clear description of what it does
- [ ] Example commands provided
- [ ] Links to live demo/docs

### After Submission
- **Reply quickly** to comments (first hour is critical)
- **Be open** to feedback, even criticism
- **Don't argue** - engage constructively
- **Share updates** if you make changes based on feedback
- **Follow up** with blog posts/posts about improvements

### Tracking Success
- Monitor upvotes and comments
- Check GitHub stars increase
- Track landing page traffic
- Note any downloads/purchases

### What NOT to do
- ❌ Ask for upvotes or shares
- ❌ Submit the same thing twice
- ❌ Delete and resubmit (looks bad)
- ❌ Get defensive about criticism
- ❌ Post when project isn't ready for users

### Alternative Titles (Tested options)
```
Show HN: File Deduplicator - CLI tool for finding similar images using perceptual hashing
Show HN: File Deduplicator v3.0 - Find duplicate and similar photos from the command line
Show HN: I built a duplicate file finder that catches similar images (not just exact duplicates)
```

**Chosen title:** "Show HN: File Deduplicator v3.0.0 - CLI tool that finds similar images using perceptual hashing"

**Why:** Clear, descriptive, includes version number (signals major update), mentions key differentiator.

---

## Pre-Submission Checklist
- [ ] Project is ready for users to try ✅
- [ ] No signup barriers ✅
- [ ] GitHub repo is public ✅
- [ ] Landing page exists ✅
- [ ] Instructions are clear ✅
- [ ] You're available to answer questions in the thread ✅

## Submission URL (Fill in after submitting)
https://news.ycombinator.com/item?id=[PASTE_ID_HERE]

---

*Created: 2026-02-11*
*Ready to submit: Yes*
