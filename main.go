package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	version                = "3.1.0"
	reportFile             = ".deduplicator_report.json"
	undoFile               = ".deduplicator_undo.json"
	maxHistory             = 100
	progressUpdateInterval  = 1 * time.Second
)

// FileHash represents a file and its hash
type FileHash struct {
	Path     string
	Size     int64
	Hash     string
	ModTime  time.Time
	PHash    string  // Perceptual hash for images
}

// Statistics tracks detailed operation metrics
type Statistics struct {
	ScanStart      time.Time
	ScanEnd        time.Time
	HashStart      time.Time
	HashEnd        time.Time
	ProcessStart   time.Time
	ProcessEnd     time.Time
	TotalFiles     int
	TotalBytes     int64
	FilesByExt     map[string]int
	DuplicateFiles int
	DuplicateBytes int64
	ImageFiles     int
}

// NewStatistics creates a new Statistics object
func NewStatistics() *Statistics {
	return &Statistics{
		FilesByExt: make(map[string]int),
	}
}

// DuplicateGroup represents a group of duplicate files
type DuplicateGroup struct {
	Hash  string
	Size  int64
	Files []FileHash
	Similarity float64 // For perceptual matches
}

// Config holds application configuration
type Config struct {
	Dir            string
	Recursive      bool
	DryRun         bool
	Verbose        bool
	Workers        int
	MinSize        int64  // Minimum file size to check (bytes)
	Interactive    bool
	MoveTo         string // Move duplicates to this folder instead of deleting
	KeepCriteria   string // "oldest", "newest", "largest", "smallest", "first", "path"
	HashAlgorithm  string // "sha256", "sha1", "md5"
	FilePattern    string // Only include files matching this pattern
	ExportReport   bool
	UndoLast       bool
	NoEmoji        bool   // Disable emoji output for cleaner logs
	// Perceptual hashing options
	PerceptualMode bool   // Enable perceptual hashing for images
	PHashAlgorithm string // "dhash", "ahash", "phash"
	SimilarityThreshold int // Hamming distance threshold (0-64, default 10)
	// Image comparison flags
	CompareImg1 string // First image for comparison
	CompareImg2 string // Second image for comparison
}

var (
	cfg Config
)

// emoji returns the emoji if NoEmoji is false, otherwise returns empty string
func emoji(e string) string {
	if cfg.NoEmoji {
		return ""
	}
	return e + " "
}

func init() {
	// Override default usage to show categorized help
	flag.Usage = customUsage

	flag.StringVar(&cfg.Dir, "dir", ".", "Directory to scan for duplicates")
	flag.BoolVar(&cfg.Recursive, "recursive", true, "Scan directories recursively")
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "Show what would be deleted without actually deleting")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Show detailed output")
	flag.IntVar(&cfg.Workers, "workers", runtime.NumCPU(), "Number of worker goroutines")
	flag.Int64Var(&cfg.MinSize, "min-size", 1024, "Minimum file size in bytes (default: 1KB)")
	flag.BoolVar(&cfg.Interactive, "interactive", false, "Ask before deleting each duplicate")
	flag.StringVar(&cfg.MoveTo, "move-to", "", "Move duplicates to this folder instead of deleting")
	flag.StringVar(&cfg.KeepCriteria, "keep", "oldest", "File to keep criteria: oldest, newest, largest, smallest, first, or path:<path>")
	flag.StringVar(&cfg.HashAlgorithm, "hash", "sha256", "Hash algorithm: sha256, sha1, or md5")
	flag.StringVar(&cfg.FilePattern, "pattern", "", "File pattern to match (e.g., *.jpg, *.pdf)")
	flag.BoolVar(&cfg.ExportReport, "export", false, "Export duplicate report to JSON file")
	flag.BoolVar(&cfg.UndoLast, "undo", false, "Undo last operation")
	flag.BoolVar(&cfg.NoEmoji, "no-emoji", false, "Disable emoji output for cleaner logs")
	
	// Perceptual hashing flags
	flag.BoolVar(&cfg.PerceptualMode, "perceptual", false, "Enable perceptual hashing for images (finds similar images, not just exact duplicates)")
	flag.StringVar(&cfg.PHashAlgorithm, "phash-algo", "dhash", "Perceptual hash algorithm: dhash (fast), ahash, phash (robust)")
	flag.IntVar(&cfg.SimilarityThreshold, "similarity", 10, "Similarity threshold (0-64). Lower = stricter. Default 10.")
	
	// Image comparison flags
	flag.StringVar(&cfg.CompareImg1, "compare", "", "Compare two images (format: img1,img2 or use with -compare-with)")
	flag.StringVar(&cfg.CompareImg2, "compare-with", "", "Second image for comparison (use with -compare)")
}

// customUsage prints categorized help text
func customUsage() {
	fmt.Fprintf(os.Stderr, "Usage: file-deduplicator [options]\n\n")
	fmt.Fprintf(os.Stderr, "A fast, parallel CLI duplicate finder with perceptual image hashing.\n\n")
	
	fmt.Fprintf(os.Stderr, "SCAN OPTIONS:\n")
	fmt.Fprintf(os.Stderr, "  -dir string\n\tDirectory to scan (default: current directory)\n")
	fmt.Fprintf(os.Stderr, "  -recursive\n\tScan subdirectories (default: true)\n")
	fmt.Fprintf(os.Stderr, "  -workers int\n\tNumber of parallel workers (default: %d)\n", runtime.NumCPU())
	fmt.Fprintf(os.Stderr, "  -min-size int\n\tSkip files smaller than this (bytes, default: 1024)\n")
	fmt.Fprintf(os.Stderr, "  -pattern string\n\tOnly match files matching this pattern (e.g., *.jpg)\n")
	
	fmt.Fprintf(os.Stderr, "\nHASH OPTIONS:\n")
	fmt.Fprintf(os.Stderr, "  -hash string\n\tAlgorithm: sha256, sha1, md5 (default: sha256)\n")
	
	fmt.Fprintf(os.Stderr, "\nPERCEPTUAL IMAGE MATCHING:\n")
	fmt.Fprintf(os.Stderr, "  -perceptual\n\tFind similar images, not just exact duplicates\n")
	fmt.Fprintf(os.Stderr, "  -phash-algo string\n\tAlgorithm: dhash, ahash, phash (default: dhash)\n")
	fmt.Fprintf(os.Stderr, "  -similarity int\n\tThreshold 0-64, lower = stricter (default: 10)\n")
	fmt.Fprintf(os.Stderr, "  -compare img1,img2\n\tCompare two specific images\n")
	fmt.Fprintf(os.Stderr, "  -compare-with string\n\tSecond image (alternative to comma syntax)\n")
	
	fmt.Fprintf(os.Stderr, "\nACTION OPTIONS:\n")
	fmt.Fprintf(os.Stderr, "  -dry-run\n\tPreview what would be deleted (no changes made)\n")
	fmt.Fprintf(os.Stderr, "  -interactive\n\tAsk before deleting each file\n")
	fmt.Fprintf(os.Stderr, "  -move-to string\n\tMove duplicates to folder instead of deleting\n")
	fmt.Fprintf(os.Stderr, "  -keep string\n\tWhich file to keep: oldest, newest, largest, smallest, path:<pattern> (default: oldest)\n")
	
	fmt.Fprintf(os.Stderr, "\nOUTPUT OPTIONS:\n")
	fmt.Fprintf(os.Stderr, "  -verbose\n\tShow detailed progress\n")
	fmt.Fprintf(os.Stderr, "  -export\n\tExport JSON report of duplicates found\n")
	fmt.Fprintf(os.Stderr, "  -no-emoji\n\tPlain text output (no emoji)\n")
	
	fmt.Fprintf(os.Stderr, "\nUTILITY:\n")
	fmt.Fprintf(os.Stderr, "  -undo\n\tView log of last deletion operation\n")
	
	fmt.Fprintf(os.Stderr, "\nEXAMPLES:\n")
	fmt.Fprintf(os.Stderr, "  file-deduplicator -dir ~/Photos -dry-run\n")
	fmt.Fprintf(os.Stderr, "  file-deduplicator -dir ~/Downloads -move-to ~/Duplicates\n")
	fmt.Fprintf(os.Stderr, "  file-deduplicator -dir ~/Photos -perceptual -similarity 8\n")
	fmt.Fprintf(os.Stderr, "  file-deduplicator -compare photo1.jpg,photo2.jpg\n")
}

func main() {
	flag.Parse()

	// Handle undo
	if cfg.UndoLast {
		if err := undoLast(); err != nil {
			log.Fatalf("❌ Error undoing: %v", err)
		}
		return
	}

	// Handle image comparison
	if cfg.CompareImg1 != "" {
		if err := compareImagesCLI(); err != nil {
			log.Fatalf("❌ Error comparing images: %v", err)
		}
		return
	}

	log.SetFlags(log.Ltime)

	log.Printf("%sFile Deduplicator v%s - Starting...", emoji("🔍"), version)
	if cfg.Verbose {
		log.Printf("%sScanning directory: %s", emoji("📁"), cfg.Dir)
		log.Printf("%sRecursive: %v", emoji("🔄"), cfg.Recursive)
		log.Printf("%sWorkers: %d", emoji("👷"), cfg.Workers)
		log.Printf("%sMin size: %d bytes", emoji("📏"), cfg.MinSize)
		log.Printf("%sHash algorithm: %s", emoji("🔐"), cfg.HashAlgorithm)
		if cfg.FilePattern != "" {
			log.Printf("%sFile pattern: %s", emoji("🎯"), cfg.FilePattern)
		}
		if cfg.MoveTo != "" {
			log.Printf("%sMove duplicates to: %s", emoji("📦"), cfg.MoveTo)
		}
		log.Printf("%sKeep criteria: %s", emoji("✋"), cfg.KeepCriteria)
		if cfg.Interactive {
			log.Printf("%sInteractive mode enabled", emoji("❓"))
		}
		if cfg.PerceptualMode {
			log.Printf("%sPerceptual mode enabled (%s, threshold: %d)", emoji("🖼️"), cfg.PHashAlgorithm, cfg.SimilarityThreshold)
		}
	}

	startTime := time.Now()

	// Scan files
	files, err := scanFiles(cfg.Dir, cfg.Recursive)
	if err != nil {
		log.Fatalf("❌ Error scanning files: %v", err)
	}

	log.Printf("%sFound %d files", emoji("📊"), len(files))

	// Filter by minimum size
	var filteredFiles []string
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			if cfg.Verbose {
				log.Printf("%sCould not stat %s: %v", emoji("⚠️"), file, err)
			}
			continue
		}
		if info.Size() >= cfg.MinSize {
			// Filter by file pattern if specified
			if cfg.FilePattern != "" {
				matched, err := filepath.Match(cfg.FilePattern, filepath.Base(file))
				if err != nil {
					log.Printf("%sInvalid pattern %s: %v", emoji("⚠️"), cfg.FilePattern, err)
					continue
				}
				if !matched {
					if cfg.Verbose {
						log.Printf("%sSkipping non-matching file: %s", emoji("🚫"), file)
					}
					continue
				}
			}
			filteredFiles = append(filteredFiles, file)
		}
	}
	log.Printf("%sAfter filters: %d files", emoji("📏"), len(filteredFiles))

	// Compute hashes in parallel
	fileHashes, err := computeHashes(filteredFiles)
	if err != nil {
		log.Fatalf("❌ Error computing hashes: %v", err)
	}
	if !cfg.Verbose {
		fmt.Fprintln(os.Stderr) // Newline after progress bar
	}
	log.Printf("%sComputed %d hashes", emoji("🔐"), len(fileHashes))

	// Find duplicates
	duplicates := findDuplicates(fileHashes)
	log.Printf("%sFound %d duplicate groups", emoji("👯"), len(duplicates))

	// Report duplicates
	reportDuplicates(duplicates)

	// Export report if requested
	if cfg.ExportReport {
		if err := exportReport(duplicates); err != nil {
			log.Printf("%sFailed to export report: %v", emoji("⚠️"), err)
		} else {
			log.Printf("%sReport exported to %s", emoji("📄"), reportFile)
		}
	}

	// Process duplicates if not dry run
	if !cfg.DryRun && len(duplicates) > 0 {
		if err := processDuplicates(duplicates); err != nil {
			log.Fatalf("❌ Error processing duplicates: %v", err)
		}
	}

	elapsed := time.Since(startTime)
	log.Printf("%sComplete in %v", emoji("✅"), elapsed)
}

func scanFiles(dir string, recursive bool) ([]string, error) {
	var files []string
	var scanned int
	var scannedMutex sync.Mutex

	// Simple progress tracker
	lastProgressUpdate := time.Now()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		scannedMutex.Lock()
		scanned++
		currentScanned := scanned
		scannedMutex.Unlock()

		// Update progress periodically
		if time.Since(lastProgressUpdate) > progressUpdateInterval {
			lastProgressUpdate = time.Now()
			if cfg.Verbose {
				log.Printf("📁 Scanned %d files...", currentScanned)
			} else {
				fmt.Fprintf(os.Stderr, "\r📁 Scanning: %d files", currentScanned)
			}
		}

		if info.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(filepath.Base(path), ".") {
				if cfg.Verbose {
					log.Printf("%sSkipping hidden directory: %s", emoji("🚫"), path)
				}
				return filepath.SkipDir
			}
			// Skip non-recursive
			if !recursive && path != dir {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(filepath.Base(path), ".") {
			if cfg.Verbose {
				log.Printf("%sSkipping hidden file: %s", emoji("🚫"), path)
			}
			return nil
		}

		files = append(files, path)
		return nil
	})

	// Final progress update
	if !cfg.Verbose {
		fmt.Fprintf(os.Stderr, "\r📁 Scanning: %d files\n", len(files))
	}

	return files, err
}

func computeHashes(files []string) ([]FileHash, error) {
	var wg sync.WaitGroup
	fileChan := make(chan string, cfg.Workers)
	resultChan := make(chan FileHash, len(files))
	errorChan := make(chan error, len(files))

	// Progress tracking
	var hashedCount int
	var hashedMutex sync.Mutex
	totalFiles := len(files)
	lastProgressUpdate := time.Now()
	startTime := time.Now()

	// Start worker goroutines
	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go worker(&wg, fileChan, resultChan, errorChan, &hashedCount, &hashedMutex, &lastProgressUpdate, totalFiles, startTime)
	}

	// Send files to workers
	go func() {
		for _, file := range files {
			fileChan <- file
		}
		close(fileChan)
	}()

	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Collect results
	var fileHashes []FileHash
	for fh := range resultChan {
		fileHashes = append(fileHashes, fh)
	}

	// Check for errors
	for err := range errorChan {
		if err != nil {
			log.Printf("%sError: %v", emoji("⚠️"), err)
		}
	}

	// Final progress update
	if !cfg.Verbose && totalFiles > 0 {
		elapsed := time.Since(startTime).Seconds()
		fmt.Fprintf(os.Stderr, "\r%s%s %d/%d (%.1f%%) Completed in %s\n",
			emoji("✅"), emoji("▏"), totalFiles, totalFiles, 100.0, formatDuration(elapsed))
	}

	return fileHashes, nil
}

func worker(wg *sync.WaitGroup, fileChan <-chan string, resultChan chan<- FileHash, errorChan chan<- error, hashedCount *int, hashedMutex *sync.Mutex, lastProgressUpdate *time.Time, totalFiles int, startTime time.Time) {
	defer wg.Done()

	for file := range fileChan {
		hasher := getHasher()
		hash, size, modTime, err := hashFile(file, hasher)
		if err != nil {
			errorChan <- fmt.Errorf("%s: %w", file, err)
			continue
		}

		// Compute perceptual hash for images if enabled
		var pHash string
		if cfg.PerceptualMode && isImageFile(file) {
			pHash, err = computePerceptualHash(file, cfg.PHashAlgorithm)
			if err != nil {
				// Log error but continue with regular hash
				if cfg.Verbose {
					log.Printf("%sCould not compute perceptual hash for %s: %v", emoji("⚠️"), file, err)
				}
			}
		}

		if cfg.Verbose {
			if pHash != "" {
				log.Printf("📄 %s: %s [phash: %s...] (%d bytes)", file, hash[:8]+"...", pHash[:8], size)
			} else {
				log.Printf("📄 %s: %s (%d bytes)", file, hash[:8]+"...", size)
			}
		}

		resultChan <- FileHash{
			Path:    file,
			Size:    size,
			Hash:    hash,
			ModTime: modTime,
			PHash:   pHash,
		}

		// Update progress
		hashedMutex.Lock()
		*hashedCount++
		currentHashed := *hashedCount
		hashedMutex.Unlock()

		// Update progress periodically
		if time.Since(*lastProgressUpdate) > progressUpdateInterval {
			*lastProgressUpdate = time.Now()
			if cfg.Verbose {
				log.Printf("🔐 Hashed %d/%d files (%.1f%%)", currentHashed, totalFiles, float64(currentHashed)*100/float64(totalFiles))
			} else {
				printProgress(currentHashed, totalFiles, startTime)
			}
		}
	}
}

// printProgress displays a progress bar with ETA
func printProgress(current, total int, startTime time.Time) {
	percentage := float64(current) * 100 / float64(total)
	barWidth := 20
	filled := int(percentage / 100 * float64(barWidth))
	empty := barWidth - filled

	// Choose bar characters based on emoji setting
	var filledChar, emptyChar, modeIcon string
	if cfg.NoEmoji {
		filledChar = "="
		emptyChar = " "
	} else {
		filledChar = "█"
		emptyChar = "░"
	}

	// Add perceptual mode indicator
	if cfg.PerceptualMode {
		if cfg.NoEmoji {
			modeIcon = "[IMG]"
		} else {
			modeIcon = emoji("🖼️")
		}
	}

	bar := "["
	for i := 0; i < filled; i++ {
		bar += filledChar
	}
	for i := 0; i < empty; i++ {
		bar += emptyChar
	}
	bar += "]"

	// Calculate ETA
	elapsed := time.Since(startTime).Seconds()
	var eta string
	if current > 0 {
		etaSeconds := float64(total-current) * (elapsed / float64(current))
		eta = formatDuration(etaSeconds)
	} else {
		eta = "..."
	}

	// Print progress with ETA and mode indicator
	fmt.Fprintf(os.Stderr, "\r%s%s%s%s %d/%d (%.1f%%) ETA: %s", emoji("🔐"), modeIcon, bar, emoji("▏"), current, total, percentage, eta)
}

// formatDuration converts seconds to a human-readable duration
func formatDuration(seconds float64) string {
	if seconds < 60 {
		return fmt.Sprintf("%.0fs", seconds)
	}
	minutes := int(seconds / 60)
	if minutes < 60 {
		return fmt.Sprintf("%dm %ds", minutes, int(seconds)%60)
	}
	hours := minutes / 60
	minutes = minutes % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func getHasher() hash.Hash {
	switch strings.ToLower(cfg.HashAlgorithm) {
	case "md5":
		return md5.New()
	case "sha1":
		return sha1.New()
	case "sha256":
		return sha256.New()
	default:
		return sha256.New()
	}
}

func hashFile(path string, hasher hash.Hash) (string, int64, time.Time, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, time.Time{}, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", 0, time.Time{}, err
	}

	if _, err := io.Copy(hasher, file); err != nil {
		return "", 0, time.Time{}, err
	}

	return hex.EncodeToString(hasher.Sum(nil)), info.Size(), info.ModTime(), nil
}

func findDuplicates(fileHashes []FileHash) []DuplicateGroup {
	// If perceptual mode is enabled, handle images differently
	if cfg.PerceptualMode {
		return findPerceptualDuplicates(fileHashes)
	}
	
	// Standard exact-match deduplication
	hashMap := make(map[string][]FileHash)

	for _, fh := range fileHashes {
		hashMap[fh.Hash] = append(hashMap[fh.Hash], fh)
	}

	var duplicates []DuplicateGroup
	for hash, files := range hashMap {
		if len(files) > 1 {
			duplicates = append(duplicates, DuplicateGroup{
				Hash:  hash,
				Size:  files[0].Size,
				Files: files,
				Similarity: 100.0, // Exact match
			})
		}
	}

	return duplicates
}

// findPerceptualDuplicates groups similar images together
func findPerceptualDuplicates(fileHashes []FileHash) []DuplicateGroup {
	var imageFiles []FileHash
	var regularFiles []FileHash
	
	// Separate images from regular files
	for _, fh := range fileHashes {
		if fh.PHash != "" {
			imageFiles = append(imageFiles, fh)
		} else {
			regularFiles = append(regularFiles, fh)
		}
	}
	
	// Group regular files by exact hash (standard dedup)
	var duplicates []DuplicateGroup
	hashMap := make(map[string][]FileHash)
	for _, fh := range regularFiles {
		hashMap[fh.Hash] = append(hashMap[fh.Hash], fh)
	}
	for hash, files := range hashMap {
		if len(files) > 1 {
			duplicates = append(duplicates, DuplicateGroup{
				Hash:  hash,
				Size:  files[0].Size,
				Files: files,
				Similarity: 100.0,
			})
		}
	}
	
	// Group images by perceptual similarity
	visited := make(map[int]bool)
	for i := 0; i < len(imageFiles); i++ {
		if visited[i] {
			continue
		}
		
		group := []FileHash{imageFiles[i]}
		visited[i] = true
		
		for j := i + 1; j < len(imageFiles); j++ {
			if visited[j] {
				continue
			}
			
			dist := hammingDistance(imageFiles[i].PHash, imageFiles[j].PHash)
			if dist >= 0 && dist <= cfg.SimilarityThreshold {
				group = append(group, imageFiles[j])
				visited[j] = true
			}
		}
		
		if len(group) > 1 {
			// Calculate average similarity
			avgSimilarity := 100.0 - (float64(cfg.SimilarityThreshold) / 64.0 * 100.0)
			if avgSimilarity < 50 {
				avgSimilarity = 50 + float64(cfg.SimilarityThreshold)
			}
			
			duplicates = append(duplicates, DuplicateGroup{
				Hash:  imageFiles[i].PHash, // Use perceptual hash as group ID
				Size:  imageFiles[i].Size,
				Files: group,
				Similarity: avgSimilarity,
			})
		}
	}
	
	return duplicates
}

func reportDuplicates(duplicates []DuplicateGroup) {
	if len(duplicates) == 0 {
		log.Println("✅ No duplicates found!")
		return
	}

	totalDuplicates := 0
	totalSpace := int64(0)
	
	// Count perceptual vs exact matches
	perceptualGroups := 0
	for _, group := range duplicates {
		if group.Similarity < 100.0 {
			perceptualGroups++
		}
	}

	if cfg.PerceptualMode && perceptualGroups > 0 {
		log.Printf("\n%sSimilar Images Found:", emoji("🖼️"))
	} else {
		log.Printf("\n%sDuplicate Files:", emoji("👯"))
	}
	log.Println(strings.Repeat("=", 70))

	for i, group := range duplicates {
		numDuplicates := len(group.Files) - 1
		space := group.Size * int64(numDuplicates)
		totalDuplicates += numDuplicates
		totalSpace += space

		keepIdx := selectFileToKeep(group)

		log.Printf("\n[%d] Hash: %s", i+1, group.Hash[:16]+"...")
		log.Printf("    Size: %s", formatBytes(group.Size))
		log.Printf("    Files: %d (keeping 1, removing %d)", len(group.Files), numDuplicates)

		// Show similarity for perceptual matches
		if group.Similarity < 100.0 {
			log.Printf("    Similarity: %.0f%% (perceptual match)", group.Similarity)
		}

		for j, fh := range group.Files {
			prefix := fmt.Sprintf("    %sKEEP", emoji("✓"))
			if j != keepIdx {
				prefix = fmt.Sprintf("    %sDELETE", emoji("✗"))
			}
			log.Printf("%s %s (modified: %s)", prefix, fh.Path, fh.ModTime.Format("2006-01-02 15:04:05"))
		}
	}

	log.Println("\n" + strings.Repeat("=", 70))
	if cfg.PerceptualMode && perceptualGroups > 0 {
		log.Printf("%sSummary: %d duplicates/similar files, %s of space can be freed (%d perceptual groups)",
			emoji("📊"), totalDuplicates, formatBytes(totalSpace), perceptualGroups)
	} else {
		log.Printf("%sSummary: %d duplicate files, %s of space can be freed",
			emoji("📊"), totalDuplicates, formatBytes(totalSpace))
	}
}

func selectFileToKeep(group DuplicateGroup) int {
	files := group.Files

	if strings.HasPrefix(cfg.KeepCriteria, "path:") {
		// Keep file matching specific path
		targetPath := strings.TrimPrefix(cfg.KeepCriteria, "path:")
		for i, fh := range files {
			if strings.Contains(fh.Path, targetPath) {
				return i
			}
		}
		return 0 // Default to first if not found
	}

	switch strings.ToLower(cfg.KeepCriteria) {
	case "oldest":
		oldestIdx := 0
		for i, fh := range files {
			if fh.ModTime.Before(files[oldestIdx].ModTime) {
				oldestIdx = i
			}
		}
		return oldestIdx

	case "newest":
		newestIdx := 0
		for i, fh := range files {
			if fh.ModTime.After(files[newestIdx].ModTime) {
				newestIdx = i
			}
		}
		return newestIdx

	case "largest":
		largestIdx := 0
		for i, fh := range files {
			if fh.Size > files[largestIdx].Size {
				largestIdx = i
			}
		}
		return largestIdx

	case "smallest":
		smallestIdx := 0
		for i, fh := range files {
			if fh.Size < files[smallestIdx].Size {
				smallestIdx = i
			}
		}
		return smallestIdx

	default:
		return 0
	}
}

func processDuplicates(duplicates []DuplicateGroup) error {
	var undoLog []UndoEntry

	// Create move directory if specified
	if cfg.MoveTo != "" {
		if err := os.MkdirAll(cfg.MoveTo, 0755); err != nil {
			return fmt.Errorf("failed to create move directory: %w", err)
		}
	}

	totalDeleted := 0
	totalSpace := int64(0)

	log.Printf("\n🗑️  %s duplicates...", map[bool]string{true: "Moving", false: "Deleting"}[cfg.MoveTo != ""])

	// Warn users about permanent deletion
	if cfg.Interactive && cfg.MoveTo == "" {
		log.Println("\n" + strings.Repeat("⚠️", 30))
		log.Println("⚠️  WARNING: Files will be PERMANENTLY deleted!")
		log.Println("⚠️  The -undo option only shows what was deleted.")
		log.Println("⚠️  Use -move-to <folder> to move files instead of deleting.")
		log.Println("⚠️" + strings.Repeat("=", 55))
		fmt.Print("Continue with permanent deletion? [y/N]: ")
		var confirm string
		fmt.Scanln(&confirm)
		if strings.ToLower(confirm) != "y" {
			log.Println("❓ Operation cancelled. No files were deleted.")
			return nil
		}
	}

	for _, group := range duplicates {
		keepIdx := selectFileToKeep(group)

		for i, fh := range group.Files {
			if i != keepIdx {
				// Interactive mode
				if cfg.Interactive {
					fmt.Printf("\nDelete %s? (%s) [y/n/q]: ", fh.Path, formatBytes(fh.Size))
					var response string
					fmt.Scanln(&response)
					if strings.ToLower(response) != "y" {
						if strings.ToLower(response) == "q" {
							log.Println("❓ Quitting...")
							return nil
						}
						continue
					}
				}

				var err error
				if cfg.MoveTo != "" {
					// Move to directory
					targetPath := filepath.Join(cfg.MoveTo, filepath.Base(fh.Path))
					// Handle name conflicts
					counter := 1
					for {
						if _, err := os.Stat(targetPath); os.IsNotExist(err) {
							break
						}
						base := filepath.Base(fh.Path)
						ext := filepath.Ext(base)
						name := strings.TrimSuffix(base, ext)
						targetPath = filepath.Join(cfg.MoveTo, fmt.Sprintf("%s_%d%s", name, counter, ext))
						counter++
					}
					err = os.Rename(fh.Path, targetPath)
					if err == nil {
						log.Printf("✓ Moved %s -> %s", fh.Path, targetPath)
					}
				} else {
					// Delete file
					err = os.Remove(fh.Path)
					if err == nil {
						log.Printf("✓ Deleted %s", fh.Path)
					}
				}

				if err != nil {
					log.Printf("❌ Failed to process %s: %v", fh.Path, err)
				} else {
					totalDeleted++
					totalSpace += fh.Size
					undoLog = append(undoLog, UndoEntry{
						Path:        fh.Path,
						Size:        fh.Size,
						ModTime:     fh.ModTime,
						Action:      "deleted",
						Timestamp:   time.Now(),
						TargetPath:  "",
					})
				}
			}
		}
	}

	log.Printf("\n✅ %s %d files, freed %s of space", map[bool]string{true: "Moved", false: "Deleted"}[cfg.MoveTo != ""], totalDeleted, formatBytes(totalSpace))

	// Save undo log
	if len(undoLog) > 0 && cfg.MoveTo == "" {
		if err := saveUndoLog(undoLog); err != nil {
			log.Printf("%sFailed to save undo log: %v", emoji("⚠️"), err)
		} else {
			log.Printf("%sUndo log saved (use -undo to view - files are NOT recoverable)", emoji("💾"))
		}
	}

	return nil
}

type UndoEntry struct {
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	ModTime    time.Time `json:"mod_time"`
	Action     string    `json:"action"`
	Timestamp  time.Time `json:"timestamp"`
	TargetPath string    `json:"target_path,omitempty"`
}

func saveUndoLog(entries []UndoEntry) error {
	return os.WriteFile(undoFile, []byte(fmt.Sprintf(`{"entries":%d,"files":%s}`,
		len(entries),
		toString(entries))), 0600)
}

func toString(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func undoLast() error {
	data, err := os.ReadFile(undoFile)
	if err != nil {
		return fmt.Errorf("no undo log found: %w", err)
	}

	log.Println("\n" + strings.Repeat("⚠️", 30))
	log.Println("⚠️  IMPORTANT: This undo log is INFORMATIONAL ONLY")
	log.Println("⚠️  Files that were deleted CANNOT be restored.")
	log.Println("⚠️  Only the metadata (what was deleted) is logged.")
	log.Println("⚠️" + strings.Repeat("=", 55))
	log.Println("")
	log.Println("💡 TIP: Next time, use -move-to <folder> to safely move duplicates")
	log.Println("💡       instead of permanently deleting them.")
	log.Println("")

	fmt.Print("View the undo log anyway? [y/N]: ")
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" {
		return nil
	}

	log.Println("")
	log.Printf("💾 Undo log contents (%s):\n", undoFile)
	log.Println(strings.Repeat("=", 70))

	var undoData map[string]interface{}
	if err := json.Unmarshal(data, &undoData); err != nil {
		return fmt.Errorf("invalid undo log: %w", err)
	}

	log.Printf("📊 Total files deleted: %d\n", undoData["entries"])
	log.Println("")

	// Display individual entries if available
	if entries, ok := undoData["files"].([]interface{}); ok {
		for i, entry := range entries {
			if e, ok := entry.(map[string]interface{}); ok {
				if i >= 10 { // Limit to 10 entries
					log.Println("...")
					break
				}
				path := e["path"].(string)
				size := int64(e["size"].(float64))
				timestamp := e["timestamp"].(string)
				log.Printf("  %s - %s - %s", path, formatBytes(size), timestamp)
			}
		}
	}

	log.Println("")
	log.Println(strings.Repeat("=", 70))
	log.Println("⚠️  These files are GONE and cannot be recovered.")
	log.Println("⚠️" + strings.Repeat("=", 55))

	return nil
}

func exportReport(duplicates []DuplicateGroup) error {
	type Report struct {
		Version      string          `json:"version"`
		Timestamp    time.Time       `json:"timestamp"`
		Config       Config          `json:"config"`
		DuplicateCount int           `json:"duplicate_count"`
		TotalSpace   int64          `json:"total_space"`
		Duplicates   []DuplicateGroup `json:"duplicates"`
	}

	totalSpace := int64(0)
	for _, group := range duplicates {
		totalSpace += group.Size * int64(len(group.Files)-1)
	}

	report := Report{
		Version:        version,
		Timestamp:      time.Now(),
		Config:         cfg,
		DuplicateCount: len(duplicates),
		TotalSpace:     totalSpace,
		Duplicates:     duplicates,
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(reportFile, data, 0644)
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// printStatistics displays detailed operation statistics
func printStatistics(stats *Statistics) {
	scanDuration := stats.ScanEnd.Sub(stats.ScanStart).Seconds()
	hashDuration := stats.HashEnd.Sub(stats.HashStart).Seconds()
	processDuration := stats.ProcessEnd.Sub(stats.ProcessStart).Seconds()
	totalDuration := stats.ProcessEnd.Sub(stats.ScanStart).Seconds()

	if cfg.Verbose || !cfg.NoEmoji {
		log.Println("")
		log.Println("📊 Detailed Statistics:")
		log.Println("─────────────────────────────────────────────────────")
		log.Printf("  Files scanned:      %d", stats.TotalFiles)
		log.Printf("  Total data size:     %s", formatBytes(stats.TotalBytes))
		log.Printf("  Duplicate files:     %d", stats.DuplicateFiles)
		log.Printf("  Duplicate size:      %s", formatBytes(stats.DuplicateBytes))
		if stats.ImageFiles > 0 {
			log.Printf("  Image files:        %d", stats.ImageFiles)
		}

		log.Println("")
		log.Println("  Time Breakdown:")
		log.Printf("    Scanning:      %s (%.1f%%)", formatDuration(scanDuration), scanDuration/totalDuration*100)
		log.Printf("    Hashing:        %s (%.1f%%)", formatDuration(hashDuration), hashDuration/totalDuration*100)
		log.Printf("    Processing:     %s (%.1f%%)", formatDuration(processDuration), processDuration/totalDuration*100)

		log.Println("")
		if stats.TotalFiles > 0 {
			scanRate := float64(stats.TotalFiles) / scanDuration
			hashRate := float64(stats.TotalFiles) / hashDuration
			byteRate := float64(stats.TotalBytes) / hashDuration

			log.Printf("  Speed:")
			log.Printf("    Scanning:      %.0f files/sec", scanRate)
			log.Printf("    Hashing:       %.0f files/sec", hashRate)
			log.Printf("    Throughput:    %s/sec", formatBytes(int64(byteRate)))
		}

		if len(stats.FilesByExt) > 0 {
			log.Println("")
			log.Println("  Files by Type:")
			sortedExts := make([]string, 0, len(stats.FilesByExt))
			for ext := range stats.FilesByExt {
				sortedExts = append(sortedExts, ext)
			}
			for i, ext := range sortedExts {
				if i > 9 { // Show top 10 file types
					break
				}
				count := stats.FilesByExt[ext]
				if ext == "" {
					ext = "(no extension)"
				}
				log.Printf("    %-5s: %d files", ext, count)
			}
		}

		log.Println("─────────────────────────────────────────────────────")
		log.Printf("  Total Time: %s\n", formatDuration(totalDuration))
	}
}

// compareImagesCLI handles the -compare flag for comparing two images
func compareImagesCLI() error {
	// Parse the compare argument (can be comma-separated or use -compare-with)
	var img1, img2 string
	
	if strings.Contains(cfg.CompareImg1, ",") {
		parts := strings.SplitN(cfg.CompareImg1, ",", 2)
		img1 = strings.TrimSpace(parts[0])
		img2 = strings.TrimSpace(parts[1])
	} else if cfg.CompareImg2 != "" {
		img1 = cfg.CompareImg1
		img2 = cfg.CompareImg2
	} else {
		return fmt.Errorf("usage: -compare img1,img2 OR -compare img1 -compare-with img2")
	}

	// Validate files exist
	for _, path := range []string{img1, img2} {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("cannot access %s: %w", path, err)
		}
		if !isImageFile(path) {
			return fmt.Errorf("%s is not a supported image file", path)
		}
	}

	log.Printf("Comparing images...")
	log.Printf("   Image 1: %s", img1)
	log.Printf("   Image 2: %s", img2)
	log.Printf("   Algorithm: %s", cfg.PHashAlgorithm)

	// Compute hashes for both images using all three algorithms
	algorithms := []string{"dhash", "ahash", "phash"}
	thresholds := map[string]int{"dhash": 10, "ahash": 12, "phash": 8}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("IMAGE COMPARISON RESULTS")
	fmt.Println(strings.Repeat("=", 70))

	for _, algo := range algorithms {
		hash1, err := computePerceptualHash(img1, algo)
		if err != nil {
			return fmt.Errorf("failed to hash %s: %w", img1, err)
		}

		hash2, err := computePerceptualHash(img2, algo)
		if err != nil {
			return fmt.Errorf("failed to hash %s: %w", img2, err)
		}

		dist := hammingDistance(hash1, hash2)
		similarity := 100.0 - (float64(dist)/64.0*100.0)
		threshold := thresholds[algo]
		isSimilar := dist <= threshold

		fmt.Printf("\n%s (%s):\n", strings.ToUpper(algo), algoDescriptions[algo])
		fmt.Printf("  Hash 1: %s...\n", hash1[:16])
		fmt.Printf("  Hash 2: %s...\n", hash2[:16])
		fmt.Printf("  Hamming Distance: %d/64\n", dist)
		fmt.Printf("  Similarity: %.1f%%\n", similarity)
		fmt.Printf("  Threshold: %d\n", threshold)
		
		if isSimilar {
			fmt.Printf("  Result: SIMILAR\n")
		} else {
			fmt.Printf("  Result: DIFFERENT\n")
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("RECOMMENDATION")
	fmt.Println(strings.Repeat("=", 70))
	
	// Use the requested algorithm for final recommendation
	reqHash1, _ := computePerceptualHash(img1, cfg.PHashAlgorithm)
	reqHash2, _ := computePerceptualHash(img2, cfg.PHashAlgorithm)
	reqDist := hammingDistance(reqHash1, reqHash2)
	reqSimilarity := 100.0 - (float64(reqDist)/64.0*100.0)
	
	if reqDist <= cfg.SimilarityThreshold {
		fmt.Printf("Images are SIMILAR (using %s, threshold %d)\n", 
			cfg.PHashAlgorithm, cfg.SimilarityThreshold)
		fmt.Printf("   Similarity: %.1f%% (distance: %d)\n", reqSimilarity, reqDist)
	} else {
		fmt.Printf("Images are DIFFERENT (using %s, threshold %d)\n",
			cfg.PHashAlgorithm, cfg.SimilarityThreshold)
		fmt.Printf("   Similarity: %.1f%% (distance: %d)\n", reqSimilarity, reqDist)
	}
	fmt.Println()

	return nil
}

var algoDescriptions = map[string]string{
	"dhash": "Difference Hash - Fast, good for near-duplicates",
	"ahash": "Average Hash - Balanced speed and accuracy",
	"phash": "Perceptual Hash - Most robust, slower",
}
