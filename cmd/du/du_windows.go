//go:build windows

package du

import (
	"io/fs"
)

// getDiskUsage returns estimated disk usage in bytes.
// On Windows, we approximate by rounding up to 4096-byte clusters (typical NTFS cluster size).
func getDiskUsage(info fs.FileInfo) int64 {
	size := info.Size()
	if size == 0 {
		return 0
	}
	// Round up to 4096-byte clusters (typical NTFS cluster size)
	return ((size + 4095) / 4096) * 4096
}
