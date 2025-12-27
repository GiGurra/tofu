package df

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Paths   []string `pos:"true" optional:"true" help:"Paths to analyze. Defaults to all mounted filesystems." default:""`
	All     bool     `short:"a" help:"Include all filesystems, including pseudo filesystems." optional:"true"`
	Human   bool     `short:"h" help:"Print sizes in human readable format." optional:"true"`
	Inode   bool     `short:"i" help:"List inode information instead of block usage." optional:"true"`
	Local   bool     `short:"l" help:"Limit listing to local filesystems." optional:"true"`
	Type    string   `short:"t" help:"Limit listing to filesystems of a specific type." default:""`
	Sort    string   `short:"S" help:"Sort by: 'used', 'available', 'percent', or 'name' (default)." default:"name" alts:"name,used,available,percent"`
	Reverse bool     `short:"r" help:"Reverse the sort order." optional:"true"`
}

type FilesystemInfo struct {
	Filesystem string
	Size       uint64
	Used       uint64
	Available  uint64
	Percent    float64
	IUsed      uint64
	IAvailable uint64
	IPercent   float64
	MountPoint string
	FSType     string
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "df",
		Short:       "Report filesystem disk space usage",
		Long:        "Report filesystem disk space usage, like the Unix df command but cross-platform.",
		ParamEnrich: common.DefaultParamEnricher(),
		InitFunc: func(params *Params, cmd *cobra.Command) error {
			cmd.Flags().BoolP("help", "", false, "help for df")
			return nil
		},
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if err := Run(params); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "df: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func Run(params *Params) error {
	var fsInfos []FilesystemInfo

	if len(params.Paths) == 0 || (len(params.Paths) == 1 && params.Paths[0] == "") {
		// Get info for all mounted filesystems
		infos, err := getAllFilesystems(params)
		if err != nil {
			return err
		}
		fsInfos = infos
	} else {
		// Get info for specific paths
		for _, path := range params.Paths {
			info, err := getFilesystemInfo(path)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "df: cannot access '%s': %v\n", path, err)
				continue
			}
			fsInfos = append(fsInfos, info)
		}
	}

	if len(fsInfos) == 0 {
		return fmt.Errorf("no filesystems found")
	}

	// Sort the results
	sortFilesystems(fsInfos, params.Sort, params.Reverse)

	// Print output
	printOutput(fsInfos, params)

	return nil
}

// getAllFilesystems returns info for all mounted filesystems, applying filters
func getAllFilesystems(params *Params) ([]FilesystemInfo, error) {
	mounts, err := getMounts()
	if err != nil {
		return nil, fmt.Errorf("failed to get mounted filesystems: %v", err)
	}

	var infos []FilesystemInfo
	seen := make(map[string]bool) // Deduplicate by mount point

	for _, mount := range mounts {
		// Skip duplicates
		if seen[mount.MountPoint] {
			continue
		}

		// Filter: skip pseudo filesystems unless -a is specified
		if !params.All && isPseudoFilesystem(mount.FSType) {
			continue
		}

		// Filter: only local filesystems if -l is specified
		if params.Local && !isLocalFilesystem(mount.FSType) {
			continue
		}

		// Filter: specific filesystem type if -t is specified
		if params.Type != "" && mount.FSType != params.Type {
			continue
		}

		info, err := getFilesystemInfoForMount(mount)
		if err != nil {
			// Skip filesystems we can't stat (permission denied, etc.)
			continue
		}

		seen[mount.MountPoint] = true
		infos = append(infos, info)
	}

	return infos, nil
}

// getFilesystemInfo returns info for a specific path
func getFilesystemInfo(path string) (FilesystemInfo, error) {
	stat, err := getStatfs(path)
	if err != nil {
		return FilesystemInfo{}, err
	}

	return statToFilesystemInfo(stat, path, path, ""), nil
}

// getFilesystemInfoForMount returns info for a mounted filesystem
func getFilesystemInfoForMount(mount MountInfo) (FilesystemInfo, error) {
	stat, err := getStatfs(mount.MountPoint)
	if err != nil {
		return FilesystemInfo{}, err
	}

	return statToFilesystemInfo(stat, mount.Device, mount.MountPoint, mount.FSType), nil
}

// statToFilesystemInfo converts statfs result to FilesystemInfo
func statToFilesystemInfo(stat interface{}, device, mountPoint, fsType string) FilesystemInfo {
	var info FilesystemInfo
	info.Filesystem = device
	info.MountPoint = mountPoint
	info.FSType = fsType

	// Handle platform-specific stat types
	switch s := stat.(type) {
	case interface{ GetBlocks() (uint64, uint64, uint64, int64) }:
		blocks, bfree, bavail, bsize := s.GetBlocks()
		info.Size = blocks * uint64(bsize)
		info.Available = bavail * uint64(bsize)
		info.Used = info.Size - (bfree * uint64(bsize))
	default:
		// Use reflection or type assertion for the actual types
		info = extractStatInfo(stat, device, mountPoint, fsType)
	}

	if info.Size > 0 {
		info.Percent = float64(info.Used) / float64(info.Size) * 100
	}

	return info
}

func sortFilesystems(infos []FilesystemInfo, sortBy string, reverse bool) {
	slices.SortFunc(infos, func(a, b FilesystemInfo) int {
		var cmp int
		switch sortBy {
		case "used":
			cmp = int(int64(a.Used) - int64(b.Used))
		case "available":
			cmp = int(int64(a.Available) - int64(b.Available))
		case "percent":
			if a.Percent < b.Percent {
				cmp = -1
			} else if a.Percent > b.Percent {
				cmp = 1
			}
		default: // "name" or empty
			cmp = strings.Compare(a.Filesystem, b.Filesystem)
		}
		if reverse {
			cmp = -cmp
		}
		return cmp
	})
}

func printOutput(infos []FilesystemInfo, params *Params) {
	if params.Inode {
		printInodeOutput(infos)
	} else if params.Human {
		printHumanOutput(infos)
	} else {
		printBlockOutput(infos)
	}
}

func printInodeOutput(infos []FilesystemInfo) {
	fmt.Printf("%-30s %12s %12s %12s %5s %-20s\n",
		"Filesystem", "Inodes", "IUsed", "IFree", "IUse%", "Mounted on")
	fmt.Println(strings.Repeat("-", 95))

	for _, info := range infos {
		totalInodes := info.IAvailable + info.IUsed
		fmt.Printf("%-30s %12d %12d %12d %4.0f%% %-20s\n",
			truncate(info.Filesystem, 30),
			totalInodes,
			info.IUsed,
			info.IAvailable,
			info.IPercent,
			info.MountPoint)
	}
}

func printHumanOutput(infos []FilesystemInfo) {
	fmt.Printf("%-30s %8s %8s %8s %5s %-20s\n",
		"Filesystem", "Size", "Used", "Avail", "Use%", "Mounted on")
	fmt.Println(strings.Repeat("-", 85))

	for _, info := range infos {
		fmt.Printf("%-30s %8s %8s %8s %4.0f%% %-20s\n",
			truncate(info.Filesystem, 30),
			formatHumanReadable(info.Size),
			formatHumanReadable(info.Used),
			formatHumanReadable(info.Available),
			info.Percent,
			info.MountPoint)
	}
}

func printBlockOutput(infos []FilesystemInfo) {
	fmt.Printf("%-30s %12s %12s %12s %5s %-20s\n",
		"Filesystem", "1K-blocks", "Used", "Available", "Use%", "Mounted on")
	fmt.Println(strings.Repeat("-", 95))

	for _, info := range infos {
		fmt.Printf("%-30s %12d %12d %12d %4.0f%% %-20s\n",
			truncate(info.Filesystem, 30),
			info.Size/1024,
			info.Used/1024,
			info.Available/1024,
			info.Percent,
			info.MountPoint)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatHumanReadable(bytes uint64) string {
	units := []string{"B", "K", "M", "G", "T", "P"}
	value := float64(bytes)

	for _, unit := range units {
		if value < 1024 {
			if value < 10 {
				return fmt.Sprintf("%.1f%s", value, unit)
			}
			return fmt.Sprintf("%.0f%s", value, unit)
		}
		value /= 1024
	}

	return fmt.Sprintf("%.0fE", value)
}
