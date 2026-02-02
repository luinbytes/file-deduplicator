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
	version                = "2.0.0"
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
}

// DuplicateGroup represents a group of duplicate files
type DuplicateGroup struct {
	Hash  string
	Size  int64
	Files []FileHash
}

// Config holds application configuration
type Config struct {
	Dir           string
	Recursive     bool
	DryRun        bool
	Verbose       bool
	Workers       int
	MinSize       int64 // Minimum file size to check (bytes)
	Interactive   bool
	MoveTo        string // Move duplicates to this folder instead of deleting
	KeepCriteria  string // "oldest", "newest", "largest", "smallest", "first", "path"
	HashAlgorithm string // "sha256", "sha1", "md5"
	FilePattern  string // Only include files matching this pattern
	ExportReport  bool
	UndoLast      bool
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
	flag.BoolVar(&cfg.Interactive, "interactive", false, "Ask before deleting each duplicate")
	flag.StringVar(&cfg.MoveTo, "move-to", "", "Move duplicates to this folder instead of deleting")
	flag.StringVar(&cfg.KeepCriteria, "keep", "oldest", "File to keep criteria: oldest, newest, largest, smallest, first, or path:<path>")
	flag.StringVar(&cfg.HashAlgorithm, "hash", "sha256", "Hash algorithm: sha256, sha1, or md5")
	flag.StringVar(&cfg.FilePattern, "pattern", "", "File pattern to match (e.g., *.jpg, *.pdf)")
	flag.BoolVar(&cfg.ExportReport, "export", false, "Export duplicate report to JSON file")
	flag.BoolVar(&cfg.UndoLast, "undo", false, "Undo last operation")
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
			log.Printf("❓ Interactive mode enabled")
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
		if err := processDuplicates(duplicates); err != nil {
			log.Fatalf("❌ Error processing duplicates: %v", err)
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

		if cfg.Verbose {
			log.Printf("📄 %s: %s (%d bytes)", file, hash[:8]+"...", size)
		}

		resultChan <- FileHash{
			Path:    file,
			Size:    size,
			Hash:    hash,
			ModTime: modTime,
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

	log.Println("\n👯 Duplicate Files:")
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

		for j, fh := range group.Files {
			prefix := "    ✓ KEEP"
			if j != keepIdx {
				prefix = "    ✗ DELETE"
			}
			log.Printf("%s %s (modified: %s)", prefix, fh.Path, fh.ModTime.Format("2006-01-02 15:04:05"))
		}
	}

	log.Println("\n" + strings.Repeat("=", 70))
	log.Printf("📊 Summary: %d duplicate files, %s of space can be freed",
		totalDuplicates, formatBytes(totalSpace))
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
