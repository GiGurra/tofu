//go:build windows

package usage

import "os"

// setupWindowResize handles terminal window resize signals.
// On Windows, SIGWINCH doesn't exist, so this is a no-op.
// The command runs quickly enough that window resize handling isn't critical.
func setupWindowResize(ptmx *os.File) {
	// No-op on Windows
}
