//go:build windows

package session

// tryFocusAttachedSession attempts to focus the terminal window that has the session attached.
// On Windows native, tmux sessions aren't typical, so this is a no-op.
func tryFocusAttachedSession(tmuxSession string) {
	// No-op
}
