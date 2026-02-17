// +build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/mattn/go-isatty"
)

// isDoubleClick returns true if the program was launched by double-click
// On Unix, we check if stdin is a TTY - if not, likely double-clicked
func isDoubleClick() bool {
	// If spawned by ourselves, don't re-spawn
	if os.Getenv("_DEDUP_SPAWNED") == "1" {
		return false
	}

	// If stdin is not a TTY, likely double-clicked (no terminal attached)
	return !isatty.IsTerminal(os.Stdin.Fd())
}

// spawnTerminal spawns a new terminal window with the executable and --tui flag
func spawnTerminal() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	// Set environment variable to prevent infinite spawn loop
	env := append(os.Environ(), "_DEDUP_SPAWNED=1")

	// Try different terminals in order of preference
	terminals := []struct {
		name string
		args []string
	}{
		{"x-terminal-emulator", []string{"-e", exe, "--tui"}},
		{"gnome-terminal", []string{"--", exe, "--tui"}},
		{"konsole", []string{"-e", exe, "--tui"}},
		{"xfce4-terminal", []string{"-e", exe + " --tui"}},
		{"xterm", []string{"-e", exe, "--tui"}},
	}

	// On macOS, use open command
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("open", "-a", "Terminal", exe, "--args", "--tui")
		cmd.Env = env
		return cmd.Start()
	}

	// Try each terminal emulator
	for _, term := range terminals {
		if _, err := exec.LookPath(term.name); err == nil {
			cmd := exec.Command(term.name, term.args...)
			cmd.Env = env
			if err := cmd.Start(); err == nil {
				return nil
			}
		}
	}

	return fmt.Errorf("no terminal emulator found")
}
