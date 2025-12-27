package mv

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Sources     []string `pos:"true" required:"true" help:"Source file(s) and destination."`
	Force       bool     `short:"f" optional:"true" help:"Do not prompt before overwriting."`
	Interactive bool     `short:"i" optional:"true" help:"Prompt before overwriting."`
	NoClobber   bool     `short:"n" optional:"true" help:"Do not overwrite an existing file."`
	Verbose     bool     `short:"v" optional:"true" help:"Explain what is being done."`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "mv",
		Short:       "Move (rename) files",
		Long:        "Rename SOURCE to DEST, or move SOURCE(s) to DIRECTORY.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			exitCode := Run(params, os.Stdin, os.Stdout, os.Stderr)
			os.Exit(exitCode)
		},
	}.ToCobra()
}

func Run(params *Params, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(params.Sources) < 2 {
		fmt.Fprintf(stderr, "mv: missing destination file operand after '%s'\n", params.Sources[0])
		return 1
	}

	sources := params.Sources[:len(params.Sources)-1]
	dest := params.Sources[len(params.Sources)-1]

	destInfo, err := os.Stat(dest)
	destIsDir := err == nil && destInfo.IsDir()

	// If multiple sources, dest must be a directory
	if len(sources) > 1 && !destIsDir {
		fmt.Fprintf(stderr, "mv: target '%s' is not a directory\n", dest)
		return 1
	}

	hadError := false
	for _, src := range sources {
		target := dest
		if destIsDir {
			target = filepath.Join(dest, filepath.Base(src))
		}

		if err := moveFile(src, target, params, stdin, stdout, stderr); err != nil {
			fmt.Fprintf(stderr, "mv: %v\n", err)
			hadError = true
		}
	}

	if hadError {
		return 1
	}
	return 0
}

func moveFile(src, dest string, params *Params, stdin io.Reader, stdout, stderr io.Writer) error {
	// Check if dest exists
	if _, err := os.Stat(dest); err == nil {
		if params.NoClobber {
			return nil // silently skip
		}
		if params.Interactive && !params.Force {
			fmt.Fprintf(stderr, "mv: overwrite '%s'? ", dest)
			var response string
			fmt.Fscanln(stdin, &response)
			if response != "y" && response != "yes" {
				return nil
			}
		}
	}

	if err := os.Rename(src, dest); err != nil {
		return fmt.Errorf("cannot move '%s' to '%s': %v", src, dest, err)
	}

	if params.Verbose {
		fmt.Fprintf(stdout, "renamed '%s' -> '%s'\n", src, dest)
	}

	return nil
}
