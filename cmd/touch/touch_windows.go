//go:build windows

package touch

import (
	"os"
	"syscall"
	"time"
)

func getAtime(info os.FileInfo) time.Time {
	stat := info.Sys().(*syscall.Win32FileAttributeData)
	return time.Unix(0, stat.LastAccessTime.Nanoseconds())
}
