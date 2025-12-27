package rmdir

import (
	"fmt"
	"io"
	"os"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Dirs    []string `pos:"true" required:"true" help:"Directories to remove."`
	Parents bool     `short:"p" optional:"true" help:"Remove DIRECTORY and its ancestors."`
	Verbose bool     `short:"v" optional:"true" help:"Output a diagnostic for every directory processed."`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "rmdir",
		Short:       "Remove empty directories",
		Long:        "Remove the DIRECTORY(ies), if they are empty.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			exitCode := Run(params, os.Stdout, os.Stderr)
			os.Exit(exitCode)
		},
	}.ToCobra()
}

func Run(params *Params, stdout, stderr io.Writer) int {
	hadError := false
	for _, dir := range params.Dirs {
		if err := removeDir(dir, params, stdout, stderr); err != nil {
			fmt.Fprintf(stderr, "rmdir: %v\n", err)
			hadError = true
		}
	}

	if hadError {
		return 1
	}
	return 0
}

func removeDir(path string, params *Params, stdout, stderr io.Writer) error {
	if params.Parents {
		return removeParents(path, params, stdout)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove '%s': %v", path, err)
	}

	if params.Verbose {
		fmt.Fprintf(stdout, "rmdir: removing directory, '%s'\n", path)
	}

	return nil
}

func removeParents(path string, params *Params, stdout io.Writer) error {
	for path != "" && path != "." && path != "/" {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to remove '%s': %v", path, err)
		}

		if params.Verbose {
			fmt.Fprintf(stdout, "rmdir: removing directory, '%s'\n", path)
		}

		// Move to parent directory
		for i := len(path) - 1; i >= 0; i-- {
			if path[i] == '/' || path[i] == os.PathSeparator {
				path = path[:i]
				break
			}
			if i == 0 {
				path = ""
			}
		}
	}

	return nil
}
