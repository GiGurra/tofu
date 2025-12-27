package tee

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
)

type Params struct {
	Files            []string `pos:"true" optional:"true" help:"Files to write to. If none specified, only write to stdout."`
	Append           bool     `short:"a" help:"Append to the given FILEs, do not overwrite."`
	IgnoreInterrupts bool     `short:"i" help:"Ignore interrupt signals (SIGINT)."`
	Silent           bool     `short:"s" help:"Silent mode: do not write to stdout, only to files."`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:         "tee",
		Short:       "Copy standard input to each FILE, and also to standard output",
		Long:        "Read from standard input and write to standard output and files simultaneously.",
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			exitCode := Run(params, os.Stdin, os.Stdout, os.Stderr)
			os.Exit(exitCode)
		},
	}.ToCobra()
}

func Run(params *Params, stdin io.Reader, stdout, stderr io.Writer) int {
	// Handle ignore interrupts flag
	if params.IgnoreInterrupts {
		signal.Ignore(syscall.SIGINT)
	}

	// Open all output files
	var writers []io.Writer
	if !params.Silent {
		writers = append(writers, stdout)
	}
	var closers []func() error

	for _, filename := range params.Files {
		flags := os.O_WRONLY | os.O_CREATE
		if params.Append {
			flags |= os.O_APPEND
		} else {
			flags |= os.O_TRUNC
		}

		f, err := os.OpenFile(filename, flags, 0644)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "tee: %s: %v\n", filename, err)
			return 1
		}
		writers = append(writers, f)
		closers = append(closers, f.Close)
	}

	// Create a MultiWriter to write to all destinations
	multiWriter := io.MultiWriter(writers...)

	// Copy stdin to all writers
	_, err := io.Copy(multiWriter, stdin)

	// Close all files
	hadError := false
	for i, closer := range closers {
		if closeErr := closer(); closeErr != nil {
			_, _ = fmt.Fprintf(stderr, "tee: error closing %s: %v\n", params.Files[i], closeErr)
			hadError = true
		}
	}

	if err != nil {
		_, _ = fmt.Fprintf(stderr, "tee: %v\n", err)
		return 1
	}

	if hadError {
		return 1
	}

	return 0
}