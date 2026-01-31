# File Deduplicator v2.0.0 🔍

A fast, parallel CLI tool to find and remove duplicate files using SHA256 hashing, now with smart selection, interactive mode, and more!

## Features

- 🚀 **Fast Parallel Processing** - Uses multiple goroutines for hashing
- 🔐 **Multiple Hash Algorithms** - Support for SHA256, SHA1, and MD5
- 📊 **Smart File Selection** - Keep oldest, newest, largest, smallest, or specific file
- ❓ **Interactive Mode** - Ask before deleting each duplicate
- 📦 **Move Instead of Delete** - Move duplicates to a safe folder
- 📁 **Pattern Filtering** - Only process files matching a pattern (e.g., `*.jpg`)
- 📄 **Export Reports** - Generate JSON reports of duplicate findings
- 🔄 **Undo Log** - Track deleted files (informational)
- 🎯 **Size Filtering** - Ignore files below a minimum size
- 🚫 **Hidden File Skipping** - Automatically skip `.hidden` files and directories
- 📈 **Progress Tracking** - See progress during large scans
- 🌳 **Recursive Scanning** - Scan directories recursively

## Installation

### Build from Source

```bash
git clone https://github.com/luinbytes/file-deduplicator.git
cd file-deduplicator
go build -o file-deduplicator main.go
```

## Usage

### Basic Usage

Find and delete duplicates:

```bash
file-deduplicator -dir /path/to/scan
```

### Dry Run

Preview what would be deleted without making changes:

```bash
file-deduplicator -dir /path/to/scan -dry-run
```

### Interactive Mode

Ask before deleting each duplicate:

```bash
file-deduplicator -dir /path/to/scan -interactive
```

### Move Duplicates

Move duplicates to a folder instead of deleting:

```bash
file-deduplicator -dir /path/to/scan -move-to /path/to/duplicates
```

### Smart File Selection

Keep the newest duplicate instead of oldest:

```bash
file-deduplicator -dir /path/to/scan -keep newest
```

Keep the largest duplicate:

```bash
file-deduplicator -dir /path/to/scan -keep largest
```

Keep a specific file (matching path):

```bash
file-deduplicator -dir /path/to/scan -keep path:/path/to/keep
```

### File Type Filtering

Only find duplicates in specific file types:

```bash
# Only JPEG images
file-deduplicator -dir /path/to/scan -pattern "*.jpg"

# Only PDF files
file-deduplicator -dir /path/to/scan -pattern "*.pdf"

# Only video files
file-deduplicator -dir /path/to/scan -pattern "*.mp4"
```

### Export Report

Generate a JSON report of duplicates:

```bash
file-deduplicator -dir /path/to/scan -export
```

Creates `.deduplicator_report.json` with detailed information.

### Hash Algorithm Selection

Use different hash algorithm (default: SHA256):

```bash
file-deduplicator -dir /path/to/scan -hash sha1
file-deduplicator -dir /path/to/scan -hash md5
```

### View Undo Log

View the log of last operation:

```bash
file-deduplicator -undo
```

Note: Undo is informational only - deleted files cannot be recovered unless you moved them.

## Options

| Option | Default | Description |
|--------|----------|-------------|
| `-dir string` | `.` | Directory to scan for duplicates |
| `-recursive` | `true` | Scan directories recursively |
| `-dry-run` | `false` | Show what would be deleted without actually deleting |
| `-verbose` | `false` | Show detailed output |
| `-workers int` | NumCPU | Number of worker goroutines |
| `-min-size int` | `1024` | Minimum file size in bytes (default: 1KB) |
| `-interactive` | `false` | Ask before deleting each duplicate |
| `-move-to string` | `""` | Move duplicates to this folder instead of deleting |
| `-keep string` | `oldest` | File to keep: oldest, newest, largest, smallest, first, or path:<path> |
| `-hash string` | `sha256` | Hash algorithm: sha256, sha1, or md5 |
| `-pattern string` | `""` | File pattern to match (e.g., `*.jpg`) |
| `-export` | `false` | Export duplicate report to JSON file |
| `-undo` | `false` | Undo last operation (informational) |

## Keep Criteria Options

| Criteria | Description |
|----------|-------------|
| `oldest` | Keep the file with oldest modification time (default) |
| `newest` | Keep the file with newest modification time |
| `largest` | Keep the file with largest size |
| `smallest` | Keep the file with smallest size |
| `first` | Keep the first file found |
| `path:<path>` | Keep file matching the specified path |

## Examples

### Find Duplicates in Home Directory

```bash
file-deduplicator -dir ~
```

Output:
```
🔍 File Deduplicator v2.0.0 - Starting...
📊 Found 1523 files
🔐 Computed 1523 hashes
👯 Found 15 duplicate groups

👯 Duplicate Files:
======================================================================

[1] Hash: a1b2c3d4e5f6...
    Size: 1.5 MB
    Files: 3 (keeping 1, removing 2)
    ✓ KEEP /home/user/docs/report.pdf (modified: 2026-01-30 14:30:00)
    ✗ DELETE /home/user/Downloads/report_copy.pdf (modified: 2026-01-31 09:15:00)
    ✗ DELETE /home/user/backup/report.pdf (modified: 2026-01-29 18:45:00)

[2] Hash: f6e5d4c3b2a1...
    Size: 456.2 KB
    Files: 2 (keeping 1, removing 1)
    ✓ KEEP /home/user/images/photo.jpg (modified: 2026-01-31 10:20:00)
    ✗ DELETE /home/user/Downloads/photo.jpg (modified: 2026-01-30 15:10:00)

======================================================================
📊 Summary: 3 duplicate files, 3.0 MB of space can be freed

🗑️  Deleting duplicates...
✓ Deleted /home/user/Downloads/report_copy.pdf
✓ Deleted /home/user/backup/report.pdf
✓ Deleted /home/user/Downloads/photo.jpg

✅ Deleted 3 files, freed 3.0 MB of space
✅ Complete in 2.3s
```

### Preview Before Deleting

```bash
file-deduplicator -dir ~/Downloads -dry-run
```

### Move Duplicates Safely

```bash
file-deduplicator -dir ~/Pictures -move-to ~/Duplicates
```

This keeps all your files safe in a `~/Duplicates` folder.

### Find Photo Duplicates Only

```bash
file-deduplicator -dir ~/Pictures -pattern "*.jpg"
file-deduplicator -dir ~/Pictures -pattern "*.png"
file-deduplicator -dir ~/Pictures -pattern "*.raw"
```

### Keep Newest Files

```bash
file-deduplicator -dir ~/Documents -keep newest
```

Useful for backup scenarios where you want the most recent version.

### Keep Largest Files

```bash
file-deduplicator -dir ~/Videos -keep largest
```

Useful for media files where quality (size) matters.

### Interactive Mode

```bash
file-deduplicator -dir ~/Downloads -interactive
```

Output:
```
Delete /home/user/Downloads/copy.pdf (1.5 MB)? [y/n/q]: y
✓ Deleted /home/user/Downloads/copy.pdf

Delete /home/user/Downloads/duplicate.jpg (456 KB)? [y/n/q]: n
Skipping /home/user/Downloads/duplicate.jpg

Delete /home/user/Downloads/backup.zip (2.1 MB)? [y/n/q]: q
❓ Quitting...
```

### Export and Review Report

```bash
# Scan and export report
file-deduplicator -dir ~/Documents -export

# View report
cat .deduplicator_report.json

# Or use jq for pretty printing
jq .deduplicator_report.json
```

### Combine Multiple Options

```bash
# Find only JPG duplicates, move to folder, keep largest
file-deduplicator -dir ~/Pictures -pattern "*.jpg" -move-to ~/Duplicates -keep largest -v

# Interactive mode with detailed output
file-deduplicator -dir ~/Downloads -interactive -v

# Quick preview of large files only
file-deduplicator -dir ~/Videos -min-size 10485760 -dry-run
```

## Hash Algorithms

### SHA256 (Default)
- **Security**: High
- **Speed**: Medium
- **Collision Probability**: Extremely low
- **Best for**: General use, security-sensitive files

### SHA1
- **Security**: Medium (deprecated for security)
- **Speed**: Fast
- **Collision Probability**: Low
- **Best for**: Legacy systems, speed-critical operations

### MD5
- **Security**: Low (not recommended for security)
- **Speed**: Very fast
- **Collision Probability**: Medium
- **Best for**: Non-critical deduplication, speed-critical

## Best Practices

1. **Dry Run First** - Always use `-dry-run` to preview changes
2. **Interactive Mode** - Use `-interactive` for important directories
3. **Move Instead of Delete** - Use `-move-to` to keep files safe
4. **Export Reports** - Use `-export` to document what was found
5. **Backup First** - Backup important data before running
6. **Pattern Filter** - Use `-pattern` to focus on specific file types
7. **Size Filter** - Use `-min-size` to ignore small files
8. **Undo Log** - Review `.deduplicator_undo.json` for reference

## Troubleshooting

### "Access Denied" Errors

Run with appropriate permissions:
```bash
# Linux/macOS
sudo file-deduplicator -dir /protected/path

# Windows
# Run Command Prompt as Administrator
file-deduplicator -dir C:\\Protected\\Path
```

### Too Many False Positives

Increase minimum file size:
```bash
file-deduplicator -dir ~/Downloads -min-size 1048576
```

### Hash Algorithm Performance

SHA256 is more secure but slower. For speed on non-critical files:
```bash
file-deduplicator -dir ~/Videos -hash md5
```

### Files Not Being Detected

Check if files match your pattern:
```bash
# List all files first
find ~/Pictures -name "*.jpg" | wc -l

# Then run deduplicator
file-deduplicator -dir ~/Pictures -pattern "*.jpg" -v
```

## Advanced Usage

### Keep Specific File in Duplicate Group

```bash
# Keep file from specific directory
file-deduplicator -dir ~/Documents -keep path:/home/user/important

# Keep file from backup
file-deduplicator -dir ~/Documents -keep path:backup
```

### Multiple Scans

```bash
# Scan different directories separately
file-deduplicator -dir ~/Pictures -pattern "*.jpg"
file-deduplicator -dir ~/Pictures -pattern "*.png"
file-deduplicator -dir ~/Videos -pattern "*.mp4"
```

### Chain Operations

```bash
# First find and report
file-deduplicator -dir ~/Downloads -export

# Review report
cat .deduplicator_report.json

# Then run with same options
file-deduplicator -dir ~/Downloads
```

## Files Created

| File | Purpose |
|-------|---------|
| `.deduplicator_report.json` | Duplicate report (with `-export`) |
| `.deduplicator_undo.json` | Operation log (for reference) |

## Performance

Typical performance (SHA256):
- ~1000 files/sec per CPU core
- 10,000 files in ~2-3s (8 cores)
- 100,000 files in ~20-30s (8 cores)

## Security Considerations

1. **Hash Collisions** - Extremely rare with SHA256
2. **File Deletion** - Use `-move-to` for safety
3. **Permissions** - Ensure write access to scanned directories
4. **Hidden Files** - Automatically skipped for safety

## License

MIT License

## Author

Created by Lu (luinbytes)
Enhanced by Lumi (Lu's AI Assistant)

---

**Clean up your files, reclaim your space! 🔍🗑️**
