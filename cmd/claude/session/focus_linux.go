//go:build linux

package session

import (
	"os"
	"os/exec"
	"strings"
)

// tryFocusAttachedSession attempts to focus the terminal window that has the session attached.
// This is best-effort and may not work on all systems.
func tryFocusAttachedSession(tmuxSession string) {
	// Get the TTY of the attached client
	tty := getTmuxClientTTY(tmuxSession)
	if tty == "" {
		return
	}

	if isWSL() {
		// On WSL, try to activate the Windows Terminal window
		tryFocusWSLWindow()
		return
	}

	// On native Linux, try xdotool to find and focus window by PID
	if _, err := exec.LookPath("xdotool"); err != nil {
		return
	}

	// Find the PID of the process controlling the TTY
	pid := findTTYOwnerPID(tty)
	if pid == "" {
		return
	}

	// Try to find and focus window associated with this PID
	// xdotool search --pid finds windows owned by a process
	cmd := exec.Command("xdotool", "search", "--pid", pid)
	output, err := cmd.Output()
	if err != nil {
		return
	}

	// Get first window ID
	windowIDs := strings.Fields(string(output))
	if len(windowIDs) == 0 {
		return
	}

	// Activate (focus) the window
	_ = exec.Command("xdotool", "windowactivate", windowIDs[0]).Run()
}

// getTmuxClientTTY returns the TTY of the first client attached to a tmux session.
func getTmuxClientTTY(tmuxSession string) string {
	// tmux list-clients -t session -F "#{client_tty}"
	cmd := exec.Command("tmux", "list-clients", "-t", tmuxSession, "-F", "#{client_tty}")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 {
		return ""
	}
	return lines[0]
}

// findTTYOwnerPID finds the PID of the process that owns a TTY.
func findTTYOwnerPID(tty string) string {
	// Use ps to find process on this TTY
	// ps -t /dev/pts/0 -o pid= gets PID of processes on that TTY
	cmd := exec.Command("ps", "-t", tty, "-o", "pid=")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	pids := strings.Fields(string(output))
	if len(pids) == 0 {
		return ""
	}
	// Return first PID (usually the shell)
	return pids[0]
}

// isWSL detects if running in Windows Subsystem for Linux.
func isWSL() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	lower := strings.ToLower(string(data))
	return strings.Contains(lower, "microsoft") || strings.Contains(lower, "wsl")
}

// tryFocusWSLWindow attempts to focus the Windows Terminal window from WSL.
func tryFocusWSLWindow() {
	// Use PowerShell to try focusing Windows Terminal
	// This is best-effort and may not work in all configurations
	script := `
$wshell = New-Object -ComObject wscript.shell
$wshell.AppActivate('Windows Terminal')
`
	_ = exec.Command("/mnt/c/Windows/System32/WindowsPowerShell/v1.0/powershell.exe",
		"-NoProfile", "-NonInteractive", "-Command", script).Run()
}
