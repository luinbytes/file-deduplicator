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

	"github.com/luinbytes/file-deduplicator/tui"
)

const (
	version                = "3.0.0"
	reportFile             = ".deduplicator_report.json"
	undoFile               = ".deduplicator_undo.json"
	maxHistory             = 100
	progressUpdateInterval  = 2 * time.Second
)

// FileHash represents a file and its hash
type FileHash struct {
	Path     string
	Size     int64
	Hash     string
	ModTime  time.Time
	PHash    string  // Perceptual hash for images
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
	TUI            bool   // Enable TUI mode (new interactive interface)
	MoveTo         string // Move duplicates to this folder instead of deleting
	KeepCriteria   string // "oldest", "newest", "largest", "smallest", "first", "path"
	HashAlgorithm  string // "sha256", "sha1", "md5"
	FilePattern    string // Only include files matching this pattern
	ExportReport   bool
	UndoLast       bool
	// Perceptual hashing options
	PerceptualMode bool   // Enable perceptual hashing for images
	PHashAlgorithm string // "dhash", "ahash", "phash"
	SimilarityThreshold int // Hamming distance threshold (0-64, default 10)
}

var (
	cfg Config
)

func init() {
	flag.StringVar(&cfg.Dir, "dir", ".", "Directory to scan for duplicates")
	flag.BoolVar(&cfg.Recursive, "recursive", true, "Scan directories recursively")
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "Show what would be deleted without actually deleting")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Show detailed output")
	flag.IntVar(&cfg.Workers, "workers", runtime.NumCPU(), "Number of worker goroutines")
	flag.Int64Var(&cfg.MinSize, "min-size", 1024, "Minimum file size in bytes (default: 1KB)")
	flag.BoolVar(&cfg.Interactive, "interactive", false, "Ask before deleting each duplicate (legacy mode)")
	flag.BoolVar(&cfg.TUI, "tui", false, "Use TUI interface for interactive deletion (recommended)")
	flag.StringVar(&cfg.MoveTo, "move-to", "", "Move duplicates to this folder instead of deleting")
	flag.StringVar(&cfg.KeepCriteria, "keep", "oldest", "File to keep criteria: oldest, newest, largest, smallest, first, or path:<path>")
	flag.StringVar(&cfg.HashAlgorithm, "hash", "sha256", "Hash algorithm: sha256, sha1, or md5")
	flag.StringVar(&cfg.FilePattern, "pattern", "", "File pattern to match (e.g., *.jpg, *.pdf)")
	flag.BoolVar(&cfg.ExportReport, "export", false, "Export duplicate report to JSON file")
	flag.BoolVar(&cfg.UndoLast, "undo", false, "Undo last operation")
	
	// Perceptual hashing flags
	flag.BoolVar(&cfg.PerceptualMode, "perceptual", false, "Enable perceptual hashing for images (finds similar images, not just exact duplicates)")
	flag.StringVar(&cfg.PHashAlgorithm, "phash-algo", "dhash", "Perceptual hash algorithm: dhash (fast), ahash, phash (robust)")
	flag.IntVar(&cfg.SimilarityThreshold, "similarity", 10, "Similarity threshold (0-64). Lower = stricter. Default 10.")
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

	log.SetFlags(log.Ltime)

	log.Printf("🔍 File Deduplicator v%s - Starting...", version)
	if cfg.Verbose {
		log.Printf("📁 Scanning directory: %s", cfg.Dir)
		log.Printf("🔄 Recursive: %v", cfg.Recursive)
		log.Printf("👷 Workers: %d", cfg.Workers)
		log.Printf("📏 Min size: %d bytes", cfg.MinSize)
		log.Printf("🔐 Hash algorithm: %s", cfg.HashAlgorithm)
		if cfg.FilePattern != "" {
			log.Printf("🎯 File pattern: %s", cfg.FilePattern)
		}
		if cfg.MoveTo != "" {
			log.Printf("📦 Move duplicates to: %s", cfg.MoveTo)
		}
		log.Printf("✋ Keep criteria: %s", cfg.KeepCriteria)
		if cfg.Interactive {
			log.Printf("❓ Interactive mode enabled (legacy)")
		}
		if cfg.TUI {
			log.Printf("🖥️  TUI mode enabled")
		}
		if cfg.PerceptualMode {
			log.Printf("🖼️  Perceptual mode enabled (%s, threshold: %d)", cfg.PHashAlgorithm, cfg.SimilarityThreshold)
		}
	}

	startTime := time.Now()

	// Scan files
	files, err := scanFiles(cfg.Dir, cfg.Recursive)
	if err != nil {
		log.Fatalf("❌ Error scanning files: %v", err)
	}

	log.Printf("📊 Found %d files", len(files))

	// Filter by minimum size
	var filteredFiles []string
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			if cfg.Verbose {
				log.Printf("⚠️  Could not stat %s: %v", file, err)
			}
			continue
		}
		if info.Size() >= cfg.MinSize {
			// Filter by file pattern if specified
			if cfg.FilePattern != "" {
				matched, err := filepath.Match(cfg.FilePattern, filepath.Base(file))
				if err != nil {
					log.Printf("⚠️  Invalid pattern %s: %v", cfg.FilePattern, err)
					continue
				}
				if !matched {
					if cfg.Verbose {
						log.Printf("🚫 Skipping non-matching file: %s", file)
					}
					continue
				}
			}
			filteredFiles = append(filteredFiles, file)
		}
	}
	log.Printf("📏 After filters: %d files", len(filteredFiles))

	// Compute hashes in parallel
	fileHashes, err := computeHashes(filteredFiles)
	if err != nil {
		log.Fatalf("❌ Error computing hashes: %v", err)
	}
	if !cfg.Verbose {
		fmt.Fprintln(os.Stderr) // Newline after progress bar
	}
	log.Printf("🔐 Computed %d hashes", len(fileHashes))

	// Find duplicates
	duplicates := findDuplicates(fileHashes)
	log.Printf("👯 Found %d duplicate groups", len(duplicates))

	// Report duplicates
	reportDuplicates(duplicates)

	// Export report if requested
	if cfg.ExportReport {
		if err := exportReport(duplicates); err != nil {
			log.Printf("⚠️  Failed to export report: %v", err)
		} else {
			log.Printf("📄 Report exported to %s", reportFile)
		}
	}

	// Process duplicates if not dry run
	if !cfg.DryRun && len(duplicates) > 0 {
		if cfg.TUI {
			if err := processDuplicatesTUI(duplicates); err != nil {
				log.Fatalf("❌ Error processing duplicates: %v", err)
			}
		} else if cfg.Interactive {
			if err := processDuplicates(duplicates); err != nil {
				log.Fatalf("❌ Error processing duplicates: %v", err)
			}
		} else {
			if err := processDuplicates(duplicates); err != nil {
				log.Fatalf("❌ Error processing duplicates: %v", err)
			}
		}
	}

	elapsed := time.Since(startTime)
	log.Printf("✅ Complete in %v", elapsed)
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
					log.Printf("🚫 Skipping hidden directory: %s", path)
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
				log.Printf("🚫 Skipping hidden file: %s", path)
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
	const progressUpdateInterval = 2 * time.Second

	// Start worker goroutines
	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go worker(&wg, fileChan, resultChan, errorChan, &hashedCount, &hashedMutex, &lastProgressUpdate, totalFiles)
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
			log.Printf("⚠️  Error: %v", err)
		}
	}

	return fileHashes, nil
}

func worker(wg *sync.WaitGroup, fileChan <-chan string, resultChan chan<- FileHash, errorChan chan<- error, hashedCount *int, hashedMutex *sync.Mutex, lastProgressUpdate *time.Time, totalFiles int) {
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
					log.Printf("⚠️  Could not compute perceptual hash for %s: %v", file, err)
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
				percentage := float64(currentHashed) * 100 / float64(totalFiles)
				fmt.Fprintf(os.Stderr, "\r🔐 Hashing: %d/%d files (%.1f%%)", currentHashed, totalFiles, percentage)
			}
		}
	}
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
		log.Println("\n🖼️  Similar Images Found:")
	} else {
		log.Println("\n👯 Duplicate Files:")
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
			prefix := "    ✓ KEEP"
			if j != keepIdx {
				prefix = "    ✗ DELETE"
			}
			log.Printf("%s %s (modified: %s)", prefix, fh.Path, fh.ModTime.Format("2006-01-02 15:04:05"))
		}
	}

	log.Println("\n" + strings.Repeat("=", 70))
	if cfg.PerceptualMode && perceptualGroups > 0 {
		log.Printf("📊 Summary: %d duplicates/similar files, %s of space can be freed (%d perceptual groups)",
			totalDuplicates, formatBytes(totalSpace), perceptualGroups)
	} else {
		log.Printf("📊 Summary: %d duplicate files, %s of space can be freed",
			totalDuplicates, formatBytes(totalSpace))
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
			log.Printf("⚠️  Failed to save undo log: %v", err)
		} else {
			log.Printf("💾 Undo log saved (use -undo to restore)")
		}
	}

	return nil
}

// processDuplicatesTUI handles duplicate processing with the new TUI interface
func processDuplicatesTUI(duplicates []DuplicateGroup) error {
	// Convert DuplicateGroup to TUI format
	tuiGroups := make([]tui.DuplicateGroup, len(duplicates))
	for i, group := range duplicates {
		files := make([]struct {
			Path    string
			Size    int64
			ModTime string
		}, len(group.Files))
		for j, f := range group.Files {
			files[j] = struct {
				Path    string
				Size    int64
				ModTime string
			}{
				Path:    f.Path,
				Size:    f.Size,
				ModTime: f.ModTime.Format("2006-01-02"),
			}
		}
		tuiGroups[i] = tui.ConvertDuplicateGroup(group.Hash, group.Size, files, group.Similarity)
	}

	// Run TUI
	filesToDelete, err := tui.Run(tuiGroups)
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Process the selected files
	var undoLog []UndoEntry
	totalDeleted := 0
	totalSpace := int64(0)

	log.Printf("\n🗑️  Deleting %d selected files...", len(filesToDelete))

	for _, path := range filesToDelete {
		// Find the file info from duplicates
		var fileInfo FileHash
		found := false
		for _, group := range duplicates {
			for _, f := range group.Files {
				if f.Path == path {
					fileInfo = f
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		if !found {
			log.Printf("⚠️  File not found in duplicates: %s", path)
			continue
		}

		if cfg.MoveTo != "" {
			// Move to directory
			targetPath := filepath.Join(cfg.MoveTo, filepath.Base(path))
			counter := 1
			for {
				if _, err := os.Stat(targetPath); os.IsNotExist(err) {
					break
				}
				base := filepath.Base(path)
				ext := filepath.Ext(base)
				name := strings.TrimSuffix(base, ext)
				targetPath = filepath.Join(cfg.MoveTo, fmt.Sprintf("%s_%d%s", name, counter, ext))
				counter++
			}
			if err := os.Rename(path, targetPath); err != nil {
				log.Printf("❌ Failed to move %s: %v", path, err)
			} else {
				log.Printf("✓ Moved %s -> %s", path, targetPath)
				totalDeleted++
				totalSpace += fileInfo.Size
			}
		} else {
			// Delete file
			if err := os.Remove(path); err != nil {
				log.Printf("❌ Failed to delete %s: %v", path, err)
			} else {
				log.Printf("✓ Deleted %s", path)
				totalDeleted++
				totalSpace += fileInfo.Size
				undoLog = append(undoLog, UndoEntry{
					Path:      path,
					Size:      fileInfo.Size,
					ModTime:   fileInfo.ModTime,
					Action:    "deleted",
					Timestamp: time.Now(),
				})
			}
		}
	}

	log.Printf("\n✅ %s %d files, freed %s of space", map[bool]string{true: "Moved", false: "Deleted"}[cfg.MoveTo != ""], totalDeleted, formatBytes(totalSpace))

	// Save undo log
	if len(undoLog) > 0 && cfg.MoveTo == "" {
		if err := saveUndoLog(undoLog); err != nil {
			log.Printf("⚠️  Failed to save undo log: %v", err)
		} else {
			log.Printf("💾 Undo log saved (use -undo to restore)")
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

	log.Printf("🔄 Undo log found. Note: Files that were deleted cannot be restored (only the metadata is logged).\n")
	log.Printf("If you used -move-to option, files are in that directory.\n")
	
	fmt.Print("Continue? [y/n]: ")
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" {
		return nil
	}

	log.Println("⚠️  Undo is informational only - deleted files cannot be recovered")
	log.Printf("💾 View undo log at: %s\n", undoFile)
	
	var undoData map[string]interface{}
	if err := json.Unmarshal(data, &undoData); err != nil {
		return fmt.Errorf("invalid undo log: %w", err)
	}
	
	log.Printf("📊 %d files were deleted\n", undoData["entries"])
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
