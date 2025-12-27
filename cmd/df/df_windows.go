//go:build windows

package df

import (
	"syscall"
	"unsafe"
)

// MountInfo holds information about a mounted filesystem
type MountInfo struct {
	Device     string
	MountPoint string
	FSType     string
}

// windowsStatfs holds filesystem statistics for Windows
type windowsStatfs struct {
	Bsize   int64
	Blocks  uint64
	Bfree   uint64
	Bavail  uint64
	Files   uint64
	Ffree   uint64
}

// extractStatInfo extracts filesystem info from windowsStatfs
func extractStatInfo(stat interface{}, device, mountPoint, fsType string) FilesystemInfo {
	s := stat.(windowsStatfs)
	info := FilesystemInfo{
		Filesystem: device,
		MountPoint: mountPoint,
		FSType:     fsType,
	}

	info.Size = s.Blocks * uint64(s.Bsize)
	info.Available = s.Bavail * uint64(s.Bsize)
	info.Used = info.Size - (s.Bfree * uint64(s.Bsize))

	if info.Size > 0 {
		info.Percent = float64(info.Used) / float64(info.Size) * 100
	}

	// Windows doesn't provide inode info
	info.IUsed = 0
	info.IAvailable = 0
	info.IPercent = 0

	return info
}

// getStatfs returns filesystem statistics for the given path
func getStatfs(path string) (windowsStatfs, error) {
	var stat windowsStatfs

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpaceEx := kernel32.NewProc("GetDiskFreeSpaceExW")

	var freeBytesAvailable, totalBytes, totalFreeBytes uint64

	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return stat, err
	}

	ret, _, err := getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)

	if ret == 0 {
		return stat, err
	}

	// Windows doesn't have a block concept, use 4096 as virtual block size
	stat.Bsize = 4096
	stat.Blocks = totalBytes / uint64(stat.Bsize)
	stat.Bfree = totalFreeBytes / uint64(stat.Bsize)
	stat.Bavail = freeBytesAvailable / uint64(stat.Bsize)
	// Windows doesn't expose inode info
	stat.Files = 0
	stat.Ffree = 0

	return stat, nil
}

// getMounts returns a list of all mounted filesystems (drives on Windows)
func getMounts() ([]MountInfo, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getLogicalDrives := kernel32.NewProc("GetLogicalDrives")
	getDriveType := kernel32.NewProc("GetDriveTypeW")
	getVolumeInformation := kernel32.NewProc("GetVolumeInformationW")

	ret, _, _ := getLogicalDrives.Call()
	drives := uint32(ret)

	var mounts []MountInfo
	for i := 0; i < 26; i++ {
		if drives&(1<<uint(i)) != 0 {
			drive := string(rune('A'+i)) + ":\\"

			drivePtr, _ := syscall.UTF16PtrFromString(drive)
			driveType, _, _ := getDriveType.Call(uintptr(unsafe.Pointer(drivePtr)))

			// Get filesystem type
			fsType := "unknown"
			fsNameBuf := make([]uint16, 256)
			ret, _, _ := getVolumeInformation.Call(
				uintptr(unsafe.Pointer(drivePtr)),
				0, 0, 0, 0, 0,
				uintptr(unsafe.Pointer(&fsNameBuf[0])),
				uintptr(len(fsNameBuf)),
			)
			if ret != 0 {
				fsType = syscall.UTF16ToString(fsNameBuf)
			}

			// Format: "C: (Fixed)" for better readability
			driveLetter := string(rune('A'+i)) + ":"
			mounts = append(mounts, MountInfo{
				Device:     driveLetter + " (" + getDriveTypeName(driveType) + ")",
				MountPoint: drive,
				FSType:     fsType,
			})
		}
	}

	return mounts, nil
}

func getDriveTypeName(driveType uintptr) string {
	switch driveType {
	case 2:
		return "Removable"
	case 3:
		return "Fixed"
	case 4:
		return "Network"
	case 5:
		return "CD-ROM"
	case 6:
		return "RAM Disk"
	default:
		return "Unknown"
	}
}

// isLocalFilesystem returns true if the filesystem is local (not network)
func isLocalFilesystem(fsType string) bool {
	return fsType != "Network"
}

// isPseudoFilesystem returns false on Windows (no pseudo filesystems)
func isPseudoFilesystem(fsType string) bool {
	return false
}