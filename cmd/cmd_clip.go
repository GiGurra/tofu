package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

var (
	clipboardWriteAll = clipboard.WriteAll
	clipboardReadAll  = clipboard.ReadAll
)

type ClipParams struct {
	Paste bool `short:"p" help:"Paste from clipboard to standard output."`
}

func ClipCmd() *cobra.Command {
	return boa.CmdT[ClipParams]{
		Use:   "clip [text]",
		Short: "Clipboard copy and paste",
		Long: `Copy to or paste from the system clipboard.

If [text] is provided, it is copied to the clipboard.
If no arguments are provided, reads from standard input until EOF.
Use -p or --paste to paste content from the clipboard to standard output.`,
		ParamEnrich: defaultParamEnricher(),
		RunFunc: func(params *ClipParams, cmd *cobra.Command, args []string) {
			if err := runClip(params, args, os.Stdin, os.Stdout); err != nil {
				fmt.Fprintf(os.Stderr, "clip: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runClip(params *ClipParams, args []string, stdin io.Reader, stdout io.Writer) error {
	if params.Paste {
		if len(args) > 0 {
			return fmt.Errorf("cannot use arguments with --paste")
		}
		text, err := clipboardReadAll()
		if err != nil {
			return err
		}
		fmt.Fprint(stdout, text)
		return nil
	}

	// Copy mode
	var text string
	if len(args) > 0 {
		text = strings.Join(args, " ")
	} else {
		// Read from stdin
		bytes, err := io.ReadAll(stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		text = string(bytes)
	}

	if err := clipboardWriteAll(text); err != nil {
		return fmt.Errorf("failed to write to clipboard: %w", err)
	}

	return nil
}
