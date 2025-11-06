package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

type FsItemType string

const (
	FsItemTypeFile FsItemType = "file"
	FsItemTypeDir  FsItemType = "dir"
	FsItemTypeAll  FsItemType = "all"
)

type FindParams struct {
	SearchTerm string       `pos:"true" help:"Term to search for in module names."`
	WorkDir    string       `short:"c" help:"The working directory to start the search from." default:"."`
	Types      []FsItemType `short:"t" help:"Types of file system items to search for (file, dir, all)." default:"all"`
}

func FindCmd() *cobra.Command {
	return boa.CmdT[FindParams]{
		Use:   "find",
		Short: "Find file system items matching a search term",
		PreExecuteFunc: func(params *FindParams, cmd *cobra.Command, args []string) error {
			if params.SearchTerm == "" {
				return fmt.Errorf("search term cannot be empty")
			}
			if len(params.Types) == 0 {
				return fmt.Errorf("at least one type must be specified")
			}
			if !ExistsAccessibleDirDir(params.WorkDir) {
				return fmt.Errorf("working directory does not exist or is not accessible: %s", params.WorkDir)
			}
			return nil
		},
		RunFunc: func(params *FindParams, cmd *cobra.Command, args []string) {
			err := filepath.WalkDir(params.WorkDir, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					// Skip paths that can't be accessed
					return nil
				}

				matchesType := false
				for _, t := range params.Types {
					switch t {
					case FsItemTypeAll:
						matchesType = true
					case FsItemTypeFile:
						if !d.IsDir() {
							matchesType = true
						}
					case FsItemTypeDir:
						if d.IsDir() {
							matchesType = true
						}
					}
				}

				if matchesType && d.Name() == params.SearchTerm {
					relPath, err := filepath.Rel(params.WorkDir, path)
					if err != nil {
						relPath = path // Fallback to full path if relative fails
					}
					fmt.Println(relPath)
				}
				return nil
			})

			if err != nil {
				panic(fmt.Errorf("error during file system walk: %w", err))
			}
		},
	}.ToCobra()
}

func ExistsAccessibleDirDir(path string) bool {
	st, err := os.Stat(path)
	if err != nil {
		return false
	}
	return st.IsDir()
}
