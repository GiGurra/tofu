//go:build linux || darwin

package ls

import (
	"io/fs"
	"os/user"
	"strconv"
	"syscall"
)

// isExecutable checks if a file is executable on Unix
func isExecutable(_ string, mode fs.FileMode) bool {
	return mode&0111 != 0
}

// FileStatInfo contains platform-specific file metadata
type FileStatInfo struct {
	Nlink  uint64 // number of hard links
	Uid    uint32
	Gid    uint32
	Inode  uint64
	Blocks int64 // number of 512-byte blocks
	Valid  bool  // whether the info is valid
}

// getFileStatInfo extracts platform-specific stat info from FileInfo
func getFileStatInfo(info fs.FileInfo) FileStatInfo {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return FileStatInfo{
			Nlink:  uint64(stat.Nlink),
			Uid:    stat.Uid,
			Gid:    stat.Gid,
			Inode:  stat.Ino,
			Blocks: stat.Blocks,
			Valid:  true,
		}
	}
	return FileStatInfo{Valid: false}
}

// getOwner returns the owner name or uid for a file
func getOwner(stat FileStatInfo, numeric bool) string {
	if !stat.Valid {
		return "?"
	}
	if numeric {
		return strconv.FormatUint(uint64(stat.Uid), 10)
	}
	if u, err := user.LookupId(strconv.FormatUint(uint64(stat.Uid), 10)); err == nil {
		return u.Username
	}
	return strconv.FormatUint(uint64(stat.Uid), 10)
}

// getGroup returns the group name or gid for a file
func getGroup(stat FileStatInfo, numeric bool, fullGroup bool) string {
	if !stat.Valid {
		return "?"
	}
	if numeric {
		return strconv.FormatUint(uint64(stat.Gid), 10)
	}
	if g, err := user.LookupGroupId(strconv.FormatUint(uint64(stat.Gid), 10)); err == nil {
		return g.Name
	}
	return strconv.FormatUint(uint64(stat.Gid), 10)
}
