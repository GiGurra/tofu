package pwd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Logical  bool `short:"L" help:"Use PWD from environment, even if it contains symlinks (default)."`
	Physical bool `short:"P" help:"Avoid all symlinks, resolve to physical path."`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "pwd",
		Short:       "Print the current working directory",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			exitCode := Run(params, os.Stdout, os.Stderr)
			if exitCode != 0 {
				os.Exit(exitCode)
			}
		},
	}.ToCobra()
}

func Run(params *Params, stdout, stderr *os.File) int {
	var wd string
	var err error

	if params.Physical {
		// Get the actual physical path, resolving all symlinks
		wd, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(stderr, "pwd: %v\n", err)
			return 1
		}
		wd, err = filepath.EvalSymlinks(wd)
		if err != nil {
			fmt.Fprintf(stderr, "pwd: %v\n", err)
			return 1
		}
	} else {
		// Logical mode (default): use PWD env var if set, otherwise os.Getwd()
		wd = os.Getenv("PWD")
		if wd == "" {
			wd, err = os.Getwd()
			if err != nil {
				fmt.Fprintf(stderr, "pwd: %v\n", err)
				return 1
			}
		}
	}

	fmt.Fprintln(stdout, wd)
	return 0
}
