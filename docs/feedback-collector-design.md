# Customer Feedback Collector

> Simple CLI tool to collect user feedback for post-launch iteration
> Date: 2026-02-09
> Ready for integration into file-deduplicator

---

## Feature: `--feedback` Flag

Adds interactive feedback collection to file-deduplicator CLI.

### What It Collects

**Interactive Prompt (via terminal):**
```
📋 Customer Feedback Collection

Help us improve file-deduplicator!

1. What's your favorite feature?
2. What's annoying or confusing?
3. What feature would make you upgrade to Team?
4. How would you rate file-deduplicator (1-5)?
5. Any suggestions or ideas?

Type your answers (press Enter to continue):

1. [Your answer]
2. [Your answer]
...
5. [Your answer]

💾 Saving feedback to: ~/.config/file-deduplicator/feedback.json
✅ Feedback saved! Thank you for your input.

Would you like to send this feedback directly to the creator?
(y/N): _
```

### Feedback Format

```json
{
  "timestamp": "2026-02-10T12:00:00Z",
  "version": "1.2.0",
  "tier": "free|pro|team",
  "feedback": {
    "favorite_feature": "Perceptual hashing",
    "pain_points": "Slow on large directories",
    "team_interest": "No, would use if API access",
    "nps_score": 4,
    "suggestions": "Add web dashboard"
  },
  "system": {
    "os": "macOS",
    "go_version": "1.21",
    "scan_stats": {
      "files_scanned": 12543,
      "duplicates_found": 342
    }
  }
}
```

---

## Implementation Steps

### 1. Add Flag to CLI

**File:** `file-deduplicator/cmd/root.go`

```go
var feedbackFlag = app.Flag("feedback", "Collect customer feedback", false)

// In root command setup
feedbackFlag.BoolVar(&collectFeedback, "feedback", false, "Collect customer feedback interactively")

// After normal command execution
if collectFeedback {
    collectFeedback()
    return
}
```

### 2. Feedback Collection Function

**File:** `file-deduplicator/feedback/feedback.go` (new file)

```go
package feedback

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/luinbytes/file-deduplicator/internal/config"
)

type Feedback struct {
	Timestamp        string `json:"timestamp"`
	Version          string `json:"version"`
	Tier             string `json:"tier"`
	FavoriteFeature  string `json:"favorite_feature,omitempty"`
	PainPoints       []string `json:"pain_points,omitempty"`
	TeamInterest     string `json:"team_interest,omitempty"`
	NPSScore         int     `json:"nps_score,omitempty"`
	Suggestions      []string `json:"suggestions,omitempty"`
	System          SystemInfo `json:"system,omitempty"`
}

type SystemInfo struct {
	OS          string   `json:"os"`
	GoVersion  string   `json:"go_version"`
	ScanStats   ScanStats `json:"scan_stats,omitempty"`
}

type ScanStats struct {
	FilesScanned   int `json:"files_scanned,omitempty"`
	DuplicatesFound int `json:"duplicates_found,omitempty"`
}

func CollectFeedback(version string, tier string, scanStats *ScanStats) error {
	// Display header
	fmt.Println(`
📋 Customer Feedback Collection

Help us improve file-deduplicator!
`)

	// Collect feedback
	var favoriteFeature, painPoints, teamInterest, suggestions string
	var npsScore int

	fmt.Print("1. What's your favorite feature? ")
	fmt.Scanln(&favoriteFeature)

	fmt.Print("2. What's annoying or confusing? ")
	fmt.Scanln(&painPoints)

	fmt.Print("3. What feature would make you upgrade to Team? ")
	fmt.Scanln(&teamInterest)

	fmt.Print("4. How would you rate file-deduplicator (1-5)? ")
	fmt.Scanln(&npsScore)

	fmt.Print("5. Any suggestions or ideas? ")
	fmt.Scanln(&suggestions)

	// Get system info
	osInfo, _ := getSystemInfo()

	// Create feedback struct
	fb := Feedback{
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   version,
		Tier:      tier,
		FavoriteFeature: favoriteFeature,
		PainPoints:       strings.Split(painPoints, ","),
		TeamInterest:     teamInterest,
		NPSScore:         npsScore,
		Suggestions:      strings.Split(suggestions, ","),
		System:          *osInfo,
	}

	// Save to file
	if err := saveFeedback(fb); err != nil {
		return fmt.Errorf("failed to save feedback: %w", err)
	}

	// Offer to send to creator
	fmt.Println(`
💾 Saving feedback to: ~/.config/file-deduplicator/feedback.json
✅ Feedback saved! Thank you for your input.

Would you like to send this feedback directly to the creator?
(y/N): `)

	var send string
	fmt.Scanln(&send)

	if strings.ToLower(send) == "y" || strings.ToLower(send) == "yes" {
		fmt.Println(`
📧 Copy this feedback and email to: luinbytes@gmail.com

Feedback:
``)
		// Print formatted feedback
		prettyPrintFeedback(fb)
		fmt.Println(`
Subject: Feedback for file-deduplicator

Paste above feedback here.
`)
	} else {
		fmt.Println("✅ Thanks for your feedback!")
	}

	return nil
}

func saveFeedback(fb Feedback) error {
	configDir := config.GetConfigDir()
	fbFile := fmt.Sprintf("%s/feedback.json", configDir)

	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load existing feedback or create new
	var feedbacks []Feedback
	data, err := os.ReadFile(fbFile)
	if err == nil {
		if err := json.Unmarshal(data, &feedbacks); err != nil {
			return fmt.Errorf("failed to parse existing feedback: %w", err)
		}
	}

	// Append new feedback
	feedbacks = append(feedbacks, fb)

	// Write back to file
	newData, err := json.MarshalIndent(feedbacks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal feedback: %w", err)
	}

	if err := os.WriteFile(fbFile, newData, 0644); err != nil {
		return fmt.Errorf("failed to write feedback file: %w", err)
	}

	return nil
}

func getSystemInfo() (*SystemInfo, error) {
	// Get OS
	osName := runtime.GOOS

	// Get Go version
	goVersion := runtime.Version()

	return &SystemInfo{
		OS:         osName,
		GoVersion:  goVersion,
	}, nil
}

func prettyPrintFeedback(fb Feedback) {
	data, _ := json.MarshalIndent(fb, "", "  ")
	fmt.Println(string(data))
}
```

### 3. Update `go.mod`

```go
// No new dependencies needed - using encoding/json, fmt, os, time, runtime
```

### 4. Add Test

**File:** `file-deduplicator/feedback/feedback_test.go`

```go
package feedback

import (
	"testing"
)

func TestCollectFeedback(t *testing.T) {
	// Mock user input - would need refactoring for testability
	// For now, just test saveFeedback function

	fb := Feedback{
		Timestamp:    "2026-02-10T12:00:00Z",
		Version:      "1.2.0",
		Tier:         "pro",
		FavoriteFeature: "Perceptual hashing",
		NPSScore:     4,
	}

	err := saveFeedback(fb)
	if err != nil {
		t.Errorf("saveFeedback failed: %v", err)
	}

	// Verify file was created
	// (Would need actual file checking logic in implementation)
}

func TestFeedbackJSONFormat(t *testing.T) {
	fb := Feedback{
		Timestamp:    "2026-02-10T12:00:00Z",
		Version:      "1.2.0",
		Tier:         "pro",
		FavoriteFeature: "Perceptual hashing",
		NPSScore:     4,
	}

	data, err := json.Marshal(fb)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify JSON structure
	var fb2 Feedback
	if err := json.Unmarshal(data, &fb2); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if fb2.Tier != fb.Tier {
		t.Error("Tier mismatch")
	}
}
```

---

## Usage Examples

### Interactive Collection
```bash
./file-deduplicator --feedback
```

**User flow:**
```
$ ./file-deduplicator --feedback

📋 Customer Feedback Collection

Help us improve file-deduplicator!

1. What's your favorite feature? Perceptual hashing
2. What's annoying or confusing? Slow on large directories
3. What feature would make you upgrade to Team? No, would use if API access
4. How would you rate file-deduplicator (1-5)? 4
5. Any suggestions or ideas? Add web dashboard

💾 Saving feedback to: ~/.config/file-deduplicator/feedback.json
✅ Feedback saved! Thank you for your input.

Would you like to send this feedback directly to the creator?
(y/N): y

📧 Copy this feedback and email to: luinbytes@gmail.com

Feedback:
{
  "timestamp": "2026-02-10T12:00:00Z",
  "version": "1.2.0",
  "tier": "pro",
  "feedback": {
    "favorite_feature": "Perceptual hashing",
    "pain_points": ["Slow on large directories"],
    "team_interest": "No, would use if API access",
    "nps_score": 4,
    "suggestions": ["Add web dashboard"]
  },
  "system": {
    "os": "darwin",
    "go_version": "go1.21",
    "scan_stats": {
      "files_scanned": 12543,
      "duplicates_found": 342
    }
  }
}

Subject: Feedback for file-deduplicator

Paste above feedback here.

✅ Thanks for your feedback!
```

### Viewing Collected Feedback

```bash
cat ~/.config/file-deduplicator/feedback.json | jq .
```

**Output:**
```json
[
  {
    "timestamp": "2026-02-10T12:00:00Z",
    "version": "1.2.0",
    "tier": "pro",
    "feedback": {
      "favorite_feature": "Perceptual hashing",
      "pain_points": ["Slow on large directories"],
      "team_interest": "No, would use if API access",
      "nps_score": 4,
      "suggestions": ["Add web dashboard"]
    },
    "system": {
      "os": "darwin",
      "go_version": "go1.21",
      "scan_stats": {
        "files_scanned": 12543,
        "duplicates_found": 342
      }
    }
  }
]
```

---

## Integration Points

### Email Templates

**Template 8 (Feedback Request):** Update to include `--feedback` flag
```
🎥 Demo Video: [Link to demo]
📚 Feedback Collection: ./file-deduplicator --feedback
```

### Post-Launch Analysis

**File:** `file-deduplicator/docs/feedback-analysis.go` (future)

```go
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type FeedbackStats struct {
	TotalFeedbacks   int    `json:"total_feedbacks"`
	AverageNPS      float64 `json:"average_nps"`
	TopFeatures     []string `json:"top_features"`
	CommonPainPoints []string `json:"common_pain_points"`
}

func AnalyzeFeedback() error {
	data, err := os.ReadFile("~/.config/file-deduplicator/feedback.json")
	if err != nil {
		return err
	}

	var feedbacks []Feedback
	if err := json.Unmarshal(data, &feedbacks); err != nil {
		return err
	}

	stats := FeedbackStats{
		TotalFeedbacks: len(feedbacks),
	}

	// Calculate average NPS
	var totalNPS int
	for _, fb := range feedbacks {
		totalNPS += fb.NPSScore
	}
	if stats.TotalFeedbacks > 0 {
		stats.AverageNPS = float64(totalNPS) / float64(stats.TotalFeedbacks)
	}

	// Extract top features
	featureCount := make(map[string]int)
	for _, fb := range feedbacks {
		featureCount[fb.FavoriteFeature]++
	}

	// Sort and get top 5
	// (Implementation details in actual code)

	// Output stats
	output, _ := json.MarshalIndent(stats, "", "  ")
	fmt.Println(string(output))

	return nil
}
```

---

## Benefits

### For Lu (Post-Launch)
- **NPS tracking** - Measure user satisfaction over time
- **Feature prioritization** - Know what users want most
- **Bug detection** - Common pain points reveal issues
- **Upgrade insights** - Team interest metrics for upselling
- **No infrastructure needed** - All feedback stored locally

### For Users
- **Easy feedback** - One command, no email required
- **Privacy** - Data stored locally, only sent if user opts in
- **Instant use** - Can use immediately after scanning

---

## Files to Create

```
file-deduplicator/
├── feedback/
│   ├── feedback.go           # Main feedback collection logic
│   ├── feedback_test.go      # Unit tests
│   └── docs/
│       ├── feedback-usage.md  # Usage examples
│       └── feedback-analysis.go  # Future analysis tool
└── cmd/
    └── root.go                 # Add --feedback flag
```

---

## Testing Checklist

- [ ] Interactive prompts work correctly
- [ ] Feedback saves to JSON file
- [ ] JSON format valid
- [ ] Tests pass
- [ ] Manual testing with real input
- [ ] Integration with existing CLI works

---

## Estimated Commit Count

- Add `--feedback` flag (1 commit)
- Implement `feedback.go` (1 commit)
- Add tests (1 commit)
- Add docs (1 commit)
- **Total:** 4 commits

**Time estimate:** 1.5-2 hours

---

*Designed: 2026-02-09*
*Ready for implementation when Lu approves*
