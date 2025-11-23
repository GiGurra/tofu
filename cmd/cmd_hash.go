package cmd

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

type HashParams struct {
	Files []string `pos:"true" optional:"true" help:"Files to hash. Read from stdin if none or '-'."`
	Algo  string   `short:"a" help:"Hash algorithm (md5, sha1, sha256, sha512)." default:"sha256"`
}

func HashCmd() *cobra.Command {
	return boa.CmdT[HashParams]{
		Use:   "hash [flags] [files...]",
		Short: "Calculate file hashes",
		Long: `Calculate cryptographic hashes for files or standard input.
Supported algorithms: md5, sha1, sha256, sha512.`, 
		ParamEnrich: defaultParamEnricher(),
		RunFunc: func(params *HashParams, cmd *cobra.Command, args []string) {
			if err := runHash(params, os.Stdout, os.Stdin); err != nil {
				fmt.Fprintf(os.Stderr, "hash: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

func runHash(params *HashParams, stdout io.Writer, stdin io.Reader) error {
	inputs := params.Files
	if len(inputs) == 0 {
		inputs = []string{"-"}
	}

	for _, input := range inputs {
		if err := processFile(input, params.Algo, stdout, stdin); err != nil {
			// Don't abort on single file error, just print to stderr
			fmt.Fprintf(os.Stderr, "hash: %v\n", err)
		}
	}

	return nil
}

func processFile(input, algo string, stdout io.Writer, stdin io.Reader) error {
	var r io.Reader
	var name string

	if input == "-" {
		r = stdin
		name = "-"
	} else {
		f, err := os.Open(input)
		if err != nil {
			return err
		}
		defer f.Close()
		r = f
		name = input
	}

	h, err := newHasher(algo)
	if err != nil {
		return err
	}

	if _, err := io.Copy(h, r); err != nil {
		return fmt.Errorf("%s: %v", name, err)
	}

	fmt.Fprintf(stdout, "%x  %s\n", h.Sum(nil), name)
	return nil
}

func newHasher(algo string) (hash.Hash, error) {
	switch algo {
	case "md5":
		return md5.New(), nil
	case "sha1":
		return sha1.New(), nil
	case "sha256":
		return sha256.New(), nil
	case "sha512":
		return sha512.New(), nil
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algo)
	}
}
