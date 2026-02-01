//go:build !windows

package session

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// attachToSession attaches to a tmux session, replacing the current process.
// Used when inbox watcher isn't needed (e.g., internal calls).
func attachToSession(tmuxSession string) error {
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("tmux not found: %w", err)
	}

	// Use syscall.Exec to replace current process with tmux attach
	args := []string{"tmux", "attach-session", "-t", tmuxSession}
	env := os.Environ()

	return syscall.Exec(tmuxPath, args, env)
}
