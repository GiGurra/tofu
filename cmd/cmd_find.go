package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

type FsItemType string

const (
	FsItemTypeFile FsItemType = "file"
	FsItemTypeDir  FsItemType = "dir"
	FsItemTypeAll  FsItemType = "all"
)

var validFsItemTypes = []FsItemType{
	FsItemTypeFile,
	FsItemTypeDir,
	FsItemTypeAll,
}

type SearchType string

const (
	SearchTypeExact    SearchType = "exact"
	SearchTypeContains SearchType = "contains"
	SearchTypePrefix   SearchType = "prefix"
	SearchTypeSuffix   SearchType = "suffix"
	SearchTypeRegex    SearchType = "regex"
)

var validSearchTypes = []SearchType{
	SearchTypeExact,
	SearchTypeContains,
	SearchTypePrefix,
	SearchTypeSuffix,
	SearchTypeRegex,
}

type FindParams struct {
	SearchTerm string       `pos:"true" help:"Term to search for in module names."`
	SearchType SearchType   `short:"s" help:"Type of search to perform." default:"contains"`
	IgnoreCase bool         `short:"i" help:"Perform a case-insensitive search." default:"false"`
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
			if !slices.Contains(validSearchTypes, params.SearchType) {
				return fmt.Errorf("invalid search type: %s", params.SearchType)
			}
			for _, t := range params.Types {
				if !slices.Contains(validFsItemTypes, t) {
					return fmt.Errorf("invalid file system item type: %s", t)
				}
			}
			return nil
		},
		RunFunc: func(params *FindParams, cmd *cobra.Command, args []string) {
			err := filepath.WalkDir(params.WorkDir, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "error accessing path %q: %v\n", path, err)
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

				if matchesType {
					switch params.SearchType {
					case SearchTypeExact:
						if !matchExact(d.Name(), params.SearchTerm, params.IgnoreCase) {
							return nil
						}
					case SearchTypeContains:
						if !matchContains(d.Name(), params.SearchTerm, params.IgnoreCase) {
							return nil
						}
					case SearchTypePrefix:
						if !matchPrefix(d.Name(), params.SearchTerm, params.IgnoreCase) {
							return nil
						}
					case SearchTypeSuffix:
						if !matchSuffix(d.Name(), params.SearchTerm, params.IgnoreCase) {
							return nil
						}
					case SearchTypeRegex:
						// Regex search not implemented
						panic(fmt.Errorf("regex search type not implemented"))
					default:
						panic(fmt.Errorf("unsupported search type: %s", params.SearchType))
					}
					fmt.Println(path)
				}
				return nil
			})

			if err != nil {
				panic(fmt.Errorf("error during file system walk: %w", err))
			}
		},
	}.ToCobra()
}

func matchExact(a, b string, ignoreCase bool) bool {
	if ignoreCase {
		return strings.EqualFold(a, b)
	} else {
		return a == b
	}
}

func matchContains(tot, substr string, ignoreCase bool) bool {
	if ignoreCase {
		tot = strings.ToLower(tot)
		substr = strings.ToLower(substr)
	}
	return strings.Contains(tot, substr)
}

func matchPrefix(tot, prefix string, ignoreCase bool) bool {
	if ignoreCase {
		tot = strings.ToLower(tot)
		prefix = strings.ToLower(prefix)
	}
	return strings.HasPrefix(tot, prefix)
}

func matchSuffix(tot, suffix string, ignoreCase bool) bool {
	if ignoreCase {
		tot = strings.ToLower(tot)
		suffix = strings.ToLower(suffix)
	}
	return strings.HasSuffix(tot, suffix)
}

func ExistsAccessibleDirDir(path string) bool {
	st, err := os.Stat(path)
	if err != nil {
		return false
	}
	return st.IsDir()
}
