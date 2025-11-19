package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

type TreeParams struct {
	Dir             string   `pos:"true" optional:"true" help:"Directory to start the tree from." default:"."`
	Depth           int      `short:"L" help:"Descend only level directories deep." default:"-1"` // -1 means infinite depth
	All             bool     `short:"a" help:"Do not ignore entries starting with ." default:"false"`
	IgnoreGitignore bool     `help:".gitignore" default:"false"`
	Exclude         []string `help:"Exclude files matching the pattern." default:"[]"`
}

func TreeCmd() *cobra.Command {
	return boa.CmdT[TreeParams]{
		Use:         "tree",
		Short:       "List contents of directories in a tree-like format",
		ParamEnrich: defaultParamEnricher(),
		RunFunc: func(params *TreeParams, cmd *cobra.Command, args []string) {
			if err := runTree(params); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "tree: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runTree(params *TreeParams) error {
	absDir, err := filepath.Abs(params.Dir)
	if err != nil {
		return fmt.Errorf("failed to resolve directory %s: %w", params.Dir, err)
	}

	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", absDir)
	}

	// Simple exclusion logic (can be expanded to gitignore later)
	isExcluded := func(path string, isDir bool) bool {
		// Handle hidden files unless -a is used
		if !params.All && strings.HasPrefix(filepath.Base(path), ".") && filepath.Base(path) != "." {
			return true
		}

		// Basic glob exclusion (can be improved with proper gitignore parsing)
		for _, pattern := range params.Exclude {
			matched, _ := filepath.Match(pattern, filepath.Base(path))
			if matched {
				return true
			}
		}
		return false
	}

	nDirs := 1
	nFiles := 0

	listingAtDir := map[string][]os.DirEntry{}
	err = filepath.WalkDir(absDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", path, err)
			return nil // Skip errors for now
		}

		relPath, err := filepath.Rel(absDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			fmt.Println(".")
			return nil
		}

		depth := strings.Count(relPath, string(os.PathSeparator))
		if relPath != "." {
			depth++ // Increment depth for all non-root items
		}
		if params.Depth != -1 && depth > params.Depth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Exclusion check
		if isExcluded(path, d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// if it is the last entry in this level, use └ instead of ├
		// This is just stupid hacky way to do it, but whatever,
		// fix later if needed
		prefix := "├── "
		parentDir := filepath.Dir(path)
		_, ok := listingAtDir[parentDir]
		if !ok {
			entries, err := os.ReadDir(parentDir)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Warning: could not read dir listing for %s: %v\n", parentDir, err)
				listingAtDir[parentDir] = []fs.DirEntry{}
			} else {
				// filter entries according to exclusion rules
				listingAtDir[parentDir] = entries
			}
		}

		listing, ok := listingAtDir[parentDir]
		if !ok {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: could not read parent dir listing for %s\n", path)
			return nil
		}
		if len(listing) > 0 {
			lastEntry := listing[len(listing)-1]
			if lastEntry.Name() == d.Name() {
				prefix = "└── "
			}
		}

		// Check for current and all parents if we are the last entry. If we are the last entry, don't print │ for that level.
		parentPath := path
		indent := ""
		for dParentDepth := depth - 1; dParentDepth > 0; dParentDepth-- {
			parentPath = filepath.Dir(parentPath)
			parentListing, ok := listingAtDir[filepath.Dir(parentPath)]
			if !ok || len(parentListing) == 0 {
				continue
			}
			lastParentEntry := parentListing[len(parentListing)-1]
			if lastParentEntry.Name() == filepath.Base(parentPath) {
				indent = "    " + indent
			} else {
				indent = "│   " + indent
			}
		}

		if d.IsDir() {
			nDirs++
		} else {
			nFiles++
		}
		fmt.Printf("%s%s%s\n", indent, prefix, d.Name())

		return nil
	})

	// print summary,
	// 5 directories, 36 files
	fmt.Printf("\n%d directories, %d files\n", nDirs, nFiles)
	return err
}
