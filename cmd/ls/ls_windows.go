//go:build windows

package ls

import (
	"io/fs"
	"os/user"
	"strings"
	"syscall"
)

// isExecutable checks if a file is executable on Windows
// Windows determines executability by file extension
func isExecutable(name string, mode fs.FileMode) bool {
	// Check Unix execute bit for cross-platform compatibility
	if mode&0111 != 0 {
		return true
	}
	// Check common Windows executable extensions
	lower := strings.ToLower(name)
	for _, ext := range []string{".exe", ".bat", ".cmd", ".com", ".ps1", ".msi"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// FileStatInfo contains platform-specific file metadata
type FileStatInfo struct {
	Nlink  uint64 // number of hard links (always 1 on Windows for simplicity)
	Uid    uint32 // not used on Windows
	Gid    uint32 // not used on Windows
	Inode  uint64 // file index on Windows
	Blocks int64  // estimated blocks based on file size
	Valid  bool   // whether the info is valid
}

// getFileStatInfo extracts platform-specific stat info from FileInfo
func getFileStatInfo(info fs.FileInfo) FileStatInfo {
	// Try to get Windows-specific file info
	if sys := info.Sys(); sys != nil {
		if winData, ok := sys.(*syscall.Win32FileAttributeData); ok {
			// Calculate file index from timestamps as a pseudo-inode
			// Windows doesn't expose inode easily without opening the file
			fileIndex := uint64(winData.CreationTime.Nanoseconds())

			// Estimate blocks (4096-byte blocks)
			size := info.Size()
			blocks := (size + 4095) / 4096

			return FileStatInfo{
				Nlink:  1, // Windows doesn't track hard links easily
				Uid:    0,
				Gid:    0,
				Inode:  fileIndex,
				Blocks: blocks,
				Valid:  true,
			}
		}
	}

	// Fallback: estimate based on size
	size := info.Size()
	blocks := (size + 4095) / 4096

	return FileStatInfo{
		Nlink:  1,
		Blocks: blocks,
		Valid:  true,
	}
}

// getOwner returns the owner name for a file on Windows
func getOwner(stat FileStatInfo, numeric bool) string {
	// On Windows, try to get the current user as owner
	// A more complete implementation would use Windows security APIs
	if u, err := user.Current(); err == nil {
		if numeric {
			return u.Uid
		}
		return u.Username
	}
	return "OWNER"
}

// getGroup returns the group name for a file on Windows
func getGroup(stat FileStatInfo, numeric bool, fullGroup bool) string {
	// On Windows, groups work differently than Unix
	// The Gid is typically a long SID like S-1-5-21-...
	if u, err := user.Current(); err == nil {
		if numeric || fullGroup {
			// Show the full SID when requested
			return u.Gid
		}
		// Try to look up group name
		if g, err := user.LookupGroupId(u.Gid); err == nil {
			return g.Name
		}
		// SID didn't resolve to a friendly name, truncate it
		if len(u.Gid) > 5 {
			return u.Gid[:5] + ".."
		}
		return u.Gid
	}
	return ".."
}

// getFileIndex retrieves the unique file index for Windows files
// This can be used as a pseudo-inode
func getFileIndex(path string) (uint64, error) {
	pathp, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}

	handle, err := syscall.CreateFile(
		pathp,
		syscall.GENERIC_READ,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_BACKUP_SEMANTICS, // Needed for directories
		0,
	)
	if err != nil {
		return 0, err
	}
	defer syscall.CloseHandle(handle)

	var fileInfo syscall.ByHandleFileInformation
	if err := syscall.GetFileInformationByHandle(handle, &fileInfo); err != nil {
		return 0, err
	}

	// Combine high and low parts of file index
	return (uint64(fileInfo.FileIndexHigh) << 32) | uint64(fileInfo.FileIndexLow), nil
}

// getFileStatInfoWithPath gets more accurate info by opening the file (for inode)
func getFileStatInfoWithPath(info fs.FileInfo, path string) FileStatInfo {
	stat := getFileStatInfo(info)

	// Try to get accurate file index
	if idx, err := getFileIndex(path); err == nil {
		stat.Inode = idx
	}

	// Try to get hard link count
	if count, err := getHardLinkCount(path); err == nil {
		stat.Nlink = count
	}

	return stat
}

// getHardLinkCount returns the number of hard links to a file
func getHardLinkCount(path string) (uint64, error) {
	pathp, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 1, err
	}

	handle, err := syscall.CreateFile(
		pathp,
		0, // No access needed for link count
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return 1, err
	}
	defer syscall.CloseHandle(handle)

	var fileInfo syscall.ByHandleFileInformation
	if err := syscall.GetFileInformationByHandle(handle, &fileInfo); err != nil {
		return 1, err
	}

	return uint64(fileInfo.NumberOfLinks), nil
}

// isHidden checks if a file has the Windows hidden attribute
func isHidden(path string) bool {
	pathp, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return false
	}

	attrs, err := syscall.GetFileAttributes(pathp)
	if err != nil {
		return false
	}

	return attrs&syscall.FILE_ATTRIBUTE_HIDDEN != 0
}
