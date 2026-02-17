# Configuration Profiles

Pre-built configuration profiles for common use cases. Use these to quickly optimize file-deduplicator for your specific workflow.

## Usage

```bash
# Use a profile
file-deduplicator -dir ~/Pictures -config profiles/photographer.json -dry-run

# Override profile settings with command-line flags
file-deduplicator -dir ~/Pictures -config profiles/photographer.json -similarity 5
```

## Available Profiles

| Profile | Description | Best For |
|---------|-------------|----------|
| **photographer** | Perceptual hashing for similar images | Photo libraries, RAW files |
| **developer** | Source code and config files | Code repositories, projects |
| **sysadmin** | Backups and system files | Backup directories, logs |
| **designer** | Design assets and exports | Graphic design files |
| **video_editor** | Large video files | Video projects, archives |
| **minimal** | Quick, lightweight scans | Small directories, fast results |
| **thorough** | Maximum accuracy | Important archives |

## Profile Details

### 📷 Photographer
- **Hash:** SHA256 + pHash (perceptual)
- **Keep:** Largest files
- **Min size:** 100KB
- **Pattern:** Image formats (jpg, png, gif, webp, raw, etc.)
- **Use when:** Managing photo libraries with potential similar/duplicate shots

### 💻 Developer
- **Hash:** SHA256
- **Keep:** Oldest files
- **Min size:** 512 bytes
- **Pattern:** Source code and config files
- **Use when:** Cleaning up duplicate code files across projects

### 🖥️ SysAdmin
- **Hash:** SHA256
- **Keep:** Newest files
- **Min size:** 1KB
- **Use when:** Managing backup directories and log archives

### 🎨 Designer
- **Hash:** SHA256 + dHash (perceptual)
- **Keep:** Largest files
- **Min size:** 5KB
- **Pattern:** Design formats (png, psd, ai, sketch, fig, etc.)
- **Use when:** Managing design assets and export folders

### 🎬 Video Editor
- **Hash:** SHA256
- **Keep:** Newest files
- **Min size:** 10MB
- **Pattern:** Video formats (mp4, mov, mkv, etc.)
- **Use when:** Managing large video libraries and project archives

### ⚡ Minimal
- **Hash:** MD5 (fastest)
- **Keep:** Oldest files
- **Min size:** 1KB
- **Workers:** 2
- **Use when:** Quick scans of small directories

### 🔍 Thorough
- **Hash:** SHA256 + pHash (perceptual)
- **Keep:** Oldest files
- **Min size:** 1 byte (everything)
- **Similarity:** 6 (strict)
- **Use when:** Maximum accuracy is required for important archives

## Creating Custom Profiles

Copy an existing profile and modify it:

```bash
cp profiles/photographer.json profiles/my-custom.json
```

Edit the JSON file with your preferred settings:

```json
{
  "name": "My Custom Profile",
  "description": "Description of your profile",
  "config": {
    "recursive": true,
    "hash": "sha256",
    "perceptual": true,
    "phash_algo": "dhash",
    "similarity": 10,
    "keep": "oldest",
    "min_size": 1024,
    "pattern": "*.jpg",
    "workers": 4
  },
  "usage": {
    "basic": "file-deduplicator -dir ~/MyFolder -config profiles/my-custom.json"
  },
  "notes": [
    "Your notes here"
  ]
}
```

## Global Profile Location

You can also save profiles to a global location:

- `~/.config/file-deduplicator/config.json` - Default config (loaded automatically)
- `~/.config/file-deduplicator/profiles/` - Custom profiles

---

*Profiles are JSON files that can include any command-line flag. Command-line flags always override profile settings.*
