//go:build linux

package session

// tryFocusAttachedSession attempts to focus the terminal window that has the session attached.
// On Linux/WSL, this is not reliably possible due to:
// - Windows preventing focus stealing (SetForegroundWindow lies)
// - Window titles being unpredictable (could be claude, tofu, script name, etc.)
// - No reliable way to map tmux sessions to window handles
//
// We just print a message and let the user find the window manually.
func tryFocusAttachedSession(tmuxSession string) {
	// Best effort: the user already got a message saying the session is attached elsewhere.
	// They'll have to find the window themselves.
}
