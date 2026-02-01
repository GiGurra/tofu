//go:build darwin

package session

import (
	"os"
	"os/exec"
	"strings"
)

// tryFocusAttachedSession attempts to focus the terminal window that has the session attached.
func tryFocusAttachedSession(tmuxSession string) {
	// Try to focus the terminal - best effort
	FocusOwnWindow()
}

// FocusOwnWindow attempts to focus the current process's terminal window.
// Uses AppleScript to activate the detected terminal application.
func FocusOwnWindow() bool {
	termApp := detectTerminalApp()
	if termApp == "" {
		return false
	}

	// Use AppleScript to activate the terminal
	script := `tell application "` + termApp + `" to activate`
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run() == nil
}

// GetOwnWindowTitle returns the title of the current terminal window.
func GetOwnWindowTitle() string {
	// Not easily available on macOS without terminal-specific APIs
	return ""
}

// detectTerminalApp tries to determine which terminal application we're running in.
func detectTerminalApp() string {
	// Check TERM_PROGRAM environment variable (set by most macOS terminals)
	termProgram := os.Getenv("TERM_PROGRAM")
	switch termProgram {
	case "Apple_Terminal":
		return "Terminal"
	case "iTerm.app":
		return "iTerm2"
	case "vscode":
		return "Visual Studio Code"
	case "Hyper":
		return "Hyper"
	case "Alacritty":
		return "Alacritty"
	case "kitty":
		return "kitty"
	case "WarpTerminal":
		return "Warp"
	}

	// Check if running inside tmux - look at parent processes
	if os.Getenv("TMUX") != "" {
		// We're in tmux, try to find the terminal from tmux client
		return detectTerminalFromTmux()
	}

	// Fallback: try common terminals
	terminals := []string{"iTerm2", "Terminal", "Alacritty", "kitty"}
	for _, term := range terminals {
		if isAppRunning(term) {
			return term
		}
	}

	return ""
}

// detectTerminalFromTmux tries to find which terminal is running the tmux client.
func detectTerminalFromTmux() string {
	// Get the tmux client tty
	cmd := exec.Command("tmux", "display-message", "-p", "#{client_tty}")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	tty := strings.TrimSpace(string(output))
	if tty == "" {
		return ""
	}

	// Try to find the process owning this tty using lsof
	cmd = exec.Command("lsof", "-t", tty)
	output, err = cmd.Output()
	if err != nil {
		return ""
	}

	// Get the first PID
	pids := strings.Fields(string(output))
	if len(pids) == 0 {
		return ""
	}

	// Get the process name
	cmd = exec.Command("ps", "-p", pids[0], "-o", "comm=")
	output, err = cmd.Output()
	if err != nil {
		return ""
	}

	procName := strings.TrimSpace(string(output))
	switch {
	case strings.Contains(procName, "iTerm"):
		return "iTerm2"
	case strings.Contains(procName, "Terminal"):
		return "Terminal"
	case strings.Contains(procName, "Alacritty"):
		return "Alacritty"
	case strings.Contains(procName, "kitty"):
		return "kitty"
	case strings.Contains(procName, "Warp"):
		return "Warp"
	}

	return ""
}

// isAppRunning checks if an application is running on macOS.
func isAppRunning(appName string) bool {
	script := `tell application "System Events" to (name of processes) contains "` + appName + `"`
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}
