package storage

import (
	"context"
	"io"
	"time"
)

// FileInfo represents a file from any storage provider
type FileInfo struct {
	ID       string    // Provider-specific ID (for local: path)
	Name     string    // File name
	Path     string    // Full path
	Size     int64     // File size in bytes
	ModTime  time.Time // Last modified time
	IsDir    bool      // Is directory
	MimeType string    // MIME type (if available)
}

// Reader provides read access to a file
type Reader interface {
	io.ReadCloser
}

// Provider defines the interface for storage providers
type Provider interface {
	// ListFiles lists all files in a directory (optionally recursive)
	ListFiles(ctx context.Context, path string, recursive bool) ([]FileInfo, error)

	// OpenFile opens a file for reading
	OpenFile(ctx context.Context, id string) (Reader, error)

	// DeleteFile deletes a file
	DeleteFile(ctx context.Context, id string) error

	// MoveFile moves a file to a new location
	MoveFile(ctx context.Context, id string, newPath string) error

	// Name returns the provider name
	Name() string

	// Close cleans up provider resources
	Close() error
}

// CloudConfig holds cloud provider configuration
type CloudConfig struct {
	GoogleDrive *GoogleDriveConfig `json:"google_drive,omitempty"`
	Dropbox     *DropboxConfig     `json:"dropbox,omitempty"`
	OneDrive    *OneDriveConfig    `json:"onedrive,omitempty"`
}

// GoogleDriveConfig holds Google Drive configuration
type GoogleDriveConfig struct {
	Enabled         bool   `json:"enabled"`
	CredentialsFile string `json:"credentials_file,omitempty"`
	TokenFile       string `json:"token_file,omitempty"`
}

// DropboxConfig holds Dropbox configuration
type DropboxConfig struct {
	Enabled     bool   `json:"enabled"`
	AccessToken string `json:"access_token,omitempty"`
}

// OneDriveConfig holds OneDrive configuration
type OneDriveConfig struct {
	Enabled      bool   `json:"enabled"`
	TenantID     string `json:"tenant_id,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
}

// ProviderType represents the type of storage provider
type ProviderType string

const (
	ProviderLocal      ProviderType = "local"
	ProviderGoogleDrive ProviderType = "google-drive"
	ProviderDropbox    ProviderType = "dropbox"
	ProviderOneDrive   ProviderType = "onedrive"
)
