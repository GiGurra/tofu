//go:build linux

package df

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// extractStatInfo extracts filesystem info from unix.Statfs_t
func extractStatInfo(stat interface{}, device, mountPoint, fsType string) FilesystemInfo {
	s := stat.(unix.Statfs_t)
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

	// Inode info - some filesystems (like drvfs) don't provide valid inode counts
	if s.Files > 0 && s.Ffree <= s.Files {
		info.IUsed = s.Files - s.Ffree
		info.IAvailable = s.Ffree
		info.IPercent = float64(info.IUsed) / float64(s.Files) * 100
	}

	return info
}

// MountInfo holds information about a mounted filesystem
type MountInfo struct {
	Device     string
	MountPoint string
	FSType     string
}

// getStatfs returns filesystem statistics for the given path
func getStatfs(path string) (unix.Statfs_t, error) {
	var stat unix.Statfs_t
	err := unix.Statfs(path, &stat)
	return stat, err
}

// getMounts returns a list of all mounted filesystems by parsing /proc/mounts
func getMounts() ([]MountInfo, error) {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		// Fallback to /etc/mtab
		file, err = os.Open("/etc/mtab")
		if err != nil {
			return nil, err
		}
	}
	defer file.Close()

	var mounts []MountInfo
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}
		mounts = append(mounts, MountInfo{
			Device:     unescapeOctal(fields[0]),
			MountPoint: unescapeOctal(fields[1]),
			FSType:     fields[2],
		})
	}

	return mounts, scanner.Err()
}

// unescapeOctal converts octal escape sequences (e.g., \040 for space, \134 for backslash)
// used in /proc/mounts back to their original characters
func unescapeOctal(s string) string {
	var result strings.Builder
	result.Grow(len(s))

	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+3 < len(s) {
			// Check if next 3 chars are octal digits
			octal := s[i+1 : i+4]
			if isOctalDigits(octal) {
				val, _ := strconv.ParseInt(octal, 8, 32)
				result.WriteByte(byte(val))
				i += 3
				continue
			}
		}
		result.WriteByte(s[i])
	}

	return result.String()
}

func isOctalDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '7' {
			return false
		}
	}
	return true
}

// isLocalFilesystem returns true if the filesystem type is local (not network)
func isLocalFilesystem(fsType string) bool {
	networkTypes := map[string]bool{
		"nfs": true, "nfs4": true, "cifs": true, "smb": true,
		"smbfs": true, "ncpfs": true, "afs": true, "gfs": true,
		"gfs2": true, "glusterfs": true, "ceph": true, "fuse.sshfs": true,
	}
	return !networkTypes[fsType]
}

// isPseudoFilesystem returns true if the filesystem is a pseudo/virtual filesystem
func isPseudoFilesystem(fsType string) bool {
	pseudoTypes := map[string]bool{
		"sysfs": true, "proc": true, "devtmpfs": true, "devpts": true,
		"tmpfs": true, "securityfs": true, "cgroup": true, "cgroup2": true,
		"pstore": true, "debugfs": true, "hugetlbfs": true, "mqueue": true,
		"fusectl": true, "configfs": true, "binfmt_misc": true, "autofs": true,
		"efivarfs": true, "tracefs": true, "bpf": true, "overlay": true,
	}
	return pseudoTypes[fsType]
}
