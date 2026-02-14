package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalProvider implements Provider for local filesystem
type LocalProvider struct {
	basePath string
}

// NewLocalProvider creates a new local filesystem provider
func NewLocalProvider(basePath string) (*LocalProvider, error) {
	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	return &LocalProvider{
		basePath: absPath,
	}, nil
}

// ListFiles lists all files in a directory (optionally recursive)
func (p *LocalProvider) ListFiles(ctx context.Context, path string, recursive bool) ([]FileInfo, error) {
	fullPath := filepath.Join(p.basePath, path)

	var files []FileInfo

	err := filepath.Walk(fullPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Get relative path
		relPath, err := filepath.Rel(p.basePath, filePath)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Skip root directory itself
		if relPath == "." {
			return nil
		}

		files = append(files, FileInfo{
			ID:      filePath, // Use full path as ID
			Name:    info.Name(),
			Path:    relPath,
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		})

		// If not recursive, skip subdirectories
		if !recursive && info.IsDir() && filePath != fullPath {
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return files, nil
}

// OpenFile opens a file for reading
func (p *LocalProvider) OpenFile(ctx context.Context, id string) (Reader, error) {
	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	file, err := os.Open(id)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// DeleteFile deletes a file
func (p *LocalProvider) DeleteFile(ctx context.Context, id string) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err := os.Remove(id); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// MoveFile moves a file to a new location
func (p *LocalProvider) MoveFile(ctx context.Context, id string, newPath string) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Ensure target directory exists
	targetDir := filepath.Dir(newPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Move file
	if err := os.Rename(id, newPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	return nil
}

// Name returns the provider name
func (p *LocalProvider) Name() string {
	return "local"
}

// Close cleans up provider resources (no-op for local)
func (p *LocalProvider) Close() error {
	return nil
}

// Helper to check if path is hidden
func isHidden(path string) bool {
	parts := strings.Split(path, string(os.PathSeparator))
	for _, part := range parts {
		if strings.HasPrefix(part, ".") && part != "." && part != ".." {
			return true
		}
	}
	return false
}

// Ensure LocalProvider implements Provider interface
var _ Provider = (*LocalProvider)(nil)
