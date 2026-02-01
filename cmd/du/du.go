package du

import (
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"

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
	Kilobytes    bool     `short:"k" help:"Print in kilobytes." optional:"true"`
	Sort         string   `short:"S" help:"Sort by: 'size' (largest last), 'name', or 'none' (fastest, streams output)." default:"size" alts:"size,name,none"`
	Reverse      bool     `short:"r" help:"Reverse the sort order." optional:"true"`
	IgnoreGit    bool     `help:"Respect .gitignore files." optional:"true"`
}

type DirNode struct {
	Path       string
	LevelSize  int64
	ChildDirs  []*DirNode
	ChildFiles []FileNode // files at this level (only used when --all)
	TotalSize  int64      // calculated later
}

type FileNode struct {
	Path string
	Size int64
}

// Entry represents a flattened entry (file or directory) for global sorting
type Entry struct {
	Path string
	Size int64
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
	} else if params.Kilobytes {
		blockSize = 1024
	}

	// -s (summarize) is equivalent to -d 0
	maxDepth := params.MaxDepth
	if params.Summarize {
		maxDepth = 0
	}

	for _, path := range params.Paths {
		// Streaming mode: print as we go, no tree building
		if params.Sort == "none" {
			onFile := func(filePath string, depth int, size int64) {
				if maxDepth == -1 || depth <= maxDepth {
					printSize(size, blockSize, params.Human, filePath)
				}
			}
			onFinish := func(nodePath string, depth int, totalSize int64) {
				if maxDepth == -1 || depth <= maxDepth {
					printSize(totalSize, blockSize, params.Human, nodePath)
				}
			}
			var fileCallback func(string, int, int64)
			if params.All {
				fileCallback = onFile
			}
			_, err := walkDir(path, apparentSize, params.All, onFinish, fileCallback)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "du: error reading '%s': %v\n", path, err)
			}
			continue
		}

		// Tree mode: build tree, then print
		rootNode, err := walkDir(path, apparentSize, params.All, nil, nil)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "du: error reading '%s': %v\n", path, err)
			continue
		}
		pruneNodesToMaxDepth(rootNode, maxDepth, 0)

		if params.Sort == "size" {
			// For size sorting, flatten into a list for global sort
			entries := flattenTree(rootNode, params.All)
			sortEntries(entries, params.Sort, params.Reverse)
			for _, e := range entries {
				printSize(e.Size, blockSize, params.Human, e.Path)
			}
		} else {
			// For name sorting or no explicit sort, use hierarchical output
			printNodes(rootNode, blockSize, params.Human, params.All)
		}
	}

	return nil
}

// flattenTree converts the tree structure into a flat list of entries for global sorting
func flattenTree(node *DirNode, includeFiles bool) []Entry {
	var entries []Entry
	flattenTreeRecursive(node, includeFiles, &entries)
	return entries
}

func flattenTreeRecursive(node *DirNode, includeFiles bool, entries *[]Entry) {
	// Add files at this level
	if includeFiles {
		for _, f := range node.ChildFiles {
			*entries = append(*entries, Entry{Path: f.Path, Size: f.Size})
		}
	}

	// Recursively add child directories
	for _, child := range node.ChildDirs {
		flattenTreeRecursive(child, includeFiles, entries)
	}

	// Add this directory itself
	*entries = append(*entries, Entry{Path: node.Path, Size: node.TotalSize})
}

// sortEntries sorts a flat list of entries by the given criteria
func sortEntries(entries []Entry, sortBy string, reverse bool) {
	switch sortBy {
	case "size":
		slices.SortFunc(entries, func(i, j Entry) int {
			if reverse {
				return int(j.Size - i.Size)
			}
			return int(i.Size - j.Size)
		})
	case "name":
		slices.SortFunc(entries, func(i, j Entry) int {
			if reverse {
				return strings.Compare(j.Path, i.Path)
			}
			return strings.Compare(i.Path, j.Path)
		})
	}
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

// walkDir walks a directory tree and either builds a tree structure (when onFinish is nil)
// or streams output via the callback (when onFinish is provided).
// In streaming mode, onFinish is called with (path, depth, totalSize) for each directory.
// When all is true and onFile is provided (streaming), onFile is called for each file with depth.
// When all is true and onFile is nil (tree mode), files are stored in ChildFiles.
func walkDir(rootPath string, apparentSize bool, all bool, onFinish func(path string, depth int, totalSize int64), onFile func(path string, depth int, size int64)) (*DirNode, error) {
	// Normalize path to handle trailing slashes (e.g., "./" -> ".")
	rootPath = filepath.Clean(rootPath)
	streaming := onFinish != nil

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
	type stackEntry struct {
		node  *DirNode
		depth int
	}
	stack := []stackEntry{{node: rootNode, depth: 0}}

	// Helper to finalize a directory: calculate TotalSize and optionally stream output
	finalizeDir := func(entry stackEntry) {
		node := entry.node
		if streaming {
			// In streaming mode, child totals are accumulated in LevelSize
			node.TotalSize = node.LevelSize
		} else {
			// In tree mode, sum up children
			var childSum int64
			for _, child := range node.ChildDirs {
				childSum += child.TotalSize
			}
			node.TotalSize = node.LevelSize + childSum
		}

		if streaming {
			onFinish(node.Path, entry.depth, node.TotalSize)
		}
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
		for len(stack) > 0 && stack[len(stack)-1].node.Path != parentPath {
			finalizeDir(stack[len(stack)-1])
			finished := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			// Add finished total to parent's LevelSize (for streaming mode where we don't keep ChildDirs)
			if streaming && len(stack) > 0 {
				stack[len(stack)-1].node.LevelSize += finished.node.TotalSize
			}
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
			// Only build tree structure when not streaming
			if !streaming {
				parent.node.ChildDirs = append(parent.node.ChildDirs, node)
			}
			stack = append(stack, stackEntry{node: node, depth: parent.depth + 1})
		} else {
			fileInfo, err := d.Info()
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "du: cannot access info for '%s': %v\n", path, err)
				return nil
			}
			fileSize := getFileSize(fileInfo)
			parent.node.LevelSize += fileSize

			// Handle --all flag: track or stream file entries
			if all {
				if onFile != nil {
					// Streaming mode: print file immediately with parent's depth
					onFile(path, parent.depth, fileSize)
				} else {
					// Tree mode: store file for later printing
					parent.node.ChildFiles = append(parent.node.ChildFiles, FileNode{Path: path, Size: fileSize})
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory '%s': %v", rootPath, err)
	}

	// Finalize remaining directories on the stack
	for len(stack) > 0 {
		finalizeDir(stack[len(stack)-1])
		finished := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if streaming && len(stack) > 0 {
			stack[len(stack)-1].node.LevelSize += finished.node.TotalSize
		}
	}

	if streaming {
		return nil, nil
	}
	return rootNode, nil
}

func printNodes(node *DirNode, blockSize int64, human bool, all bool) {
	// Print files at this level first (if --all)
	if all {
		for _, f := range node.ChildFiles {
			printSize(f.Size, blockSize, human, f.Path)
		}
	}

	// Recursively print child directories
	for _, child := range node.ChildDirs {
		printNodes(child, blockSize, human, all)
	}

	// Print this directory's total last
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
			// Match GNU du format: show decimal for values >= 10, or always for M and above
			if value < 10 && unit != "B" {
				return fmt.Sprintf("%.1f%s", value, unit)
			}
			return fmt.Sprintf("%.0f%s", math.Ceil(value), unit)
		}
		value /= 1024
	}

	return fmt.Sprintf("%.0fE", value)
}
