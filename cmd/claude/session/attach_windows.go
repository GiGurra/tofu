//go:build windows

package session

import "fmt"

// attachToSession is not supported on Windows native (no tmux).
// Sessions should be managed differently on Windows.
func attachToSession(tmuxSession string) error {
	return fmt.Errorf("tmux sessions not supported on Windows native")
}
