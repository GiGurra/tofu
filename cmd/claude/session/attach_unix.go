//go:build !windows

package session

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// attachToSession attaches to a tmux session, replacing the current process.
func attachToSession(tmuxSession string) error {
	return attachToSessionWithFlags(tmuxSession, false)
}

// attachToSessionWithFlags attaches to a tmux session with optional force flag.
// If force is true, uses -d to detach other clients.
func attachToSessionWithFlags(tmuxSession string, force bool) error {
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("tmux not found: %w", err)
	}

	// Use syscall.Exec to replace current process with tmux attach
	args := []string{"tmux", "attach-session", "-t", tmuxSession}
	if force {
		args = []string{"tmux", "attach-session", "-d", "-t", tmuxSession}
	}
	env := os.Environ()

	return syscall.Exec(tmuxPath, args, env)
}
