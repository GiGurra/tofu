//go:build !windows

package usage

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
)

// setupWindowResize handles terminal window resize signals.
// On Unix, we listen for SIGWINCH and update the PTY size accordingly.
func setupWindowResize(ptmx *os.File) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if ws, err := pty.GetsizeFull(os.Stdin); err == nil {
				_ = pty.Setsize(ptmx, ws)
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial size sync
}
