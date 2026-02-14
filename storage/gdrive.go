package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// GoogleDriveProvider implements Provider for Google Drive
type GoogleDriveProvider struct {
	service     *drive.Service
	tokenFile   string
}

// NewGoogleDriveProvider creates a new Google Drive provider
func NewGoogleDriveProvider(ctx context.Context, credentialsFile, tokenFile string) (*GoogleDriveProvider, error) {
	// Expand home directory if needed
	tokenFile = expandHome(tokenFile)
	credentialsFile = expandHome(credentialsFile)

	// Read credentials file
	credBytes, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	// Parse OAuth2 config
	config, err := google.ConfigFromJSON(credBytes, drive.DriveScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	// Load or create token
	token, err := loadToken(tokenFile)
	if err != nil {
		// Token doesn't exist, need to authenticate
		token, err = getTokenFromWeb(ctx, config)
		if err != nil {
			return nil, fmt.Errorf("failed to get token: %w", err)
		}

		// Save token for future use
		if err := saveToken(tokenFile, token); err != nil {
			return nil, fmt.Errorf("failed to save token: %w", err)
		}
	}

	// Create Drive service
	service, err := drive.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx, token)))
	if err != nil {
		return nil, fmt.Errorf("failed to create Drive service: %w", err)
	}

	return &GoogleDriveProvider{
		service:   service,
		tokenFile: tokenFile,
	}, nil
}

// ListFiles lists all files in a directory (optionally recursive)
func (p *GoogleDriveProvider) ListFiles(ctx context.Context, path string, recursive bool) ([]FileInfo, error) {
	var files []FileInfo

	// For Google Drive, we need to:
	// 1. Find the folder ID for the given path
	// 2. List all files in that folder
	// 3. If recursive, traverse subfolders

	folderID, err := p.getFolderID(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to find folder: %w", err)
	}

	// List files in folder
	if err := p.listFilesRecursive(ctx, folderID, path, recursive, &files); err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return files, nil
}

// listFilesRecursive recursively lists files
func (p *GoogleDriveProvider) listFilesRecursive(ctx context.Context, folderID, currentPath string, recursive bool, files *[]FileInfo) error {
	query := fmt.Sprintf("'%s' in parents and trashed = false", folderID)

	pageToken := ""
	for {
		call := p.service.Files.List().
			Q(query).
			Fields("nextPageToken, files(id, name, size, modifiedTime, mimeType, parents").
			PageToken(pageToken)

		result, err := call.Do()
		if err != nil {
			return fmt.Errorf("failed to list files: %w", err)
		}

		for _, file := range result.Files {
			filePath := filepath.Join(currentPath, file.Name)

			modTime, _ := file.ModifiedTime.Parse()

			info := FileInfo{
				ID:       file.Id,
				Name:     file.Name,
				Path:     filePath,
				Size:     file.Size,
				ModTime:  modTime,
				IsDir:    file.MimeType == "application/vnd.google-apps.folder",
				MimeType: file.MimeType,
			}

			*files = append(*files, info)

			// Recurse into subfolders if recursive
			if recursive && info.IsDir {
				if err := p.listFilesRecursive(ctx, file.Id, filePath, recursive, files); err != nil {
					return err
				}
			}
		}

		if result.NextPageToken == "" {
			break
		}
		pageToken = result.NextPageToken
	}

	return nil
}

// getFolderID gets the folder ID for a given path
func (p *GoogleDriveProvider) getFolderID(ctx context.Context, path string) (string, error) {
	if path == "" || path == "/" {
		return "root", nil
	}

	// Split path into components
	parts := filepath.SplitList(path)
	if len(parts) == 0 {
		return "root", nil
	}

	// Start from root and traverse
	parentID := "root"
	for _, part := range parts {
		query := fmt.Sprintf("name = '%s' and '%s' in parents and mimeType = 'application/vnd.google-apps.folder' and trashed = false",
			part, parentID)

		result, err := p.service.Files.List().Q(query).Fields("files(id)").Do()
		if err != nil {
			return "", fmt.Errorf("failed to find folder %s: %w", part, err)
		}

		if len(result.Files) == 0 {
			return "", fmt.Errorf("folder not found: %s", part)
		}

		parentID = result.Files[0].Id
	}

	return parentID, nil
}

// OpenFile opens a file for reading
func (p *GoogleDriveProvider) OpenFile(ctx context.Context, id string) (Reader, error) {
	resp, err := p.service.Files.Get(id).Download()
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return resp.Body, nil
}

// DeleteFile deletes a file
func (p *GoogleDriveProvider) DeleteFile(ctx context.Context, id string) error {
	if err := p.service.Files.Delete(id).Do(); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// MoveFile moves a file to a new location
func (p *GoogleDriveProvider) MoveFile(ctx context.Context, id string, newPath string) error {
	// Get new parent folder ID
	parentID, err := p.getFolderID(ctx, filepath.Dir(newPath))
	if err != nil {
		return fmt.Errorf("failed to find target folder: %w", err)
	}

	// Update file with new parent and name
	_, err = p.service.Files.Update(id, &drive.File{
		Name: filepath.Base(newPath),
	}).AddParents(parentID).Do()

	if err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	return nil
}

// Name returns the provider name
func (p *GoogleDriveProvider) Name() string {
	return "google-drive"
}

// Close cleans up provider resources
func (p *GoogleDriveProvider) Close() error {
	// Nothing to close for Drive service
	return nil
}

// Helper functions

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func loadToken(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(file string, token *oauth2.Token) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}

func getTokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	// Generate auth URL
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Printf("\n🔐 Go to the following link in your browser:\n%s\n\n", authURL)
	fmt.Print("Enter authorization code: ")

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, fmt.Errorf("failed to read authorization code: %w", err)
	}

	// Exchange code for token
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	return token, nil
}

// Ensure GoogleDriveProvider implements Provider interface
var _ Provider = (*GoogleDriveProvider)(nil)
