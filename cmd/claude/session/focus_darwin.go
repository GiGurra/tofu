//go:build darwin

package session

// tryFocusAttachedSession attempts to focus the terminal window that has the session attached.
// On macOS, this is not reliably possible without knowing which terminal app and window.
func tryFocusAttachedSession(tmuxSession string) {
	// Best effort: the user already got a message saying the session is attached elsewhere.
}

// FocusOwnWindow attempts to focus the current process's terminal window.
// On macOS, this would require AppleScript and knowing which terminal app we're in.
func FocusOwnWindow() bool {
	// Could potentially use AppleScript to focus Terminal.app or iTerm2,
	// but we'd need to identify which terminal process we're running in.
	return false
}

// GetOwnWindowTitle returns the title of the current terminal window.
// On macOS, this would require terminal-specific queries.
func GetOwnWindowTitle() string {
	return ""
}
