//go:build linux || darwin

package du

import (
	"io/fs"
	"syscall"
)

// getDiskUsage returns actual disk usage in bytes using syscall.Stat_t.Blocks
func getDiskUsage(info fs.FileInfo) int64 {
	if sys, ok := info.Sys().(*syscall.Stat_t); ok {
		// Blocks are in 512-byte units
		return sys.Blocks * 512
	}
	// Fallback: round up to 4096-byte blocks
	if info.Size() == 0 {
		return 0
	}
	return ((info.Size() + 4095) / 4096) * 4096
}
