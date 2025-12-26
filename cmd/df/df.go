package df

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

type Params struct {
	Paths   []string `pos:"true" optional:"true" help:"Paths to analyze. Defaults to all mounted filesystems." default:""`
	All     bool     `short:"a" help:"Include all filesystems, including pseudo filesystems." optional:"true"`
	Human   bool     `short:"h" help:"Print sizes in human readable format." optional:"true"`
	Inode   bool     `short:"i" help:"List inode information instead of block usage." optional:"true"`
	Local   bool     `short:"l" help:"Limit listing to local filesystems." optional:"true"`
	Type    string   `short:"t" help:"List only filesystems of a specific type." default:""`
	Sort    string   `short:"S" help:"Sort by: 'used', 'available', 'percent' (default: filesystem)." default:""`
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
				fmt.Fprintf(os.Stderr, "df: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func Run(params *Params) error {
	var statfsInfos []FilesystemInfo

	if len(params.Paths) == 0 {
		// Get info for all mounted filesystems
		infos, err := getFilesystemsInfo(params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "df: %v\n", err)
			return err
		}
		statfsInfos = infos
	} else {
		// Get info for specific paths
		for _, path := range params.Paths {
			info, err := getPathInfo(path, params)
			if err != nil {
				fmt.Fprintf(os.Stderr, "df: cannot access '%s': %v\n", path, err)
				continue
			}
			statfsInfos = append(statfsInfos, info)
		}
	}

	if len(statfsInfos) == 0 {
		return fmt.Errorf("no filesystems found")
	}

	// Sort the results
	sortFilesystems(statfsInfos, params.Sort, params.Reverse)

	// Print header
	if params.Inode {
		fmt.Printf("%-30s %10s %10s %10s %4s %-20s\n",
			"Filesystem", "Inodes", "IUsed", "IFree", "IUse%", "Mounted on")
		fmt.Println(strings.Repeat("-", 85))

		for _, info := range statfsInfos {
			fmt.Printf("%-30s %10d %10d %10d %3.0f%% %-20s\n",
				info.Filesystem,
				info.IAvailable+info.IUsed,
				info.IUsed,
				info.IAvailable,
				info.IPercent,
				info.MountPoint)
		}
	} else {
		if params.Human {
			fmt.Printf("%-30s %8s %8s %8s %4s %-20s\n",
				"Filesystem", "Size", "Used", "Avail", "Use%", "Mounted on")
			fmt.Println(strings.Repeat("-", 85))

			for _, info := range statfsInfos {
				fmt.Printf("%-30s %8s %8s %8s %3.0f%% %-20s\n",
					info.Filesystem,
					formatHumanReadable(info.Size),
					formatHumanReadable(info.Used),
					formatHumanReadable(info.Available),
					info.Percent,
					info.MountPoint)
			}
		} else {
			// Print in 1K blocks (like traditional df)
			fmt.Printf("%-30s %10s %10s %10s %4s %-20s\n",
				"Filesystem", "1K-blocks", "Used", "Available", "Use%", "Mounted on")
			fmt.Println(strings.Repeat("-", 95))

			for _, info := range statfsInfos {
				size := info.Size / 1024
				used := info.Used / 1024
				avail := info.Available / 1024

				fmt.Printf("%-30s %10d %10d %10d %3.0f%% %-20s\n",
					info.Filesystem,
					size,
					used,
					avail,
					info.Percent,
					info.MountPoint)
			}
		}
	}

	return nil
}

func getFilesystemsInfo(params *Params) ([]FilesystemInfo, error) {
	// This is a placeholder implementation
	// On Unix systems, we should parse /proc/mounts or similar
	// For now, just get info for root
	var infos []FilesystemInfo

	if err := os.Chdir("/"); err == nil {
		if info, err := getPathInfo("/", params); err == nil {
			infos = append(infos, info)
		}
	}

	return infos, nil
}

func getPathInfo(path string, params *Params) (FilesystemInfo, error) {
	var statBuf unix.Statfs_t
	err := unix.Statfs(path, &statBuf)
	if err != nil {
		return FilesystemInfo{}, err
	}

	info := FilesystemInfo{
		Filesystem: path,
		Size:       statBuf.Blocks * uint64(statBuf.Bsize),
		Available:  statBuf.Bavail * uint64(statBuf.Bsize),
		MountPoint: path,
	}

	info.Used = info.Size - (statBuf.Bfree * uint64(statBuf.Bsize))

	if info.Size > 0 {
		info.Percent = float64(info.Used) / float64(info.Size) * 100
	}

	// Inode info
	total := statBuf.Files
	free := statBuf.Ffree
	info.IAvailable = free
	info.IUsed = total - free

	if total > 0 {
		info.IPercent = float64(info.IUsed) / float64(total) * 100
	}

	return info, nil
}

func sortFilesystems(infos []FilesystemInfo, sortBy string, reverse bool) {
	switch sortBy {
	case "used":
		sort.Slice(infos, func(i, j int) bool {
			if reverse {
				return infos[i].Used < infos[j].Used
			}
			return infos[i].Used > infos[j].Used
		})
	case "available":
		sort.Slice(infos, func(i, j int) bool {
			if reverse {
				return infos[i].Available < infos[j].Available
			}
			return infos[i].Available > infos[j].Available
		})
	case "percent":
		sort.Slice(infos, func(i, j int) bool {
			if reverse {
				return infos[i].Percent < infos[j].Percent
			}
			return infos[i].Percent > infos[j].Percent
		})
	default:
		// Sort by filesystem name
		sort.Slice(infos, func(i, j int) bool {
			if reverse {
				return infos[i].Filesystem > infos[j].Filesystem
			}
			return infos[i].Filesystem < infos[j].Filesystem
		})
	}
}

func formatHumanReadable(bytes uint64) string {
	units := []string{"B", "K", "M", "G", "T", "P"}
	value := float64(bytes)

	for _, unit := range units {
		if value < 1024 {
			return fmt.Sprintf("%.0f%s", value, unit)
		}
		value /= 1024
	}

	return fmt.Sprintf("%.0fE", value)
}
