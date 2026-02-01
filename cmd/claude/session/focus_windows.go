//go:build windows

package session

// tryFocusAttachedSession attempts to focus the terminal window that has the session attached.
// On Windows native, this is not easily doable without more complex Win32 API calls.
func tryFocusAttachedSession(tmuxSession string) {
	// Windows native typically doesn't have tmux sessions in the same way
	// This is a no-op for now
}
