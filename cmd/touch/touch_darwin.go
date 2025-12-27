//go:build darwin

package touch

import (
	"os"
	"syscall"
	"time"
)

func getAtime(info os.FileInfo) time.Time {
	stat := info.Sys().(*syscall.Stat_t)
	return time.Unix(stat.Atimespec.Sec, stat.Atimespec.Nsec)
}
