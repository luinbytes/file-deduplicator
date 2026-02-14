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

// Edge case tests

func TestScanFilesEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Don't create any files - test empty directory
	files, err := scanFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	if len(files) != 0 {
		t.Errorf("scanFiles() found %d files in empty directory, want 0", len(files))
	}
}

func TestScanFilesNestedEmptyDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested empty directories
	subDirs := []string{
		filepath.Join(tmpDir, "level1"),
		filepath.Join(tmpDir, "level1", "level2"),
		filepath.Join(tmpDir, "level1", "level2", "level3"),
	}

	for _, dir := range subDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Should scan successfully with no files
	files, err := scanFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	if len(files) != 0 {
		t.Errorf("scanFiles() found %d files in nested empty directories, want 0", len(files))
	}
}

func TestScanFilesSymlinks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a real file
	realFile := filepath.Join(tmpDir, "real.txt")
	if err := os.WriteFile(realFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create real file: %v", err)
	}

	// Create a symlink to the real file
	symlink := filepath.Join(tmpDir, "link.txt")
	if err := os.Symlink(realFile, symlink); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Symlink should be followed and file should be found
	files, err := scanFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	// Both real file and symlink should be found
	// (symlinks are followed by filepath.Walk)
	if len(files) < 1 {
		t.Errorf("scanFiles() found %d files, want at least 1", len(files))
	}
}

func TestScanFilesSymlinkToDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory with a file
	realDir := filepath.Join(tmpDir, "realdir")
	if err := os.Mkdir(realDir, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(realDir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Create a symlink to the directory
	linkDir := filepath.Join(tmpDir, "linkdir")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Should handle symlinked directories
	files, err := scanFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	// Should find the file (may appear twice due to symlink)
	if len(files) < 1 {
		t.Errorf("scanFiles() found %d files, want at least 1", len(files))
	}
}

func TestScanFilesBrokenSymlink(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a broken symlink (points to non-existent file)
	brokenLink := filepath.Join(tmpDir, "broken")
	if err := os.Symlink("/nonexistent/file", brokenLink); err != nil {
		t.Fatalf("Failed to create broken symlink: %v", err)
	}

	// Note: filepath.Walk reports broken symlinks, but they can't be hashed
	// The scanFiles function will report them, but hashFile will fail
	files, err := scanFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	// Broken symlinks may be found by scan, but should fail when hashed
	// This is acceptable behavior - the error is caught during hashing
	// We just verify no panic/crash occurs
	_ = files // Files may be found, but will fail hashing
}

func TestHashFilePermissionDenied(t *testing.T) {
	// Skip on Windows - permission handling differs
	if os.PathSeparator == '\\' {
		t.Skip("Skipping on Windows")
	}

	tmpDir := t.TempDir()
	restrictedFile := filepath.Join(tmpDir, "restricted.txt")

	// Create a file with no read permissions
	if err := os.WriteFile(restrictedFile, []byte("content"), 0000); err != nil {
		t.Fatalf("Failed to create restricted file: %v", err)
	}

	// Try to hash the file - should fail gracefully
	hasher := sha256.New()
	_, _, _, err := hashFile(restrictedFile, hasher)
	if err == nil {
		// Restore permissions before cleanup
		os.Chmod(restrictedFile, 0644)
		t.Error("hashFile() should fail on file with no read permissions")
	}

	// Restore permissions before cleanup
	os.Chmod(restrictedFile, 0644)
}

func TestScanFilesHiddenDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create visible file
	if err := os.WriteFile(filepath.Join(tmpDir, "visible.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create visible file: %v", err)
	}

	// Create hidden directory with file
	hiddenDir := filepath.Join(tmpDir, ".hidden")
	if err := os.Mkdir(hiddenDir, 0755); err != nil {
		t.Fatalf("Failed to create hidden directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hiddenDir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file in hidden directory: %v", err)
	}

	// Hidden directories should be skipped
	files, err := scanFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	if len(files) != 1 {
		t.Errorf("scanFiles() found %d files, want 1 (hidden directory should be skipped)", len(files))
	}
}

func TestScanFilesHiddenFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create visible file
	if err := os.WriteFile(filepath.Join(tmpDir, "visible.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create visible file: %v", err)
	}

	// Create hidden file
	if err := os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create hidden file: %v", err)
	}

	// Hidden files should be skipped
	files, err := scanFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	if len(files) != 1 {
		t.Errorf("scanFiles() found %d files, want 1 (hidden file should be skipped)", len(files))
	}
}

func TestSelectFileToKeep(t *testing.T) {
	now := time.Now()
	files := []FileHash{
		{Path: "/path/oldest.txt", Size: 100, ModTime: now.Add(-24 * time.Hour)},
		{Path: "/path/newest.txt", Size: 200, ModTime: now},
		{Path: "/path/largest.txt", Size: 500, ModTime: now.Add(-12 * time.Hour)},
		{Path: "/path/smallest.txt", Size: 50, ModTime: now.Add(-6 * time.Hour)},
	}

	group := DuplicateGroup{Files: files}

	tests := []struct {
		name     string
		criteria string
		wantPath string
	}{
		{"oldest", "oldest", "/path/oldest.txt"},
		{"newest", "newest", "/path/newest.txt"},
		{"largest", "largest", "/path/largest.txt"},
		{"smallest", "smallest", "/path/smallest.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg.KeepCriteria = tt.criteria
			idx := selectFileToKeep(group)
			if group.Files[idx].Path != tt.wantPath {
				t.Errorf("selectFileToKeep(%s) = %s, want %s", tt.criteria, group.Files[idx].Path, tt.wantPath)
			}
		})
	}
}

func TestSelectFileToKeepPath(t *testing.T) {
	files := []FileHash{
		{Path: "/path/to/file1.txt"},
		{Path: "/path/to/file2.txt"},
		{Path: "/another/path/file3.txt"},
	}

	group := DuplicateGroup{Files: files}

	cfg.KeepCriteria = "path:another"
	idx := selectFileToKeep(group)
	if group.Files[idx].Path != "/another/path/file3.txt" {
		t.Errorf("selectFileToKeep(path:another) = %s, want /another/path/file3.txt", group.Files[idx].Path)
	}
}

func TestFormatFileError(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		err      error
		contains string
	}{
		{"permission denied", "/test/file", os.ErrPermission, "Permission denied"},
		{"file not found", "/test/file", os.ErrNotExist, "File not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatFileError(tt.path, tt.err)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("formatFileError() = %s, should contain %s", got, tt.contains)
			}
		})
	}
}

func TestFindDuplicatesNoDuplicates(t *testing.T) {
	// All files have unique hashes
	fileHashes := []FileHash{
		{Path: "/a.txt", Size: 100, Hash: "hash1"},
		{Path: "/b.txt", Size: 100, Hash: "hash2"},
		{Path: "/c.txt", Size: 100, Hash: "hash3"},
	}

	duplicates := findDuplicates(fileHashes)
	if len(duplicates) != 0 {
		t.Errorf("findDuplicates() found %d groups, want 0", len(duplicates))
	}
}

func TestFindDuplicatesSingleFile(t *testing.T) {
	// Only one file - can't have duplicates
	fileHashes := []FileHash{
		{Path: "/a.txt", Size: 100, Hash: "hash1"},
	}

	duplicates := findDuplicates(fileHashes)
	if len(duplicates) != 0 {
		t.Errorf("findDuplicates() found %d groups with single file, want 0", len(duplicates))
	}
}
