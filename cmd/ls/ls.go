package ls

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Paths          []string `pos:"true" optional:"true" help:"Files or directories to list." default:"."`
	All            bool     `short:"a" help:"Do not ignore entries starting with ."`
	AlmostAll      bool     `short:"A" help:"Do not list implied . and .."`
	Long           bool     `short:"l" help:"Use a long listing format."`
	HumanReadable  bool     `short:"h" help:"With -l, print sizes like 1K 234M 2G etc."`
	OnePerLine     bool     `short:"1" help:"List one file per line."`
	Reverse        bool     `short:"r" help:"Reverse order while sorting."`
	SortByTime     bool     `short:"t" help:"Sort by time, newest first."`
	SortBySize     bool     `short:"S" help:"Sort by file size, largest first."`
	NoSort         bool     `short:"U" help:"Do not sort; list entries in directory order."`
	Classify       bool     `short:"F" help:"Append indicator (one of */=>@|) to entries."`
	Directory      bool     `short:"d" help:"List directories themselves, not their contents."`
	Recursive      bool     `short:"R" help:"List subdirectories recursively."`
	Inode          bool     `short:"i" help:"Print the index number of each file."`
	Size           bool     `short:"s" help:"Print the allocated size of each file, in blocks."`
	Color          string   `help:"Colorize the output: 'always', 'auto', or 'never'." default:"auto" alts:"always,auto,never"`
	GroupDirsFirst bool     `help:"Group directories before files."`
	NoGroup        bool     `short:"G" help:"In a long listing, don't print group names."`
	NumericUidGid  bool     `short:"n" help:"Like -l, but list numeric user and group IDs."`
	FullGroup      bool     `help:"Show full group identifier (e.g., Windows SID)."`
}

type fileEntry struct {
	name    string
	info    fs.FileInfo
	path    string // full path for recursive
	linkDst string // symlink destination
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "ls",
		Short:       "List directory contents",
		Long:        "List information about the FILEs (the current directory by default).",
		ParamEnrich: common.DefaultParamEnricher(),
		InitFunc: func(params *Params, cmd *cobra.Command) error {
			cmd.Flags().BoolP("help", "", false, "help for ls")
			return nil
		},
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			exitCode := Run(params, os.Stdout, os.Stderr)
			os.Exit(exitCode)
		},
	}.ToCobra()
}

func LlCmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "ll",
		Short:       "List directory contents in long format (alias for ls -l)",
		Long:        "List information about the FILEs in long format (alias for ls -l).",
		ParamEnrich: common.DefaultParamEnricher(),
		InitFunc: func(params *Params, cmd *cobra.Command) error {
			cmd.Flags().BoolP("help", "", false, "help for ll")
			return nil
		},
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			params.Long = true
			exitCode := Run(params, os.Stdout, os.Stderr)
			os.Exit(exitCode)
		},
	}.ToCobra()
}

func LaCmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "la",
		Short:       "List all directory contents in long format (alias for ls -la)",
		Long:        "List information about all FILEs including hidden files in long format (alias for ls -la).",
		ParamEnrich: common.DefaultParamEnricher(),
		InitFunc: func(params *Params, cmd *cobra.Command) error {
			cmd.Flags().BoolP("help", "", false, "help for la")
			return nil
		},
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			params.Long = true
			params.All = true
			exitCode := Run(params, os.Stdout, os.Stderr)
			os.Exit(exitCode)
		},
	}.ToCobra()
}

func Run(params *Params, stdout, stderr io.Writer) int {
	// -n implies -l
	if params.NumericUidGid {
		params.Long = true
	}

	// Determine if we should use color
	useColor := shouldUseColor(params.Color, stdout)

	hadError := false
	paths := params.Paths
	if len(paths) == 0 {
		paths = []string{"."}
	}

	// Separate files and directories
	var files []fileEntry
	var dirs []string

	for _, path := range paths {
		info, err := os.Lstat(path)
		if err != nil {
			fmt.Fprintf(stderr, "ls: cannot access '%s': %v\n", path, err)
			hadError = true
			continue
		}

		if info.IsDir() && !params.Directory {
			dirs = append(dirs, path)
		} else {
			linkDst := ""
			if info.Mode()&os.ModeSymlink != 0 {
				linkDst, _ = os.Readlink(path)
			}
			files = append(files, fileEntry{name: filepath.Base(path), info: info, path: path, linkDst: linkDst})
		}
	}

	// Print files first
	if len(files) > 0 {
		sortEntries(files, params)
		printEntries(files, params, stdout, useColor, "")
	}

	// Print directories
	multipleTargets := len(files) > 0 || len(dirs) > 1
	for i, dir := range dirs {
		if multipleTargets {
			if i > 0 || len(files) > 0 {
				fmt.Fprintln(stdout)
			}
			fmt.Fprintf(stdout, "%s:\n", dir)
		}

		if err := listDirectory(dir, params, stdout, stderr, useColor, ""); err != nil {
			fmt.Fprintf(stderr, "ls: %v\n", err)
			hadError = true
		}
	}

	if hadError {
		return 1
	}
	return 0
}

func listDirectory(dir string, params *Params, stdout, stderr io.Writer, useColor bool, prefix string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var fileEntries []fileEntry
	for _, entry := range entries {
		name := entry.Name()

		// Filter hidden files
		if !params.All && !params.AlmostAll && strings.HasPrefix(name, ".") {
			continue
		}
		if params.AlmostAll && (name == "." || name == "..") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			fmt.Fprintf(stderr, "ls: cannot access '%s': %v\n", filepath.Join(dir, name), err)
			continue
		}

		linkDst := ""
		if info.Mode()&os.ModeSymlink != 0 {
			linkDst, _ = os.Readlink(filepath.Join(dir, name))
		}

		fileEntries = append(fileEntries, fileEntry{
			name:    name,
			info:    info,
			path:    filepath.Join(dir, name),
			linkDst: linkDst,
		})
	}

	// Add . and .. if -a is specified
	if params.All {
		if dotInfo, err := os.Lstat(dir); err == nil {
			fileEntries = append([]fileEntry{{name: ".", info: dotInfo, path: dir}}, fileEntries...)
		}
		if parentInfo, err := os.Lstat(filepath.Join(dir, "..")); err == nil {
			// Insert after .
			if len(fileEntries) > 0 {
				fileEntries = append(fileEntries[:1], append([]fileEntry{{name: "..", info: parentInfo, path: filepath.Join(dir, "..")}}, fileEntries[1:]...)...)
			}
		}
	}

	sortEntries(fileEntries, params)
	printEntries(fileEntries, params, stdout, useColor, prefix)

	// Handle recursive listing
	if params.Recursive {
		for _, entry := range fileEntries {
			if entry.info.IsDir() && entry.name != "." && entry.name != ".." {
				fmt.Fprintln(stdout)
				fmt.Fprintf(stdout, "%s:\n", entry.path)
				if err := listDirectory(entry.path, params, stdout, stderr, useColor, ""); err != nil {
					fmt.Fprintf(stderr, "ls: cannot open directory '%s': %v\n", entry.path, err)
				}
			}
		}
	}

	return nil
}

func sortEntries(entries []fileEntry, params *Params) {
	if params.NoSort {
		return
	}

	slices.SortFunc(entries, func(a, b fileEntry) int {
		// Group directories first if requested
		if params.GroupDirsFirst {
			aIsDir := a.info.IsDir()
			bIsDir := b.info.IsDir()
			if aIsDir && !bIsDir {
				return -1
			}
			if !aIsDir && bIsDir {
				return 1
			}
		}

		var cmp int
		if params.SortByTime {
			cmp = b.info.ModTime().Compare(a.info.ModTime()) // newest first
		} else if params.SortBySize {
			cmp = int(b.info.Size() - a.info.Size()) // largest first
		} else {
			cmp = strings.Compare(strings.ToLower(a.name), strings.ToLower(b.name))
		}

		if params.Reverse {
			cmp = -cmp
		}
		return cmp
	})
}

func printEntries(entries []fileEntry, params *Params, stdout io.Writer, useColor bool, prefix string) {
	if params.Long {
		printLongFormat(entries, params, stdout, useColor)
	} else if params.OnePerLine {
		for _, entry := range entries {
			printName(entry, params, stdout, useColor)
			fmt.Fprintln(stdout)
		}
	} else {
		// Simple column format (one per line for now, can enhance later)
		for _, entry := range entries {
			printName(entry, params, stdout, useColor)
			fmt.Fprintln(stdout)
		}
	}
}

func printLongFormat(entries []fileEntry, params *Params, stdout io.Writer, useColor bool) {
	// Calculate column widths
	var maxLinks, maxOwner, maxGroup, maxSize, maxInode, maxBlocks int
	for _, entry := range entries {
		stat := getFileStatInfo(entry.info)
		if stat.Valid {
			linkStr := strconv.FormatUint(stat.Nlink, 10)
			if len(linkStr) > maxLinks {
				maxLinks = len(linkStr)
			}

			owner := getOwner(stat, params.NumericUidGid)
			if len(owner) > maxOwner {
				maxOwner = len(owner)
			}

			if !params.NoGroup {
				group := getGroup(stat, params.NumericUidGid, params.FullGroup)
				if len(group) > maxGroup {
					maxGroup = len(group)
				}
			}

			if params.Inode {
				inodeStr := strconv.FormatUint(stat.Inode, 10)
				if len(inodeStr) > maxInode {
					maxInode = len(inodeStr)
				}
			}

			if params.Size {
				blocks := (stat.Blocks + 1) / 2 // Convert 512-byte blocks to 1K blocks
				blockStr := strconv.FormatInt(blocks, 10)
				if len(blockStr) > maxBlocks {
					maxBlocks = len(blockStr)
				}
			}
		}

		sizeStr := formatSize(entry.info.Size(), params.HumanReadable)
		if len(sizeStr) > maxSize {
			maxSize = len(sizeStr)
		}
	}

	for _, entry := range entries {
		var line strings.Builder

		stat := getFileStatInfo(entry.info)
		if stat.Valid {
			if params.Inode {
				fmt.Fprintf(&line, "%*d ", maxInode, stat.Inode)
			}

			if params.Size {
				blocks := (stat.Blocks + 1) / 2
				fmt.Fprintf(&line, "%*d ", maxBlocks, blocks)
			}

			// Mode
			line.WriteString(modeString(entry.info.Mode()))
			line.WriteString(" ")

			// Links
			fmt.Fprintf(&line, "%*d ", maxLinks, stat.Nlink)

			// Owner
			owner := getOwner(stat, params.NumericUidGid)
			fmt.Fprintf(&line, "%-*s ", maxOwner, owner)

			// Group
			if !params.NoGroup {
				group := getGroup(stat, params.NumericUidGid, params.FullGroup)
				fmt.Fprintf(&line, "%-*s ", maxGroup, group)
			}
		} else {
			line.WriteString(modeString(entry.info.Mode()))
			line.WriteString(" ")
		}

		// Size
		sizeStr := formatSize(entry.info.Size(), params.HumanReadable)
		fmt.Fprintf(&line, "%*s ", maxSize, sizeStr)

		// Time
		line.WriteString(formatTime(entry.info.ModTime()))
		line.WriteString(" ")

		fmt.Fprint(stdout, line.String())
		printName(entry, params, stdout, useColor)

		// Symlink destination
		if entry.linkDst != "" {
			fmt.Fprintf(stdout, " -> %s", entry.linkDst)
		}

		fmt.Fprintln(stdout)
	}
}

func printName(entry fileEntry, params *Params, stdout io.Writer, useColor bool) {
	name := entry.name

	if useColor {
		name = colorize(name, entry.info)
	}

	fmt.Fprint(stdout, name)

	if params.Classify {
		fmt.Fprint(stdout, classifyChar(entry.name, entry.info))
	}
}

func classifyChar(name string, info fs.FileInfo) string {
	mode := info.Mode()
	switch {
	case mode.IsDir():
		return "/"
	case mode&os.ModeSymlink != 0:
		return "@"
	case mode&os.ModeNamedPipe != 0:
		return "|"
	case mode&os.ModeSocket != 0:
		return "="
	case isExecutable(name, mode):
		return "*"
	default:
		return ""
	}
}

func modeString(mode fs.FileMode) string {
	var buf [10]byte

	// File type
	switch {
	case mode.IsDir():
		buf[0] = 'd'
	case mode&os.ModeSymlink != 0:
		buf[0] = 'l'
	case mode&os.ModeNamedPipe != 0:
		buf[0] = 'p'
	case mode&os.ModeSocket != 0:
		buf[0] = 's'
	case mode&os.ModeDevice != 0:
		if mode&os.ModeCharDevice != 0 {
			buf[0] = 'c'
		} else {
			buf[0] = 'b'
		}
	default:
		buf[0] = '-'
	}

	// Owner permissions
	perm := mode.Perm()
	buf[1] = permChar(perm&0400 != 0, 'r')
	buf[2] = permChar(perm&0200 != 0, 'w')
	if mode&os.ModeSetuid != 0 {
		if perm&0100 != 0 {
			buf[3] = 's'
		} else {
			buf[3] = 'S'
		}
	} else {
		buf[3] = permChar(perm&0100 != 0, 'x')
	}

	// Group permissions
	buf[4] = permChar(perm&0040 != 0, 'r')
	buf[5] = permChar(perm&0020 != 0, 'w')
	if mode&os.ModeSetgid != 0 {
		if perm&0010 != 0 {
			buf[6] = 's'
		} else {
			buf[6] = 'S'
		}
	} else {
		buf[6] = permChar(perm&0010 != 0, 'x')
	}

	// Other permissions
	buf[7] = permChar(perm&0004 != 0, 'r')
	buf[8] = permChar(perm&0002 != 0, 'w')
	if mode&os.ModeSticky != 0 {
		if perm&0001 != 0 {
			buf[9] = 't'
		} else {
			buf[9] = 'T'
		}
	} else {
		buf[9] = permChar(perm&0001 != 0, 'x')
	}

	return string(buf[:])
}

func permChar(set bool, c byte) byte {
	if set {
		return c
	}
	return '-'
}

func formatSize(size int64, human bool) string {
	if !human {
		return strconv.FormatInt(size, 10)
	}

	units := []string{"", "K", "M", "G", "T", "P"}
	value := float64(size)

	for _, unit := range units {
		if value < 1024 {
			if unit == "" {
				return fmt.Sprintf("%d", int(value))
			}
			if value < 10 {
				return fmt.Sprintf("%.1f%s", value, unit)
			}
			return fmt.Sprintf("%.0f%s", value, unit)
		}
		value /= 1024
	}
	return fmt.Sprintf("%.0fE", value)
}

func formatTime(t time.Time) string {
	now := time.Now()
	sixMonthsAgo := now.AddDate(0, -6, 0)

	if t.Before(sixMonthsAgo) || t.After(now) {
		return t.Format("Jan _2  2006")
	}
	return t.Format("Jan _2 15:04")
}

func shouldUseColor(colorOpt string, stdout io.Writer) bool {
	switch colorOpt {
	case "always":
		return true
	case "never":
		return false
	default: // "auto"
		if f, ok := stdout.(*os.File); ok {
			stat, _ := f.Stat()
			return (stat.Mode() & os.ModeCharDevice) != 0
		}
		return false
	}
}

func colorize(name string, info fs.FileInfo) string {
	mode := info.Mode()

	var colorCode string
	switch {
	case mode.IsDir():
		colorCode = "\033[1;34m" // Bold blue
	case mode&os.ModeSymlink != 0:
		colorCode = "\033[1;36m" // Bold cyan
	case mode&os.ModeNamedPipe != 0:
		colorCode = "\033[33m" // Yellow
	case mode&os.ModeSocket != 0:
		colorCode = "\033[1;35m" // Bold magenta
	case isExecutable(name, mode):
		colorCode = "\033[1;32m" // Bold green
	default:
		return name // No color for regular files
	}

	return colorCode + name + "\033[0m"
}
