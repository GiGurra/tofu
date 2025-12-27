package rm

import (
	"fmt"
	"io"
	"os"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Files       []string `pos:"true" required:"true" help:"Files or directories to remove."`
	Recursive   bool     `short:"r" optional:"true" help:"Remove directories and their contents recursively."`
	Force       bool     `short:"f" optional:"true" help:"Ignore nonexistent files and arguments, never prompt."`
	Interactive bool     `short:"i" optional:"true" help:"Prompt before every removal."`
	Dir         bool     `short:"d" optional:"true" help:"Remove empty directories."`
	Verbose     bool     `short:"v" optional:"true" help:"Explain what is being done."`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "rm",
		Short:       "Remove files or directories",
		Long:        "Remove (unlink) the FILE(s).",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			exitCode := Run(params, os.Stdin, os.Stdout, os.Stderr)
			os.Exit(exitCode)
		},
	}.ToCobra()
}

func Run(params *Params, stdin io.Reader, stdout, stderr io.Writer) int {
	hadError := false
	for _, file := range params.Files {
		if err := removeFile(file, params, stdin, stdout, stderr); err != nil {
			if !params.Force {
				fmt.Fprintf(stderr, "rm: %v\n", err)
				hadError = true
			}
		}
	}

	if hadError {
		return 1
	}
	return 0
}

func removeFile(path string, params *Params, stdin io.Reader, stdout, stderr io.Writer) error {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) && params.Force {
			return nil
		}
		return fmt.Errorf("cannot remove '%s': %v", path, err)
	}

	if info.IsDir() {
		if !params.Recursive && !params.Dir {
			return fmt.Errorf("cannot remove '%s': Is a directory", path)
		}

		if params.Dir && !params.Recursive {
			// Only remove if empty
			entries, err := os.ReadDir(path)
			if err != nil {
				return fmt.Errorf("cannot read directory '%s': %v", path, err)
			}
			if len(entries) > 0 {
				return fmt.Errorf("cannot remove '%s': Directory not empty", path)
			}
		}
	}

	if params.Interactive && !params.Force {
		prompt := "remove"
		if info.IsDir() {
			prompt = "remove directory"
		}
		fmt.Fprintf(stderr, "rm: %s '%s'? ", prompt, path)
		var response string
		fmt.Fscanln(stdin, &response)
		if response != "y" && response != "yes" {
			return nil
		}
	}

	var removeErr error
	if params.Recursive {
		removeErr = os.RemoveAll(path)
	} else {
		removeErr = os.Remove(path)
	}

	if removeErr != nil {
		return fmt.Errorf("cannot remove '%s': %v", path, removeErr)
	}

	if params.Verbose {
		fmt.Fprintf(stdout, "removed '%s'\n", path)
	}

	return nil
}
