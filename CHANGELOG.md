# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [3.1.0] - 2026-02-09

### Fixed
- **Perceptual hashing not detecting filtered/edited images** (P1 Bug)
  - Images with brightness/contrast adjustments now correctly detected as similar
  - Instagram-style color filters now handled properly
  - Added gamma correction preprocessing for brightness normalization
  - Added histogram equalization for color distribution normalization
  - Improved blur to work on color channels before grayscale conversion

### Added
- **New preprocessing pipeline** for robust perceptual hashing:
  - `applyGammaCorrection()` - Normalizes brightness/contrast adjustments
  - `normalizeHistogram()` - Equalizes color distribution across RGB channels
  - `applyColorBlur()` - Applies blur to color channels before grayscale
  - `preprocessImage()` - Orchestrates all preprocessing steps
  - `PreprocessingOptions` - Configurable preprocessing configuration

- **Image comparison CLI command**:
  - `-compare img1,img2` - Compare two specific images
  - `-compare img1 -compare-with img2` - Alternative syntax
  - Outputs hash comparison for all three algorithms
  - Shows similarity percentage and recommendation

- **New API functions**:
  - `AdaptiveThreshold()` - Returns algorithm-specific threshold recommendations
  - `CompareImages()` - Detailed comparison between two image files

- **Comprehensive test coverage**:
  - `TestFilteredImages` - Tests brightness, contrast, saturation, warm/cool filters
  - `TestResizedImages` - Tests detection across different dimensions
  - `TestCroppedImages` - Tests center crop scenarios
  - `TestAdaptiveThreshold` - Tests threshold recommendation function
  - `TestPreprocessingOptions` - Tests different preprocessing configurations

### Changed
- Improved default similarity thresholds based on new preprocessing:
  - dHash: 10 (unchanged)
  - aHash: 12 (was 10)
  - pHash: 8 (was 10)
- Updated README with improved perceptual hashing documentation
- Added filter detection examples to documentation

### Technical Details
- Processing order changed from:
  ```
  image → grayscale → blur → resize → hash
  ```
  to:
  ```
  image → gamma correction → histogram normalization → color blur → grayscale → resize → hash
  ```

### Test Results
All filter simulations now pass with >90% similarity:
- Brightness adjustments: 93-100% similarity
- Contrast adjustments: 96-100% similarity
- Saturation filters: 96-100% similarity
- Color temperature (warm/cool): 96-100% similarity

## [3.0.0] - Previous Release

### Added
- Initial perceptual hashing implementation
- dHash, aHash, pHash algorithms
- Perceptual mode with similarity threshold
- Hybrid deduplication (standard + perceptual)

[3.1.0]: https://github.com/luinbytes/file-deduplicator/compare/v3.0.0...v3.1.0
