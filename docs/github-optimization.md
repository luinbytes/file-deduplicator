# GitHub Optimization Tasks

## Repository Topics to Add

GitHub topics help with discoverability. Add these to the repo:

```
go, golang, cli, duplicate-file-finder, deduplication,
perceptual-hashing, image-similarity, file-management,
storage, disk-cleanup, sha256, hashing, parallel-processing,
cross-platform, command-line, utility
```

## Badges to Add to README

```markdown
![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![GitHub Stars](https://img.shields.io/github/stars/luinbytes/file-deduplicator?style=social)
![GitHub Downloads](https://img.shields.io/github/downloads/luinbytes/file-deduplicator/total)
```

## Comparison Section to Add

```markdown
## Comparison with Other Tools

| Feature | file-deduplicator | fdupes | dupeGuru | rmlint |
|---------|------------------|--------|----------|--------|
| **Exact duplicates** | ✅ | ✅ | ✅ | ✅ |
| **Similar images** | ✅ | ❌ | Limited | ❌ |
| **Perceptual hashing** | ✅ (3 algos) | ❌ | ❌ | ❌ |
| **Cross-platform** | ✅ | ✅ | Limited | Linux only |
| **CLI** | ✅ | ✅ | GUI only | ✅ |
| **Parallel processing** | ✅ | ❌ | ❌ | ✅ |
| **Performance** | 🚀 Fast | 🐢 Slow | 🐢 Slow | 🚀 Fast |
| **Algorithm choice** | ✅ | ❌ | ❌ | ❌ |

**Why file-deduplicator?**
- Only tool with perceptual image deduplication
- Fastest cross-platform solution (Go + goroutines)
- Scriptable and automation-friendly
- Active development and open source
```

## Performance Benchmarks to Add

```markdown
## Performance

Tested on ~10,000 files (mix of images and documents):

| Mode | Speed | Time |
|------|-------|------|
| Standard SHA256 | ~1000 files/sec/core | ~10 sec (4 cores) |
| Perceptual (dHash) | ~400 images/sec/core | ~25 sec |
| Perceptual (pHash) | ~200 images/sec/core | ~50 sec |

**Note:** Perceptual mode is slower due to image decoding, but finds more duplicates.

System: Linux 6.5, 4-core CPU, NVMe SSD
```

## SEO Keywords for README

Add these keywords naturally in the README:
- "duplicate file finder"
- "find similar images"
- "perceptual image deduplication"
- "photo deduplication tool"
- "remove duplicate files"
- "clean up photo library"
- "image similarity detection"

## Social Proof Section

Once you have:
- GitHub stars: Add star count badge
- Users: Add "Used by X people" (if known)
- Testimonials: Add quotes from users

```markdown
## Used By

- [ ] Add early adopter testimonials
- [ ] Add company logos (if applicable)
```

## GitHub Actions to Add

```yaml
# .github/workflows/build.yml
name: Build and Test

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Build
        run: go build -v
      - name: Test
        run: go test -v ./...
```

## Release Notes Template

```markdown
## [v3.0.0] - 2026-02-11

### Added
- ✨ Perceptual image deduplication
- ✨ Multiple hash algorithms (dHash, aHash, pHash)
- ✨ Configurable similarity threshold
- ✨ Hybrid mode (standard + perceptual)

### Changed
- 📝 Improved README with more examples
- 📝 Added comparison section

### Fixed
- 🐛 Fixed PNG/JPEG decoder imports
- 🐛 Fixed image decoding errors
```

## GitHub Release Checklist

When creating a GitHub release:

- [ ] Tag version: `git tag -a v3.0.0 -m "Release v3.0.0"`
- [ ] Push tag: `git push origin v3.0.0`
- [ ] Upload binaries (Win/Mac/Linux × amd64/arm64)
- [ ] Write release notes
- [ ] Link to landing page
- [ ] Add "Show HN" mention (if applicable)

## Contributing Section Enhancement

```markdown
## Contributing

Contributions welcome! Areas where help is needed:

- **New hash algorithms**: Want to add a new perceptual hash algo?
- **Performance**: Optimize image decoding or hash computation
- **Documentation**: Improve examples and guides
- **Tests**: Add more test cases for edge cases
- **Platforms**: Test on different OS and architectures

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.
```

## GitHub Issue Templates

Create `.github/ISSUE_TEMPLATE/bug_report.md`:

```markdown
---
name: Bug report
about: Create a report to help us improve
title: '[BUG] '
labels: bug
---

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Run '....'
3. Scroll down to '....'
4. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Environment:**
 - OS: [e.g. Ubuntu 22.04]
 - Version: [e.g. v3.0.0]
 - Command: [paste command used]

**Additional context**
Add any other context about the problem here.
```

Create `.github/ISSUE_TEMPLATE/feature_request.md`:

```markdown
---
name: Feature request
about: Suggest an idea for this project
title: '[FEATURE] '
labels: enhancement
---

**Is your feature request related to a problem? Please describe.**
A clear and concise description of what the problem is.

**Describe the solution you'd like**
A clear and concise description of what you want to happen.

**Describe alternatives you've considered**
A clear and concise description of any alternative solutions or features you've considered.

**Additional context**
Add any other context or screenshots about the feature request here.
```

## GitHub Actions for Releases

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Build binaries
        run: |
          # Build for all platforms
          make build-all
      - name: Release
        uses: softprops/action-gh-release@v1
        with:
          files: |
            dist/file-deduplicator-*
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

*Created: 2026-02-11*
*Priority: Medium - Helps with discoverability but not urgent for launch*
