package cmd

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

type Base64Params struct {
	Files     []string `pos:"true" optional:"true" help:"File to process. If none specified or -, read from standard input." default:"-"`
	Decode    bool     `short:"d" help:"Decode data."`
	UrlSafe   bool     `short:"u" help:"Use URL-safe character set (alias for --alphabet url)."`
	NoPadding bool     `short:"r" help:"Do not write padding characters (raw) when encoding. Handle unpadded input when decoding."`
	Alphabet  string   `short:"a" help:"Custom 64-character alphabet or predefined set (standard, url)." default:"standard" optional:"true"`
}

func Base64Cmd() *cobra.Command {
	return boa.CmdT[Base64Params]{
		Use:         "base64",
		Short:       "Base64 encode or decode data",
		ParamEnrich: defaultParamEnricher(),
		RunFunc: func(params *Base64Params, cmd *cobra.Command, args []string) {
			if err := runBase64(params, os.Stdout, os.Stdin); err != nil {
				fmt.Fprintf(os.Stderr, "base64: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runBase64(params *Base64Params, stdout io.Writer, stdin io.Reader) error {
	// Determine encoding
	var enc *base64.Encoding

	// Resolve alphabet
	alphabet := params.Alphabet
	if params.UrlSafe {
		alphabet = "url"
	}

	switch alphabet {
	case "standard", "":
		if params.NoPadding {
			enc = base64.RawStdEncoding
		} else {
			enc = base64.StdEncoding
		}
	case "url":
		if params.NoPadding {
			enc = base64.RawURLEncoding
		} else {
			enc = base64.URLEncoding
		}
	default:
		if len(alphabet) != 64 {
			return fmt.Errorf("custom alphabet must be exactly 64 characters long")
		}
		enc = base64.NewEncoding(alphabet)
		if params.NoPadding {
			enc = enc.WithPadding(base64.NoPadding)
		}
	}

	// Setup input
	var readers []io.Reader
	if len(params.Files) == 0 {
		readers = append(readers, stdin)
	} else {
		for _, file := range params.Files {
			if file == "-" {
				readers = append(readers, stdin)
			} else {
				f, err := os.Open(file)
				if err != nil {
					return err
				}
				defer f.Close()
				readers = append(readers, f)
			}
		}
	}
	reader := io.MultiReader(readers...)

	if params.Decode {
		// Decoding
		decoder := base64.NewDecoder(enc, reader)
		_, err := io.Copy(stdout, decoder)
		return err
	} else {
		// Encoding
		encoder := base64.NewEncoder(enc, stdout)
		_, err := io.Copy(encoder, reader)
		if err != nil {
			encoder.Close()
			return err
		}
		// Must close encoder to flush padding
		if err := encoder.Close(); err != nil {
			return err
		}
		// Add a trailing newline for terminal friendliness
		_, err = fmt.Fprintln(stdout)
		return err
	}
}
