package du

import (
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Paths        []string `pos:"true" optional:"true" help:"Paths to analyze. Defaults to current directory." default:"."`
	Summarize    bool     `short:"s" help:"Display only the total for each path." optional:"true"`
	All          bool     `short:"a" help:"Write counts for all files, not just directories." optional:"true"`
	Human        bool     `short:"h" help:"Print sizes in human readable format (B, KB, MB, GB, etc.)." optional:"true"`
	MaxDepth     int      `short:"d" help:"Print the total for a directory only if it is N or fewer levels deep." default:"-1"`
	Bytes        bool     `short:"b" help:"Apparent size in bytes (equivalent to --apparent-size --block-size=1)." optional:"true"`
	ApparentSize bool     `help:"Print apparent sizes rather than disk usage." optional:"true"`
	Killobytes   bool     `short:"k" help:"Print in kilobytes." optional:"true"`
	Sort         string   `short:"S" help:"Sort by: 'size' (largest last) or 'name'." default:"size"`
	Reverse      bool     `short:"r" help:"Reverse the sort order." optional:"true"`
	IgnoreGit    bool     `help:"Respect .gitignore files." optional:"true"`
}

type DirNode struct {
	Path      string
	LevelSize int64
	ChildDirs []*DirNode
	TotalSize int64 // calculated later
}

func Cmd() *cobra.Command {
	cmd := boa.CmdT[Params]{
		Use:         "du",
		Short:       "Estimate file and directory space usage",
		Long:        "Estimate file and directory space usage, like the Unix du command but cross-platform.",
		ParamEnrich: common.DefaultParamEnricher(),
		InitFunc: func(params *Params, cmd *cobra.Command) error {
			cmd.Flags().BoolP("help", "", false, "help for du")
			return nil
		},
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if err := Run(params); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "du: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
	return cmd
}

func Run(params *Params) error {
	blockSize := int64(1024) // Default 1024-byte blocks
	apparentSize := params.ApparentSize

	if params.Bytes {
		blockSize = 1
		apparentSize = true // -b implies --apparent-size
	} else if params.Killobytes {
		blockSize = 1024
	}

	// -s (summarize) is equivalent to -d 0
	maxDepth := params.MaxDepth
	if params.Summarize {
		maxDepth = 0
	}

	for _, path := range params.Paths {

		rootNode, err := walkDir(path, apparentSize)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "du: error reading '%s': %v\n", path, err)
			continue
		}
		pruneNodesToMaxDepth(rootNode, maxDepth, 0)
		sortNodes(rootNode, params.Sort, params.Reverse)
		printNodes(rootNode, blockSize, params.Human)

	}

	return nil
}

func sortNodes(node *DirNode, sortBy string, reverse bool) {
	// Sort the results
	if sortBy == "size" {

		slices.SortFunc(node.ChildDirs, func(i, j *DirNode) int {
			if reverse {
				return int(j.TotalSize - i.TotalSize)
			}
			return int(i.TotalSize - j.TotalSize)
		})

		for _, child := range node.ChildDirs {
			sortNodes(child, sortBy, reverse)
		}

	} else if sortBy == "name" {
		slices.SortFunc(node.ChildDirs, func(i, j *DirNode) int {
			if reverse {
				return strings.Compare(j.Path, i.Path)
			}

			return strings.Compare(i.Path, j.Path)
		})

		for _, child := range node.ChildDirs {
			sortNodes(child, sortBy, reverse)
		}
	}

}

// getDiskUsage returns actual disk usage in bytes using syscall.Stat_t.Blocks
func getDiskUsage(info fs.FileInfo) int64 {
	if sys, ok := info.Sys().(*syscall.Stat_t); ok {
		// Blocks are in 512-byte units
		return sys.Blocks * 512
	}
	// Fallback for systems without Stat_t: round up to 4096-byte blocks
	if info.Size() == 0 {
		return 0
	}
	return ((info.Size() + 4095) / 4096) * 4096
}

func pruneNodesToMaxDepth(node *DirNode, maxDepth int, currentDepth int) {
	if maxDepth != -1 && currentDepth >= maxDepth {
		node.ChildDirs = nil
		return
	}
	for _, child := range node.ChildDirs {
		pruneNodesToMaxDepth(child, maxDepth, currentDepth+1)
	}
}

func walkDir(rootPath string, apparentSize bool) (*DirNode, error) {

	// Helper to get the right size based on mode
	getFileSize := func(info fs.FileInfo) int64 {
		if apparentSize {
			return info.Size()
		}
		return getDiskUsage(info)
	}

	// Helper for directory size - in apparent size mode, directories don't count
	getDirSize := func(info fs.FileInfo) int64 {
		if apparentSize {
			return 0 // du --apparent-size doesn't count directory entries
		}
		return getDiskUsage(info)
	}

	// Get disk usage for the root directory itself
	rootInfo, err := os.Stat(rootPath)
	if err != nil {
		return nil, fmt.Errorf("cannot stat root path '%s': %v", rootPath, err)
	}

	rootNode := &DirNode{
		Path:      rootPath,
		LevelSize: getDirSize(rootInfo),
	}

	// Stack-based approach: since WalkDir is depth-first, we use a stack
	// that mirrors the traversal. Push on dir entry, pop when we leave.
	stack := []*DirNode{rootNode}

	// Helper to finalize a directory: calculate TotalSize from LevelSize + children
	finalizeDir := func(node *DirNode) {
		var childSum int64
		for _, child := range node.ChildDirs {
			childSum += child.TotalSize
		}
		node.TotalSize = node.LevelSize + childSum
	}

	err = filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "du: cannot access '%s': %v\n", path, err)
			return nil // Skip errors
		}

		if path == rootPath {
			return nil // Root already on stack
		}

		parentPath := filepath.Dir(path)

		// Pop finished directories until stack top is our parent
		for len(stack) > 0 && stack[len(stack)-1].Path != parentPath {
			finalizeDir(stack[len(stack)-1])
			stack = stack[:len(stack)-1]
		}

		if len(stack) == 0 {
			panic("Bug: stack empty, parent not found for " + path)
		}

		parent := stack[len(stack)-1]

		if d.IsDir() {
			dirInfo, err := d.Info()
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "du: cannot access info for '%s': %v\n", path, err)
				return nil
			}

			node := &DirNode{
				Path:      path,
				LevelSize: getDirSize(dirInfo),
			}
			parent.ChildDirs = append(parent.ChildDirs, node)
			stack = append(stack, node)
		} else {
			fileInfo, err := d.Info()
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "du: cannot access info for '%s': %v\n", path, err)
				return nil
			}
			parent.LevelSize += getFileSize(fileInfo)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory '%s': %v", rootPath, err)
	}

	// Finalize remaining directories on the stack
	for len(stack) > 0 {
		finalizeDir(stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return rootNode, nil
}

func printNodes(node *DirNode, blockSize int64, human bool) {
	for _, child := range node.ChildDirs {
		printNodes(child, blockSize, human)
	}
	printSize(node.TotalSize, blockSize, human, node.Path)
}

func printSize(size int64, blockSize int64, human bool, path string) {
	if human {
		fmt.Printf("%s\t%s\n", formatHumanReadable(size), path)
	} else {
		blocks := (size + blockSize - 1) / blockSize // Round up
		fmt.Printf("%d\t%s\n", blocks, path)
	}
}

func formatHumanReadable(bytes int64) string {
	units := []string{"B", "K", "M", "G", "T", "P"}
	value := float64(bytes)

	for _, unit := range units {
		if value < 1024 {
			return fmt.Sprintf("%.0f%s", math.Ceil(value), unit)
		}
		value /= 1024
	}

	return fmt.Sprintf("%.0fE", value)
}
