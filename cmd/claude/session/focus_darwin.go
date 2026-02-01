//go:build darwin

package session

// tryFocusAttachedSession attempts to focus the terminal window that has the session attached.
// On macOS, this is not reliably possible without knowing which terminal app and window.
func tryFocusAttachedSession(tmuxSession string) {
	// Best effort: the user already got a message saying the session is attached elsewhere.
}
