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
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

type Params struct {
	Paths      []string `pos:"true" optional:"true" help:"Paths to analyze. Defaults to current directory." default:"."`
	Summarize  bool     `short:"s" help:"Display only the total for each path." optional:"true"`
	All        bool     `short:"a" help:"Write counts for all files, not just directories." optional:"true"`
	Human      bool     `short:"h" help:"Print sizes in human readable format (B, KB, MB, GB, etc.)." optional:"true"`
	MaxDepth   int      `short:"d" help:"Print the total for a directory only if it is N or fewer levels deep." default:"-1"`
	Bytes      bool     `short:"b" help:"Print in bytes (default is 1024-byte blocks)." optional:"true"`
	Killobytes bool     `short:"k" help:"Print in kilobytes." optional:"true"`
	Sort       string   `short:"S" help:"Sort by: 'size' (largest last) or 'name'." default:"size"`
	Reverse    bool     `short:"r" help:"Reverse the sort order." optional:"true"`
	IgnoreGit  bool     `help:"Respect .gitignore files." optional:"true"`
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

	if params.Bytes {
		blockSize = 1
	} else if params.Killobytes {
		blockSize = 1024
	}

	for _, path := range params.Paths {

		rootNode, err := walkDir(path)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "du: error reading '%s': %v\n", path, err)
			continue
		}
		aggregateNodeSizesOnDisk(rootNode)
		pruneNodesToMaxDepth(rootNode, params.MaxDepth, 0)
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

func determineBlockSize(rootPath string) int64 {
	// Placeholder for more complex logic if needed
	st, err := os.Stat(rootPath)
	if err != nil {
		panic("Bug: cannot stat root path " + rootPath)
	}

	return st.Size()
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

func walkDir(rootPath string) (*DirNode, error) {

	nodeLkup := make(map[string]*DirNode)

	blockSize := determineBlockSize(rootPath)

	rootNode := &DirNode{
		Path:      rootPath,
		LevelSize: blockSize,
	}
	nodeLkup[rootPath] = rootNode

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "du: cannot access '%s': %v\n", path, err)
			return nil // Skip errors
		}

		if d.IsDir() {

			if path == rootPath {
				return nil // Skip root processing here
			}

			parentPath := filepath.Dir(path)
			parentNode, ok := nodeLkup[parentPath]
			if !ok {
				panic("Bug: parent not found for dir " + path)
			}

			currentNode, ok := nodeLkup[path]
			if !ok {
				currentNode = &DirNode{
					Path:      path,
					LevelSize: blockSize,
				}
				nodeLkup[path] = currentNode
				parentNode.ChildDirs = append(parentNode.ChildDirs, currentNode)
			}

		} else {

			parentPath := filepath.Dir(path)
			parentNode := nodeLkup[parentPath]
			if parentNode == nil {
				panic("Bug: parent not found for " + path)
			}
			fileInfo, err := d.Info()
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "du: cannot access info for '%s': %v\n", path, err)
				return nil
			}
			rawSize := fileInfo.Size()
			parentNode.LevelSize += roundToNearestBlockAbove(rawSize, blockSize)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory '%s': %v", rootPath, err)
	}

	return rootNode, nil
}

func roundToNearestBlockAbove(size int64, blockSize int64) int64 {
	if size == 0 {
		return blockSize
	}
	remainder := size % blockSize
	if remainder == 0 {
		return size
	}
	wholeParts := size / blockSize
	return (wholeParts + 1) * blockSize
}

func printNodes(node *DirNode, blockSize int64, human bool) {
	for _, child := range node.ChildDirs {
		printNodes(child, blockSize, human)
	}
	printSize(node.TotalSize, blockSize, human, node.Path)
}

func aggregateNodeSizesOnDisk(node *DirNode) {
	childSum := lo.SumBy(node.ChildDirs, func(n *DirNode) int64 {
		aggregateNodeSizesOnDisk(n)
		return n.TotalSize
	})
	node.TotalSize = node.LevelSize + childSum
}

/* old code

absPath, err := filepath.Abs(path)
if err != nil {
	fmt.Fprintf(os.Stderr, "du: cannot access '%s': %v\n", path, err)
	continue
}

info, err := os.Stat(absPath)
if err != nil {
	fmt.Fprintf(os.Stderr, "du: cannot access '%s': %v\n", path, err)
	continue
}

if info.IsDir() {
	if params.Summarize {
		size, err := calculateDirSize(absPath, params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "du: error reading '%s': %v\n", absPath, err)
			continue
		}
		printSize(size, blockSize, params.Human, path)
	} else {
		sizes, err := walkDir(absPath, params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "du: error reading '%s': %v\n", absPath, err)
			continue
		}
		printSizes(sizes, blockSize, params.Human, params.Sort, params.Reverse)
	}
} else {
	// Single file
	size := info.Size()
	printSize(size, blockSize, params.Human, path)
}
*/

/*
func calculateDirSize(dir string, params *Params) (int64, error) {
	var totalSize int64

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Calculate depth
		relPath, _ := filepath.Rel(dir, path)
		depth := strings.Count(relPath, string(os.PathSeparator))
		if params.MaxDepth != -1 && depth >= params.MaxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return nil
			}
			totalSize += info.Size()
		}

		return nil
	})

	return totalSize, err
}

func walkDir(dir string, params *Params) ([]DirSize, error) {
	sizeMap := make(map[string]int64)

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(dir, path)
		depth := strings.Count(relPath, string(os.PathSeparator))

		if params.MaxDepth != -1 && depth >= params.MaxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() || params.All {
			info, err := d.Info()
			if err != nil {
				return nil
			}

			sizeMap[path] = info.Size()
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Build aggregated sizes (each directory includes its contents)
	aggregated := make(map[string]int64)
	for path := range sizeMap {
		size, _ := calculateDirSize(path, &Params{MaxDepth: -1})
		aggregated[path] = size
	}

	var result []DirSize
	for path, size := range aggregated {
		result = append(result, DirSize{
			Path: path,
			Size: size,
		})
	}

	return result, nil
}


func printSizes(sizes []DirSize, blockSize int64, human bool, sortBy string, reverse bool) {
	// Sort the results
	if sortBy == "size" {
		sort.Slice(sizes, func(i, j int) bool {
			if reverse {
				return sizes[i].Size > sizes[j].Size
			}
			return sizes[i].Size < sizes[j].Size
		})
	} else if sortBy == "name" {
		sort.Slice(sizes, func(i, j int) bool {
			if reverse {
				return sizes[i].Path > sizes[j].Path
			}
			return sizes[i].Path < sizes[j].Path
		})
	}

	for _, ds := range sizes {
		printSize(ds.Size, blockSize, human, ds.Path)
	}
}

*/

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
