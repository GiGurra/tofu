//go:build darwin

package df

import (
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

	// Inode info
	info.IUsed = s.Files - s.Ffree
	info.IAvailable = s.Ffree

	if s.Files > 0 {
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

// getMounts returns a list of all mounted filesystems using getfsstat
func getMounts() ([]MountInfo, error) {
	// Get the number of mounted filesystems
	n, err := unix.Getfsstat(nil, unix.MNT_NOWAIT)
	if err != nil {
		return nil, err
	}

	// Get the actual filesystem stats
	buf := make([]unix.Statfs_t, n)
	_, err = unix.Getfsstat(buf, unix.MNT_NOWAIT)
	if err != nil {
		return nil, err
	}

	var mounts []MountInfo
	for _, stat := range buf {
		mounts = append(mounts, MountInfo{
			Device:     byteSliceToString(stat.Mntfromname[:]),
			MountPoint: byteSliceToString(stat.Mntonname[:]),
			FSType:     byteSliceToString(stat.Fstypename[:]),
		})
	}

	return mounts, nil
}

// byteSliceToString converts a null-terminated byte slice to a string
func byteSliceToString(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
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
