// Package tui provides an interactive terminal UI for file deduplication
package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			PaddingLeft(2).
			PaddingRight(2)

	itemStyle = lipgloss.NewStyle().PaddingLeft(4)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	checkedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	uncheckedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA"))

	previewStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1)
)

// FileInfo represents a file in the duplicate group
type FileInfo struct {
	Path     string
	Size     int64
	ModTime  string
	Selected bool
}

// DuplicateGroup represents a group of duplicate files
type DuplicateGroup struct {
	Hash       string
	Size       int64
	Files      []FileInfo
	Similarity float64
}

// keyMap defines keybindings for the TUI
type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Toggle   key.Binding
	ToggleAll key.Binding
	Confirm  key.Binding
	Quit     key.Binding
	Help     key.Binding
	Preview  key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Toggle: key.NewBinding(
		key.WithKeys("space", " "),
		key.WithHelp("space", "toggle selection"),
	),
	ToggleAll: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "toggle all"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm deletion"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q/esc", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Preview: key.NewBinding(
		key.WithKeys("p", "tab"),
		key.WithHelp("p/tab", "toggle preview"),
	),
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Toggle, k.ToggleAll},
		{k.Confirm, k.Preview, k.Help, k.Quit},
	}
}

// Model is the TUI state
type Model struct {
	groups          []DuplicateGroup
	currentGroup    int
	cursor          int
	showHelp        bool
	showPreview     bool
	confirmed       bool
	quitting        bool
	width           int
	height          int
	keys            keyMap
	help            help.Model
	filesToDelete   []string
	statusMsg       string
}

// New creates a new TUI model
func New(groups []DuplicateGroup) Model {
	return Model{
		groups:        groups,
		currentGroup:  0,
		cursor:        0,
		showHelp:      false,
		showPreview:   false,
		keys:          keys,
		help:          help.New(),
		filesToDelete: []string{},
	}
}

// Init initializes the TUI
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and user input
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp

		case key.Matches(msg, m.keys.Preview):
			m.showPreview = !m.showPreview

		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, m.keys.Down):
			if m.currentGroup < len(m.groups) {
				group := m.groups[m.currentGroup]
				if m.cursor < len(group.Files)-1 {
					m.cursor++
				}
			}

		case key.Matches(msg, m.keys.Toggle):
			if m.currentGroup < len(m.groups) {
				group := &m.groups[m.currentGroup]
				if m.cursor < len(group.Files) {
					group.Files[m.cursor].Selected = !group.Files[m.cursor].Selected
					m.updateStatus()
				}
			}

		case key.Matches(msg, m.keys.ToggleAll):
			if m.currentGroup < len(m.groups) {
				group := &m.groups[m.currentGroup]
				// Check if all are selected
				allSelected := true
				for i := range group.Files {
					if !group.Files[i].Selected {
						allSelected = false
						break
					}
				}
				// Toggle all to opposite state
				for i := range group.Files {
					group.Files[i].Selected = !allSelected
				}
				m.updateStatus()
			}

		case key.Matches(msg, m.keys.Confirm):
			if m.currentGroup < len(m.groups) {
				group := m.groups[m.currentGroup]
				for _, file := range group.Files {
					if file.Selected {
						m.filesToDelete = append(m.filesToDelete, file.Path)
					}
				}
				m.currentGroup++
				m.cursor = 0
				
				if m.currentGroup >= len(m.groups) {
					m.confirmed = true
					return m, tea.Quit
				}
			}
		}
	}

	return m, nil
}

// updateStatus updates the status message
func (m *Model) updateStatus() {
	if m.currentGroup >= len(m.groups) {
		return
	}
	group := m.groups[m.currentGroup]
	selected := 0
	for _, f := range group.Files {
		if f.Selected {
			selected++
		}
	}
	m.statusMsg = fmt.Sprintf("Selected: %d/%d", selected, len(group.Files))
}

// View renders the TUI
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	if m.confirmed {
		return m.renderConfirmation()
	}

	if len(m.groups) == 0 {
		return "No duplicates found!\n"
	}

	if m.currentGroup >= len(m.groups) {
		return m.renderConfirmation()
	}

	var s strings.Builder

	// Header
	s.WriteString(titleStyle.Render(fmt.Sprintf(" File Deduplicator v3.0.0 ")))
	s.WriteString("\n\n")

	// Group info
	group := m.groups[m.currentGroup]
	s.WriteString(headerStyle.Render(fmt.Sprintf("Duplicate Group %d/%d", m.currentGroup+1, len(m.groups))))
	s.WriteString("\n")
	
	if group.Similarity < 100.0 {
		s.WriteString(infoStyle.Render(fmt.Sprintf("Similarity: %.0f%% | Size: %s", group.Similarity, formatBytes(group.Size))))
	} else {
		s.WriteString(infoStyle.Render(fmt.Sprintf("Exact match | Size: %s", formatBytes(group.Size))))
	}
	s.WriteString("\n\n")

	// File list
	s.WriteString(m.renderFileList(group))
	s.WriteString("\n")

	// Status
	if m.statusMsg != "" {
		s.WriteString(infoStyle.Render(m.statusMsg))
		s.WriteString("\n")
	}

	// Help
	if m.showHelp {
		s.WriteString("\n")
		s.WriteString(m.help.FullHelpView(m.keys.FullHelp()))
	} else {
		s.WriteString("\n")
		s.WriteString(m.help.ShortHelpView(m.keys.ShortHelp()))
	}

	return s.String()
}

// renderFileList renders the list of files in the current group
func (m Model) renderFileList(group DuplicateGroup) string {
	var s strings.Builder

	for i, file := range group.Files {
		var line strings.Builder

		// Checkbox
		if file.Selected {
			line.WriteString(checkedStyle.Render("[✓] "))
		} else {
			line.WriteString(uncheckedStyle.Render("[ ] "))
		}

		// File name
		filename := filepath.Base(file.Path)
		if i == m.cursor {
			line.WriteString(selectedItemStyle.Render("> " + filename))
		} else {
			line.WriteString(itemStyle.Render(filename))
		}

		// File info
		info := fmt.Sprintf(" (%s, %s)", formatBytes(file.Size), file.ModTime)
		line.WriteString(infoStyle.Render(info))

		s.WriteString(line.String())
		s.WriteString("\n")
	}

	return s.String()
}

// renderConfirmation renders the final confirmation screen
func (m Model) renderConfirmation() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(" Confirmation "))
	s.WriteString("\n\n")

	if len(m.filesToDelete) == 0 {
		s.WriteString("No files selected for deletion.\n")
	} else {
		s.WriteString(fmt.Sprintf("About to delete %d files:\n\n", len(m.filesToDelete)))
		for i, path := range m.filesToDelete {
			if i >= 10 {
				s.WriteString(fmt.Sprintf("... and %d more\n", len(m.filesToDelete)-10))
				break
			}
			s.WriteString(fmt.Sprintf("  • %s\n", path))
		}
		s.WriteString("\nPress Enter to confirm, or Esc to cancel.\n")
	}

	return s.String()
}

// GetFilesToDelete returns the list of files marked for deletion
func (m Model) GetFilesToDelete() []string {
	return m.filesToDelete
}

// Run starts the TUI and returns the selected files to delete
func Run(groups []DuplicateGroup) ([]string, error) {
	p := tea.NewProgram(New(groups), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		return nil, err
	}

	model := m.(Model)
	return model.GetFilesToDelete(), nil
}

// formatBytes formats bytes into human-readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ConvertDuplicateGroup converts the main package's DuplicateGroup to TUI format
func ConvertDuplicateGroup(hash string, size int64, files []struct {
	Path    string
	Size    int64
	ModTime string
}, similarity float64) DuplicateGroup {
	convertedFiles := make([]FileInfo, len(files))
	for i, f := range files {
		convertedFiles[i] = FileInfo{
			Path:     f.Path,
			Size:     f.Size,
			ModTime:  f.ModTime,
			Selected: false,
		}
	}
	return DuplicateGroup{
		Hash:       hash,
		Size:       size,
		Files:      convertedFiles,
		Similarity: similarity,
	}
}
