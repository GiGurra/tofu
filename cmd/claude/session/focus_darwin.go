//go:build darwin

package session

import (
	"os/exec"
)

// tryFocusAttachedSession attempts to focus the terminal window that has the session attached.
// On macOS, we use AppleScript to bring Terminal.app to the front.
func tryFocusAttachedSession(tmuxSession string) {
	// Try to activate Terminal.app (most common terminal on macOS)
	script := `tell application "Terminal" to activate`
	_ = exec.Command("osascript", "-e", script).Run()
}
