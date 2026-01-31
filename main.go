package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
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
	Dir        string
	Recursive  bool
	DryRun     bool
	Verbose    bool
	Workers    int
	MinSize    int64 // Minimum file size to check (bytes)
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
}

func main() {
	flag.Parse()

	log.SetFlags(log.Ltime)

	log.Println("🔍 File Deduplicator - Starting...")
	if cfg.Verbose {
		log.Printf("📁 Scanning directory: %s", cfg.Dir)
		log.Printf("🔄 Recursive: %v", cfg.Recursive)
		log.Printf("👷 Workers: %d", cfg.Workers)
		log.Printf("📏 Min size: %d bytes", cfg.MinSize)
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
			filteredFiles = append(filteredFiles, file)
		}
	}
	log.Printf("📏 After size filter: %d files", len(filteredFiles))

	// Compute hashes in parallel
	fileHashes, err := computeHashes(filteredFiles)
	if err != nil {
		log.Fatalf("❌ Error computing hashes: %v", err)
	}
	log.Printf("🔐 Computed %d hashes", len(fileHashes))

	// Find duplicates
	duplicates := findDuplicates(fileHashes)
	log.Printf("👯 Found %d duplicate groups", len(duplicates))

	// Report duplicates
	reportDuplicates(duplicates)

	// Delete duplicates if not dry run
	if !cfg.DryRun && len(duplicates) > 0 {
		if err := deleteDuplicates(duplicates); err != nil {
			log.Fatalf("❌ Error deleting duplicates: %v", err)
		}
	}

	elapsed := time.Since(startTime)
	log.Printf("✅ Complete in %v", elapsed)
}

func scanFiles(dir string, recursive bool) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
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

	return files, err
}

func computeHashes(files []string) ([]FileHash, error) {
	var wg sync.WaitGroup
	fileChan := make(chan string, cfg.Workers)
	resultChan := make(chan FileHash, len(files))
	errorChan := make(chan error, len(files))

	// Start worker goroutines
	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go worker(&wg, fileChan, resultChan, errorChan)
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

func worker(wg *sync.WaitGroup, fileChan <-chan string, resultChan chan<- FileHash, errorChan chan<- error) {
	defer wg.Done()

	for file := range fileChan {
		hash, size, modTime, err := hashFile(file)
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
	}
}

func hashFile(path string) (string, int64, time.Time, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, time.Time{}, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", 0, time.Time{}, err
	}

	hasher := sha256.New()
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
	log.Println(strings.Repeat("=", 60))

	for i, group := range duplicates {
		numDuplicates := len(group.Files) - 1
		space := group.Size * int64(numDuplicates)
		totalDuplicates += numDuplicates
		totalSpace += space

		log.Printf("\n[%d] Hash: %s", i+1, group.Hash[:16]+"...")
		log.Printf("    Size: %s", formatBytes(group.Size))
		log.Printf("    Files: %d (keeping 1, removing %d)", len(group.Files), numDuplicates)

		// Find the oldest file to keep
		oldestIdx := 0
		for i, fh := range group.Files {
			if fh.ModTime.Before(group.Files[oldestIdx].ModTime) {
				oldestIdx = i
			}
		}

		for j, fh := range group.Files {
			prefix := "    ✓ KEEP"
			if j != oldestIdx {
				prefix = "    ✗ DELETE"
			}
			log.Printf("%s %s (modified: %s)", prefix, fh.Path, fh.ModTime.Format("2006-01-02 15:04:05"))
		}
	}

	log.Println("\n" + strings.Repeat("=", 60))
	log.Printf("📊 Summary: %d duplicate files, %s of space can be freed",
		totalDuplicates, formatBytes(totalSpace))
}

func deleteDuplicates(duplicates []DuplicateGroup) error {
	totalDeleted := 0
	totalSpace := int64(0)

	log.Println("\n🗑️  Deleting duplicates...")

	for _, group := range duplicates {
		// Find the oldest file to keep
		oldestIdx := 0
		for i, fh := range group.Files {
			if fh.ModTime.Before(group.Files[oldestIdx].ModTime) {
				oldestIdx = i
			}
		}

		for i, fh := range group.Files {
			if i != oldestIdx {
				if err := os.Remove(fh.Path); err != nil {
					log.Printf("❌ Failed to delete %s: %v", fh.Path, err)
				} else {
					log.Printf("✓ Deleted %s", fh.Path)
					totalDeleted++
					totalSpace += fh.Size
				}
			}
		}
	}

	log.Printf("\n✅ Deleted %d files, freed %s of space", totalDeleted, formatBytes(totalSpace))
	return nil
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
