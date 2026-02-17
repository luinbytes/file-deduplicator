# Cloud Storage Integration Design

## Overview

Add support for scanning and deduplicating files on cloud storage services:
- Google Drive
- Dropbox
- OneDrive

## Architecture

### Storage Abstraction Layer

Create a `storage` package with a common interface:

```go
package storage

// FileInfo represents a file from any storage provider
type FileInfo struct {
    ID       string    // Provider-specific ID
    Name     string    // File name
    Path     string    // Full path
    Size     int64     // File size in bytes
    ModTime  time.Time // Last modified time
    IsDir    bool      // Is directory
}

// Reader provides read access to a file
type Reader interface {
    Read(p []byte) (n int, err error)
    Close() error
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
}
```

### Configuration

Add new CLI flags:

```bash
# Cloud storage source
file-deduplicator -cloud drive://path/to/folder

# Authentication
file-deduplicator -cloud-auth google credentials.json

# Download for processing (if cloud doesn't support streaming)
file-deduplicator -cloud-drive -cloud-download-dir /tmp/cloud-files
```

Config file support (`.deduprc.json`):

```json
{
  "cloud_providers": {
    "google_drive": {
      "enabled": true,
      "credentials_file": "~/.config/file-deduplicator/google-credentials.json",
      "token_file": "~/.config/file-deduplicator/google-token.json"
    },
    "dropbox": {
      "enabled": false,
      "access_token": ""
    },
    "onedrive": {
      "enabled": false,
      "tenant_id": "",
      "client_id": "",
      "client_secret": ""
    }
  }
}
```

## Implementation Plan

### Phase 1: Core Infrastructure (Current Session)
1. Create `storage` package with interface
2. Implement local filesystem provider (for testing)
3. Add CLI flags for cloud support
4. Update file scanning to use abstraction layer

### Phase 2: Google Drive (Next Session)
1. Add Google Drive provider implementation
2. OAuth2 authentication flow
3. File listing and streaming
4. Test with real Google Drive

### Phase 3: Dropbox & OneDrive (Future)
1. Dropbox provider
2. OneDrive provider
3. Unified authentication management

## Challenges

### Authentication
- Each provider uses different auth mechanisms
- Need to store tokens securely
- Token refresh handling

### Performance
- Cloud APIs have rate limits
- Network latency
- Streaming vs downloading files

### Cost
- API calls may incur costs
- Bandwidth charges
- Need to minimize API calls

## Implementation Strategy

### Hybrid Approach (Recommended)
Instead of pure cloud-to-cloud, use a hybrid approach:

1. **List** files from cloud (metadata only)
2. **Download** files to local temp directory for hashing
3. **Process** using existing deduplication logic
4. **Upload/Move** changes back to cloud

This leverages the existing, battle-tested deduplication code while adding cloud support.

### Pure Cloud Approach (Future)
For advanced users, implement pure cloud-to-cloud:
- Stream files directly from cloud for hashing
- Batch API operations for deletes/moves
- Parallel cloud API calls

## Dependencies

```go
// Google Drive
google.golang.org/api/drive/v3
google.golang.org/api/option
golang.org/x/oauth2/google

// Dropbox
github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files

// OneDrive
github.com/microsoftgraph/msgraph-sdk-go
github.com/Azure/azure-sdk-for-go/sdk/azidentity
```

## Next Steps

1. ✅ Research APIs (done)
2. ⏳ Create storage interface (in progress)
3. ⏳ Implement local provider (for testing)
4. ⏳ Add CLI flags
5. ⏳ Implement Google Drive provider
6. ⏳ Test with real account
7. ⏳ Implement Dropbox & OneDrive
8. ⏳ Documentation and examples

## Questions for Lu

1. Should we support pure cloud-to-cloud, or just sync-to-local approach?
2. Priority order for providers? (Drive, Dropbox, OneDrive)
3. Should this be a premium feature or free?
