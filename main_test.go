package main

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHashFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("Hello, World!")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hasher := sha256.New()
	hash, size, modTime, err := hashFile(testFile, hasher)
	if err != nil {
		t.Fatalf("hashFile() error = %v", err)
	}

	// Verify hash is not empty
	if hash == "" {
		t.Error("hashFile() returned empty hash")
	}

	// Verify size matches content length
	if size != int64(len(content)) {
		t.Errorf("hashFile() size = %d, want %d", size, len(content))
	}

	// Verify mod time is recent (within last minute)
	if time.Since(modTime) > time.Minute {
		t.Error("hashFile() returned old modification time")
	}

	// Verify hash is deterministic
	hasher2 := sha256.New()
	hash2, _, _, err := hashFile(testFile, hasher2)
	if err != nil {
		t.Fatalf("hashFile() error = %v", err)
	}
	if hash != hash2 {
		t.Error("hashFile() returned different hashes for same file")
	}
}

func TestFindDuplicates(t *testing.T) {
	fileHashes := []FileHash{
		{
			Path:    "/path/to/file1.txt",
			Size:    100,
			Hash:    "hash1",
			ModTime: time.Now(),
		},
		{
			Path:    "/path/to/file2.txt",
			Size:    100,
			Hash:    "hash1",
			ModTime: time.Now(),
		},
		{
			Path:    "/path/to/file3.txt",
			Size:    200,
			Hash:    "hash2",
			ModTime: time.Now(),
		},
		{
			Path:    "/path/to/file4.txt",
			Size:    100,
			Hash:    "hash1",
			ModTime: time.Now(),
		},
	}

	duplicates := findDuplicates(fileHashes)

	if len(duplicates) != 1 {
		t.Errorf("findDuplicates() found %d groups, want 1", len(duplicates))
	}

	if len(duplicates) > 0 {
		group := duplicates[0]
		if group.Hash != "hash1" {
			t.Errorf("findDuplicates() hash = %s, want hash1", group.Hash)
		}
		if len(group.Files) != 3 {
			t.Errorf("findDuplicates() files = %d, want 3", len(group.Files))
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{"Bytes", 500, "500 B"},
		{"Kilobytes", 1536, "1.5 KB"},
		{"Megabytes", 1572864, "1.5 MB"},
		{"Gigabytes", 1610612736, "1.5 GB"},
		{"Terabytes", 1649267441664, "1.5 TB"},
		{"Exact KB", 1024, "1.0 KB"},
		{"Zero", 0, "0 B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBytes(tt.bytes)
			if got != tt.want {
				t.Errorf("formatBytes(%d) = %s, want %s", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestScanFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []struct {
		name    string
		content string
		hidden  bool
	}{
		{"file1.txt", "content1", false},
		{"file2.txt", "content2", false},
		{".hiddenfile", "hidden", true},
	}

	for _, tf := range testFiles {
		path := filepath.Join(tmpDir, tf.name)
		if err := os.WriteFile(path, []byte(tf.content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create a subdirectory with files
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file3.txt"), []byte("content3"), 0644); err != nil {
		t.Fatalf("Failed to create subdirectory file: %v", err)
	}

	// Test recursive scan
	files, err := scanFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	// Should find 3 files (file1, file2, and file3 in subdir)
	// Hidden files (starting with .) should be skipped
	if len(files) != 3 {
		t.Errorf("scanFiles() found %d files, want 3", len(files))
	}

	// Verify hidden files (starting with .) are excluded
	for _, file := range files {
		basename := filepath.Base(file)
		if strings.HasPrefix(basename, ".") {
			t.Errorf("scanFiles() should exclude hidden files starting with dot, found: %s", basename)
		}
	}
}

func TestScanFilesNonRecursive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	if err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a subdirectory with files
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create subdirectory file: %v", err)
	}

	// Test non-recursive scan
	files, err := scanFiles(tmpDir, false)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	// Should only find 1 file (file1)
	if len(files) != 1 {
		t.Errorf("scanFiles() found %d files, want 1", len(files))
	}

	if filepath.Base(files[0]) != "file1.txt" {
		t.Errorf("scanFiles() found wrong file: %s", files[0])
	}
}

// Benchmark tests
func BenchmarkHashFile(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := make([]byte, 1024*1024) // 1MB
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasher := sha256.New()
		_, _, _, _ = hashFile(testFile, hasher)
	}
}

func BenchmarkFindDuplicates(b *testing.B) {
	fileHashes := make([]FileHash, 1000)
	for i := 0; i < 1000; i++ {
		fileHashes[i] = FileHash{
			Path:    "/path/to/file.txt",
			Size:    100,
			Hash:    "hash" + string(rune(i%10)),
			ModTime: time.Now(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = findDuplicates(fileHashes)
	}
}
