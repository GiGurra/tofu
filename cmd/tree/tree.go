package tree

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Dir     string   `pos:"true" optional:"true" help:"Directory to start the tree from." default:"."`
	Depth   int      `short:"L" help:"Descend only level directories deep." default:"-1"` // -1 means infinite depth
	All     bool     `short:"a" help:"Do not ignore entries starting with ." default:"false"`
	Exclude []string `help:"Exclude files matching the pattern." default:"[]"`
}

type counters struct {
	dirs  int
	files int
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "tree",
		Short:       "List contents of directories in a tree-like format",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if err := Run(params); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "tree: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func Run(params *Params) error {
	absDir, err := filepath.Abs(params.Dir)
	if err != nil {
		return fmt.Errorf("failed to resolve directory %s: %w", params.Dir, err)
	}

	info, err := os.Stat(absDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", absDir)
	}
	if err != nil {
		return fmt.Errorf("failed to stat directory %s: %w", absDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", absDir)
	}

	// Print root directory
	fmt.Println(params.Dir)

	c := &counters{dirs: 1, files: 0}
	printTree(absDir, "", 1, params, c)

	fmt.Printf("\n%d directories, %d files\n", c.dirs, c.files)
	return nil
}

// printTree recursively prints directory contents in tree format.
// prefix is the indentation string for the current level.
// depth is the current depth (1-based, root children are depth 1).
func printTree(dirPath string, prefix string, depth int, params *Params, c *counters) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: cannot read directory %s: %v\n", dirPath, err)
		return
	}

	// Filter entries according to exclusion rules
	filtered := filterEntries(entries, dirPath, params)

	for i, entry := range filtered {
		isLast := i == len(filtered)-1

		// Choose connector based on whether this is the last entry
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		fmt.Printf("%s%s%s\n", prefix, connector, entry.Name())

		if entry.IsDir() {
			c.dirs++

			// Recurse into subdirectory if within depth limit
			if params.Depth == -1 || depth < params.Depth {
				// Extend prefix: use "│   " if more siblings follow, "    " if last
				childPrefix := prefix
				if isLast {
					childPrefix += "    "
				} else {
					childPrefix += "│   "
				}
				printTree(filepath.Join(dirPath, entry.Name()), childPrefix, depth+1, params, c)
			}
		} else {
			c.files++
		}
	}
}

// filterEntries filters directory entries based on exclusion rules.
func filterEntries(entries []fs.DirEntry, dirPath string, params *Params) []fs.DirEntry {
	var filtered []fs.DirEntry
	for _, entry := range entries {
		if !isExcluded(entry.Name(), dirPath, entry.IsDir(), params) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// isExcluded checks if an entry should be excluded based on params.
func isExcluded(name string, dirPath string, isDir bool, params *Params) bool {
	// Hidden files (starting with .) unless -a is used
	if !params.All && strings.HasPrefix(name, ".") {
		return true
	}

	// Check exclusion patterns
	for _, pattern := range params.Exclude {
		// Try matching just the name
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
		// Try matching the full path for patterns with path separators
		if strings.Contains(pattern, string(os.PathSeparator)) || strings.Contains(pattern, "/") {
			fullPath := filepath.Join(dirPath, name)
			if matched, _ := filepath.Match(pattern, fullPath); matched {
				return true
			}
		}
	}

	return false
}
