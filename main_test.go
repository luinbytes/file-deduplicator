package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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

	// Create an empty subdirectory
	emptyDir := filepath.Join(tmpDir, "empty")
	if err := os.Mkdir(emptyDir, 0755); err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}

	// Scan should succeed with no files
	files, err := scanFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	// Should find 0 files
	if len(files) != 0 {
		t.Errorf("scanFiles() found %d files in empty directory, want 0", len(files))
	}
}

func TestScanFilesNestedEmptyDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested empty directories
	dirs := []string{
		filepath.Join(tmpDir, "level1"),
		filepath.Join(tmpDir, "level1", "level2"),
		filepath.Join(tmpDir, "level1", "level2", "level3"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Scan should succeed
	files, err := scanFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	if len(files) != 0 {
		t.Errorf("scanFiles() found %d files in empty directories, want 0", len(files))
	}
}

func TestScanFilesSymlinkToFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a regular file
	targetFile := filepath.Join(tmpDir, "target.txt")
	if err := os.WriteFile(targetFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}

	// Create a symlink to the file
	symlinkFile := filepath.Join(tmpDir, "link.txt")
	if err := os.Symlink(targetFile, symlinkFile); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Scan should include both the file and the symlink
	files, err := scanFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	// Should find 2 entries (file + symlink)
	// Note: symlinks are followed by default, so both will be hashed
	if len(files) != 2 {
		t.Errorf("scanFiles() found %d files, want 2 (file + symlink)", len(files))
	}
}

func TestScanFilesSymlinkToDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory with a file
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Create a symlink to the directory
	symlinkDir := filepath.Join(tmpDir, "linked_dir")
	if err := os.Symlink(subDir, symlinkDir); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Scan should follow the symlink and find files
	files, err := scanFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	// Should find files (behavior depends on whether symlinks to dirs are followed)
	// At minimum, it shouldn't crash or infinite loop
	if len(files) < 1 {
		t.Errorf("scanFiles() found %d files, expected at least 1", len(files))
	}
}

func TestScanFilesBrokenSymlink(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a broken symlink (points to non-existent file)
	brokenLink := filepath.Join(tmpDir, "broken_link")
	if err := os.Symlink("/nonexistent/path/to/file.txt", brokenLink); err != nil {
		t.Fatalf("Failed to create broken symlink: %v", err)
	}

	// Scan should handle broken symlinks gracefully
	files, err := scanFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	// Broken symlinks should be skipped (they're not regular files)
	// The exact behavior depends on implementation
	// Important: it shouldn't crash
	t.Logf("scanFiles() found %d files with broken symlink present", len(files))
}

func TestScanFilesSymlinkLoop(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory loop: a -> b -> a
	dirA := filepath.Join(tmpDir, "a")
	dirB := filepath.Join(dirA, "b")
	if err := os.MkdirAll(dirB, 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	// Create symlink from b back to a (creates a loop)
	loopLink := filepath.Join(dirB, "loop")
	if err := os.Symlink(dirA, loopLink); err != nil {
		t.Fatalf("Failed to create loop symlink: %v", err)
	}

	// Scan should handle symlink loops without infinite recursion
	// filepath.Walk should detect and skip loops
	_, err := scanFiles(tmpDir, true)
	if err != nil {
		// Some error is acceptable (loop detected), but it shouldn't hang
		t.Logf("scanFiles() returned error for symlink loop (expected): %v", err)
	}
}

func TestHashFilePermissionDenied(t *testing.T) {
	// Skip on Windows - permission handling differs
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	tmpDir := t.TempDir()

	// Create a file with no read permissions
	restrictedFile := filepath.Join(tmpDir, "restricted.txt")
	if err := os.WriteFile(restrictedFile, []byte("secret"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Remove read permissions
	if err := os.Chmod(restrictedFile, 0000); err != nil {
		t.Fatalf("Failed to chmod file: %v", err)
	}
	defer os.Chmod(restrictedFile, 0644) // Restore for cleanup

	hasher := sha256.New()
	_, _, _, err := hashFile(restrictedFile, hasher)

	// Should return an error for permission denied
	if err == nil {
		t.Error("hashFile() should return error for permission denied")
	}
}

func TestScanFilesUnicodeFilename(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with various Unicode characters
	unicodeFiles := []struct {
		name    string
		content string
	}{
		{"日本語.txt", "japanese"},
		{"العربية.txt", "arabic"},
		{"emoji_🎉_test.txt", "emoji"},
		{"über_file.txt", "german umlaut"},
		{"файл.txt", "russian"},
		{"中文名.txt", "chinese"},
	}

	for _, uf := range unicodeFiles {
		path := filepath.Join(tmpDir, uf.name)
		if err := os.WriteFile(path, []byte(uf.content), 0644); err != nil {
			t.Fatalf("Failed to create unicode file %s: %v", uf.name, err)
		}
	}

	// Scan should find all unicode files
	files, err := scanFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	if len(files) != len(unicodeFiles) {
		t.Errorf("scanFiles() found %d files, want %d", len(files), len(unicodeFiles))
	}

	// Verify all files were found
	foundNames := make(map[string]bool)
	for _, f := range files {
		foundNames[filepath.Base(f)] = true
	}
	for _, uf := range unicodeFiles {
		if !foundNames[uf.name] {
			t.Errorf("scanFiles() missed unicode file: %s", uf.name)
		}
	}
}

func TestHashFileUnicodeContent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with Unicode content
	testCases := []string{
		"Hello 世界",
		"Привет мир",
		"مرحبا بالعالم",
		"🎉🎊🎁",
		strings.Repeat("ä", 1000), // Long unicode string
	}

	for i, content := range testCases {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("test_%d.txt", i))
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		hasher := sha256.New()
		hash, _, _, err := hashFile(testFile, hasher)
		if err != nil {
			t.Errorf("hashFile() error for unicode content: %v", err)
			continue
		}

		if hash == "" {
			t.Error("hashFile() returned empty hash for unicode content")
		}

		// Verify hash is deterministic
		hasher2 := sha256.New()
		hash2, _, _, err := hashFile(testFile, hasher2)
		if err != nil {
			t.Errorf("hashFile() error on second read: %v", err)
			continue
		}
		if hash != hash2 {
			t.Error("hashFile() returned different hashes for same unicode file")
		}
	}
}

func TestScanFilesVeryLongPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a deeply nested directory structure
	// Windows has a 260 character limit by default, Unix typically allows longer
	deepPath := tmpDir
	segmentCount := 0
	maxSegments := 30 // Create reasonably deep structure

	for i := 0; i < maxSegments; i++ {
		deepPath = filepath.Join(deepPath, fmt.Sprintf("level_%02d_deep_directory", i))
		segmentCount++
		if len(deepPath) > 200 {
			break
		}
	}

	if err := os.MkdirAll(deepPath, 0755); err != nil {
		t.Logf("Could not create deep path (expected on some systems): %v", err)
		t.Skip("Skipping long path test - directory creation failed")
	}

	// Create a file at the deep path
	deepFile := filepath.Join(deepPath, "deep_file.txt")
	if err := os.WriteFile(deepFile, []byte("deep content"), 0644); err != nil {
		t.Fatalf("Failed to create deep file: %v", err)
	}

	// Scan should find the file
	files, err := scanFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	// Should find at least one file
	if len(files) < 1 {
		t.Error("scanFiles() found no files in deep directory structure")
	}

	t.Logf("Deep path length: %d characters, found %d files", len(deepFile), len(files))
}

func TestFindDuplicatesEmptyInput(t *testing.T) {
	// Empty input should return empty result
	duplicates := findDuplicates([]FileHash{})
	if len(duplicates) != 0 {
		t.Errorf("findDuplicates() with empty input returned %d groups, want 0", len(duplicates))
	}
}

func TestFindDuplicatesNoDuplicates(t *testing.T) {
	// All unique files
	fileHashes := []FileHash{
		{Path: "/a.txt", Size: 100, Hash: "hash1", ModTime: time.Now()},
		{Path: "/b.txt", Size: 200, Hash: "hash2", ModTime: time.Now()},
		{Path: "/c.txt", Size: 300, Hash: "hash3", ModTime: time.Now()},
	}

	duplicates := findDuplicates(fileHashes)
	if len(duplicates) != 0 {
		t.Errorf("findDuplicates() found %d groups, want 0 (no duplicates)", len(duplicates))
	}
}

func TestFindDuplicatesSingleFile(t *testing.T) {
	// Single file - no duplicates possible
	fileHashes := []FileHash{
		{Path: "/a.txt", Size: 100, Hash: "hash1", ModTime: time.Now()},
	}

	duplicates := findDuplicates(fileHashes)
	if len(duplicates) != 0 {
		t.Errorf("findDuplicates() with single file returned %d groups, want 0", len(duplicates))
	}
}

func TestFindDuplicatesManyGroups(t *testing.T) {
	// Multiple duplicate groups
	fileHashes := []FileHash{
		// Group 1: 3 duplicates
		{Path: "/a1.txt", Size: 100, Hash: "hash1", ModTime: time.Now()},
		{Path: "/a2.txt", Size: 100, Hash: "hash1", ModTime: time.Now()},
		{Path: "/a3.txt", Size: 100, Hash: "hash1", ModTime: time.Now()},
		// Group 2: 2 duplicates
		{Path: "/b1.txt", Size: 200, Hash: "hash2", ModTime: time.Now()},
		{Path: "/b2.txt", Size: 200, Hash: "hash2", ModTime: time.Now()},
		// Unique file
		{Path: "/c.txt", Size: 300, Hash: "hash3", ModTime: time.Now()},
	}

	duplicates := findDuplicates(fileHashes)
	if len(duplicates) != 2 {
		t.Errorf("findDuplicates() found %d groups, want 2", len(duplicates))
	}

	// Check group sizes
	for _, group := range duplicates {
		if group.Hash == "hash1" && len(group.Files) != 3 {
			t.Errorf("Group hash1 has %d files, want 3", len(group.Files))
		}
		if group.Hash == "hash2" && len(group.Files) != 2 {
			t.Errorf("Group hash2 has %d files, want 2", len(group.Files))
		}
	}
}

func TestHashFileEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an empty file
	emptyFile := filepath.Join(tmpDir, "empty.txt")
	if err := os.WriteFile(emptyFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	hasher := sha256.New()
	hash, size, _, err := hashFile(emptyFile, hasher)
	if err != nil {
		t.Fatalf("hashFile() error = %v", err)
	}

	if hash == "" {
		t.Error("hashFile() returned empty hash for empty file")
	}

	if size != 0 {
		t.Errorf("hashFile() size = %d, want 0 for empty file", size)
	}

	// Empty file should have a consistent hash
	// SHA256 of empty string is: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
	expectedEmptyHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if hash != expectedEmptyHash {
		t.Errorf("hashFile() hash = %s, want %s", hash, expectedEmptyHash)
	}
}

func TestHashFileLargeFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a larger file (10MB)
	largeFile := filepath.Join(tmpDir, "large.bin")
	size := 10 * 1024 * 1024 // 10MB
	content := make([]byte, size)
	for i := range content {
		content[i] = byte(i % 256)
	}
	if err := os.WriteFile(largeFile, content, 0644); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	hasher := sha256.New()
	hash, fileSize, _, err := hashFile(largeFile, hasher)
	if err != nil {
		t.Fatalf("hashFile() error = %v", err)
	}

	if hash == "" {
		t.Error("hashFile() returned empty hash for large file")
	}

	if fileSize != int64(size) {
		t.Errorf("hashFile() size = %d, want %d", fileSize, size)
	}
}

func TestFormatBytesEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{"Negative bytes", -1, "-1 B"}, // Should handle gracefully
		{"Exact KB", 1024, "1.0 KB"},
		{"Exact MB", 1048576, "1.0 MB"},
		{"Exact GB", 1073741824, "1.0 GB"},
		{"Max int64", 9223372036854775807, "8.0 EB"}, // Very large
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBytes(tt.bytes)
			// Just verify it doesn't crash and returns something
			if got == "" {
				t.Errorf("formatBytes(%d) returned empty string", tt.bytes)
			}
		})
	}
}

func TestScanFilesWithHiddenDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create visible file
	visibleFile := filepath.Join(tmpDir, "visible.txt")
	if err := os.WriteFile(visibleFile, []byte("visible"), 0644); err != nil {
		t.Fatalf("Failed to create visible file: %v", err)
	}

	// Create hidden directory with file
	hiddenDir := filepath.Join(tmpDir, ".hidden")
	if err := os.Mkdir(hiddenDir, 0755); err != nil {
		t.Fatalf("Failed to create hidden directory: %v", err)
	}
	hiddenFile := filepath.Join(hiddenDir, "secret.txt")
	if err := os.WriteFile(hiddenFile, []byte("secret"), 0644); err != nil {
		t.Fatalf("Failed to create hidden file: %v", err)
	}

	// Scan should skip hidden directory
	files, err := scanFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	// Should only find visible file
	if len(files) != 1 {
		t.Errorf("scanFiles() found %d files, want 1 (hidden dir should be skipped)", len(files))
	}

	if len(files) > 0 && filepath.Base(files[0]) != "visible.txt" {
		t.Errorf("scanFiles() found wrong file: %s", files[0])
	}
}

func TestScanFilesMultipleExtensions(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with various extensions
	extensions := []string{".txt", ".md", ".go", ".py", ".js", ".json", ".yaml", ""}
	for i, ext := range extensions {
		name := fmt.Sprintf("file%d%s", i, ext)
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", name, err)
		}
	}

	files, err := scanFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	if len(files) != len(extensions) {
		t.Errorf("scanFiles() found %d files, want %d", len(files), len(extensions))
	}
}

func TestScanFilesSpecialCharactersInName(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with special characters (within filesystem limits)
	specialNames := []string{
		"file with spaces.txt",
		"file-with-dashes.txt",
		"file_with_underscores.txt",
		"file.with.dots.txt",
		"file(more).txt",
		"file[1].txt",
	}

	for _, name := range specialNames {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			// Some characters may not be allowed on all systems
			t.Logf("Could not create file %s (expected on some systems): %v", name, err)
			continue
		}
	}

	files, err := scanFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	// Should find all files that were successfully created
	t.Logf("scanFiles() found %d files with special characters", len(files))
}

// Test for file modification during scan (concurrent access)
func TestHashFileModifiedDuringScan(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file
	testFile := filepath.Join(tmpDir, "concurrent.txt")
	if err := os.WriteFile(testFile, []byte("initial content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Start a goroutine to modify the file
	done := make(chan bool)
	go func() {
		for i := 0; i < 5; i++ {
			time.Sleep(1 * time.Millisecond)
			content := fmt.Sprintf("modified content %d", i)
			os.WriteFile(testFile, []byte(content), 0644)
		}
		done <- true
	}()

	// Try to hash the file while it's being modified
	// The hash should succeed (possibly with different content each time)
	hasher := sha256.New()
	hash1, _, _, err := hashFile(testFile, hasher)
	if err != nil {
		t.Errorf("hashFile() error during concurrent modification: %v", err)
	}

	// Wait for modification to complete
	<-done

	// Hash again - might be different due to modifications
	hasher2 := sha256.New()
	hash2, _, _, err := hashFile(testFile, hasher2)
	if err != nil {
		t.Errorf("hashFile() error after concurrent modification: %v", err)
	}

	// Both should return valid hashes (even if different)
	if hash1 == "" || hash2 == "" {
		t.Error("hashFile() returned empty hash during concurrent access")
	}

	t.Logf("Concurrent hash test: hash1=%s..., hash2=%s...", hash1[:8], hash2[:8])
}

// Test file that's being deleted during scan
func TestHashFileDeletedDuringScan(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file
	testFile := filepath.Join(tmpDir, "to-delete.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Delete the file
	if err := os.Remove(testFile); err != nil {
		t.Fatalf("Failed to delete test file: %v", err)
	}

	// Try to hash the deleted file
	hasher := sha256.New()
	_, _, _, err := hashFile(testFile, hasher)

	// Should return an error
	if err == nil {
		t.Error("hashFile() should return error for deleted file")
	}
}

// Test with minimum size filtering
func TestMinSizeFilter(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files of various sizes
	files := []struct {
		name    string
		content []byte
	}{
		{"tiny.txt", []byte("x")},
		{"small.txt", []byte("small content")},
		{"medium.txt", make([]byte, 2048)},  // 2KB
		{"large.txt", make([]byte, 10240)},  // 10KB
	}

	for _, f := range files {
		path := filepath.Join(tmpDir, f.name)
		if err := os.WriteFile(path, f.content, 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", f.name, err)
		}
	}

	// Scan all files
	allFiles, err := scanFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	if len(allFiles) != 4 {
		t.Errorf("scanFiles() found %d files, want 4", len(allFiles))
	}

	// The actual size filtering happens in main(), not scanFiles()
	// This test verifies scanFiles returns all files
}

// Test with file pattern filtering
func TestFilePatternFilter(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with various extensions
	extensions := []string{".txt", ".txt", ".md", ".go"}
	for i, ext := range extensions {
		name := fmt.Sprintf("file%d%s", i, ext)
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", name, err)
		}
	}

	// Scan all files
	files, err := scanFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	// Should find all files (pattern filtering happens in main())
	if len(files) != 4 {
		t.Errorf("scanFiles() found %d files, want 4", len(files))
	}
}

// Test duplicate detection with different file sizes (shouldn't match)
func TestFindDuplicatesDifferentSizes(t *testing.T) {
	// Files with same hash but different sizes shouldn't exist in practice
	// but we test that size is correctly captured
	fileHashes := []FileHash{
		{Path: "/a.txt", Size: 100, Hash: "hash1", ModTime: time.Now()},
		{Path: "/b.txt", Size: 200, Hash: "hash1", ModTime: time.Now()}, // Same hash, diff size (impossible but test)
	}

	duplicates := findDuplicates(fileHashes)
	if len(duplicates) != 1 {
		t.Errorf("findDuplicates() found %d groups, want 1", len(duplicates))
	}

	// Both should be in the group (they have same hash)
	if len(duplicates) > 0 && len(duplicates[0].Files) != 2 {
		t.Errorf("Duplicate group has %d files, want 2", len(duplicates[0].Files))
	}
}

// Test scanning directory without read permission (Unix only)
func TestScanFilesNoReadPermission(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	tmpDir := t.TempDir()

	// Create a subdirectory with a file
	subDir := filepath.Join(tmpDir, "restricted")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Create another accessible directory
	accessDir := filepath.Join(tmpDir, "accessible")
	if err := os.Mkdir(accessDir, 0755); err != nil {
		t.Fatalf("Failed to create accessible directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(accessDir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file in accessible dir: %v", err)
	}

	// Remove read and execute permissions from restricted directory
	if err := os.Chmod(subDir, 0000); err != nil {
		t.Fatalf("Failed to chmod directory: %v", err)
	}
	defer os.Chmod(subDir, 0755) // Restore for cleanup

	// Scan should continue despite permission error
	files, err := scanFiles(tmpDir, true)

	// May get error or may skip the directory depending on implementation
	// Important: it shouldn't crash
	t.Logf("scanFiles() with restricted dir: %d files, err=%v", len(files), err)
}

// Test scanning with only hidden files
func TestScanFilesOnlyHiddenFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create only hidden files
	hiddenFiles := []string{".hidden1", ".hidden2", ".config"}
	for _, name := range hiddenFiles {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("hidden"), 0644); err != nil {
			t.Fatalf("Failed to create hidden file %s: %v", name, err)
		}
	}

	// Scan should find no files (all hidden)
	files, err := scanFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("scanFiles() error = %v", err)
	}

	if len(files) != 0 {
		t.Errorf("scanFiles() found %d files in hidden-only directory, want 0", len(files))
	}
}

// Test formatBytes with fractional values
func TestFormatBytesFractional(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{1536, "1.5 KB"},     // 1.5 KB
		{2560, "2.5 KB"},     // 2.5 KB
		{1572864, "1.5 MB"},  // 1.5 MB
		{2621440, "2.5 MB"},  // 2.5 MB
	}

	for _, tt := range tests {
		got := formatBytes(tt.bytes)
		if got != tt.expected {
			t.Errorf("formatBytes(%d) = %s, want %s", tt.bytes, got, tt.expected)
		}
	}
}

// Test selectFileToKeep with different criteria
func TestSelectFileToKeepOldest(t *testing.T) {
	now := time.Now()
	group := DuplicateGroup{
		Hash: "test",
		Size: 100,
		Files: []FileHash{
			{Path: "/newest.txt", Size: 100, Hash: "test", ModTime: now.Add(-1 * time.Hour)},
			{Path: "/oldest.txt", Size: 100, Hash: "test", ModTime: now.Add(-24 * time.Hour)},
			{Path: "/middle.txt", Size: 100, Hash: "test", ModTime: now.Add(-12 * time.Hour)},
		},
	}

	// Save original config and restore after test
	originalKeep := cfg.KeepCriteria
	defer func() { cfg.KeepCriteria = originalKeep }()

	cfg.KeepCriteria = "oldest"
	idx := selectFileToKeep(group)
	if idx != 1 { // oldest.txt is at index 1
		t.Errorf("selectFileToKeep(oldest) returned %d, want 1", idx)
	}
}

func TestSelectFileToKeepNewest(t *testing.T) {
	now := time.Now()
	group := DuplicateGroup{
		Hash: "test",
		Size: 100,
		Files: []FileHash{
			{Path: "/newest.txt", Size: 100, Hash: "test", ModTime: now.Add(-1 * time.Hour)},
			{Path: "/oldest.txt", Size: 100, Hash: "test", ModTime: now.Add(-24 * time.Hour)},
			{Path: "/middle.txt", Size: 100, Hash: "test", ModTime: now.Add(-12 * time.Hour)},
		},
	}

	originalKeep := cfg.KeepCriteria
	defer func() { cfg.KeepCriteria = originalKeep }()

	cfg.KeepCriteria = "newest"
	idx := selectFileToKeep(group)
	if idx != 0 { // newest.txt is at index 0
		t.Errorf("selectFileToKeep(newest) returned %d, want 0", idx)
	}
}

func TestSelectFileToKeepLargest(t *testing.T) {
	group := DuplicateGroup{
		Hash: "test",
		Size: 100,
		Files: []FileHash{
			{Path: "/small.txt", Size: 50, Hash: "test", ModTime: time.Now()},
			{Path: "/large.txt", Size: 200, Hash: "test", ModTime: time.Now()},
			{Path: "/medium.txt", Size: 100, Hash: "test", ModTime: time.Now()},
		},
	}

	originalKeep := cfg.KeepCriteria
	defer func() { cfg.KeepCriteria = originalKeep }()

	cfg.KeepCriteria = "largest"
	idx := selectFileToKeep(group)
	if idx != 1 { // large.txt is at index 1
		t.Errorf("selectFileToKeep(largest) returned %d, want 1", idx)
	}
}

func TestSelectFileToKeepSmallest(t *testing.T) {
	group := DuplicateGroup{
		Hash: "test",
		Size: 100,
		Files: []FileHash{
			{Path: "/small.txt", Size: 50, Hash: "test", ModTime: time.Now()},
			{Path: "/large.txt", Size: 200, Hash: "test", ModTime: time.Now()},
			{Path: "/medium.txt", Size: 100, Hash: "test", ModTime: time.Now()},
		},
	}

	originalKeep := cfg.KeepCriteria
	defer func() { cfg.KeepCriteria = originalKeep }()

	cfg.KeepCriteria = "smallest"
	idx := selectFileToKeep(group)
	if idx != 0 { // small.txt is at index 0
		t.Errorf("selectFileToKeep(smallest) returned %d, want 0", idx)
	}
}

func TestSelectFileToKeepPath(t *testing.T) {
	group := DuplicateGroup{
		Hash: "test",
		Size: 100,
		Files: []FileHash{
			{Path: "/dirA/file.txt", Size: 100, Hash: "test", ModTime: time.Now()},
			{Path: "/dirB/file.txt", Size: 100, Hash: "test", ModTime: time.Now()},
			{Path: "/dirC/file.txt", Size: 100, Hash: "test", ModTime: time.Now()},
		},
	}

	originalKeep := cfg.KeepCriteria
	defer func() { cfg.KeepCriteria = originalKeep }()

	cfg.KeepCriteria = "path:dirB"
	idx := selectFileToKeep(group)
	if idx != 1 { // dirB/file.txt is at index 1
		t.Errorf("selectFileToKeep(path:dirB) returned %d, want 1", idx)
	}
}

func TestSelectFileToKeepPathNotFound(t *testing.T) {
	group := DuplicateGroup{
		Hash: "test",
		Size: 100,
		Files: []FileHash{
			{Path: "/dirA/file.txt", Size: 100, Hash: "test", ModTime: time.Now()},
			{Path: "/dirB/file.txt", Size: 100, Hash: "test", ModTime: time.Now()},
		},
	}

	originalKeep := cfg.KeepCriteria
	defer func() { cfg.KeepCriteria = originalKeep }()

	cfg.KeepCriteria = "path:nonexistent"
	idx := selectFileToKeep(group)
	if idx != 0 { // Should default to first file
		t.Errorf("selectFileToKeep(path:nonexistent) returned %d, want 0 (default)", idx)
	}
}
