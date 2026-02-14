// +build windows

package main

import (
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procGetConsoleProcessList = kernel32.NewProc("GetConsoleProcessList")
)

// isDoubleClick returns true if the program was launched by double-click
// On Windows, GetConsoleProcessList returns 1 when double-clicked (only our process)
// and > 1 when run from an existing terminal
func isDoubleClick() bool {
	var processes [2]uint32
	ret, _, _ := procGetConsoleProcessList.Call(
		uintptr(unsafe.Pointer(&processes[0])),
		uintptr(2),
	)
	return ret == 1
}

// spawnTerminal spawns a new terminal window with the executable and --tui flag
func spawnTerminal() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	// Set environment variable to prevent infinite spawn loop
	cmd := exec.Command("cmd", "/c", "start", "", exe, "--tui")
	cmd.Env = append(os.Environ(), "_DEDUP_SPAWNED=1")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_CONSOLE,
	}
	return cmd.Start()
}
