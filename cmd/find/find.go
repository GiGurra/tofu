package find

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type FsItemType string

const (
	FsItemTypeFile FsItemType = "file"
	FsItemTypeDir  FsItemType = "dir"
	FsItemTypeAll  FsItemType = "all"
)

type SearchType string

const (
	SearchTypeExact    SearchType = "exact"
	SearchTypeContains SearchType = "contains"
	SearchTypePrefix   SearchType = "prefix"
	SearchTypeSuffix   SearchType = "suffix"
	SearchTypeRegex    SearchType = "regex"
)

type Params struct {
	SearchTerm string       `pos:"true" help:"Term to search for in module names."`
	SearchType SearchType   `short:"s" help:"Type of search to perform (exact,contains,prefix,suffix,regex)." default:"contains" alts:"exact,contains,prefix,suffix,regex"`
	IgnoreCase bool         `short:"i" help:"Perform a case-insensitive search." default:"false"`
	WorkDir    string       `short:"c" help:"The working directory to start the search from." default:"."`
	Types      []FsItemType `short:"t" help:"Types of file system items to search for (file,dir,all)." default:"all" alts:"file,dir,all"`
	Quiet      bool         `short:"q" help:"Suppress error messages." default:"false"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "find",
		Short:       "Find file system items matching a search term",
		ParamEnrich: common.DefaultParamEnricher(),
		PreExecuteFunc: func(params *Params, cmd *cobra.Command, args []string) error {
			if params.SearchTerm == "" {
				return fmt.Errorf("search term cannot be empty")
			}
			if len(params.Types) == 0 {
				return fmt.Errorf("at least one type must be specified")
			}
			if !ExistsAccessibleDir(params.WorkDir) {
				return fmt.Errorf("working directory does not exist or is not accessible: %s", params.WorkDir)
			}
			return nil
		},
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			Run(params, os.Stdout, os.Stderr)
		},
	}.ToCobra()
}

func Run(params *Params, stdout, stderr io.Writer) {
	var precompiledRegex *regexp.Regexp
	if params.SearchType == SearchTypeRegex {
		var err error
		precompiledRegex, err = regexp.Compile(params.SearchTerm)
		if err != nil {
			panic(fmt.Errorf("invalid regex pattern: %w", err))
		}
	}
	err := filepath.WalkDir(params.WorkDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if !params.Quiet {
				_, _ = fmt.Fprintf(stderr, "error accessing path %q: %v\n", path, err)
			}
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
				if !MatchExact(d.Name(), params.SearchTerm, params.IgnoreCase) {
					return nil
				}
			case SearchTypeContains:
				if !MatchContains(d.Name(), params.SearchTerm, params.IgnoreCase) {
					return nil
				}
			case SearchTypePrefix:
				if !MatchPrefix(d.Name(), params.SearchTerm, params.IgnoreCase) {
					return nil
				}
			case SearchTypeSuffix:
				if !MatchSuffix(d.Name(), params.SearchTerm, params.IgnoreCase) {
					return nil
				}
			case SearchTypeRegex:
				if precompiledRegex == nil || !MatchRegex(d.Name(), precompiledRegex) {
					return nil
				}
			default:
				panic(fmt.Errorf("unsupported search type: %s", params.SearchType))
			}
			fmt.Fprintln(stdout, path)
		}
		return nil
	})

	if err != nil {
		panic(fmt.Errorf("error during file system walk: %w", err))
	}
}

func MatchRegex(tot string, precompiledRegex *regexp.Regexp) bool {
	return precompiledRegex.MatchString(tot)
}

func MatchExact(a, b string, ignoreCase bool) bool {
	if ignoreCase {
		return strings.EqualFold(a, b)
	} else {
		return a == b
	}
}

func MatchContains(tot, substr string, ignoreCase bool) bool {
	if ignoreCase {
		tot = strings.ToLower(tot)
		substr = strings.ToLower(substr)
	}
	return strings.Contains(tot, substr)
}

func MatchPrefix(tot, prefix string, ignoreCase bool) bool {
	if ignoreCase {
		tot = strings.ToLower(tot)
		prefix = strings.ToLower(prefix)
	}
	return strings.HasPrefix(tot, prefix)
}

func MatchSuffix(tot, suffix string, ignoreCase bool) bool {
	if ignoreCase {
		tot = strings.ToLower(tot)
		suffix = strings.ToLower(suffix)
	}
	return strings.HasSuffix(tot, suffix)
}

func ExistsAccessibleDir(path string) bool {
	st, err := os.Stat(path)
	if err != nil {
		return false
	}
	return st.IsDir()
}
