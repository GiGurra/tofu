package cp

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
	Recursive   bool     `short:"r" optional:"true" help:"Copy directories recursively."`
	Force       bool     `short:"f" optional:"true" help:"If an existing destination file cannot be opened, remove it and try again."`
	Interactive bool     `short:"i" optional:"true" help:"Prompt before overwriting."`
	NoClobber   bool     `short:"n" optional:"true" help:"Do not overwrite an existing file."`
	Verbose     bool     `short:"v" optional:"true" help:"Explain what is being done."`
	Preserve    bool     `short:"p" optional:"true" help:"Preserve mode, ownership, and timestamps."`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "cp",
		Short:       "Copy files and directories",
		Long:        "Copy SOURCE to DEST, or multiple SOURCE(s) to DIRECTORY.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			exitCode := Run(params, os.Stdin, os.Stdout, os.Stderr)
			os.Exit(exitCode)
		},
	}.ToCobra()
}

func Run(params *Params, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(params.Sources) < 2 {
		fmt.Fprintf(stderr, "cp: missing destination file operand after '%s'\n", params.Sources[0])
		return 1
	}

	sources := params.Sources[:len(params.Sources)-1]
	dest := params.Sources[len(params.Sources)-1]

	destInfo, err := os.Stat(dest)
	destIsDir := err == nil && destInfo.IsDir()

	// If multiple sources, dest must be a directory
	if len(sources) > 1 && !destIsDir {
		fmt.Fprintf(stderr, "cp: target '%s' is not a directory\n", dest)
		return 1
	}

	hadError := false
	for _, src := range sources {
		target := dest
		if destIsDir {
			target = filepath.Join(dest, filepath.Base(src))
		}

		if err := copyPath(src, target, params, stdin, stdout, stderr); err != nil {
			fmt.Fprintf(stderr, "cp: %v\n", err)
			hadError = true
		}
	}

	if hadError {
		return 1
	}
	return 0
}

func copyPath(src, dest string, params *Params, stdin io.Reader, stdout, stderr io.Writer) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("cannot stat '%s': %v", src, err)
	}

	if srcInfo.IsDir() {
		if !params.Recursive {
			return fmt.Errorf("omitting directory '%s'", src)
		}
		return copyDir(src, dest, params, stdin, stdout, stderr)
	}

	return copyFile(src, dest, srcInfo, params, stdin, stdout, stderr)
}

func copyFile(src, dest string, srcInfo os.FileInfo, params *Params, stdin io.Reader, stdout, stderr io.Writer) error {
	// Check if dest exists
	if _, err := os.Stat(dest); err == nil {
		if params.NoClobber {
			return nil // silently skip
		}
		if params.Interactive && !params.Force {
			fmt.Fprintf(stderr, "cp: overwrite '%s'? ", dest)
			var response string
			fmt.Fscanln(stdin, &response)
			if response != "y" && response != "yes" {
				return nil
			}
		}
		if params.Force {
			os.Remove(dest)
		}
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("cannot open '%s': %v", src, err)
	}
	defer srcFile.Close()

	mode := srcInfo.Mode()
	if !params.Preserve {
		mode = 0644
	}

	destFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("cannot create '%s': %v", dest, err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("error copying to '%s': %v", dest, err)
	}

	if params.Preserve {
		os.Chtimes(dest, srcInfo.ModTime(), srcInfo.ModTime())
	}

	if params.Verbose {
		fmt.Fprintf(stdout, "'%s' -> '%s'\n", src, dest)
	}

	return nil
}

func copyDir(src, dest string, params *Params, stdin io.Reader, stdout, stderr io.Writer) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dest, srcInfo.Mode()); err != nil {
		return fmt.Errorf("cannot create directory '%s': %v", dest, err)
	}

	if params.Verbose {
		fmt.Fprintf(stdout, "'%s' -> '%s'\n", src, dest)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("cannot read directory '%s': %v", src, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if err := copyPath(srcPath, destPath, params, stdin, stdout, stderr); err != nil {
			return err
		}
	}

	return nil
}
